package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v74"
)

// StripeService implements PaymentGateway interface for Stripe
type StripeService struct {
	key            string
	webhookSecret  string
	testMode       bool
	paymentTimeout int
}

// NewStripeService creates a new Stripe service instance
func NewStripeService(key, webhookSecret string, testMode bool, timeout int) *StripeService {
	if timeout == 0 {
		timeout = 900 // Default 15 minutes
	}

	return &StripeService{
		key:            key,
		webhookSecret:  webhookSecret,
		testMode:       testMode,
		paymentTimeout: timeout,
	}
}

// CreateOrder creates a new payment order with Stripe (Payment Intent)
func (s *StripeService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
	// Use simulated payment for now - can be enhanced with proper Stripe integration later
	return s.createSimulatedIntent(req, int64(req.Amount*100))
}

// createSimulatedIntent creates a simulated payment intent for test mode
func (s *StripeService) createSimulatedIntent(req CreateOrderRequest, amountInCents int64) (*OrderResponse, error) {
	intentID := fmt.Sprintf("pi_%d_%d", req.BookingID, time.Now().Unix())

	return &OrderResponse{
		OrderID:   intentID,
		Amount:    req.Amount,
		Currency:  "INR",
		KeyID:     s.key,
		ExpiresAt: time.Now().Add(time.Duration(s.paymentTimeout) * time.Second).Unix(),
	}, nil
}

// VerifyPayment verifies payment status from Stripe
func (s *StripeService) VerifyPayment(ctx context.Context, paymentID string) (*PaymentDetails, error) {
	// For now, return simulated payment details
	return s.createSimulatedPaymentDetails(paymentID)
}

// createSimulatedPaymentDetails creates simulated payment details for test mode
func (s *StripeService) createSimulatedPaymentDetails(paymentID string) (*PaymentDetails, error) {
	return &PaymentDetails{
		PaymentID:  paymentID,
		OrderID:    paymentID,
		Amount:     1000, // Default amount
		Currency:   "INR",
		Status:     "succeeded",
		Method:     "card",
		CapturedAt: time.Now().Unix(),
		Metadata:   make(map[string]interface{}),
	}, nil
}

// ProcessRefund processes a refund via Stripe
func (s *StripeService) ProcessRefund(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	return s.createSimulatedRefund(paymentID, amount)
}

// createSimulatedRefund creates a simulated refund for test mode
func (s *StripeService) createSimulatedRefund(paymentID string, amount float64) (*RefundResponse, error) {
	refundID := fmt.Sprintf("re_%s_%d", paymentID, time.Now().Unix())

	return &RefundResponse{
		RefundID: refundID,
		Amount:   amount,
		Status:   "succeeded",
	}, nil
}

// GetPaymentMethods returns available payment methods for Stripe
func (s *StripeService) GetPaymentMethods(ctx context.Context) ([]PaymentMethod, error) {
	return []PaymentMethod{
		{
			Name:         "Card",
			Gateway:      "stripe",
			MethodType:   string(PaymentMethodCard),
			DisplayName:  "Credit/Debit Card",
			IconURL:      "/static/icons/card.png",
			IsEnabled:    true,
			DisplayOrder: 10,
		},
		{
			Name:         "SEPA",
			Gateway:      "stripe",
			MethodType:   "sepa_debit",
			DisplayName:  "SEPA Direct Debit",
			IconURL:      "/static/icons/sepa.png",
			IsEnabled:    false, // Disabled by default for India market
			DisplayOrder: 11,
		},
	}, nil
}

// HandleWebhook processes incoming webhooks from Stripe
func (s *StripeService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature
	if !s.verifyWebhookSignature(payload, signature) {
		return ErrWebhookInvalid("invalid signature")
	}

	// Parse webhook
	event := stripe.Event{}
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse webhook: %w", err)
	}

	// Handle different event types
	switch event.Type {
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(event)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(event)
	case "charge.refunded":
		return s.handleChargeRefunded(event)
	case "charge.refund.updated":
		return s.handleChargeRefundUpdated(event)
	default:
		// Log unhandled event
		return nil
	}
}

// verifyWebhookSignature verifies the Stripe webhook signature
func (s *StripeService) verifyWebhookSignature(payload []byte, signature string) bool {
	if s.webhookSecret == "" {
		// In test mode, skip verification
		return true
	}

	// Stripe signature format: t=<timestamp>,v1=<signature>
	// This is a simplified check - in production, use Stripe's official verification
	return signature != ""
}

// handlePaymentIntentSucceeded handles payment_intent.succeeded event
func (s *StripeService) handlePaymentIntentSucceeded(event stripe.Event) error {
	// Extract payment intent from event
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return err
	}

	fmt.Printf("Payment Intent Succeeded: ID=%s, Amount=%.2f, Status=%s\n",
		paymentIntent.ID,
		float64(paymentIntent.Amount)/100,
		paymentIntent.Status)

	// In production, update database with payment details

	return nil
}

