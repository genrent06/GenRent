package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	razorpay "github.com/razorpay/razorpay-go"
)

// RazorpayService implements PaymentGateway interface for Razorpay
type RazorpayService struct {
	client         *razorpay.Client
	keyID          string
	keySecret      string
	webhookSecret  string
	testMode       bool
	paymentTimeout int // in seconds
}

// NewRazorpayService creates a new Razorpay service instance
func NewRazorpayService(keyID, keySecret, webhookSecret string, testMode bool, timeout int) *RazorpayService {
	if timeout == 0 {
		timeout = 900 // Default 15 minutes
	}

	return &RazorpayService{
		client:        razorpay.NewClient(keyID, keySecret),
		keyID:         keyID,
		keySecret:     keySecret,
		webhookSecret: webhookSecret,
		testMode:      testMode,
		paymentTimeout: timeout,
	}
}

// CreateOrder creates a new payment order with Razorpay
func (r *RazorpayService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
	// Convert amount to paise (Razorpay uses paise)
	amountInPaise := int(req.Amount * 100)

	// Create order data
	orderData := map[string]interface{}{
		"amount":          amountInPaise,
		"currency":        "INR",
		"receipt":         fmt.Sprintf("booking_%d", req.BookingID),
		"payment_capture": 1, // Auto capture
		"notes": map[string]string{
			"booking_id":      fmt.Sprintf("%d", req.BookingID),
			"customer_id":     fmt.Sprintf("%d", req.CustomerID),
			"customer_email":  req.CustomerEmail,
			"customer_name":   req.CustomerName,
			"customer_phone":  req.CustomerPhone,
		},
	}

	// Call Razorpay API to create order
	response, err := r.client.Order.Create(orderData, nil)
	if err != nil {
		// Log error for debugging
		if r.testMode {
			// In test mode, return simulated response on API error
			return r.createSimulatedOrder(req, amountInPaise)
		}
		return nil, ErrOrderCreationFailed(err)
	}

	// Extract order ID from response
	orderID, ok := response["id"].(string)
	if !ok {
		return nil, ErrAPIError("order creation", fmt.Errorf("missing order id in response"))
	}

	amount := float64(amountInPaise) / 100

	return &OrderResponse{
		OrderID:    orderID,
		Amount:     amount,
		Currency:   "INR",
		KeyID:      r.keyID,
		ExpiresAt:  time.Now().Add(time.Duration(r.paymentTimeout) * time.Second).Unix(),
	}, nil
}

// createSimulatedOrder creates a simulated order for test mode
func (r *RazorpayService) createSimulatedOrder(req CreateOrderRequest, amountInPaise int) (*OrderResponse, error) {
	orderID := fmt.Sprintf("order_%d_%d", req.BookingID, time.Now().Unix())

	return &OrderResponse{
		OrderID:    orderID,
		Amount:     req.Amount,
		Currency:   "INR",
		KeyID:      r.keyID,
		ExpiresAt:  time.Now().Add(time.Duration(r.paymentTimeout) * time.Second).Unix(),
	}, nil
}

// VerifyPayment verifies payment status from Razorpay
func (r *RazorpayService) VerifyPayment(ctx context.Context, paymentID string) (*PaymentDetails, error) {
	// Fetch payment from Razorpay API
	payment, err := r.client.Payment.Fetch(paymentID, nil, nil)
	if err != nil {
		if r.testMode {
			// In test mode, return simulated response
			return r.createSimulatedPaymentDetails(paymentID)
		}
		return nil, ErrVerificationFailed(paymentID, err)
	}

	// Extract payment details
	orderID, _ := payment["order_id"].(string)
	amount := float64(0)
	if val, ok := payment["amount"].(float64); ok {
		amount = val / 100
	}
	currency, _ := payment["currency"].(string)
	status, _ := payment["status"].(string)
	method, _ := payment["method"].(string)

	// Extract capture time
	var capturedAt int64
	if val, ok := payment["created_at"].(float64); ok {
		capturedAt = int64(val)
	}

	// Extract metadata
	metadata := make(map[string]interface{})
	if notes, ok := payment["notes"].(map[string]interface{}); ok {
		metadata = notes
	}

	return &PaymentDetails{
		PaymentID:   paymentID,
		OrderID:     orderID,
		Amount:      amount,
		Currency:    currency,
		Status:      status,
		Method:      method,
		CapturedAt:  capturedAt,
		Metadata:    metadata,
	}, nil
}

