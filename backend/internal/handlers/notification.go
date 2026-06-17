package handlers

import (
	"fmt"
	"genrent/internal/middleware"
	"genrent/internal/models"
	"genrent/internal/services/email"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// emailCfg is set once at startup via InitEmail.
var emailCfg email.Config

// InitEmail wires the email config into the handlers package.
func InitEmail(cfg email.Config) {
	emailCfg = cfg
}

// GetNotifications — returns all notifications for the logged-in user (unread first)
func GetNotifications(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var notifications []models.Notification
		db.Where("user_id = ?", userID).
			Order("read ASC, created_at DESC").
			Limit(50).
			Find(&notifications)

		var unreadCount int64
		db.Model(&models.Notification{}).Where("user_id = ? AND read = false", userID).Count(&unreadCount)

		c.JSON(http.StatusOK, gin.H{
			"notifications": notifications,
			"unread_count":  unreadCount,
		})
	}
}

// MarkNotificationRead — marks a single notification as read
func MarkNotificationRead(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		result := db.Model(&models.Notification{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("read", true)
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
	}
}

// MarkAllNotificationsRead — marks all notifications for user as read
func MarkAllNotificationsRead(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		db.Model(&models.Notification{}).Where("user_id = ? AND read = false", userID).Update("read", true)
		c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
	}
}

// createNotif saves an in-app notification and optionally sends an email.
func createNotif(db *gorm.DB, userID uint, bookingID uint, notifType models.NotificationType, title, message string) {
	bid := bookingID
	db.Create(&models.Notification{
		UserID:    userID,
		BookingID: &bid,
		Type:      notifType,
		Title:     title,
		Message:   message,
	})
}

// sendBookingEmail resolves the user's email from the DB and dispatches the
// correct template based on the notification type.
func sendBookingEmail(db *gorm.DB, userID uint, notifType models.NotificationType, data email.EmailData) {
	if !emailCfg.Enabled {
		return
	}

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		return
	}
	data.To = user.Email
	data.ToName = user.Name

	var htmlBody string
	var subject string

	switch notifType {
	case models.NotifBookingRequested:
		subject = fmt.Sprintf("New Booking Request #%d — GenRent", data.BookingID)
		htmlBody = email.BookingRequested(data)

	case models.NotifBookingAccepted:
		subject = fmt.Sprintf("Booking #%d Accepted — Pay Advance Now", data.BookingID)
		htmlBody = email.BookingAccepted(data)

	case models.NotifBookingRejected:
		subject = fmt.Sprintf("Booking #%d Update — GenRent", data.BookingID)
		htmlBody = email.BookingRejected(data)

	case models.NotifAdvancePaid:
		subject = fmt.Sprintf("Payment Received for Booking #%d", data.BookingID)
		htmlBody = email.PaymentReceived(data)

	case models.NotifDispatched:
		subject = fmt.Sprintf("Your Generator is On the Way! — Booking #%d", data.BookingID)
		htmlBody = email.GeneratorDispatched(data)

	case models.NotifDelivered:
		subject = fmt.Sprintf("Delivery Confirmed — Payment Released #%d", data.BookingID)
		htmlBody = email.DeliveryConfirmed(data)

	case models.NotifCancelled:
		subject = fmt.Sprintf("Booking #%d Cancelled — GenRent", data.BookingID)
		htmlBody = email.BookingCancelled(data)

	case models.NotifCompleted:
		subject = fmt.Sprintf("Booking #%d Completed — Thank You!", data.BookingID)
		htmlBody = email.BookingCompleted(data)

	default:
		return
	}

	data.Subject = subject
	email.Send(emailCfg, data, htmlBody)
}
