package notification

import (
	"fmt"
	"log"
	"time"

	"genrent/internal/models"
	"genrent/internal/services/websocket"

	"gorm.io/gorm"
)

// NotificationService handles all notification operations
type NotificationService struct {
	db     *gorm.DB
	wsHub  *websocket.Hub
	fcmService *FCMService
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB, wsHub *websocket.Hub, fcmService *FCMService) *NotificationService {
	return &NotificationService{
		db:         db,
		wsHub:      wsHub,
		fcmService: fcmService,
	}
}

// CreateNotification creates a new notification for a user
func (s *NotificationService) CreateNotification(userID uint64, notifType, title, message string, data map[string]interface{}) error {
	// Check user notification preferences
	pref, err := s.getUserPreferences(userID)
	if err == nil {
		if !pref.InAppEnabled {
			return nil // User has disabled notifications
		}
	}

	// Create notification record
	notification := &models.Notification{
		UserID:  uint(userID),
		Type:    models.NotificationType(notifType),
		Title:   title,
		Message: message,
		Read:    false,
	}

	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send real-time notification if user is online
	if s.wsHub.IsUserOnline(userID) {
		s.wsHub.SendToUser(userID, "new_notification", map[string]interface{}{
			"notification_id": notification.ID,
			"type":           notifType,
			"title":          title,
			"message":        message,
			"data":           data,
			"created_at":     notification.CreatedAt.Unix(),
		})
	} else if pref.PushEnabled {
		// Send push notification if user is offline
		s.sendPushNotification(userID, title, message, data)
	}

	return nil
}

// SendBookingNotification sends a booking-related notification
func (s *NotificationService) SendBookingNotification(bookingID uint64, notifType models.NotificationType, recipientID uint64, details map[string]interface{}) error {
	// Get booking details
	var booking struct {
		ID         uint64
		CustomerID uint64
		VendorID   uint64
		Status     string
		EquipmentName string
		StartDate  time.Time
		EndDate    time.Time
	}

	err := s.db.Table("bookings").
		Select("bookings.id, bookings.customer_id, equipment.vendor_id, equipment.name as equipment_name, bookings.status, bookings.start_date, bookings.end_date").
		Joins("JOIN equipment ON equipment.id = bookings.equipment_id").
		Where("bookings.id = ?", bookingID).
		First(&booking).Error

	if err != nil {
		return fmt.Errorf("failed to get booking details: %w", err)
	}

	// Determine vendor or customer
	var isVendor bool
	if booking.VendorID == recipientID {
		isVendor = true
	}

	// Create notification title and message based on type
	var title, message string
	var link string

	switch notifType {
	case models.NotifBookingRequested:
		if isVendor {
			title = "🔔 New Booking Request!"
			message = fmt.Sprintf("You have a new booking request for %s", booking.EquipmentName)
			link = fmt.Sprintf("/vendor/bookings/%d", bookingID)
		} else {
			return nil // Customer doesn't get notified of their own request
		}

	case models.NotifBookingAccepted:
		if isVendor {
			return nil // Vendor doesn't get notified of their own acceptance
		} else {
			title = "✅ Booking Accepted!"
			message = fmt.Sprintf("Your booking for %s has been accepted. Please pay advance to confirm.", booking.EquipmentName)
			link = fmt.Sprintf("/customer/bookings/%d", bookingID)
		}

	case models.NotifBookingRejected:
		if isVendor {
			return nil
		} else {
			title = "❌ Booking Update"
			message = fmt.Sprintf("Your booking request for %s could not be accepted.", booking.EquipmentName)
			link = fmt.Sprintf("/customer/bookings/%d", bookingID)
		}

	case models.NotifAdvancePaid:
		if isVendor {
			title = "💰 Advance Payment Received!"
			message = fmt.Sprintf("Advance payment received for booking #%d. Equipment will be dispatched soon.", bookingID)
			link = fmt.Sprintf("/vendor/bookings/%d", bookingID)
		} else {
			title = "✅ Payment Confirmed!"
			message = "Your advance payment has been received. Booking is now confirmed."
			link = fmt.Sprintf("/customer/bookings/%d", bookingID)
		}

	case models.NotifDispatched:
		if isVendor {
			return nil
		} else {
			title = "🚚 Equipment Dispatched!"
			message = fmt.Sprintf("Your %s has been dispatched and will reach you soon.", booking.EquipmentName)
			link = fmt.Sprintf("/customer/bookings/%d", bookingID)
		}

	case models.NotifDelivered:
		if isVendor {
			title = "📦 Equipment Delivered!"
			message = "Equipment has been delivered to customer. Payment will be released after confirmation."
			link = fmt.Sprintf("/vendor/bookings/%d", bookingID)
		} else {
			title = "📦 Equipment Delivered!"
			message = "Equipment has been delivered. Please confirm after inspection."
			link = fmt.Sprintf("/customer/bookings/%d", bookingID)
		}

	case models.NotifCompleted:
		if isVendor {
			title = "✅ Booking Completed!"
			message = fmt.Sprintf("Booking #%d has been completed. Payment released to your wallet.", bookingID)
			link = fmt.Sprintf("/vendor/bookings/%d", bookingID)
		} else {
			title = "✅ Booking Completed!"
			message = "Your booking has been completed successfully. Thank you for using GenRent!"
			link = fmt.Sprintf("/customer/bookings/%d", bookingID)
		}

	case models.NotifCancelled:
		title = "🔄 Booking Cancelled"
		message = fmt.Sprintf("Booking #%d has been cancelled. Refund will be processed if applicable.", bookingID)
		link = fmt.Sprintf("/bookings/%d", bookingID)
	}

	notificationData := map[string]interface{}{
		"booking_id":   bookingID,
		"type":         string(notifType),
		"link":        link,
		"equipment_name": booking.EquipmentName,
		"start_date":   booking.StartDate.Unix(),
		"end_date":     booking.EndDate.Unix(),
	}

	for k, v := range details {
		notificationData[k] = v
	}

	return s.CreateNotification(recipientID, string(notifType), title, message, notificationData)
}

