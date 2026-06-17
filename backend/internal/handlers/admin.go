package handlers

import (
	"fmt"
	"genrent/internal/models"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminListVendors(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset := (page - 1) * limit
		verified := c.Query("verified")

		query := db.Preload("User")
		if verified == "true" {
			query = query.Where("verified = ?", true)
		} else if verified == "false" {
			query = query.Where("verified = ?", false)
		}

		var vendors []models.Vendor
		var total int64
		query.Model(&models.Vendor{}).Count(&total)
		query.Limit(limit).Offset(offset).Find(&vendors)

		c.JSON(http.StatusOK, gin.H{"vendors": vendors, "total": total, "page": page})
	}
}

func AdminVerifyVendor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var vendor models.Vendor
		if result := db.First(&vendor, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
			return
		}
		db.Model(&vendor).Update("verified", true)
		c.JSON(http.StatusOK, gin.H{"message": "vendor verified successfully"})
	}
}

func AdminRejectVendor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var vendor models.Vendor
		if result := db.First(&vendor, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
			return
		}
		db.Model(&vendor).Update("verified", false)
		c.JSON(http.StatusOK, gin.H{"message": "vendor rejected"})
	}
}

func AdminListBookings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset := (page - 1) * limit
		status := c.Query("status")

		query := db.Preload("Customer").Preload("Generator").Preload("Generator.Vendor")
		if status != "" {
			query = query.Where("bookings.status = ?", status)
		}

		var bookings []models.Booking
		var total int64
		query.Model(&models.Booking{}).Count(&total)
		query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&bookings)

		c.JSON(http.StatusOK, gin.H{"bookings": bookings, "total": total, "page": page})
	}
}

func AdminGetStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var totalVendors, verifiedVendors, totalGenerators, availableGenerators int64
		var totalBookings, requestedBookings, completedBookings, cancelledBookings int64
		var totalCustomers int64

		db.Model(&models.Vendor{}).Count(&totalVendors)
		db.Model(&models.Vendor{}).Where("verified = ?", true).Count(&verifiedVendors)
		db.Model(&models.Generator{}).Count(&totalGenerators)
		db.Model(&models.Generator{}).Where("availability_status = ?", models.StatusAvailable).Count(&availableGenerators)
		db.Model(&models.Booking{}).Count(&totalBookings)
		db.Model(&models.Booking{}).Where("status = ?", models.BookingRequested).Count(&requestedBookings)
		db.Model(&models.Booking{}).Where("status = ?", models.BookingCompleted).Count(&completedBookings)
		db.Model(&models.Booking{}).Where("status = ?", models.BookingCancelled).Count(&cancelledBookings)
		db.Model(&models.User{}).Where("role = ?", models.RoleCustomer).Count(&totalCustomers)

		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		var totalRevenue, revenueToday, revenueMonth float64
		db.Model(&models.Booking{}).Where("status = ?", models.BookingCompleted).
			Select("COALESCE(SUM(total_price), 0)").Scan(&totalRevenue)
		db.Model(&models.Booking{}).Where("status = ? AND completed_at >= ?", models.BookingCompleted, todayStart).
			Select("COALESCE(SUM(total_price), 0)").Scan(&revenueToday)
		db.Model(&models.Booking{}).Where("status = ? AND completed_at >= ?", models.BookingCompleted, monthStart).
			Select("COALESCE(SUM(total_price), 0)").Scan(&revenueMonth)

		// Platform fee revenue — SUM of platform_revenue table (net of refunds)
		var platformTotal, platformToday, platformMonth float64
		db.Model(&models.PlatformRevenue{}).
			Select("COALESCE(SUM(amount), 0)").Scan(&platformTotal)
		db.Model(&models.PlatformRevenue{}).
			Where("created_at >= ?", todayStart).
			Select("COALESCE(SUM(amount), 0)").Scan(&platformToday)
		db.Model(&models.PlatformRevenue{}).
			Where("created_at >= ?", monthStart).
			Select("COALESCE(SUM(amount), 0)").Scan(&platformMonth)

		c.JSON(http.StatusOK, gin.H{
			"vendors":    gin.H{"total": totalVendors, "verified": verifiedVendors, "pending": totalVendors - verifiedVendors},
			"generators": gin.H{"total": totalGenerators, "available": availableGenerators},
			"bookings":   gin.H{"total": totalBookings, "requested": requestedBookings, "completed": completedBookings, "cancelled": cancelledBookings},
			"customers":  totalCustomers,
			"revenue":    gin.H{"all_time": totalRevenue, "today": revenueToday, "month": revenueMonth},
			"platform_fee": gin.H{"all_time": platformTotal, "today": platformToday, "month": platformMonth},
		})
	}
}

func AdminListGenerators(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset := (page - 1) * limit

		var generators []models.Generator
		var total int64
		db.Model(&models.Generator{}).Count(&total)
		db.Preload("Vendor").Preload("Vendor.User").Limit(limit).Offset(offset).Find(&generators)

		c.JSON(http.StatusOK, gin.H{"generators": generators, "total": total, "page": page})
	}
}

func AdminUpdateGeneratorStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body struct {
			Status models.AvailabilityStatus `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}
		db.Model(&models.Generator{}).Where("id = ?", id).Update("availability_status", body.Status)
		c.JSON(http.StatusOK, gin.H{"message": "generator status updated"})
	}
}

// AdminForceCancel — admin force-cancels any booking, refunds escrow if advance was paid
func AdminForceCancel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body struct {
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var booking models.Booking
		if result := db.Preload("Generator").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if booking.Status == models.BookingCompleted || booking.Status == models.BookingCancelled {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("booking already in '%s' state", booking.Status)})
			return
		}

		reason := body.Reason
		if reason == "" {
			reason = "Force cancelled by admin"
		}

		if booking.AdvancePaid {
			refundEscrow(db, booking)
		}

		db.Model(&booking).Updates(map[string]interface{}{
			"status":        models.BookingCancelled,
			"cancel_reason": reason,
		})
		db.Model(&models.Generator{}).Where("id = ?", booking.GeneratorID).
			Updates(map[string]interface{}{
				"availability_status": models.StatusAvailable,
				"reservation_expiry":  nil,
			})

		auditLog(db, 0, "admin_force_cancel", "booking", booking.ID, string(booking.Status), string(models.BookingCancelled), c.ClientIP())

		createNotif(db, booking.CustomerID, booking.ID, models.NotifCancelled,
			"Booking Cancelled by Admin",
			fmt.Sprintf("Your booking has been cancelled by admin. Reason: %s", reason))

		c.JSON(http.StatusOK, gin.H{
			"message":  "Booking force-cancelled",
			"refunded": booking.AdvancePaid,
		})
	}
}

// AdminReleaseEscrow — admin manually releases escrow to vendor (resolves dispute in vendor's favor)
func AdminReleaseEscrow(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var booking models.Booking
		if result := db.Preload("Generator").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if !booking.AdvancePaid {
			c.JSON(http.StatusConflict, gin.H{"error": "no advance payment to release"})
			return
		}

		now := time.Now()
		db.Model(&booking).Updates(map[string]interface{}{
			"status":       models.BookingDelivered,
			"otp_verified": true,
			"delivered_at": now,
		})
		releaseEscrow(db, booking)

		auditLog(db, 0, "admin_release_escrow", "booking", booking.ID, "advance_paid", "escrow_released", c.ClientIP())

		c.JSON(http.StatusOK, gin.H{"message": "Escrow manually released to vendor by admin"})
	}
}

// AdminRefundCustomer — admin force-refunds customer (resolves dispute in customer's favor)
func AdminRefundCustomer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body struct {
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var booking models.Booking
		if result := db.Preload("Generator").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if !booking.AdvancePaid {
			c.JSON(http.StatusConflict, gin.H{"error": "no advance payment to refund"})
			return
		}

		refundEscrow(db, booking)
		reason := body.Reason
		if reason == "" {
			reason = "Admin ordered refund"
		}
		db.Model(&booking).Updates(map[string]interface{}{
			"status":        models.BookingCancelled,
			"cancel_reason": reason,
		})
		db.Model(&models.Generator{}).Where("id = ?", booking.GeneratorID).
			Updates(map[string]interface{}{
				"availability_status": models.StatusAvailable,
				"reservation_expiry":  nil,
			})

		auditLog(db, 0, "admin_refund_customer", "booking", booking.ID, "advance_paid", "refunded", c.ClientIP())

		createNotif(db, booking.CustomerID, booking.ID, models.NotifCancelled,
			"Refund Issued by Admin",
			fmt.Sprintf("Admin has ordered a refund for your booking. Reason: %s", reason))

		c.JSON(http.StatusOK, gin.H{"message": "Customer refunded and booking cancelled"})
	}
}

// AdminListWithdrawals — list all withdrawal requests (filter by status)
func AdminListWithdrawals(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", "pending")

		var withdrawals []models.WithdrawalRequest
		query := db.Preload("Vendor").Preload("Vendor.User")
		if status != "all" {
			query = query.Where("status = ?", status)
		}
		query.Order("created_at ASC").Find(&withdrawals)

		c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
	}
}

// AdminApproveWithdrawal — mark withdrawal as approved, create debit transaction
func AdminApproveWithdrawal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body struct {
			Note string `json:"note"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var wr models.WithdrawalRequest
		if db.Preload("Vendor").Preload("Vendor.User").First(&wr, id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "withdrawal request not found"})
			return
		}
		if wr.Status != models.WithdrawalPending {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("withdrawal already %s", wr.Status)})
			return
		}

		now := time.Now()
		db.Model(&wr).Updates(map[string]interface{}{
			"status":       models.WithdrawalApproved,
			"admin_note":   body.Note,
			"processed_at": now,
		})

		// Clear from withdrawal_hold (balance was already deducted on request)
		var wallet models.VendorWallet
		if db.Where("vendor_id = ?", wr.VendorID).First(&wallet).Error == nil {
			db.Model(&wallet).UpdateColumn("withdrawal_hold_balance",
				gorm.Expr("withdrawal_hold_balance - ?", wr.Amount))
			db.Create(&models.WalletTransaction{
				WalletID:    wallet.ID,
				Amount:      wr.Amount,
				Type:        models.WalletWithdrawalCompleted,
				Description: fmt.Sprintf("Withdrawal #%d completed — ₹%.0f paid to %s %s", wr.ID, wr.Amount, wr.BankName, last4(wr.AccountNo)),
			})
		}

		auditLog(db, 0, "withdrawal_approved", "withdrawal_request", wr.ID, "pending", "approved", c.ClientIP())

		createNotif(db, wr.Vendor.UserID, 0, models.NotifWithdrawalApproved,
			"Withdrawal Approved!",
			fmt.Sprintf("₹%.0f has been approved and transferred to your %s account ending %s.",
				wr.Amount, wr.BankName, last4(wr.AccountNo)))

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Withdrawal #%d approved — ₹%.0f transferred", wr.ID, wr.Amount)})
	}
}