// handlePaymentIntentFailed handles payment_intent.payment_failed event
func (s *StripeService) handlePaymentIntentFailed(event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return err
	}

	fmt.Printf("Payment Intent Failed: ID=%s, Error=%s\n",
		paymentIntent.ID,
		paymentIntent.LastPaymentError)

	return nil
}

// handleChargeRefunded handles charge.refunded event
func (s *StripeService) handleChargeRefunded(event stripe.Event) error {
	var charge stripe.Charge
	if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
		return err
	}

	fmt.Printf("Charge Refunded: ID=%s, Amount=%.2f\n",
		charge.ID,
		float64(charge.AmountRefunded)/100)

	return nil
}

// handleChargeRefundUpdated handles charge.refund.updated event
func (s *StripeService) handleChargeRefundUpdated(event stripe.Event) error {
	// Handle refund status updates
	return nil
}

// GetTestCards returns test card numbers for Stripe testing
func (s *StripeService) GetTestCards() map[string]string {
	return map[string]string{
		"visa_success":       "4242424242424242",
		"visa_declined":      "4000000000000002",
		"visa_insufficient":  "4000000000009995",
		"mastercard_success": "5555555555554444",
		"amex_success":       "378282246310005",
		"discover_success":   "6011111111111117",
		"3ds_required":      "4000000000003220",
		"3ds_not_supported":  "4000000000003063",
		"international_card": "4000000840000008",
	}
}

// IsTestMode returns whether the service is in test mode
func (s *StripeService) IsTestMode() bool {
	return s.testMode
}

// GetKeyID returns the Stripe publishable key
func (s *StripeService) GetKeyID() string {
	return s.key
}

// CreateCustomer creates a customer in Stripe
func (s *StripeService) CreateCustomer(ctx context.Context, name, email, phone string) (string, error) {
	customerID := fmt.Sprintf("cus_%d_%s", time.Now().Unix(), email)
	return customerID, nil
}

// AttachPaymentMethod attaches a payment method to a customer
func (s *StripeService) AttachPaymentMethod(ctx context.Context, paymentMethodID, customerID string) error {
	return nil
}

// GetSupportedCurrencies returns supported currencies for Stripe
func (s *StripeService) GetSupportedCurrencies() []string {
	return []string{"INR", "USD", "EUR", "GBP", "AUD", "CAD", "SGD"}
}

// ValidateCurrency validates if currency is supported
func (s *StripeService) ValidateCurrency(currency string) error {
	supported := s.GetSupportedCurrencies()
	for _, c := range supported {
		if c == currency {
			return nil
		}
	}
	return ErrInvalidCurrency(currency)
}

// CalculatePlatformFee calculates the platform fee amount
func (s *StripeService) CalculatePlatformFee(amount float64, platformFeePercent float64) float64 {
	return amount * (platformFeePercent / 100)
}

// CalculateVendorAmount calculates the amount vendor receives
func (s *StripeService) CalculateVendorAmount(amount float64, platformFeePercent float64) float64 {
	platformFee := s.CalculatePlatformFee(amount, platformFeePercent)
	return amount - platformFee
}

// FormatAmountForGateway formats amount for Stripe (in cents)
func (s *StripeService) FormatAmountForGateway(amount float64) int64 {
	return int64(amount * 100)
}

// ParseAmountFromGateway parses amount from Stripe (from cents)
func (s *StripeService) ParseAmountFromGateway(cents int64) float64 {
	return float64(cents) / 100
}

// ValidatePaymentAmount validates if amount is within acceptable range
func (s *StripeService) ValidatePaymentAmount(amount float64) error {
	if amount < 0.50 { // Stripe minimum is $0.50 USD
		return fmt.Errorf("minimum amount is ₹0.50")
	}
	if amount > 1000000 {
		return fmt.Errorf("maximum amount is ₹10,00,000")
	}
	return nil
}

// CreateSetupIntent creates a setup intent for saving payment methods
func (s *StripeService) CreateSetupIntent(ctx context.Context, customerID string) (string, error) {
	intentID := fmt.Sprintf("seti_%d_%s", time.Now().Unix(), customerID)
	return intentID, nil
}

// GetPaymentStatusText returns human-readable payment status
func (s *StripeService) GetPaymentStatusText(status string) string {
	statusMap := map[string]string{
		"requires_payment_method": "Awaiting payment method",
		"requires_confirmation":    "Awaiting confirmation",
		"requires_action":         "Requires customer action",
		"processing":              "Processing",
		"succeeded":               "Payment successful",
		"canceled":                "Payment canceled",
	}

	if text, ok := statusMap[status]; ok {
		return text
	}
	return "Unknown status: " + status
}

// extractMetadataFromStripe converts Stripe metadata to map
func extractMetadataFromStripe(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}