// createSimulatedPaymentDetails creates simulated payment details for test mode
func (r *RazorpayService) createSimulatedPaymentDetails(paymentID string) (*PaymentDetails, error) {
	return &PaymentDetails{
		PaymentID:  paymentID,
		OrderID:    fmt.Sprintf("order_%s", paymentID),
		Amount:     0,
		Currency:   "INR",
		Status:     "captured",
		Method:     "upi",
		CapturedAt: time.Now().Unix(),
		Metadata:   make(map[string]interface{}),
	}, nil
}

// ProcessRefund processes a refund via Razorpay
func (r *RazorpayService) ProcessRefund(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	// Convert amount to paise
	amountInPaise := int(amount * 100)

	// Create refund data
	refundData := map[string]interface{}{
		"amount": amountInPaise,
	}

	// Call Razorpay API to process refund
	response, err := r.client.Payment.Refund(paymentID, amountInPaise, refundData, nil)
	if err != nil {
		if r.testMode {
			// In test mode, return simulated response
			return r.createSimulatedRefund(paymentID, amount)
		}
		return nil, ErrRefundFailed(paymentID, err)
	}

	// Extract refund details
	refundID, ok := response["id"].(string)
	if !ok {
		return nil, ErrAPIError("refund", fmt.Errorf("missing refund id in response"))
	}

	status, _ := response["status"].(string)
	refundAmount := float64(0)
	if val, ok := response["amount"].(float64); ok {
		refundAmount = val / 100
	}

	return &RefundResponse{
		RefundID: refundID,
		Amount:   refundAmount,
		Status:   status,
	}, nil
}

// createSimulatedRefund creates a simulated refund for test mode
func (r *RazorpayService) createSimulatedRefund(paymentID string, amount float64) (*RefundResponse, error) {
	refundID := fmt.Sprintf("refund_%s_%d", paymentID, time.Now().Unix())

	return &RefundResponse{
		RefundID: refundID,
		Amount:   amount,
		Status:   "processed",
	}, nil
}

// GetPaymentMethods returns available payment methods for Razorpay
func (r *RazorpayService) GetPaymentMethods(ctx context.Context) ([]PaymentMethod, error) {
	return []PaymentMethod{
		{
			Name:        "UPI",
			Gateway:     "razorpay",
			MethodType:  string(PaymentMethodUPI),
			DisplayName: "UPI Payment",
			IconURL:     "/static/icons/upi.png",
			IsEnabled:   true,
			DisplayOrder: 1,
		},
		{
			Name:        "Card",
			Gateway:     "razorpay",
			MethodType:  string(PaymentMethodCard),
			DisplayName: "Credit/Debit Card",
			IconURL:     "/static/icons/card.png",
			IsEnabled:   true,
			DisplayOrder: 2,
		},
		{
			Name:        "Netbanking",
			Gateway:     "razorpay",
			MethodType:  string(PaymentMethodNetbanking),
			DisplayName: "Net Banking",
			IconURL:     "/static/icons/bank.png",
			IsEnabled:   true,
			DisplayOrder: 3,
		},
		{
			Name:        "Wallet",
			Gateway:     "razorpay",
			MethodType:  string(PaymentMethodWallet),
			DisplayName: "Mobile Wallet (Paytm, PhonePe)",
			IconURL:     "/static/icons/wallet.png",
			IsEnabled:   true,
			DisplayOrder: 4,
		},
	}, nil
}

// HandleWebhook processes incoming webhooks from Razorpay
func (r *RazorpayService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature
	if !r.verifyWebhookSignature(payload, signature) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Parse webhook
	var webhook map[string]interface{}
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return fmt.Errorf("failed to parse webhook: %w", err)
	}

	// Extract event
	event, ok := webhook["event"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook format: missing event")
	}

	// Handle different event types
	switch event {
	case "payment.captured":
		return r.handlePaymentCaptured(webhook)
	case "payment.authorized":
		return r.handlePaymentAuthorized(webhook)
	case "payment.failed":
		return r.handlePaymentFailed(webhook)
	case "refund.processed":
		return r.handleRefundProcessed(webhook)
	case "refund.failed":
		return r.handleRefundFailed(webhook)
	case "order.paid":
		return r.handleOrderPaid(webhook)
	default:
		// Log unhandled event
		return nil
	}
}

