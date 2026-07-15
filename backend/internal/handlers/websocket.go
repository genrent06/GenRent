package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/services/chat"
	"genrent/internal/services/notification"
	"genrent/internal/services/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	clientManager *websocket.ClientManager
	chatService  *chat.ChatService
	notifService *notification.NotificationService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *websocket.Hub, db *gorm.DB) *WebSocketHandler {
	return &WebSocketHandler{
		clientManager: websocket.NewClientManager(hub),
		chatService:  chat.NewChatService(db, hub),
		notifService: notification.NewNotificationService(db, hub, nil),
	}
}

// HandleWebSocket handles WebSocket connection requests
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Set user ID in context for WebSocket handler
	c.Set("user_id", userID)
	h.clientManager.HandleWebSocket(c)
}

// GetOnlineUsers returns list of currently online users
func (h *WebSocketHandler) GetOnlineUsers(c *gin.Context) {
	onlineUsers := h.clientManager.GetConnectedUsers()

	c.JSON(http.StatusOK, gin.H{
		"online_users": onlineUsers,
		"count":        len(onlineUsers),
	})
}

// GetOnlineUserCount returns count of online users
func (h *WebSocketHandler) GetOnlineUserCount(c *gin.Context) {
	count := h.clientManager.GetClientCount()

	c.JSON(http.StatusOK, gin.H{
		"online_count": count,
	})
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *WebSocketHandler) BroadcastMessage(c *gin.Context) {
	var req struct {
		Type string                 `json:"type" binding:"required"`
		Data map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.clientManager.BroadcastMessage(req.Type, req.Data)

	c.JSON(http.StatusOK, gin.H{"message": "broadcast sent"})
}

// SendToUser sends a message to a specific user
func (h *WebSocketHandler) SendToUser(c *gin.Context) {
	var req struct {
		UserID uint64                 `json:"user_id" binding:"required"`
		Type   string                 `json:"type" binding:"required"`
		Data   map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.clientManager.SendToUser(req.UserID, req.Type, req.Data)

	c.JSON(http.StatusOK, gin.H{"message": "message sent"})
}

// SendToRoom sends a message to all clients in a room
func (h *WebSocketHandler) SendToRoom(c *gin.Context) {
	var req struct {
		RoomID string                 `json:"room_id" binding:"required"`
		Type   string                 `json:"type" binding:"required"`
		Data   map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.clientManager.SendToRoom(req.RoomID, req.Type, req.Data)

	c.JSON(http.StatusOK, gin.H{"message": "message sent to room"})
}

// ChatHandler handles chat-related requests
type ChatHandler struct {
	chatService  *chat.ChatService
	notifService *notification.NotificationService
	wsHub        *websocket.Hub
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *gorm.DB, wsHub *websocket.Hub) *ChatHandler {
	return &ChatHandler{
		chatService:  chat.NewChatService(db, wsHub),
		notifService: notification.NewNotificationService(db, wsHub, nil),
		wsHub:        wsHub,
	}
}

// CreateConversation creates a new conversation
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		VendorID  uint64  `json:"vendor_id" binding:"required"`
		BookingID *uint64 `json:"booking_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.chatService.CreateConversation(req.VendorID, uint64(userID), req.BookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conversation)
}

// GetConversations retrieves all conversations for the user
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	role := c.DefaultQuery("role", "customer")
	conversations, err := h.chatService.GetConversations(uint64(userID), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
	})
}

// GetConversation retrieves a specific conversation
func (h *ChatHandler) GetConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	conversation, err := h.chatService.GetConversation(conversationID, uint64(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// GetMessages retrieves messages for a conversation
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationIDStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	messages, total, err := h.chatService.GetMessages(conversationID, uint64(userID), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":  messages,
		"total":     total,
		"page":      page,
		"per_page":  perPage,
	})
}

// SendMessage sends a message in a conversation
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ConversationID uint64 `json:"conversation_id" binding:"required"`
		ReceiverID     uint64 `json:"receiver_id" binding:"required"`
		Message        string `json:"message" binding:"required"`
		MessageType    string `json:"message_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MessageType == "" {
		req.MessageType = "text"
	}

	message, err := h.chatService.SendMessage(req.ConversationID, uint64(userID), req.ReceiverID, req.Message, req.MessageType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// MarkMessagesAsRead marks messages as read
func (h *ChatHandler) MarkMessagesAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationIDStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	if err := h.chatService.MarkMessagesAsRead(conversationID, uint64(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "messages marked as read"})
}

// GetUnreadCount returns unread message count
func (h *ChatHandler) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.chatService.GetUnreadCount(uint64(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// ArchiveConversation archives a conversation
func (h *ChatHandler) ArchiveConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	if err := h.chatService.ArchiveConversation(conversationID, uint64(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation archived"})
}

// DeleteConversation deletes a conversation
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	if err := h.chatService.DeleteConversation(conversationID, uint64(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted"})
}

// SearchMessages searches for messages
func (h *ChatHandler) SearchMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	messages, total, err := h.chatService.SearchMessages(uint64(userID), query, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// GetConversationStats returns statistics for a conversation
func (h *ChatHandler) GetConversationStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	stats, err := h.chatService.GetConversationStats(conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SetTypingStatus sets typing status for a user
func (h *ChatHandler) SetTypingStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationIDStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	var req struct {
		IsTyping bool `json:"is_typing"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.chatService.SetTypingStatus(conversationID, uint64(userID), req.IsTyping)

	c.JSON(http.StatusOK, gin.H{"message": "typing status updated"})
}
