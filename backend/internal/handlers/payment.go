package handlers

import (
	"crypto/sha256"
	"fmt"
	"genrent/internal/middleware"
	"genrent/internal/models"
	"genrent/internal/services/email"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// hashOTP returns the SHA256 hex of the plain OTP — only the hash is stored in DB.
func hashOTP(plain string) string {
	h := sha256.Sum256([]byte(plain))
	return fmt.Sprintf("%x", h)
}

type ProcessPaymentRequest struct {
	BookingID uint                 `json:"booking_id" binding:"required"`
	Method    models.PaymentMethod `json:"method" binding:"required"`
}

// GetPaymentDetails — returns advance breakdown before payment
func GetPaymentDetails(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("booking_id")
		userID := middleware.GetUserID(c)

		var booking models.Booking
		if result := db.Preload("Generator").Preload("Generator.Vendor").
			First(&booking, bookingID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if !booking.CanTransitionTo(models.BookingAdvancePaid) {
			c.JSON(http.StatusConflict, gin.H{
				"error":  fmt.Sprintf("payment not allowed for booking in '%s' state — vendor must accept first", booking.Status),
				"status": booking.Status,
			})
			return
		}
		if booking.AdvancePaid {
			c.JSON(http.StatusConflict, gin.H{"error": "advance already paid for this booking"})
			return
		}

		advance := booking.AdvanceAmount
		escrowHold := advance                             // full 30% held in platform escrow
		vendorShare := math.Round(advance*0.5*100) / 100 // 15% → vendor (after delivery)
		platformFee := math.Round(advance*0.5*100) / 100 // 15% → platform

		c.JSON(http.StatusOK, gin.H{
			"booking":         booking,
			"total_amount":    booking.TotalPrice,
			"advance_amount":  advance,
			"escrow_hold":     escrowHold,
			"vendor_share":    vendorShare,
			"platform_fee":    platformFee,
			"note":            "Vendor receives payment ONLY after delivery is confirmed by customer OTP",
			"payment_methods": []string{"upi", "card", "netbanking", "wallet", "cash"},
		})
	}
}

// ProcessPayment — customer pays advance, money goes to escrow (NOT to vendor yet)
func ProcessPayment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req ProcessPaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var booking models.Booking
		if result := db.Preload("Generator").Preload("Generator.Vendor").
			Preload("Equipment").Preload("Equipment.Vendor").
			First(&booking, req.BookingID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		// Idempotency: if payment already completed for this booking, return cached response
		var existingPayment models.Payment
		if db.Where("booking_id = ? AND status = ?", req.BookingID, models.PaymentCompleted).
			First(&existingPayment).Error == nil {
			c.JSON(http.StatusOK, gin.H{
				"message":        "Payment already processed",
				"transaction_id": existingPayment.TransactionID,
				"advance_paid":   existingPayment.AdvanceAmount,
				"escrow_hold":    existingPayment.VendorAmount,
				"platform_fee":   existingPayment.PlatformFee,
				"booking_id":     req.BookingID,
				"idempotent":     true,
			})
			return
		}

		if !booking.CanTransitionTo(models.BookingAdvancePaid) {
			c.JSON(http.StatusConflict, gin.H{
				"error":  fmt.Sprintf("payment not allowed for booking in '%s' state — vendor must accept first", booking.Status),
				"status": booking.Status,
			})
			return
		}
		if booking.AdvancePaid {
			c.JSON(http.StatusConflict, gin.H{"error": "advance already paid"})
			return
		}

		advance := booking.AdvanceAmount
		vendorEscrow := math.Round(advance*0.5*100) / 100 // 15% held for vendor
		platformFee := math.Round(advance*0.5*100) / 100  // 15% platform keeps

		txID := fmt.Sprintf("TXN%d%04d", time.Now().Unix(), rand.Intn(9999))
		now := time.Now()

		payment := models.Payment{
			BookingID:     booking.ID,
			TotalAmount:   booking.TotalPrice,
			AdvanceAmount: advance,
			VendorAmount:  vendorEscrow,
			PlatformFee:   platformFee,
			Method:        req.Method,
			Status:        models.PaymentCompleted,
			TransactionID: txID,
			PaidAt:        &now,
		}
		if result := db.Create(&payment); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record payment"})
			return
		}

		// Mark booking as advance_paid
		db.Model(&booking).Updates(map[string]interface{}{
			"advance_paid": true,
			"status":       models.BookingAdvancePaid,
		})

		auditLog(db, userID, "payment_processed", "booking", booking.ID, string(models.BookingAccepted), string(models.BookingAdvancePaid), c.ClientIP())

		// Resolve vendor from either Equipment or Generator booking
		var vendorID uint
		var vendorUserID uint
		if booking.EquipmentID != nil && booking.Equipment != nil {
			vendorID = booking.Equipment.VendorID
			vendorUserID = booking.Equipment.Vendor.UserID
		} else if booking.GeneratorID != nil && booking.Generator != nil {
			vendorID = booking.Generator.VendorID
			vendorUserID = booking.Generator.Vendor.UserID
		}

		// Notify vendor that advance was paid
		if vendorUserID != 0 {
			createNotif(db, vendorUserID, booking.ID,
				models.NotifAdvancePaid,
				"Advance Payment Received!",
				fmt.Sprintf("Customer paid advance of ₹%.0f for booking #%d. Please dispatch the equipment.", booking.AdvanceAmount, booking.ID))
		}

		// Add vendor's share to ESCROW hold (NOT balance — released only after OTP)
		var wallet models.VendorWallet
		result := db.Where("vendor_id = ?", vendorID).First(&wallet)
		if result.Error != nil {
			wallet = models.VendorWallet{VendorID: vendorID, Balance: 0, HoldBalance: 0}
			db.Create(&wallet)
		}
		db.Model(&wallet).UpdateColumn("hold_balance", gorm.Expr("hold_balance + ?", vendorEscrow))

		escrowTx := models.WalletTransaction{
			WalletID:    wallet.ID,
			BookingID:   &booking.ID,
			Amount:      vendorEscrow,
			Type:        models.WalletEscrowHold,
			Description: fmt.Sprintf("Escrow hold for booking #%d (released after delivery OTP)", booking.ID),
		}
		db.Create(&escrowTx)

		// Record platform revenue — completely separate from vendor wallets
		db.Create(&models.PlatformRevenue{
			PaymentID:   payment.ID,
			BookingID:   booking.ID,
			Amount:      platformFee,
			Type:        models.PlatformCommission,
			Description: fmt.Sprintf("15%% commission on booking #%d advance ₹%.0f", booking.ID, advance),
		})

		c.JSON(http.StatusOK, gin.H{
			"message":        "Payment successful! Advance held in escrow. Vendor will dispatch your generator.",
			"transaction_id": txID,
			"advance_paid":   advance,
			"escrow_hold":    vendorEscrow,
			"platform_fee":   platformFee,
			"booking_id":     booking.ID,
			"note":           "Vendor receives ₹" + fmt.Sprintf("%.2f", vendorEscrow) + " only after you confirm delivery via OTP",
		})
	}
}