// AdminRejectWithdrawal — reject and return amount to vendor balance
func AdminRejectWithdrawal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body struct {
			Note string `json:"note" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rejection reason is required"})
			return
		}

		var wr models.WithdrawalRequest
		if db.Preload("Vendor").Preload("Vendor.User").First(&wr, id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "withdrawal request not found"})
			return
		}
		if wr.Status != models.WithdrawalPending {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("withdrawal already %s", wr.Status)})
			return
		}

		now := time.Now()
		db.Model(&wr).Updates(map[string]interface{}{
			"status":       models.WithdrawalRejected,
			"admin_note":   body.Note,
			"processed_at": now,
		})

		// Return: withdrawal_hold → balance
		var wallet models.VendorWallet
		if db.Where("vendor_id = ?", wr.VendorID).First(&wallet).Error == nil {
			db.Model(&wallet).Updates(map[string]interface{}{
				"balance":                 gorm.Expr("balance + ?", wr.Amount),
				"withdrawal_hold_balance": gorm.Expr("withdrawal_hold_balance - ?", wr.Amount),
			})
			db.Create(&models.WalletTransaction{
				WalletID:    wallet.ID,
				Amount:      wr.Amount,
				Type:        models.WalletWithdrawalRefund,
				Description: fmt.Sprintf("Withdrawal #%d rejected — ₹%.0f returned to balance. Reason: %s", wr.ID, wr.Amount, body.Note),
			})
		}

		auditLog(db, 0, "withdrawal_rejected", "withdrawal_request", wr.ID, "pending", "rejected", c.ClientIP())

		createNotif(db, wr.Vendor.UserID, 0, models.NotifWithdrawalRejected,
			"Withdrawal Rejected",
			fmt.Sprintf("Your withdrawal of ₹%.0f was rejected. Reason: %s. Amount returned to your wallet.", wr.Amount, body.Note))

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Withdrawal #%d rejected — ₹%.0f returned to vendor wallet", wr.ID, wr.Amount)})
	}
}

func last4(s string) string {
	if len(s) <= 4 {
		return s
	}
	return "..." + s[len(s)-4:]
}

// AdminPenalizeVendor — admin manually deducts penalty from vendor wallet
func AdminPenalizeVendor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id") // vendor ID
		var body struct {
			Amount float64 `json:"amount" binding:"required,gt=0"`
			Reason string  `json:"reason" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var vendor models.Vendor
		if result := db.First(&vendor, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
			return
		}

		var wallet models.VendorWallet
		if db.Where("vendor_id = ?", vendor.ID).First(&wallet).Error != nil {
			wallet = models.VendorWallet{VendorID: vendor.ID}
			db.Create(&wallet)
			db.Where("vendor_id = ?", vendor.ID).First(&wallet)
		}

		if wallet.Balance >= body.Amount {
			db.Model(&wallet).UpdateColumn("balance", gorm.Expr("balance - ?", body.Amount))
		}
		db.Create(&models.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      body.Amount,
			Type:        models.WalletDebit,
			Description: fmt.Sprintf("Admin penalty: %s", body.Reason),
		})

		auditLog(db, 0, "admin_penalize_vendor", "vendor", vendor.ID, "0", fmt.Sprintf("%.0f", body.Amount), c.ClientIP())

		createNotif(db, vendor.UserID, 0, models.NotifVendorPenalized,
			"Penalty Applied by Admin",
			fmt.Sprintf("₹%.0f has been deducted from your wallet by admin. Reason: %s", body.Amount, body.Reason))

		c.JSON(http.StatusOK, gin.H{
			"message":  fmt.Sprintf("₹%.0f penalty applied to vendor wallet", body.Amount),
			"vendor":   vendor.CompanyName,
			"deducted": math.Min(body.Amount, wallet.Balance),
		})
	}
}
