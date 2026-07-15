package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	SenderID  string                 `json:"sender_id,omitempty"`
	RoomID    string                 `json:"room_id,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	UserID uint64
	Conn   *websocket.Conn
	Send   chan *Message
	Rooms  map[string]bool
	mu     sync.RWMutex
}

// Hub maintains active client connections and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[string]*Client

	// Client rooms (user_id -> client)
	userClients map[uint64]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients
	broadcast chan *Message

	// Send message to specific user
	sendUser chan *UserMessage

	// Send message to specific room
	sendRoom chan *RoomMessage

	// Room subscriptions
	rooms map[string]map[string]bool

	mu sync.RWMutex
}

// UserMessage represents a message to a specific user
type UserMessage struct {
	UserID  uint64
	Message *Message
}

// RoomMessage represents a message to a specific room
type RoomMessage struct {
	RoomID  string
	Message *Message
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		userClients: make(map[uint64]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *Message),
		sendUser:    make(chan *UserMessage, 256),
		sendRoom:    make(chan *RoomMessage, 256),
		rooms:       make(map[string]map[string]bool),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)

		case userMessage := <-h.sendUser:
			h.handleSendUser(userMessage)

		case roomMessage := <-h.sendRoom:
			h.handleSendRoom(roomMessage)
		}
	}
}

// handleRegister registers a new client
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client
	h.userClients[client.UserID] = client

	log.Printf("[WebSocket] Client registered: %s (User: %d)", client.ID, client.UserID)

	// Send welcome message
	welcome := &Message{
		Type: "system",
		Data: map[string]interface{}{
			"message": "Connected to GenRent real-time service",
			"user_id": client.UserID,
		},
		Timestamp: time.Now(),
	}
	select {
	case client.Send <- welcome:
	default:
		close(client.Send)
		delete(h.clients, client.ID)
		delete(h.userClients, client.UserID)
	}
}

// handleUnregister unregisters a client
func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		delete(h.userClients, client.UserID)
		close(client.Send)

		// Remove from all rooms
		for roomID := range client.Rooms {
			if roomUsers, ok := h.rooms[roomID]; ok {
				delete(roomUsers, client.ID)
				if len(roomUsers) == 0 {
					delete(h.rooms, roomID)
				}
			}
		}

		log.Printf("[WebSocket] Client unregistered: %s (User: %d)", client.ID, client.UserID)
	}
}

// handleBroadcast broadcasts a message to all clients
func (h *Hub) handleBroadcast(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client.ID)
		}
	}
}

// handleSendUser sends a message to a specific user
func (h *Hub) handleSendUser(userMessage *UserMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.userClients[userMessage.UserID]; ok {
		select {
		case client.Send <- userMessage.Message:
		default:
			log.Printf("[WebSocket] Failed to send message to user %d (channel full)", userMessage.UserID)
		}
	} else {
		log.Printf("[WebSocket] User %d not connected", userMessage.UserID)
	}
}

// handleSendRoom sends a message to all clients in a room
func (h *Hub) handleSendRoom(roomMessage *RoomMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomUsers, ok := h.rooms[roomMessage.RoomID]; ok {
		for clientID := range roomUsers {
			if client, ok := h.clients[clientID]; ok {
				select {
				case client.Send <- roomMessage.Message:
				default:
					close(client.Send)
					delete(h.clients, clientID)
				}
			}
		}
	}
}

// RegisterClient registers a new client
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// Broadcast broadcasts a message to all connected clients
func (h *Hub) Broadcast(messageType string, data map[string]interface{}) {
	message := &Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.broadcast <- message
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uint64, messageType string, data map[string]interface{}) {
	message := &Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.sendUser <- &UserMessage{
		UserID:  userID,
		Message: message,
	}
}

// SendToRoom sends a message to all clients in a room
func (h *Hub) SendToRoom(roomID, messageType string, data map[string]interface{}) {
	message := &Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
		RoomID:    roomID,
	}
	h.sendRoom <- &RoomMessage{
		RoomID:  roomID,
		Message: message,
	}
}

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add room to client's rooms
	client.mu.Lock()
	if client.Rooms == nil {
		client.Rooms = make(map[string]bool)
	}
	client.Rooms[roomID] = true
	client.mu.Unlock()

	// Add client to room
	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[string]bool)
	}
	h.rooms[roomID][client.ID] = true

	// Notify room
	h.SendToRoom(roomID, "user_joined", map[string]interface{}{
		"user_id": client.UserID,
		"room_id": roomID,
	})

	log.Printf("[WebSocket] User %d joined room: %s", client.UserID, roomID)
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove room from client's rooms
	client.mu.Lock()
	delete(client.Rooms, roomID)
	client.mu.Unlock()

	// Remove client from room
	if roomUsers, ok := h.rooms[roomID]; ok {
		delete(roomUsers, client.ID)
		if len(roomUsers) == 0 {
			delete(h.rooms, roomID)
		}
	}

	// Notify room
	h.SendToRoom(roomID, "user_left", map[string]interface{}{
		"user_id": client.UserID,
		"room_id": roomID,
	})

	log.Printf("[WebSocket] User %d left room: %s", client.UserID, roomID)
}

// GetOnlineUsers returns the count of online users
func (h *Hub) GetOnlineUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userClients)
}

// GetUsersInRoom returns the count of users in a specific room
func (h *Hub) GetUsersInRoom(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomUsers, ok := h.rooms[roomID]; ok {
		return len(roomUsers)
	}
	return 0
}

// BroadcastToUsers sends a message to multiple specific users
func (h *Hub) BroadcastToUsers(userIDs []uint64, messageType string, data map[string]interface{}) {
	for _, userID := range userIDs {
		h.SendToUser(userID, messageType, data)
	}
}

// SendJSON sends a JSON message to a specific user
func (h *Hub) SendJSON(userID uint64, messageType string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return err
	}

	h.SendToUser(userID, messageType, dataMap)
	return nil
}

// NotifyBookingUpdate sends booking status update to relevant users
func (h *Hub) NotifyBookingUpdate(bookingID uint64, status string, customerID, vendorID uint64) {
	data := map[string]interface{}{
		"booking_id": bookingID,
		"status":     status,
		"timestamp":  time.Now().Unix(),
	}

	// Notify customer
	h.SendToUser(customerID, "booking_update", data)

	// Notify vendor
	h.SendToUser(vendorID, "booking_update", data)

	log.Printf("[WebSocket] Booking update sent: Booking=%d, Status=%s", bookingID, status)
}

// NotifyNewMessage sends new message notification to recipient
func (h *Hub) NotifyNewMessage(conversationID string, senderID, recipientID uint64, message string) {
	data := map[string]interface{}{
		"conversation_id": conversationID,
		"sender_id":       senderID,
		"message":         message,
		"timestamp":       time.Now().Unix(),
	}

	h.SendToUser(recipientID, "new_message", data)

	log.Printf("[WebSocket] New message notification sent to user %d", recipientID)
}

// NotifyAvailabilityUpdate sends equipment availability update
func (h *Hub) NotifyAvailabilityUpdate(equipmentID uint64, availableQty int, vendorID uint64) {
	data := map[string]interface{}{
		"equipment_id":      equipmentID,
		"available_qty":     availableQty,
		"timestamp":         time.Now().Unix(),
	}

	// Notify vendor
	h.SendToUser(vendorID, "availability_update", data)

	// Broadcast to booking room
	roomID := fmt.Sprintf("equipment_%d", equipmentID)
	h.SendToRoom(roomID, "availability_update", data)

	log.Printf("[WebSocket] Availability update sent: Equipment=%d, Qty=%d", equipmentID, availableQty)
}

// GetConnectedUsers returns list of connected user IDs
func (h *Hub) GetConnectedUsers() []uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uint64, 0, len(h.userClients))
	for userID := range h.userClients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID uint64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.userClients[userID]
	return ok
}

// DisconnectUser disconnects a specific user
func (h *Hub) DisconnectUser(userID uint64) {
	h.mu.RLock()
	client, ok := h.userClients[userID]
	h.mu.RUnlock()

	if ok {
		h.UnregisterClient(client)
	}
}

// GetRoomCount returns the number of active rooms
func (h *Hub) GetRoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// CreateRoom creates a new room and optionally adds users
func (h *Hub) CreateRoom(roomID string, userIDs []uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[string]bool)
	}

	// Add users to room
	for _, userID := range userIDs {
		if client, ok := h.userClients[userID]; ok {
			h.rooms[roomID][client.ID] = true

			// Add room to client's rooms
			client.mu.Lock()
			if client.Rooms == nil {
				client.Rooms = make(map[string]bool)
			}
			client.Rooms[roomID] = true
			client.mu.Unlock()
		}
	}

	log.Printf("[WebSocket] Room created: %s with %d users", roomID, len(userIDs))
}

// GetRoomInfo returns information about a specific room
func (h *Hub) GetRoomInfo(roomID string) map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	info := map[string]interface{}{
		"room_id":       roomID,
		"user_count":    0,
		"users":         []uint64{},
		"exists":        false,
	}

	if roomUsers, ok := h.rooms[roomID]; ok {
		info["exists"] = true
		info["user_count"] = len(roomUsers)

		users := make([]uint64, 0)
		for clientID := range roomUsers {
			if client, ok := h.clients[clientID]; ok {
				users = append(users, client.UserID)
			}
		}
		info["users"] = users
	}

	return info
}
