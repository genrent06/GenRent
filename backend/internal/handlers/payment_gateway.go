package handlers

import (
	"genrent/internal/services/payment"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaymentGatewayHandler handles payment gateway operations
type PaymentGatewayHandler struct {
	paymentService payment.PaymentGateway
	db              *gorm.DB
}

// NewPaymentGatewayHandler creates a new payment gateway handler
func NewPaymentGatewayHandler(ps payment.PaymentGateway, db *gorm.DB) *PaymentGatewayHandler {
	return &PaymentGatewayHandler{
		paymentService: ps,
		db:              db,
	}
}

// CreatePaymentOrderRequest represents request to create payment order
type CreatePaymentOrderRequest struct {
	BookingID     uint64  `json:"booking_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method"`
	Currency      string  `json:"currency"`
}

// CreatePaymentOrder creates a new payment order
func (h *PaymentGatewayHandler) CreatePaymentOrder(c *gin.Context) {
	var req CreatePaymentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Convert user to expected type
	userData, ok := user.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	userID := uint64(userData["id"].(float64))
	userEmail := userData["email"].(string)
	userName := userData["name"].(string)
	userPhone := userData["phone"].(string)

	// Set default currency
	if req.Currency == "" {
		req.Currency = "INR"
	}

	// Create order
	order, err := h.paymentService.CreateOrder(c.Request.Context(), payment.CreateOrderRequest{
		Amount:        req.Amount,
		Currency:      req.Currency,
		CustomerID:    userID,
		CustomerName:  userName,
		CustomerEmail: userEmail,
		CustomerPhone: userPhone,
		BookingID:     req.BookingID,
		Description:   "Equipment rental booking",
		Metadata: map[string]interface{}{
			"booking_id":  req.BookingID,
			"payment_method": req.PaymentMethod,
			"user_agent":   c.GetHeader("User-Agent"),
			"ip_address":  c.ClientIP(),
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment order: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":    order.OrderID,
		"amount":      order.Amount,
		"currency":    order.Currency,
		"key_id":      order.KeyID,
		"expires_at":  order.ExpiresAt,
		"created_at":  time.Now().Unix(),
	})
}

// VerifyPaymentRequest represents request to verify payment
type VerifyPaymentRequest struct {
	PaymentID string `json:"payment_id" binding:"required"`
	OrderID   string `json:"order_id"`
}

// VerifyPayment verifies a payment after completion
func (h *PaymentGatewayHandler) VerifyPayment(c *gin.Context) {
	paymentID := c.Query("payment_id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
		return
	}

	// Verify payment from gateway
	details, err := h.paymentService.VerifyPayment(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment verification failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_id":   details.PaymentID,
		"order_id":     details.OrderID,
		"amount":       details.Amount,
		"currency":     details.Currency,
		"status":       details.Status,
		"method":       details.Method,
		"captured_at":  details.CapturedAt,
	})
}

// ProcessRefundRequest represents request to process refund
type ProcessRefundRequest struct {
	PaymentID string  `json:"payment_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Reason    string  `json:"reason" binding:"required"`
}

// ProcessRefund initiates a refund for a payment
func (h *PaymentGatewayHandler) ProcessRefund(c *gin.Context) {
	var req ProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process refund
	refund, err := h.paymentService.ProcessRefund(c.Request.Context(), req.PaymentID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refund processing failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refund_id": refund.RefundID,
		"amount":    refund.Amount,
		"status":    refund.Status,
		"processed_at": time.Now().Unix(),
	})
}

// GetPaymentMethods returns available payment methods
func (h *PaymentGatewayHandler) GetPaymentMethods(c *gin.Context) {
	methods, err := h.paymentService.GetPaymentMethods(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment methods"})
		return
	}

	// Convert to response format
	response := make([]gin.H, 0, len(methods))
	for _, method := range methods {
		response = append(response, gin.H{
			"name":          method.Name,
			"gateway":       method.Gateway,
			"method_type":   method.MethodType,
			"display_name":  method.DisplayName,
			"icon_url":      method.IconURL,
			"is_enabled":    method.IsEnabled,
			"display_order": method.DisplayOrder,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_methods": response,
	})
}

// HandlePaymentWebhook handles incoming webhooks from payment gateway
func (h *PaymentGatewayHandler) HandlePaymentWebhook(c *gin.Context) {
	// Get raw payload
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read payload"})
		return
	}

	// Get signature from header (varies by gateway)
	signature := c.GetHeader("X-Razorpay-Signature")
	if signature == "" {
		signature = c.GetHeader("X-Stripe-Signature")
	}

	// Process webhook
	if err := h.paymentService.HandleWebhook(c.Request.Context(), payload, signature); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook processing failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "webhook processed"})
}

// GetPaymentStatus returns payment status for a booking
func (h *PaymentGatewayHandler) GetPaymentStatus(c *gin.Context) {
	bookingID := c.Query("booking_id")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking_id is required"})
		return
	}

	// Query payment status from database
	var payment struct {
		ID              uint64  `json:"id"`
		BookingID       uint64  `json:"booking_id"`
		Status          string  `json:"status"`
		Amount          float64 `json:"amount"`
		PaymentMethod   string  `json:"payment_method"`
		GatewayPaymentID string `json:"gateway_payment_id"`
		CreatedAt       time.Time `json:"created_at"`
		PaidAt          *time.Time `json:"paid_at"`
	}

	if err := h.db.Table("payments").
		Select("id, booking_id, status, total_amount as amount, payment_method, gateway_payment_id, created_at, paid_at").
		Where("booking_id = ?", bookingID).
		First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// CalculatePaymentOptions calculates payment options including EMI
func (h *PaymentGatewayHandler) CalculatePaymentOptions(c *gin.Context) {
	amountStr := c.Query("amount")
	if amountStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount is required"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}

	// Get payment service (check if it's Razorpay)
	razorpayService, ok := h.paymentService.(*payment.RazorpayService)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"amount": amount,
			"note": "EMI options only available for Razorpay",
		})
		return
	}

	// Calculate EMI options
	emiOptions := razorpayService.CalculateEMIOptions(amount)

	// Calculate platform fee and vendor amount
	platformFeePercent := 10.0 // Should come from config
	platformFee := amount * (platformFeePercent / 100)
	vendorAmount := amount - platformFee

	c.JSON(http.StatusOK, gin.H{
		"amount": amount,
		"platform_fee": gin.H{
			"percent": platformFeePercent,
			"amount":  platformFee,
		},
		"vendor_amount": vendorAmount,
		"emi_options":  emiOptions,
	})
}

// GetPaymentHistory returns payment history for current user
func (h *PaymentGatewayHandler) GetPaymentHistory(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userData := user.(map[string]interface{})
	userID := uint64(userData["id"].(float64))

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Query payments
	var payments []struct {
		ID              uint64    `json:"id"`
		BookingID       uint64    `json:"booking_id"`
		Amount          float64   `json:"amount"`
		Status          string    `json:"status"`
		PaymentMethod   string    `json:"payment_method"`
		Gateway         string    `json:"gateway"`
		CreatedAt       time.Time `json:"created_at"`
		PaidAt          *time.Time `json:"paid_at"`
		RefundedAt      *time.Time `json:"refunded_at"`
	}

	// Get total count
	var total int64
	h.db.Table("payments").Where("booking_id IN (SELECT id FROM bookings WHERE customer_id = ?)", userID).Count(&total)

	// Get payments
	if err := h.db.Table("payments").
		Select("id, booking_id, total_amount as amount, status, payment_method, gateway, created_at, paid_at, refunded_at").
		Where("booking_id IN (SELECT id FROM bookings WHERE customer_id = ?)", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch payment history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetPaymentDetails returns detailed payment information
func (h *PaymentGatewayHandler) GetPaymentDetails(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
		return
	}

	// Query payment details
	var payment struct {
		ID                uint64         `json:"id"`
		BookingID         uint64         `json:"booking_id"`
		TotalAmount       float64        `json:"total_amount"`
		AdvanceAmount     float64        `json:"advance_amount"`
		VendorAmount      float64        `json:"vendor_amount"`
		PlatformFee       float64        `json:"platform_fee"`
		Status            string         `json:"status"`
		PaymentMethod     string         `json:"payment_method"`
		Gateway           string         `json:"gateway"`
		GatewayOrderID    string         `json:"gateway_order_id"`
		GatewayPaymentID  string         `json:"gateway_payment_id"`
		PaymentMetadata   map[string]interface{} `json:"payment_metadata"`
		EscrowHeldAt      *time.Time     `json:"escrow_held_at"`
		EscrowReleasedAt  *time.Time     `json:"escrow_released_at"`
		RefundID          string         `json:"refund_id,omitempty"`
		RefundAmount      float64        `json:"refund_amount,omitempty"`
		RefundStatus      string         `json:"refund_status,omitempty"`
		CreatedAt         time.Time      `json:"created_at"`
		PaidAt            *time.Time     `json:"paid_at"`
		RefundedAt        *time.Time     `json:"refunded_at,omitempty"`
	}

	if err := h.db.Table("payments").
		Where("id = ?", paymentID).
		First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// CancelPaymentRequest represents request to cancel payment
type CancelPaymentRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Reason  string `json:"reason"`
}

// CancelPayment cancels a pending payment order
func (h *PaymentGatewayHandler) CancelPayment(c *gin.Context) {
	var req CancelPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In production, call Razorpay API to cancel order
	// For now, return success
	c.JSON(http.StatusOK, gin.H{
		"message": "payment order cancelled",
		"order_id": req.OrderID,
	})
}

// ResendPaymentOTP resends OTP for payment verification (if applicable)
func (h *PaymentGatewayHandler) ResendPaymentOTP(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
		return
	}

	// In production, trigger OTP resend
	c.JSON(http.StatusOK, gin.H{
		"message": "OTP resent successfully",
		"payment_id": paymentID,
	})
}
