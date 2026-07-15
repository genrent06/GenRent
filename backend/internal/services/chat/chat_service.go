package chat

import (
	"fmt"
	"time"

	"genrent/internal/models"
	"genrent/internal/services/websocket"

	"gorm.io/gorm"
)

// ChatService handles chat operations
type ChatService struct {
	db              *gorm.DB
	wsHub          *websocket.Hub
}

// NewChatService creates a new chat service
func NewChatService(db *gorm.DB, wsHub *websocket.Hub) *ChatService {
	return &ChatService{
		db:     db,
		wsHub:  wsHub,
	}
}

// CreateConversation creates a new conversation between vendor and customer
func (s *ChatService) CreateConversation(vendorID, customerID uint64, bookingID *uint64) (*models.Conversation, error) {
	// Check if conversation already exists
	var existingConv models.Conversation
	query := s.db.Where("vendor_id = ? AND customer_id = ?", vendorID, customerID)
	if bookingID != nil {
		query = query.Where("booking_id = ?", *bookingID)
	}

	err := query.First(&existingConv).Error
	if err == nil {
		return &existingConv, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing conversation: %w", err)
	}

	// Create new conversation
	conversation := &models.Conversation{
		VendorID:     vendorID,
		CustomerID:   customerID,
		BookingID:    bookingID,
		Status:       "active",
		VendorRead:   true,
		CustomerRead: true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(conversation).Error; err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Create chat room for WebSocket
	roomID := fmt.Sprintf("conversation_%d", conversation.ID)
	s.wsHub.CreateRoom(roomID, []uint64{vendorID, customerID})

	return conversation, nil
}

// GetConversation retrieves a conversation by ID
func (s *ChatService) GetConversation(conversationID uint64, userID uint64) (*models.Conversation, error) {
	var conversation models.Conversation
	err := s.db.Where("id = ? AND (vendor_id = ? OR customer_id = ?)", conversationID, userID, userID).
		First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// GetConversations retrieves all conversations for a user
func (s *ChatService) GetConversations(userID uint64, role string) ([]models.Conversation, error) {
	var conversations []models.Conversation

	query := s.db.Preload("Booking").
		Where("status = ?", "active")

	if role == "vendor" {
		query = query.Where("vendor_id = ?", userID)
	} else {
		query = query.Where("customer_id = ?", userID)
	}

	err := query.Order("last_message_at DESC").Find(&conversations).Error
	return conversations, err
}

// SendMessage sends a message in a conversation
func (s *ChatService) SendMessage(conversationID, senderID, receiverID uint64, message string, messageType string) (*models.Message, error) {
	// Verify conversation exists and user is participant
	var conversation models.Conversation
	err := s.db.Where("id = ? AND (vendor_id = ? OR customer_id = ?)", conversationID, senderID, senderID).
		First(&conversation).Error
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	// Create message
	msg := &models.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Message:        message,
		MessageType:    messageType,
		IsRead:         false,
		SentAt:         time.Now(),
		CreatedAt:      time.Now(),
	}

	if err := s.db.Create(msg).Error; err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation
	now := time.Now()
	updateData := map[string]interface{}{
		"last_message":   message,
		"last_message_at": now,
	}

	// Update read status based on sender
	if senderID == conversation.VendorID {
		updateData["customer_read"] = false
	} else {
		updateData["vendor_read"] = false
	}

	s.db.Model(&conversation).Updates(updateData)

	// Send real-time notification
	roomID := fmt.Sprintf("conversation_%d", conversationID)
	messageData := map[string]interface{}{
		"conversation_id": conversationID,
		"message_id":      msg.ID,
		"sender_id":       senderID,
		"message":         message,
		"message_type":    messageType,
		"sent_at":         msg.SentAt.Unix(),
	}

	s.wsHub.SendToRoom(roomID, "new_message", messageData)

	// Send push notification if recipient is offline
	if !s.wsHub.IsUserOnline(receiverID) {
		s.sendPushNotification(receiverID, "New Message", message, map[string]interface{}{
			"conversation_id": conversationID,
			"sender_id":       senderID,
		})
	}

	return msg, nil
}

// GetMessages retrieves messages for a conversation
func (s *ChatService) GetMessages(conversationID uint64, userID uint64, page, perPage int) ([]models.Message, int64, error) {
	// Verify user is participant in conversation
	var conversation models.Conversation
	err := s.db.Where("id = ? AND (vendor_id = ? OR customer_id = ?)", conversationID, userID, userID).
		First(&conversation).Error
	if err != nil {
		return nil, 0, fmt.Errorf("conversation not found: %w", err)
	}

	var messages []models.Message
	var total int64

	offset := (page - 1) * perPage

	// Get total count
	s.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID).Count(&total)

	// Get messages
	err = s.db.Where("conversation_id = ?", conversationID).
		Order("sent_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&messages).Error

	// Reverse order for display
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, total, err
}

// MarkMessagesAsRead marks messages as read for a user
func (s *ChatService) MarkMessagesAsRead(conversationID, userID uint64) error {
	// Verify conversation exists
	var conversation models.Conversation
	err := s.db.Where("id = ? AND (vendor_id = ? OR customer_id = ?)", conversationID, userID, userID).
		First(&conversation).Error
	if err != nil {
		return fmt.Errorf("conversation not found: %w", err)
	}

	now := time.Now()

	// Update messages
	updateData := map[string]interface{}{
		"is_read": true,
		"read_at": now,
	}

	err = s.db.Model(&models.Message{}).
		Where("conversation_id = ? AND receiver_id = ? AND is_read = ?", conversationID, userID, false).
		Updates(updateData).Error
	if err != nil {
		return err
	}

	// Update conversation read status
	convUpdateData := map[string]interface{}{}
	if userID == conversation.VendorID {
		convUpdateData["vendor_read"] = true
	} else {
		convUpdateData["customer_read"] = true
	}

	err = s.db.Model(&conversation).Updates(convUpdateData).Error
	if err != nil {
		return err
	}

	// Notify other user
	roomID := fmt.Sprintf("conversation_%d", conversationID)
	s.wsHub.SendToRoom(roomID, "messages_read", map[string]interface{}{
		"conversation_id": conversationID,
		"user_id":        userID,
		"read_at":        now.Unix(),
	})

	return nil
}

// GetUnreadCount returns the count of unread messages for a user
func (s *ChatService) GetUnreadCount(userID uint64) (int64, error) {
	var count int64
	err := s.db.Model(&models.Message{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// UpdateConversationStatus updates conversation status
func (s *ChatService) UpdateConversationStatus(conversationID, userID uint64, status string) error {
	// Verify user is participant
	var conversation models.Conversation
	err := s.db.Where("id = ? AND (vendor_id = ? OR customer_id = ?)", conversationID, userID, userID).
		First(&conversation).Error
	if err != nil {
		return fmt.Errorf("conversation not found: %w", err)
	}

	// Update status
	err = s.db.Model(&conversation).Update("status", status).Error
	if err != nil {
		return err
	}

	// Notify participants
	roomID := fmt.Sprintf("conversation_%d", conversationID)
	s.wsHub.SendToRoom(roomID, "conversation_status_update", map[string]interface{}{
		"conversation_id": conversationID,
		"status":         status,
		"updated_by":      userID,
	})

	return nil
}

// ArchiveConversation archives a conversation
func (s *ChatService) ArchiveConversation(conversationID, userID uint64) error {
	return s.UpdateConversationStatus(conversationID, userID, "archived")
}

// sendPushNotification sends a push notification to a user
func (s *ChatService) sendPushNotification(userID uint64, title, body string, data map[string]interface{}) error {
	// Get user's active devices
	var devices []models.DeviceRegistration
	err := s.db.Where("user_id = ? AND is_active = ?", userID, true).
		Find(&devices).Error
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return nil
	}

	// Check user notification preferences
	var pref models.NotificationPreference
	err = s.db.Where("user_id = ?", userID).First(&pref).Error
	if err == nil {
		if !pref.PushEnabled || !pref.MessageAlerts {
			return nil
		}
	}

	// Send push notification to all devices
	for _, device := range devices {
		// In production, integrate with FCM/APNS
		// For now, just log
		fmt.Printf("[Push] To: %d, Device: %s, Title: %s, Body: %s", userID, device.DeviceToken, title, body)
	}

	return nil
}

// CreateSystemMessage creates a system message in a conversation
func (s *ChatService) CreateSystemMessage(conversationID uint64, message string) error {
	msg := &models.Message{
		ConversationID: conversationID,
		SenderID:       0, // System messages have sender_id = 0
		ReceiverID:     0,
		Message:        message,
		MessageType:    "system",
		IsRead:         true,
		SentAt:         time.Now(),
	}

	if err := s.db.Create(msg).Error; err != nil {
		return err
	}

	// Notify room
	roomID := fmt.Sprintf("conversation_%d", conversationID)
	s.wsHub.SendToRoom(roomID, "system_message", map[string]interface{}{
		"conversation_id": conversationID,
		"message":        message,
		"sent_at":        msg.SentAt.Unix(),
	})

	return nil
}

// GetTypingStatus handles typing status updates
func (s *ChatService) GetTypingStatus(conversationID uint64) map[uint64]bool {
	// In production, use Redis to track typing status
	// For now, return empty map
	return make(map[uint64]bool)
}

// SetTypingStatus sets typing status for a user
func (s *ChatService) SetTypingStatus(conversationID, userID uint64, isTyping bool) {
	roomID := fmt.Sprintf("conversation_%d", conversationID)
	s.wsHub.SendToRoom(roomID, "typing_status", map[string]interface{}{
		"conversation_id": conversationID,
		"user_id":        userID,
		"is_typing":      isTyping,
	})
}

// UploadAttachment handles file attachments in messages
func (s *ChatService) UploadAttachment(conversationID, senderID, receiverID uint64, attachmentURL, filename string) (*models.Message, error) {
	messageText := fmt.Sprintf("📎 File: %s", filename)
	return s.SendMessage(conversationID, senderID, receiverID, messageText, "file")
}

// DeleteConversation deletes a conversation and all messages
func (s *ChatService) DeleteConversation(conversationID, userID uint64) error {
	// Verify user is participant
	var conversation models.Conversation
	err := s.db.Where("id = ? AND (vendor_id = ? OR customer_id = ?)", conversationID, userID, userID).
		First(&conversation).Error
	if err != nil {
		return fmt.Errorf("conversation not found: %w", err)
	}

	// Delete messages
	err = s.db.Where("conversation_id = ?", conversationID).Delete(&models.Message{}).Error
	if err != nil {
		return err
	}

	// Delete conversation
	err = s.db.Delete(&conversation).Error
	if err != nil {
		return err
	}

	// Notify participants
	roomID := fmt.Sprintf("conversation_%d", conversationID)
	s.wsHub.SendToRoom(roomID, "conversation_deleted", map[string]interface{}{
		"conversation_id": conversationID,
		"deleted_by":      userID,
	})

	return nil
}

// SearchMessages searches for messages across conversations
func (s *ChatService) SearchMessages(userID uint64, query string, page, perPage int) ([]models.Message, int64, error) {
	var messages []models.Message
	var total int64

	offset := (page - 1) * perPage

	// Get conversations user is part of
	var conversationIDs []uint64
	s.db.Model(&models.Conversation{}).
		Where("vendor_id = ? OR customer_id = ?", userID, userID).
		Pluck("id", &conversationIDs)

	if len(conversationIDs) == 0 {
		return messages, 0, nil
	}

	// Search messages
	searchQuery := s.db.Model(&models.Message{}).
		Where("conversation_id IN ?", conversationIDs).
		Where("message ILIKE ?", "%"+query+"%").
		Where("message_type = ?", "text")

	searchQuery.Count(&total)

	err := searchQuery.Order("sent_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

// GetConversationStats returns statistics for a conversation
func (s *ChatService) GetConversationStats(conversationID uint64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Message count
	var messageCount int64
	s.db.Model(&models.Message{}).Where("conversation_id = ?", conversationID).Count(&messageCount)
	stats["message_count"] = messageCount

	// Unread count for vendor
	var vendorUnread int64
	s.db.Model(&models.Message{}).
		Where("conversation_id = ? AND receiver_id = (SELECT vendor_id FROM conversations WHERE id = ?) AND is_read = ?", conversationID, conversationID, false).
		Count(&vendorUnread)
	stats["vendor_unread"] = vendorUnread

	// Unread count for customer
	var customerUnread int64
	s.db.Model(&models.Message{}).
		Where("conversation_id = ? AND receiver_id = (SELECT customer_id FROM conversations WHERE id = ?) AND is_read = ?", conversationID, conversationID, false).
		Count(&customerUnread)
	stats["customer_unread"] = customerUnread

	// Last message time
	var conversation models.Conversation
	s.db.Where("id = ?", conversationID).First(&conversation)
	stats["last_message_at"] = conversation.LastMessageAt
	stats["last_message"] = conversation.LastMessage

	return stats, nil
}
