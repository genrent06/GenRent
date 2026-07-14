package payment

import "context"

// PaymentGateway defines the interface for payment gateway implementations
type PaymentGateway interface {
	// CreateOrder initiates a payment order with the gateway
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error)

	// VerifyPayment verifies payment status from the gateway
	VerifyPayment(ctx context.Context, paymentID string) (*PaymentDetails, error)

	// ProcessRefund processes a refund for a payment
	ProcessRefund(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error)

	// GetPaymentMethods returns available payment methods
	GetPaymentMethods(ctx context.Context) ([]PaymentMethod, error)

	// HandleWebhook processes incoming webhook from payment gateway
	HandleWebhook(ctx context.Context, payload []byte, signature string) error
}

// CreateOrderRequest contains details for creating a payment order
type CreateOrderRequest struct {
	Amount        float64
	Currency      string
	CustomerID    uint64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	BookingID     uint64
	Description   string
	Metadata      map[string]interface{}
}

// OrderResponse contains the created order details
type OrderResponse struct {
	OrderID    string
	Amount     float64
	Currency   string
	PaymentURL string
	ExpiresAt  int64
	KeyID      string // Public key for frontend
}

// PaymentDetails contains verified payment information
type PaymentDetails struct {
	PaymentID   string
	OrderID     string
	Amount      float64
	Currency    string
	Status      string
	Method      string
	Description string
	Metadata    map[string]interface{}
	CapturedAt  int64
}

// RefundResponse contains refund processing details
type RefundResponse struct {
	RefundID string
	Amount   float64
	Status   string
}

// PaymentMethod represents an available payment method
type PaymentMethod struct {
	Name         string
	Gateway      string
	MethodType   string
	DisplayName  string
	IconURL      string
	IsEnabled    bool
	DisplayOrder int
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending     PaymentStatus = "pending"
	PaymentStatusInitiated   PaymentStatus = "initiated"
	PaymentStatusPaid        PaymentStatus = "paid"
	PaymentStatusFailed      PaymentStatus = "failed"
	PaymentStatusRefunded    PaymentStatus = "refunded"
	PaymentStatusEscrow      PaymentStatus = "escrow"
	PaymentStatusCompleted   PaymentStatus = "completed"
	PaymentStatusCancelled   PaymentStatus = "cancelled"
)

// GatewayType represents supported payment gateways
type GatewayType string

const (
	GatewayRazorpay GatewayType = "razorpay"
	GatewayStripe   GatewayType = "stripe"
)

// PaymentMethodType represents payment method types
type PaymentMethodType string

const (
	PaymentMethodUPI        PaymentMethodType = "upi"
	PaymentMethodCard       PaymentMethodType = "card"
	PaymentMethodNetbanking PaymentMethodType = "netbanking"
	PaymentMethodWallet     PaymentMethodType = "wallet"
	PaymentMethodCardStripe PaymentMethodType = "card" // Stripe uses same type
)

// EscrowService handles escrow operations
type EscrowService interface {
	// HoldFunds holds payment in escrow until delivery confirmation
	HoldFunds(ctx context.Context, paymentID uint64, bookingID uint64) error

	// ReleaseFunds releases funds to vendor after successful delivery
	ReleaseFunds(ctx context.Context, bookingID uint64) error

	// ProcessRefund processes refund for cancelled bookings
	ProcessRefund(ctx context.Context, bookingID uint64, reason string, amount float64) error

	// GetEscrowStatus returns current escrow status
	GetEscrowStatus(ctx context.Context, bookingID uint64) (*EscrowStatus, error)
}

// EscrowStatus represents the current escrow status
type EscrowStatus struct {
	BookingID       uint64
	PaymentID       uint64
	Amount          float64
	Status          PaymentStatus
	HeldAt          *int64
	ReleasedAt      *int64
	VendorAmount    float64
	PlatformFee     float64
	RefundAmount    float64
	RefundProcessed bool
}

// RefundService handles refund operations
type RefundService interface {
	// InitiateRefund initiates a refund request
	InitiateRefund(ctx context.Context, req RefundRequest) (*RefundResponse, error)

	// GetRefundStatus returns refund status
	GetRefundStatus(ctx context.Context, refundID string) (*RefundStatus, error)

	// ProcessPendingRefunds processes all pending refunds
	ProcessPendingRefunds(ctx context.Context) error
}

// RefundRequest contains details for initiating a refund
type RefundRequest struct {
	PaymentID string
	Amount    float64
	Reason    string
	RefundID  string // Optional, for existing refunds
}

// RefundStatus represents the status of a refund
type RefundStatus struct {
	RefundID     string
	PaymentID    string
	Amount       float64
	Status       string
	Reason       string
	ProcessedAt  *int64
	EstimatedAt  *int64 // Estimated settlement date
}

// WebhookEvent represents a payment webhook event
type WebhookEvent struct {
	Event      string                 // Event type (e.g., "payment.captured")
	Timestamp int64                  // Event timestamp
	Data       map[string]interface{} // Event payload
}