// SendBulkNotifications sends notifications to multiple users
func (s *NotificationService) SendBulkNotifications(userIDs []uint64, notifType, title, message string, data map[string]interface{}) error {
	for _, userID := range userIDs {
		if err := s.CreateNotification(userID, notifType, title, message, data); err != nil {
			log.Printf("Failed to send notification to user %d: %v", userID, err)
		}
	}
	return nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID, userID uint64) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(userID uint64) error {
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Update("read", true).Error
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID uint64) (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// GetNotifications retrieves notifications for a user
func (s *NotificationService) GetNotifications(userID uint64, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.db.Where("user_id = ?", userID).
		Order("read ASC, created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID, userID uint64) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// CleanOldNotifications removes notifications older than specified days
func (s *NotificationService) CleanOldNotifications(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return s.db.Where("created_at < ? AND read = ?", cutoffDate, true).
		Delete(&models.Notification{}).Error
}

// getUserPreferences retrieves user notification preferences
func (s *NotificationService) getUserPreferences(userID uint64) (*models.NotificationPreference, error) {
	var pref models.NotificationPreference
	err := s.db.Where("user_id = ?", userID).First(&pref).Error
	if err != nil {
		// Return default preferences if not found
		return &models.NotificationPreference{
			UserID:         userID,
			EmailEnabled:   true,
			SMSEnabled:     false,
			PushEnabled:    true,
			InAppEnabled:   true,
			BookingUpdates: true,
			MessageAlerts:  true,
			Promotions:     false,
			Reviews:        true,
			Availability:   true,
		}, nil
	}
	return &pref, nil
}

// sendPushNotification sends a push notification via FCM/APNS
func (s *NotificationService) sendPushNotification(userID uint64, title, message string, data map[string]interface{}) error {
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

	// Check if quiet hours are enabled
	pref, err := s.getUserPreferences(userID)
	if err == nil && pref.QuietHoursEnd != "" && pref.QuietHoursStart != "" {
		currentTime := time.Now()
		currentHour := currentTime.Hour()

		// Parse quiet hours (simple check for current hour)
		// In production, use proper timezone handling
		startHour := 22 // Default 10 PM
		endHour := 8    // Default 8 AM

		if (currentHour >= startHour || currentHour < endHour) {
			// During quiet hours, only send critical notifications
			if data["type"] != "critical" {
				return nil
			}
		}
	}

	// Send push notification to all devices
	for _, device := range devices {
		if s.fcmService != nil {
			err := s.fcmService.SendPush(device.DeviceToken, device.Platform, title, message, data)
			if err != nil {
				log.Printf("Failed to send push to device %s: %v", device.DeviceToken, err)
			}
		}
	}

	return nil
}

// CreateNotificationPreferences creates default notification preferences for a user
func (s *NotificationService) CreateNotificationPreferences(userID uint64) error {
	pref := &models.NotificationPreference{
		UserID:         userID,
		EmailEnabled:   true,
		SMSEnabled:     false,
		PushEnabled:    true,
		InAppEnabled:   true,
		BookingUpdates: true,
		MessageAlerts:  true,
		Promotions:     false,
		Reviews:        true,
		Availability:   true,
		QuietHoursStart: "22:00",
		QuietHoursEnd:   "08:00",
		Timezone:       "Asia/Kolkata",
	}

	return s.db.Create(pref).Error
}

// UpdateNotificationPreferences updates user notification preferences
func (s *NotificationService) UpdateNotificationPreferences(userID uint64, updates map[string]interface{}) error {
	return s.db.Model(&models.NotificationPreference{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

// GetNotificationPreferences retrieves user notification preferences
func (s *NotificationService) GetNotificationPreferences(userID uint64) (*models.NotificationPreference, error) {
	var pref models.NotificationPreference
	err := s.db.Where("user_id = ?", userID).First(&pref).Error
	if err != nil {
		return nil, err
	}
	return &pref, nil
}
