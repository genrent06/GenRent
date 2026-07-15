package workers

import (
	"fmt"
	"genrent/internal/models"
	"log"
	"math"
	"time"

	"gorm.io/gorm"
)

const PenaltyNoDispatch = 500.0

// StartExpiryWorker runs a background goroutine that checks for expired bookings every 5 minutes
func StartExpiryWorker(db *gorm.DB) {
	go func() {
		runExpiryCheck(db)
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			runExpiryCheck(db)
		}
	}()
	log.Println("[Worker] Booking expiry worker started (checks every 5 minutes)")
}

func runExpiryCheck(db *gorm.DB) {
	now := time.Now()

	// 1. Auto-cancel requested bookings where vendor didn't respond > 2 hours
	var requestedExpired []models.Booking
	db.Where("status = ? AND created_at < ?", models.BookingRequested, now.Add(-2*time.Hour)).
		Find(&requestedExpired)
	for _, b := range requestedExpired {
		db.Model(&b).Updates(map[string]interface{}{
			"status":        models.BookingCancelled,
			"cancel_reason": "Auto-cancelled: vendor did not respond within 2 hours",
		})
		db.Model(&models.Generator{}).Where("id = ?", b.GeneratorID).
			Update("availability_status", models.StatusAvailable)
		workerNotify(db, b.CustomerID, &b.ID, models.NotifCancelled,
			"Booking Auto-Cancelled",
			"Vendor did not respond within 2 hours. No charge has been made.")
		log.Printf("[Expiry] Cancelled booking #%d — vendor response timeout", b.ID)
	}

	// 2. Auto-cancel accepted bookings where customer didn't pay advance > 1 hour
	var acceptedExpired []models.Booking
	db.Where("status = ? AND accepted_at IS NOT NULL AND accepted_at < ?",
		models.BookingAccepted, now.Add(-1*time.Hour)).
		Find(&acceptedExpired)
	for _, b := range acceptedExpired {
		db.Model(&b).Updates(map[string]interface{}{
			"status":        models.BookingCancelled,
			"cancel_reason": "Auto-cancelled: advance payment not received within 1 hour of acceptance",
		})
		db.Model(&models.Generator{}).Where("id = ?", b.GeneratorID).
			Update("availability_status", models.StatusAvailable)
		// Mark any pending payment session for this booking as expired
		db.Model(&models.Payment{}).
			Where("booking_id = ? AND status = ?", b.ID, models.PaymentPending).
			Update("status", models.PaymentExpired)
		workerNotify(db, b.CustomerID, &b.ID, models.NotifCancelled,
			"Booking Auto-Cancelled",
			"Advance payment was not completed within 1 hour. Booking has been cancelled.")
		log.Printf("[Expiry] Cancelled booking #%d — payment timeout", b.ID)
	}

	// 3. Auto-cancel advance_paid bookings where vendor didn't dispatch > 3 hours + penalize
	var dispatchExpired []models.Booking
	db.Preload("Generator").
		Where("status = ? AND advance_paid = true AND updated_at < ?",
			models.BookingAdvancePaid, now.Add(-3*time.Hour)).
		Find(&dispatchExpired)
	for _, b := range dispatchExpired {
		workerRefundEscrow(db, b)
		db.Model(&b).Updates(map[string]interface{}{
			"status":        models.BookingCancelled,
			"cancel_reason": "Auto-cancelled: vendor did not dispatch within 3 hours. Full refund issued.",
		})
		db.Model(&models.Generator{}).Where("id = ?", b.GeneratorID).
			Update("availability_status", models.StatusAvailable)
		workerPenalizeVendor(db, b.Generator.VendorID, &b.ID, PenaltyNoDispatch,
			"Failed to dispatch generator within 3 hours of advance payment")
		workerNotify(db, b.CustomerID, &b.ID, models.NotifCancelled,
			"Booking Cancelled — Full Refund",
			fmt.Sprintf("Vendor failed to dispatch your generator. Advance of ₹%.0f has been refunded.", b.AdvanceAmount))
		log.Printf("[Expiry] Cancelled booking #%d — dispatch timeout, vendor penalized ₹%.0f", b.ID, PenaltyNoDispatch)
	}

	// 4. Release expired generator reservations (30-min lock)
	rowsReleased := db.Model(&models.Generator{}).
		Where("availability_status = ? AND reservation_expiry IS NOT NULL AND reservation_expiry < ?",
			models.StatusReserved, now).
		Updates(map[string]interface{}{
			"availability_status": models.StatusAvailable,
			"reservation_expiry":  nil,
		}).RowsAffected
	if rowsReleased > 0 {
		log.Printf("[Expiry] Released %d expired generator reservations", rowsReleased)
	}
}

func workerRefundEscrow(db *gorm.DB, booking models.Booking) {
	vendorAmount := math.Round(booking.AdvanceAmount*0.5*100) / 100
	var gen models.Generator
	if db.First(&gen, booking.GeneratorID).Error != nil {
		return
	}
	var wallet models.VendorWallet
	if db.Where("vendor_id = ?", gen.VendorID).First(&wallet).Error != nil {
		return
	}
	db.Model(&wallet).UpdateColumn("hold_balance", gorm.Expr("hold_balance - ?", vendorAmount))
	db.Create(&models.WalletTransaction{
		WalletID:    wallet.ID,
		BookingID:   &booking.ID,
		Amount:      vendorAmount,
		Type:        models.WalletDebit,
		Description: fmt.Sprintf("Escrow refunded — booking #%d auto-cancelled (vendor no-dispatch)", booking.ID),
	})
}

func workerPenalizeVendor(db *gorm.DB, vendorID uint, bookingID *uint, amount float64, reason string) {
	var wallet models.VendorWallet
	if db.Where("vendor_id = ?", vendorID).First(&wallet).Error != nil {
		wallet = models.VendorWallet{VendorID: vendorID}
		db.Create(&wallet)
		db.Where("vendor_id = ?", vendorID).First(&wallet)
	}
	if wallet.Balance >= amount {
		db.Model(&wallet).UpdateColumn("balance", gorm.Expr("balance - ?", amount))
	}
	db.Create(&models.WalletTransaction{
		WalletID:    wallet.ID,
		BookingID:   bookingID,
		Amount:      amount,
		Type:        models.WalletDebit,
		Description: fmt.Sprintf("Penalty: %s", reason),
	})

	var vendor models.Vendor
	if db.First(&vendor, vendorID).Error != nil {
		return
	}
	db.Model(&vendor).UpdateColumn("cancelled_bookings", gorm.Expr("cancelled_bookings + 1"))

	if vendor.TotalBookings > 0 {
		deliveryRate := float64(vendor.SuccessfulDeliveries) / float64(vendor.TotalBookings) * 5.0
		var score float64
		if vendor.TotalRatings > 0 {
			score = deliveryRate*0.6 + vendor.AverageRating*0.4
		} else {
			score = deliveryRate
		}
		db.Model(&vendor).UpdateColumn("reliability_score", math.Round(score*100)/100)
	}

	workerNotify(db, vendor.UserID, bookingID, models.NotifVendorPenalized,
		"Penalty Deducted from Wallet",
		fmt.Sprintf("₹%.0f has been deducted from your wallet. Reason: %s", amount, reason))
}

func workerNotify(db *gorm.DB, userID uint, bookingID *uint, notifType models.NotificationType, title, message string) {
	db.Create(&models.Notification{
		UserID:    userID,
		BookingID: bookingID,
		Type:      notifType,
		Title:     title,
		Message:   message,
	})
}
