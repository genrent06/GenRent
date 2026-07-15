package handlers

import (
	"genrent/internal/services/payment"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// EscrowHandler handles escrow operations
type EscrowHandler struct {
	escrowService payment.EscrowService
	db            *gorm.DB
}

// NewEscrowHandler creates a new escrow handler
func NewEscrowHandler(es payment.EscrowService, db *gorm.DB) *EscrowHandler {
	return &EscrowHandler{
		escrowService: es,
		db:            db,
	}
}

// HoldFundsRequest represents request to hold funds in escrow
type HoldFundsRequest struct {
	PaymentID uint64 `json:"payment_id" binding:"required"`
	BookingID uint64 `json:"booking_id" binding:"required"`
}

// HoldFunds holds payment in escrow until delivery confirmation
func (h *EscrowHandler) HoldFunds(c *gin.Context) {
	var req HoldFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hold funds
	if err := h.escrowService.HoldFunds(c.Request.Context(), req.PaymentID, req.BookingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hold funds in escrow: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "funds held in escrow successfully",
		"payment_id": req.PaymentID,
		"booking_id": req.BookingID,
		"held_at":    time.Now().Unix(),
	})
}

// ReleaseFundsRequest represents request to release funds from escrow
type ReleaseFundsRequest struct {
	BookingID uint64 `json:"booking_id" binding:"required"`
}

// ReleaseFunds releases funds to vendor after successful delivery
func (h *EscrowHandler) ReleaseFunds(c *gin.Context) {
	var req ReleaseFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Release funds
	if err := h.escrowService.ReleaseFunds(c.Request.Context(), req.BookingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to release funds: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "funds released successfully",
		"booking_id": req.BookingID,
		"released_at": time.Now().Unix(),
	})
}

// ProcessRefundRequest represents request to process refund
type ProcessEscrowRefundRequest struct {
	BookingID uint64  `json:"booking_id" binding:"required"`
	Reason    string  `json:"reason" binding:"required"`
	Amount    float64 `json:"amount"`
}

// ProcessRefund processes refund for cancelled bookings
func (h *EscrowHandler) ProcessRefund(c *gin.Context) {
	var req ProcessEscrowRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process refund
	if err := h.escrowService.ProcessRefund(c.Request.Context(), req.BookingID, req.Reason, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process refund: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "refund processed successfully",
		"booking_id": req.BookingID,
		"refunded_at": time.Now().Unix(),
	})
}

// GetEscrowStatus returns current escrow status for a booking
func (h *EscrowHandler) GetEscrowStatus(c *gin.Context) {
	bookingIDStr := c.Param("booking_id")
	if bookingIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking_id is required"})
		return
	}

	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking_id"})
		return
	}

	// Get escrow status
	status, err := h.escrowService.GetEscrowStatus(c.Request.Context(), bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "escrow status not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"booking_id":        status.BookingID,
		"payment_id":        status.PaymentID,
		"amount":            status.Amount,
		"status":            string(status.Status),
		"held_at":           status.HeldAt,
		"released_at":       status.ReleasedAt,
		"vendor_amount":     status.VendorAmount,
		"platform_fee":      status.PlatformFee,
		"refund_amount":     status.RefundAmount,
		"refund_processed":  status.RefundProcessed,
	})
}

// RefundHandler handles refund automation operations
type RefundHandler struct {
	refundService payment.RefundService
	db            *gorm.DB
}

// NewRefundHandler creates a new refund handler
func NewRefundHandler(rs payment.RefundService, db *gorm.DB) *RefundHandler {
	return &RefundHandler{
		refundService: rs,
		db:            db,
	}
}

