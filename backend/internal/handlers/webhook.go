package handlers

import (
	"fmt"
	"genrent/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WebhookPaymentPayload is the shape of an incoming payment gateway callback
type WebhookPaymentPayload struct {
	Event         string  `json:"event"`          // "payment.captured" | "payment.failed" | "refund.processed"
	TransactionID string  `json:"transaction_id"` // gateway tx ID
	BookingID     uint    `json:"booking_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"` // "success" | "failed"
	GatewayRef    string  `json:"gateway_ref"`
	// TODO: add HMAC signature field for real gateway verification
}

// HandlePaymentWebhook handles async callbacks from the payment gateway.
// In production this endpoint must verify the gateway HMAC/signature before processing.
func HandlePaymentWebhook(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- Signature verification placeholder ---
		// sig := c.GetHeader("X-Razorpay-Signature")
		// if !verifyHMAC(sig, webhookSecret, body) {
		//     c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		//     return
		// }

		var payload WebhookPaymentPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
			return
		}

		switch payload.Event {
		case "payment.captured":
			handlePaymentCaptured(db, c, payload)
		case "payment.failed":
			handlePaymentFailed(db, c, payload)
		case "refund.processed":
			handleRefundProcessed(db, c, payload)
		default:
			// Unknown event — acknowledge receipt so gateway doesn't retry
			c.JSON(http.StatusOK, gin.H{"message": "event acknowledged", "event": payload.Event})
		}
	}
}

func handlePaymentCaptured(db *gorm.DB, c *gin.Context, p WebhookPaymentPayload) {
	if p.BookingID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking_id required"})
		return
	}

	// Find or create payment record for this booking
	var payment models.Payment
	result := db.Where("booking_id = ? AND transaction_id = ?", p.BookingID, p.TransactionID).First(&payment)
	if result.Error == nil {
		// Already recorded — idempotent ack
		c.JSON(http.StatusOK, gin.H{"message": "already processed", "idempotent": true})
		return
	}

	var booking models.Booking
	if db.Preload("Generator.Vendor").First(&booking, p.BookingID).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	now := time.Now()
	advance := booking.AdvanceAmount
	vendorEscrow := advance * 0.5
	platformFee := advance * 0.5

	newPayment := models.Payment{
		BookingID:     booking.ID,
		TotalAmount:   booking.TotalPrice,
		AdvanceAmount: advance,
		VendorAmount:  vendorEscrow,
		PlatformFee:   platformFee,
		Method:        models.PaymentMethod(p.GatewayRef),
		Status:        models.PaymentCompleted,
		TransactionID: p.TransactionID,
		PaidAt:        &now,
	}
	db.Create(&newPayment)

	db.Model(&booking).Updates(map[string]interface{}{
		"advance_paid": true,
		"status":       models.BookingAdvancePaid,
	})

	// Update escrow hold on vendor wallet
	var wallet models.VendorWallet
	if db.Where("vendor_id = ?", booking.Generator.VendorID).First(&wallet).Error != nil {
		wallet = models.VendorWallet{VendorID: booking.Generator.VendorID}
		db.Create(&wallet)
	}
	db.Model(&wallet).UpdateColumn("hold_balance", gorm.Expr("hold_balance + ?", vendorEscrow))
	db.Create(&models.WalletTransaction{
		WalletID:    wallet.ID,
		BookingID:   &booking.ID,
		Amount:      vendorEscrow,
		Type:        models.WalletEscrowHold,
		Description: fmt.Sprintf("Escrow hold via webhook for booking #%d", booking.ID),
	})

	auditLog(db, 0, "webhook_payment_captured", "booking", booking.ID,
		string(models.BookingAccepted), string(models.BookingAdvancePaid), c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "payment captured and booking updated"})
}

func handlePaymentFailed(db *gorm.DB, c *gin.Context, p WebhookPaymentPayload) {
	var booking models.Booking
	if db.First(&booking, p.BookingID).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	auditLog(db, 0, "webhook_payment_failed", "booking", booking.ID,
		"payment_initiated", "payment_failed", c.ClientIP())

	createNotif(db, booking.CustomerID, booking.ID, models.NotifCancelled,
		"Payment Failed",
		fmt.Sprintf("Your payment for booking #%d failed. Please try again.", booking.ID))

	c.JSON(http.StatusOK, gin.H{"message": "payment failure recorded"})
}

func handleRefundProcessed(db *gorm.DB, c *gin.Context, p WebhookPaymentPayload) {
	var booking models.Booking
	if db.First(&booking, p.BookingID).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	auditLog(db, 0, "webhook_refund_processed", "booking", booking.ID,
		"refund_initiated", "refund_completed", c.ClientIP())

	createNotif(db, booking.CustomerID, booking.ID, models.NotifCancelled,
		"Refund Processed",
		fmt.Sprintf("Your refund of ₹%.0f for booking #%d has been processed.", p.Amount, booking.ID))

	c.JSON(http.StatusOK, gin.H{"message": "refund webhook recorded"})
}