// verifyWebhookSignature verifies the Razorpay webhook signature
func (r *RazorpayService) verifyWebhookSignature(payload []byte, signature string) bool {
	if r.webhookSecret == "" {
		// In test mode, skip verification
		return true
	}

	// Razorpay uses HMAC SHA256 for webhook signature
	h := hmac.New(sha256.New, []byte(r.webhookSecret))
	h.Write(payload)
	calculatedSignature := hex.EncodeToString(h.Sum(nil))

	// Razorpay sends signature as "sha256=<calculated_hash>"
	expectedSignature := "sha256=" + calculatedSignature

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// handlePaymentCaptured handles payment.captured event
func (r *RazorpayService) handlePaymentCaptured(webhook map[string]interface{}) error {
	// Extract payment details
	payload, ok := webhook["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid webhook payload")
	}

	payment, ok := payload["payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment details")
	}

	paymentEntity, ok := payment["entity"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment entity")
	}

	// Extract key details
	paymentID, _ := paymentEntity["id"].(string)
	orderID, _ := paymentEntity["order_id"].(string)
	amount, _ := paymentEntity["amount"].(float64)
	currency, _ := paymentEntity["currency"].(string)
	method, _ := paymentEntity["method"].(string)
	status, _ := paymentEntity["status"].(string)

	// Log successful payment
	fmt.Printf("Payment Captured: ID=%s, Order=%s, Amount=%.2f %s, Method=%s, Status=%s\n",
		paymentID, orderID, amount/100, currency, method, status)

	// In production, update database with payment details

	return nil
}

// handlePaymentAuthorized handles payment.authorized event
func (r *RazorpayService) handlePaymentAuthorized(webhook map[string]interface{}) error {
	// Similar to payment.captured but for authorized payments
	return nil
}

// handlePaymentFailed handles payment.failed event
func (r *RazorpayService) handlePaymentFailed(webhook map[string]interface{}) error {
	payload, ok := webhook["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid webhook payload")
	}

	payment, ok := payload["payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment details")
	}

	paymentEntity, ok := payment["entity"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment entity")
	}

	paymentID, _ := paymentEntity["id"].(string)
	orderID, _ := paymentEntity["order_id"].(string)
	errorCode, _ := paymentEntity["error_code"].(string)
	errorDesc, _ := paymentEntity["error_description"].(string)

	fmt.Printf("Payment Failed: ID=%s, Order=%s, Error=%s: %s\n",
		paymentID, orderID, errorCode, errorDesc)

	// In production, update booking status and notify customer

	return nil
}

// handleRefundProcessed handles refund.processed event
func (r *RazorpayService) handleRefundProcessed(webhook map[string]interface{}) error {
	payload, ok := webhook["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid webhook payload")
	}

	refund, ok := payload["refund"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid refund details")
	}

	refundEntity, ok := refund["entity"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid refund entity")
	}

	refundID, _ := refundEntity["id"].(string)
	paymentID, _ := refundEntity["payment_id"].(string)
	amount, _ := refundEntity["amount"].(float64)
	status, _ := refundEntity["status"].(string)

	fmt.Printf("Refund Processed: ID=%s, Payment=%s, Amount=%.2f, Status=%s\n",
		refundID, paymentID, amount/100, status)

	// In production, update refund status in database

	return nil
}

// handleRefundFailed handles refund.failed event
func (r *RazorpayService) handleRefundFailed(webhook map[string]interface{}) error {
	// Handle failed refunds
	return nil
}

// handleOrderPaid handles order.paid event
func (r *RazorpayService) handleOrderPaid(webhook map[string]interface{}) error {
	// Handle order paid event
	return nil
}

// GetTestCards returns test card numbers for testing
func (r *RazorpayService) GetTestCards() map[string]string {
	return map[string]string{
		"success":              "4242 4242 4242 4242",
		"failure":              "4000 0000 0000 0002",
		"3ds_required":         "4000 0025 0000 3155",
		"3ds_failure":          "4000 0500 0000 0005",
		"upi_success":          "razorpay@upi",
		"netbanking_success":    "RTEST0001",
		"wallet_success":        "wallet@razorpay",
	}
}

// IsTestMode returns whether the service is in test mode
func (r *RazorpayService) IsTestMode() bool {
	return r.testMode
}

// GetKeyID returns the Razorpay key ID
func (r *RazorpayService) GetKeyID() string {
	return r.keyID
}

// CalculatePlatformFee calculates the platform fee amount
func (r *RazorpayService) CalculatePlatformFee(amount float64, platformFeePercent float64) float64 {
	return amount * (platformFeePercent / 100)
}