const (
	minWithdrawalAmount = 1000.0   // ₹1,000 minimum
	dailyWithdrawalLimit = 50000.0 // ₹50,000 per day
)

// GetVendorWallet — balance + escrow + withdrawal_hold + transactions
func GetVendorWallet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var wallet models.VendorWallet
		result := db.Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(50)
		}).Where("vendor_id = ?", vendor.ID).First(&wallet)

		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{
				"balance":                0,
				"hold_balance":           0,
				"withdrawal_hold_balance": 0,
				"transactions":           []interface{}{},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"balance":                wallet.Balance,
			"hold_balance":           wallet.HoldBalance,
			"withdrawal_hold_balance": wallet.WithdrawalHoldBalance,
			"transactions":           wallet.Transactions,
		})
	}
}

// SaveBankAccount — vendor adds a bank account for withdrawals
func SaveBankAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if db.Where("user_id = ?", userID).First(&vendor).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var body struct {
			BankName    string `json:"bank_name" binding:"required"`
			AccountNo   string `json:"account_no" binding:"required"`
			IFSC        string `json:"ifsc" binding:"required"`
			AccountName string `json:"account_name" binding:"required"`
			IsPrimary   bool   `json:"is_primary"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		// Max 5 bank accounts per vendor
		var count int64
		db.Model(&models.VendorBankAccount{}).Where("vendor_id = ?", vendor.ID).Count(&count)
		if count >= 5 {
			c.JSON(http.StatusConflict, gin.H{"error": "maximum 5 bank accounts allowed — delete one first"})
			return
		}

		if body.IsPrimary {
			db.Model(&models.VendorBankAccount{}).Where("vendor_id = ?", vendor.ID).
				Update("is_primary", false)
		}

		acc := models.VendorBankAccount{
			VendorID:    vendor.ID,
			BankName:    body.BankName,
			AccountNo:   body.AccountNo,
			IFSC:        body.IFSC,
			AccountName: body.AccountName,
			IsPrimary:   body.IsPrimary,
		}
		db.Create(&acc)

		c.JSON(http.StatusCreated, gin.H{"message": "Bank account saved", "account": acc})
	}
}

// GetBankAccounts — vendor's saved bank accounts
func GetBankAccounts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if db.Where("user_id = ?", userID).First(&vendor).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var accounts []models.VendorBankAccount
		db.Where("vendor_id = ?", vendor.ID).Order("is_primary DESC, created_at ASC").Find(&accounts)
		c.JSON(http.StatusOK, gin.H{"accounts": accounts})
	}
}

// DeleteBankAccount — vendor removes a saved bank account
func DeleteBankAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		var vendor models.Vendor
		if db.Where("user_id = ?", userID).First(&vendor).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		result := db.Where("id = ? AND vendor_id = ?", id, vendor.ID).Delete(&models.VendorBankAccount{})
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "bank account not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Bank account removed"})
	}
}

const withdrawalCooldownMinutes = 60 // 1 hour between requests

// RequestWithdrawal — step 1: validate, check cooldown/limits, generate OTP, send email.
// Money does NOT move yet — that happens on OTP confirmation.
func RequestWithdrawal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if db.Where("user_id = ?", userID).First(&vendor).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var req struct {
			Amount        float64 `json:"amount" binding:"required,gt=0"`
			BankAccountID uint    `json:"bank_account_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		req.Amount = math.Round(req.Amount*100) / 100

		if req.Amount < minWithdrawalAmount {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("minimum withdrawal amount is ₹%.0f", minWithdrawalAmount),
			})
			return
		}

		var bankAcc models.VendorBankAccount
		if db.Where("id = ? AND vendor_id = ?", req.BankAccountID, vendor.ID).First(&bankAcc).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "bank account not found"})
			return
		}

		var wallet models.VendorWallet
		if db.Where("vendor_id = ?", vendor.ID).First(&wallet).Error != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "no wallet found — earn from bookings first"})
			return
		}

		if req.Amount > wallet.Balance {
			c.JSON(http.StatusConflict, gin.H{
				"error":     fmt.Sprintf("insufficient balance — available: ₹%.2f", wallet.Balance),
				"available": wallet.Balance,
			})
			return
		}

		// Cooldown check: block if any request (including expired OTPs) was created in the last hour.
		// Expired OTPs are included so vendors can't bypass the limit by letting OTPs expire repeatedly.
		cooldownSince := time.Now().Add(-time.Duration(withdrawalCooldownMinutes) * time.Minute)
		var recentCount int64
		db.Model(&models.WithdrawalRequest{}).
			Where("vendor_id = ? AND status IN ? AND created_at >= ?",
				vendor.ID, []string{
					string(models.WithdrawalOTPPending),
					string(models.WithdrawalExpired),
					string(models.WithdrawalPending),
				}, cooldownSince).
			Count(&recentCount)
		if recentCount > 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("please wait %d minutes between withdrawal requests", withdrawalCooldownMinutes),
			})
			return
		}

		// Daily limit check
		today := time.Now().Truncate(24 * time.Hour)
		var todayTotal float64
		db.Model(&models.WithdrawalRequest{}).
			Where("vendor_id = ? AND status IN ? AND created_at >= ?",
				vendor.ID, []string{"pending", "approved", "paid"}, today).
			Select("COALESCE(SUM(amount), 0)").Scan(&todayTotal)
		if todayTotal+req.Amount > dailyWithdrawalLimit {
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("daily withdrawal limit is ₹%.0f — used ₹%.0f today",
					dailyWithdrawalLimit, todayTotal),
				"daily_limit": dailyWithdrawalLimit,
				"used_today":  todayTotal,
			})
			return
		}

		// Generate 6-digit OTP — store only the hash (plain OTP goes to email, never to DB)
		otpPlain := fmt.Sprintf("%06d", rand.Intn(1000000))
		otpExpiry := time.Now().Add(10 * time.Minute)

		// Create withdrawal in otp_pending state (money NOT moved yet)
		wr := models.WithdrawalRequest{
			VendorID:      vendor.ID,
			BankAccountID: &bankAcc.ID,
			Amount:        req.Amount,
			Status:        models.WithdrawalOTPPending,
			BankName:      bankAcc.BankName,
			AccountNo:     bankAcc.AccountNo,
			IFSC:          bankAcc.IFSC,
			AccountName:   bankAcc.AccountName,
			OTPCode:       hashOTP(otpPlain), // hash stored, never plain
			OTPExpiresAt:  &otpExpiry,
		}
		db.Create(&wr)

		// Look up vendor's user for email
		var user models.User
		db.First(&user, vendor.UserID)

		// Send OTP email with plain OTP (non-blocking) — plain OTP only ever lives in transit
		email.Send(emailCfg, email.EmailData{
			To:         user.Email,
			ToName:     user.Name,
			Subject:    fmt.Sprintf("Withdrawal OTP — ₹%.0f from GenRent Wallet", req.Amount),
			VendorName: vendor.CompanyName,
			Amount:     req.Amount,
			OTP:        otpPlain,
		}, email.WithdrawalOTPEmail(email.EmailData{
			VendorName: vendor.CompanyName,
			Amount:     req.Amount,
			OTP:        otpPlain,
		}))

		// In-app notification
		createNotif(db, vendor.UserID, 0, models.NotifWithdrawalOTPSent,
			"Withdrawal OTP Sent",
			fmt.Sprintf("An OTP has been sent to your email to confirm withdrawal of ₹%.0f. It expires in 10 minutes.", req.Amount))

		auditLog(db, userID, "withdrawal_otp_sent", "withdrawal_request", wr.ID, "", fmt.Sprintf("%.0f", req.Amount), c.ClientIP())

		c.JSON(http.StatusCreated, gin.H{
			"message":       "OTP sent to your registered email. Enter it to confirm the withdrawal.",
			"withdrawal_id": wr.ID,
			"otp_expires_in": "10 minutes",
		})
	}
}

