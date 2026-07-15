package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// EscrowServiceImpl implements EscrowService interface
type EscrowServiceImpl struct {
	db              *gorm.DB
	paymentService  PaymentGateway
	platformFeeRate float64 // e.g., 10.0 for 10%
}

// NewEscrowService creates a new escrow service instance
func NewEscrowService(db *gorm.DB, paymentService PaymentGateway, platformFeeRate float64) *EscrowServiceImpl {
	return &EscrowServiceImpl{
		db:              db,
		paymentService:  paymentService,
		platformFeeRate: platformFeeRate,
	}
}

// HoldFunds holds payment in escrow until delivery confirmation
func (e *EscrowServiceImpl) HoldFunds(ctx context.Context, paymentID uint64, bookingID uint64) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		var payment struct {
			ID     uint64
			Status string
		}

		if err := tx.Table("payments").
			Select("id, status").
			Where("id = ?", paymentID).
			First(&payment).Error; err != nil {
			return fmt.Errorf("payment not found: %w", err)
		}

		if payment.Status != "paid" && payment.Status != "pending" {
			return fmt.Errorf("payment not eligible for escrow: current status %s", payment.Status)
		}

		// Update payment status to escrow
		now := time.Now()
		if err := tx.Table("payments").
			Where("id = ?", paymentID).
			Updates(map[string]interface{}{
				"status":        "escrow",
				"escrow_held_at": now,
			}).Error; err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		// Create transaction record
		if err := e.createTransaction(tx, paymentID, bookingID, "escrow_hold", 0, nil); err != nil {
			return err
		}

		return nil
	})
}

// ReleaseFunds releases funds to vendor after successful delivery
func (e *EscrowServiceImpl) ReleaseFunds(ctx context.Context, bookingID uint64) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		var payment struct {
			ID              uint64
			TotalAmount     float64
			Status          string
			VendorID        uint64
			EscrowReleasedAt *time.Time
		}

		// Find payment for this booking
		if err := tx.Table("payments").
			Select("id, total_amount, status, vendor_id, escrow_released_at").
			Where("booking_id = ?", bookingID).
			First(&payment).Error; err != nil {
			return fmt.Errorf("payment not found: %w", err)
		}

		if payment.Status != "escrow" {
			return fmt.Errorf("payment not in escrow: current status %s", payment.Status)
		}

		if payment.EscrowReleasedAt != nil {
			return fmt.Errorf("funds already released for this booking")
		}

		// Calculate platform fee and vendor amount
		platformFee := payment.TotalAmount * (e.platformFeeRate / 100)
		vendorAmount := payment.TotalAmount - platformFee

		// Update payment
		now := time.Now()
		if err := tx.Table("payments").
			Where("id = ?", payment.ID).
			Updates(map[string]interface{}{
				"status":           "completed",
				"vendor_amount":    vendorAmount,
				"platform_fee":     platformFee,
				"escrow_released_at": now,
			}).Error; err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		// Credit vendor wallet
		var wallet struct {
			ID     uint64
			Balance float64
		}

		// Get or create vendor wallet
		if err := tx.Table("vendor_wallets").
			Select("id, balance").
			Where("vendor_id = ?", payment.VendorID).
			First(&wallet).Error; err != nil {
			// Create wallet if doesn't exist
			if err := tx.Table("vendor_wallets").
				Create(map[string]interface{}{
					"vendor_id": payment.VendorID,
					"balance":   0,
				}).Error; err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}
			wallet.ID = payment.VendorID // Will be set by DB
			wallet.Balance = 0
		}

		// Update wallet balance
		if err := tx.Table("vendor_wallets").
			Where("vendor_id = ?", payment.VendorID).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance + ?", vendorAmount),
			}).Error; err != nil {
			return fmt.Errorf("failed to update wallet: %w", err)
		}

		// Create transaction record
		if err := e.createTransaction(tx, payment.ID, bookingID, "credit", vendorAmount, map[string]interface{}{
			"type":        "escrow_release",
			"platform_fee": platformFee,
			"vendor_id":   payment.VendorID,
		}); err != nil {
			return err
		}

		return nil
	})
}

