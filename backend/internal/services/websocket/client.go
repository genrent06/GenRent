package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// ClientManager manages WebSocket client connections
type ClientManager struct {
	hub          *Hub
	clients      map[string]*Client
	mu           sync.RWMutex
	nextClientID int
}

// NewClientManager creates a new client manager
func NewClientManager(hub *Hub) *ClientManager {
	return &ClientManager{
		hub:     hub,
		clients: make(map[string]*Client),
	}
}

// HandleWebSocket handles WebSocket connection requests
func (cm *ClientManager) HandleWebSocket(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDUint, ok := userID.(uint64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade error: %v", err)
		return
	}

	// Generate unique client ID
	cm.mu.Lock()
	cm.nextClientID++
	clientID := c.ClientIP() + "-" + string(rune(cm.nextClientID))
	cm.mu.Unlock()

	// Create client
	client := &Client{
		ID:     clientID,
		UserID: userIDUint,
		Conn:   conn,
		Send:   make(chan *Message, 256),
		Rooms:  make(map[string]bool),
	}

	// Register client
	cm.hub.RegisterClient(client)
	cm.clients[clientID] = client

	// Start client pumps
	go client.writePump()
	go client.readPump(cm.hub)

	log.Printf("[WebSocket] Client connected: %s (User: %d)", clientID, userIDUint)
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error: %v", err)
			}
			break
		}

		// Parse incoming message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[WebSocket] JSON parse error: %v", err)
			continue
		}

		// Set sender ID
		msg.SenderID = c.ID

		// Handle message types
		switch msg.Type {
		case "ping":
			// Respond with pong
			c.Send <- &Message{
				Type:      "pong",
				Timestamp: time.Now(),
			}

		case "join_room":
			// Join a room
			if roomID, ok := msg.Data["room_id"].(string); ok {
				hub.JoinRoom(c, roomID)
			}

		case "leave_room":
			// Leave a room
			if roomID, ok := msg.Data["room_id"].(string); ok {
				hub.LeaveRoom(c, roomID)
			}

		case "broadcast":
			// Broadcast to all clients
			msg.Type = "broadcast"
			hub.Broadcast(msg.Type, msg.Data)

		case "send_message":
			// Send message to specific user
			if recipientID, ok := msg.Data["recipient_id"].(float64); ok {
				hub.SendToUser(uint64(recipientID), msg.Type, msg.Data)
			}

		case "room_message":
			// Send message to room
			if roomID, ok := msg.Data["room_id"].(string); ok {
				hub.SendToRoom(roomID, msg.Type, msg.Data)
			}

		default:
			log.Printf("[WebSocket] Unknown message type: %s", msg.Type)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Convert message to JSON
			jsonMessage, err := json.Marshal(message)
			if err != nil {
				log.Printf("[WebSocket] JSON marshal error: %v", err)
				continue
			}

			// Send message
			if err := c.Conn.WriteMessage(websocket.TextMessage, jsonMessage); err != nil {
				log.Printf("[WebSocket] Write error: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// DisconnectClient disconnects a specific client
func (cm *ClientManager) DisconnectClient(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, ok := cm.clients[clientID]; ok {
		cm.hub.UnregisterClient(client)
		delete(cm.clients, clientID)
	}
}

// DisconnectUser disconnects all connections for a specific user
func (cm *ClientManager) DisconnectUser(userID uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for clientID, client := range cm.clients {
		if client.UserID == userID {
			cm.hub.UnregisterClient(client)
			delete(cm.clients, clientID)
		}
	}
}

// GetConnectedUsers returns list of connected user IDs
func (cm *ClientManager) GetConnectedUsers() []uint64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	userMap := make(map[uint64]bool)
	for _, client := range cm.clients {
		userMap[client.UserID] = true
	}

	users := make([]uint64, 0, len(userMap))
	for userID := range userMap {
		users = append(users, userID)
	}

	return users
}

// GetClientCount returns the number of connected clients
func (cm *ClientManager) GetClientCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.clients)
}

// BroadcastMessage broadcasts a message to all connected clients
func (cm *ClientManager) BroadcastMessage(messageType string, data map[string]interface{}) {
	cm.hub.Broadcast(messageType, data)
}

// SendToUser sends a message to a specific user
func (cm *ClientManager) SendToUser(userID uint64, messageType string, data map[string]interface{}) {
	cm.hub.SendToUser(userID, messageType, data)
}

// SendToRoom sends a message to all clients in a room
func (cm *ClientManager) SendToRoom(roomID, messageType string, data map[string]interface{}) {
	cm.hub.SendToRoom(roomID, messageType, data)
}