// ConfirmWithdrawalOTP — step 2: vendor submits OTP, money moves to withdrawal_hold, status → pending
func ConfirmWithdrawalOTP(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		wrID := c.Param("id")

		var vendor models.Vendor
		if db.Where("user_id = ?", userID).First(&vendor).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var body struct {
			OTP string `json:"otp" binding:"required,len=6"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var wr models.WithdrawalRequest
		if db.Where("id = ? AND vendor_id = ?", wrID, vendor.ID).First(&wr).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "withdrawal request not found"})
			return
		}

		if wr.Status != models.WithdrawalOTPPending {
			c.JSON(http.StatusConflict, gin.H{"error": "withdrawal is not awaiting OTP confirmation"})
			return
		}

		if wr.OTPExpiresAt == nil || time.Now().After(*wr.OTPExpiresAt) {
			db.Model(&wr).Update("status", models.WithdrawalExpired)
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "OTP has expired — please start a new withdrawal request",
			})
			return
		}

		// Compare against stored hash — plain OTP is never in DB
		if wr.OTPCode != hashOTP(body.OTP) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid OTP"})
			return
		}

		// OTP verified — now move money: balance → withdrawal_hold
		var wallet models.VendorWallet
		if db.Where("vendor_id = ?", vendor.ID).First(&wallet).Error != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "wallet not found"})
			return
		}

		if wr.Amount > wallet.Balance {
			c.JSON(http.StatusConflict, gin.H{
				"error":     fmt.Sprintf("insufficient balance — available: ₹%.2f", wallet.Balance),
				"available": wallet.Balance,
			})
			return
		}

		db.Model(&wallet).Updates(map[string]interface{}{
			"balance":                 gorm.Expr("balance - ?", wr.Amount),
			"withdrawal_hold_balance": gorm.Expr("withdrawal_hold_balance + ?", wr.Amount),
		})

		db.Create(&models.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      wr.Amount,
			Type:        models.WalletWithdrawalHold,
			Description: fmt.Sprintf("Withdrawal #%d — ₹%.0f held, pending admin approval", wr.ID, wr.Amount),
		})

		// Clear OTP and transition to pending
		db.Model(&wr).Updates(map[string]interface{}{
			"status":         models.WithdrawalPending,
			"otp_code":       "",
			"otp_expires_at": nil,
		})

		auditLog(db, userID, "withdrawal_confirmed", "withdrawal_request", wr.ID, "otp_pending", "pending", c.ClientIP())

		// In-app notification
		createNotif(db, vendor.UserID, 0, models.NotifWithdrawalPending,
			"Withdrawal Under Review",
			fmt.Sprintf("Your withdrawal of ₹%.0f to %s (···%s) is confirmed and pending admin approval. Processing within 1–2 business days.",
				wr.Amount, wr.BankName, last4(wr.AccountNo)))

		c.JSON(http.StatusOK, gin.H{
			"message":       "Withdrawal confirmed! Admin will process within 1–2 business days.",
			"withdrawal_id": wr.ID,
			"amount":        wr.Amount,
		})
	}
}

// GetWithdrawals — vendor's own withdrawal history
func GetWithdrawals(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if db.Where("user_id = ?", userID).First(&vendor).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var withdrawals []models.WithdrawalRequest
		db.Where("vendor_id = ?", vendor.ID).
			Order("created_at DESC").
			Limit(50).
			Find(&withdrawals)

		c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
	}
}