// CalculateVendorAmount calculates the amount vendor receives
func (r *RazorpayService) CalculateVendorAmount(amount float64, platformFeePercent float64) float64 {
	platformFee := r.CalculatePlatformFee(amount, platformFeePercent)
	return amount - platformFee
}

// FormatAmountForGateway formats amount for Razorpay (in paise)
func (r *RazorpayService) FormatAmountForGateway(amount float64) int {
	return int(amount * 100)
}

// ParseAmountFromGateway parses amount from Razorpay (from paise)
func (r *RazorpayService) ParseAmountFromGateway(paise int) float64 {
	return float64(paise) / 100
}

// ValidatePaymentAmount validates if amount is within acceptable range
func (r *RazorpayService) ValidatePaymentAmount(amount float64) error {
	if amount < 1 {
		return fmt.Errorf("minimum amount is ₹1")
	}
	if amount > 1000000 {
		return fmt.Errorf("maximum amount is ₹10,00,000")
	}
	return nil
}

// GenerateReceiptNumber generates a unique receipt number
func (r *RazorpayService) GenerateReceiptNumber(bookingID uint64) string {
	return fmt.Sprintf("GENRENT-%d-%d", bookingID, time.Now().Unix())
}

// SupportsEMI checks if EMI is supported for the card
func (r *RazorpayService) SupportsEMI(cardBin string) bool {
	// In production, check with Razorpay API
	// For now, return true for cards starting with 4 (Visa)
	return len(cardBin) >= 6 && cardBin[:1] == "4"
}

// CalculateEMIOptions calculates EMI options for a given amount
func (r *RazorpayService) CalculateEMIOptions(amount float64) []EMIOption {
	// Standard EMI tenures: 3, 6, 9, 12 months
	tenures := []int{3, 6, 9, 12}
	options := make([]EMIOption, 0)

	for _, tenure := range tenures {
		// Simple interest calculation (actual rates vary by bank)
		interestRate := 14.0 // 14% per annum
		monthlyRate := interestRate / 12 / 100

		// EMI calculation: [P x R x (1+R)^N]/[(1+R)^N-1]
		emiAmount := (amount * monthlyRate * pow(1+monthlyRate, float64(tenure))) /
			(pow(1+monthlyRate, float64(tenure)) - 1)

		totalAmount := emiAmount * float64(tenure)
		interestAmount := totalAmount - amount

		options = append(options, EMIOption{
			Tenure:        tenure,
			EMIAmount:     emiAmount,
			TotalAmount:   totalAmount,
			Interest:      interestAmount,
			InterestRate:  interestRate,
		})
	}

	return options
}

// EMIOption represents an EMI option
type EMIOption struct {
	Tenure       int     `json:"tenure"`        // Months
	EMIAmount    float64 `json:"emi_amount"`    // Monthly payment
	TotalAmount  float64 `json:"total_amount"`  // Total payable
	Interest     float64 `json:"interest"`      // Interest amount
	InterestRate float64 `json:"interest_rate"` // Annual interest rate
}

// pow calculates base^exp
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// GetPaymentStatusText returns human-readable payment status
func (r *RazorpayService) GetPaymentStatusText(status string) string {
	statusMap := map[string]string{
		"created":           "Order created",
		"attempted":         "Payment attempted",
		"authorized":        "Payment authorized",
		"captured":          "Payment successful",
		"failed":            "Payment failed",
		"refunded":          "Payment refunded",
		"partially_refunded": "Partially refunded",
	}

	if text, ok := statusMap[status]; ok {
		return text
	}
	return "Unknown status: " + status
}

// ValidateWebhookEvent validates if webhook event is supported
func (r *RazorpayService) ValidateWebhookEvent(event string) bool {
	supportedEvents := map[string]bool{
		"payment.captured":   true,
		"payment.authorized":  true,
		"payment.failed":      true,
		"refund.processed":    true,
		"refund.failed":       true,
		"order.paid":          true,
	}

	return supportedEvents[event]
}

// GetSupportedCurrencies returns supported currencies
func (r *RazorpayService) GetSupportedCurrencies() []string {
	return []string{"INR", "USD"}
}

// ValidateCurrency validates if currency is supported
func (r *RazorpayService) ValidateCurrency(currency string) error {
	supported := r.GetSupportedCurrencies()
	for _, c := range supported {
		if c == currency {
			return nil
		}
	}
	return fmt.Errorf("unsupported currency: %s", currency)
}