// ProcessRefund processes refund for cancelled bookings
func (e *EscrowServiceImpl) ProcessRefund(ctx context.Context, bookingID uint64, reason string, amount float64) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		var payment struct {
			ID              uint64
			TotalAmount     float64
			Status          string
			Gateway         string
			GatewayPaymentID string
			RefundID        *string
			RefundAmount     float64
			RefundStatus     string
		}

		if err := tx.Table("payments").
			Select("id, total_amount, status, gateway, gateway_payment_id, refund_id, refund_amount, refund_status").
			Where("booking_id = ?", bookingID).
			First(&payment).Error; err != nil {
			return fmt.Errorf("payment not found: %w", err)
		}

		// Check if payment is refundable
		if payment.Status != "escrow" && payment.Status != "paid" {
			return fmt.Errorf("payment not refundable: current status %s", payment.Status)
		}

		// If amount not specified, refund full amount
		if amount == 0 {
			amount = payment.TotalAmount
		}

		// Process refund via gateway
		var refundResponse *RefundResponse
		var err error

		if payment.Gateway == "razorpay" || payment.Gateway == "" {
			// Will use the configured payment service
			refundResponse, err = e.paymentService.ProcessRefund(ctx, payment.GatewayPaymentID, amount)
		} else if payment.Gateway == "stripe" {
			refundResponse, err = e.paymentService.ProcessRefund(ctx, payment.GatewayPaymentID, amount)
		}

		if err != nil {
			return fmt.Errorf("gateway refund failed: %w", err)
		}

		// Update payment with refund details
		now := time.Now()
		if err := tx.Table("payments").
			Where("id = ?", payment.ID).
			Updates(map[string]interface{}{
				"status":       "refunded",
				"refund_id":    refundResponse.RefundID,
				"refund_amount": refundResponse.Amount,
				"refund_status": refundResponse.Status,
				"refunded_at":   now,
			}).Error; err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		// Create transaction record
		if err := e.createTransaction(tx, payment.ID, bookingID, "debit", refundResponse.Amount, map[string]interface{}{
			"type":   "refund",
			"reason": reason,
		}); err != nil {
			return err
		}

		return nil
	})
}

// GetEscrowStatus returns current escrow status
func (e *EscrowServiceImpl) GetEscrowStatus(ctx context.Context, bookingID uint64) (*EscrowStatus, error) {
	var payment struct {
		ID               uint64
		BookingID        uint64
		TotalAmount      float64
		Status           string
		EscrowHeldAt     *time.Time
		EscrowReleasedAt *time.Time
		VendorAmount     float64
		PlatformFee      float64
		RefundID         *string
		RefundAmount     float64
	}

	if err := e.db.Table("payments").
		Select("id, booking_id, total_amount, status, escrow_held_at, escrow_released_at, vendor_amount, platform_fee, refund_id, refund_amount").
		Where("booking_id = ?", bookingID).
		First(&payment).Error; err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	status := &EscrowStatus{
		BookingID:       payment.BookingID,
		PaymentID:       payment.ID,
		Amount:          payment.TotalAmount,
		Status:          PaymentStatus(payment.Status),
		VendorAmount:    payment.VendorAmount,
		PlatformFee:     payment.PlatformFee,
		RefundAmount:    payment.RefundAmount,
		RefundProcessed: payment.RefundID != nil,
	}

	if payment.EscrowHeldAt != nil {
		timestamp := payment.EscrowHeldAt.Unix()
		status.HeldAt = &timestamp
	}

	if payment.EscrowReleasedAt != nil {
		timestamp := payment.EscrowReleasedAt.Unix()
		status.ReleasedAt = &timestamp
	}

	return status, nil
}

// createTransaction creates a transaction record in the payment_transactions table
func (e *EscrowServiceImpl) createTransaction(tx *gorm.DB, paymentID, bookingID uint64, transType string, amount float64, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)

	transaction := map[string]interface{}{
		"payment_id":       paymentID,
		"booking_id":       bookingID,
		"transaction_type": transType,
		"amount":           amount,
		"gateway":          "escrow",
		"status":           "completed",
		"metadata":         metadataJSON,
		"created_at":       time.Now(),
	}

	return tx.Table("payment_transactions").Create(transaction).Error
}

// RefundServiceImpl handles refund operations
type RefundServiceImpl struct {
	db              *gorm.DB
	paymentService  PaymentGateway
	escrowService   EscrowService
	autoProcess     bool
}

// NewRefundService creates a new refund service instance
func NewRefundService(db *gorm.DB, paymentService PaymentGateway, escrowService EscrowService, autoProcess bool) *RefundServiceImpl {
	return &RefundServiceImpl{
		db:             db,
		paymentService:  paymentService,
		escrowService:   escrowService,
		autoProcess:     autoProcess,
	}
}