// InitiateRefundRequest represents request to initiate a refund
type InitiateRefundRequest struct {
	PaymentID string  `json:"payment_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Reason    string  `json:"reason" binding:"required"`
}

// InitiateRefund initiates a refund request
func (h *RefundHandler) InitiateRefund(c *gin.Context) {
	var req InitiateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initiate refund
	response, err := h.refundService.InitiateRefund(c.Request.Context(), payment.RefundRequest{
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
		Reason:    req.Reason,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate refund: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refund_id":   response.RefundID,
		"amount":      response.Amount,
		"status":      response.Status,
		"initiated_at": time.Now().Unix(),
	})
}

// GetRefundStatus returns refund status
func (h *RefundHandler) GetRefundStatus(c *gin.Context) {
	refundID := c.Param("refund_id")
	if refundID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refund_id is required"})
		return
	}

	// Get refund status
	status, err := h.refundService.GetRefundStatus(c.Request.Context(), refundID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "refund not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refund_id":      status.RefundID,
		"payment_id":     status.PaymentID,
		"amount":         status.Amount,
		"status":         status.Status,
		"reason":         status.Reason,
		"processed_at":   status.ProcessedAt,
		"estimated_at":   status.EstimatedAt,
	})
}

// ProcessPendingRefunds processes all pending refund requests
func (h *RefundHandler) ProcessPendingRefunds(c *gin.Context) {
	// Process pending refunds
	if err := h.refundService.ProcessPendingRefunds(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process pending refunds: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "pending refunds processed successfully",
	})
}

// CalculatePartialRefund calculates partial refund amount based on cancellation policy
func (h *RefundHandler) CalculatePartialRefund(c *gin.Context) {
	bookingIDStr := c.Query("booking_id")
	hoursBeforeStr := c.Query("hours_before_start")

	if bookingIDStr == "" || hoursBeforeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking_id and hours_before_start are required"})
		return
	}

	bookingID, err := strconv.ParseUint(bookingIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking_id"})
		return
	}

	hoursBeforeStart, err := strconv.Atoi(hoursBeforeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hours_before_start"})
		return
	}

	// Check if refund service supports partial refund calculation
	refundServiceImpl, ok := h.refundService.(*payment.RefundServiceImpl)
	if !ok {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "partial refund calculation not available"})
		return
	}

	// Calculate partial refund
	amount, policy := refundServiceImpl.CalculatePartialRefund(bookingID, hoursBeforeStart)

	c.JSON(http.StatusOK, gin.H{
		"booking_id":        bookingID,
		"hours_before_start": hoursBeforeStart,
		"refund_amount":     amount,
		"policy":            policy,
	})
}

// GetVendorWalletBalance returns vendor wallet balance
func (h *EscrowHandler) GetVendorWalletBalance(c *gin.Context) {
	vendorIDStr := c.Param("vendor_id")
	if vendorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
		return
	}

	vendorID, err := strconv.ParseUint(vendorIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor_id"})
		return
	}

	// Query vendor wallet balance
	var wallet struct {
		VendorID uint64    `json:"vendor_id"`
		Balance  float64   `json:"balance"`
		UpdateAt time.Time `json:"updated_at"`
	}

	if err := h.db.Table("vendor_wallets").
		Select("vendor_id, balance, updated_at").
		Where("vendor_id = ?", vendorID).
		First(&wallet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// GetVendorTransactions returns vendor transaction history
func (h *EscrowHandler) GetVendorTransactions(c *gin.Context) {
	vendorIDStr := c.Param("vendor_id")
	if vendorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
		return
	}

	vendorID, err := strconv.ParseUint(vendorIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor_id"})
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Query vendor transactions
	var transactions []struct {
		ID             uint64      `json:"id"`
		PaymentID      uint64      `json:"payment_id"`
		BookingID      uint64      `json:"booking_id"`
		TransactionType string     `json:"transaction_type"`
		Amount         float64     `json:"amount"`
		Status         string      `json:"status"`
		CreatedAt      time.Time   `json:"created_at"`
	}

	// Get total count
	var total int64
	h.db.Table("payment_transactions").
		Where("metadata->>'vendor_id' = ?", vendorID).
		Count(&total)

	// Get transactions
	if err := h.db.Table("payment_transactions").
		Select("id, payment_id, booking_id, transaction_type, amount, status, created_at").
		Where("metadata->>'vendor_id' = ?", vendorID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}