// InitiateRefund initiates a refund request
func (r *RefundServiceImpl) InitiateRefund(ctx context.Context, req RefundRequest) (*RefundResponse, error) {
	var result *RefundResponse

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var payment struct {
			ID              uint64
			BookingID       uint64
			TotalAmount     float64
			Status          string
			Gateway         string
			GatewayPaymentID string
		}

		if err := tx.Table("payments").
			Select("id, booking_id, total_amount, status, gateway, gateway_payment_id").
			Where("id = ?", req.PaymentID).
			First(&payment).Error; err != nil {
			return fmt.Errorf("payment not found: %w", err)
		}

		// Check if already refunded
		if payment.Status == "refunded" {
			return fmt.Errorf("payment already refunded")
		}

		// Create refund request record
		refundRecord := map[string]interface{}{
			"payment_id":    req.PaymentID,
			"booking_id":    payment.BookingID,
			"requested_by":  0, // Will be set from context
			"amount":        req.Amount,
			"reason":        req.Reason,
			"status":        "pending",
			"created_at":    time.Now(),
		}

		if err := tx.Table("refund_requests").Create(refundRecord).Error; err != nil {
			return fmt.Errorf("failed to create refund request: %w", err)
		}

		// Process refund if auto-process is enabled
		if r.autoProcess {
			refundResponse, err := r.paymentService.ProcessRefund(ctx, payment.GatewayPaymentID, req.Amount)
			if err != nil {
				return fmt.Errorf("gateway refund failed: %w", err)
			}

			// Store the result for return
			result = refundResponse

			// Update payment
			now := time.Now()
			if err := tx.Table("payments").
				Where("id = ?", req.PaymentID).
				Updates(map[string]interface{}{
					"status":       "refunded",
					"refund_id":    refundResponse.RefundID,
					"refund_amount": refundResponse.Amount,
					"refund_status": refundResponse.Status,
					"refunded_at":   now,
				}).Error; err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}

			// Update refund request
			if err := tx.Table("refund_requests").
				Where("payment_id = ?", req.PaymentID).
				Updates(map[string]interface{}{
					"status":       "completed",
					"processed_at": now,
				}).Error; err != nil {
				return err
			}

			return nil
		}

		return nil
	})

	// If no result was set, create a default response
	if result == nil {
		result = &RefundResponse{
			RefundID: fmt.Sprintf("refund_%d", req.PaymentID),
			Amount:   req.Amount,
			Status:   "pending",
		}
	}

	return result, err
}

// GetRefundStatus returns refund status
func (r *RefundServiceImpl) GetRefundStatus(ctx context.Context, refundID string) (*RefundStatus, error) {
	var refund struct {
		PaymentID      uint64
		Amount         float64
		Status         string
		Reason         string
		ProcessedAt    *time.Time
		GatewayRefundID string
		CreatedAt      time.Time
	}

	if err := r.db.Table("refund_requests").
		Select("payment_id, amount, status, reason, processed_at, gateway_refund_id, created_at").
		Where("gateway_refund_id = ?", refundID).
		First(&refund).Error; err != nil {
		return nil, fmt.Errorf("refund not found: %w", err)
	}

	status := &RefundStatus{
		RefundID: refundID,
		PaymentID: fmt.Sprintf("%d", refund.PaymentID),
		Amount:    refund.Amount,
		Status:    refund.Status,
		Reason:    refund.Reason,
	}

	if refund.ProcessedAt != nil {
		timestamp := refund.ProcessedAt.Unix()
		status.ProcessedAt = &timestamp
	}

	// Estimated settlement (T+2 for most refunds)
	estimatedAt := refund.CreatedAt.Add(48 * time.Hour)
	timestamp := estimatedAt.Unix()
	status.EstimatedAt = &timestamp

	return status, nil
}

// ProcessPendingRefunds processes all pending refund requests
func (r *RefundServiceImpl) ProcessPendingRefunds(ctx context.Context) error {
	var pendingRefunds []struct {
		ID         uint64
		PaymentID  uint64
		Amount     float64
		Reason     string
	}

	if err := r.db.Table("refund_requests").
		Select("id, payment_id, amount, reason").
		Where("status = ?", "pending").
		Find(&pendingRefunds).Error; err != nil {
		return fmt.Errorf("failed to fetch pending refunds: %w", err)
	}

	for _, refund := range pendingRefunds {
		_, err := r.InitiateRefund(ctx, RefundRequest{
			PaymentID: fmt.Sprintf("%d", refund.PaymentID),
			Amount:    refund.Amount,
			Reason:    refund.Reason,
		})
		if err != nil {
			// Log error but continue with other refunds
			fmt.Printf("Failed to process refund %d: %v\n", refund.ID, err)
		}
	}

	return nil
}

// CalculatePartialRefund calculates partial refund amount based on cancellation policy
func (r *RefundServiceImpl) CalculatePartialRefund(bookingID uint64, hoursBeforeStart int) (float64, string) {
	var payment struct {
		TotalAmount float64
	}

	if err := r.db.Table("payments").
		Select("total_amount").
		Where("booking_id = ?", bookingID).
		First(&payment).Error; err != nil {
		return 0, "payment not found"
	}

	// Refund policy:
	// - Full refund if cancelled 48+ hours before start
	// - 75% refund if cancelled 24-48 hours before
	// - 50% refund if cancelled 12-24 hours before
	// - 25% refund if cancelled 6-12 hours before
	// - No refund if cancelled less than 6 hours before

	var refundPercent float64
	var policy string

	switch {
	case hoursBeforeStart >= 48:
		refundPercent = 100
		policy = "Full refund (cancelled 48+ hours before start)"
	case hoursBeforeStart >= 24:
		refundPercent = 75
		policy = "75% refund (cancelled 24-48 hours before start)"
	case hoursBeforeStart >= 12:
		refundPercent = 50
		policy = "50% refund (cancelled 12-24 hours before start)"
	case hoursBeforeStart >= 6:
		refundPercent = 25
		policy = "25% refund (cancelled 6-12 hours before start)"
	default:
		refundPercent = 0
		policy = "No refund (cancelled less than 6 hours before start)"
	}

	refundAmount := payment.TotalAmount * (refundPercent / 100)

	return refundAmount, policy
}
