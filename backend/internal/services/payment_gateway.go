package services

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

// PaymentResult is returned by every gateway after initiating or verifying payment
type PaymentResult struct {
	TransactionID string
	Status        string // "success" | "failed" | "pending"
	GatewayRef    string // gateway-specific reference ID
	Message       string
}

// RefundResult is returned after a refund request
type RefundResult struct {
	RefundID string
	Status   string // "success" | "pending" | "failed"
	Message  string
}

// PaymentGateway is the abstraction all providers must satisfy
type PaymentGateway interface {
	Name() string
	InitiatePayment(bookingID uint, amount float64, method string) (*PaymentResult, error)
	VerifyPayment(transactionID string) (*PaymentResult, error)
	RefundPayment(transactionID string, amount float64, reason string) (*RefundResult, error)
}

// ---- Mock Gateway (development / testing) ----

type MockGateway struct{}

func NewMockGateway() PaymentGateway { return &MockGateway{} }
func (g *MockGateway) Name() string  { return "mock" }

func (g *MockGateway) InitiatePayment(bookingID uint, amount float64, method string) (*PaymentResult, error) {
	txID := fmt.Sprintf("MOCK%d%04d", time.Now().Unix(), rand.Intn(9999))
	return &PaymentResult{TransactionID: txID, Status: "success", GatewayRef: "mock-ref-" + txID, Message: "Mock payment successful"}, nil
}

func (g *MockGateway) VerifyPayment(transactionID string) (*PaymentResult, error) {
	return &PaymentResult{TransactionID: transactionID, Status: "success", Message: "Mock verification successful"}, nil
}

func (g *MockGateway) RefundPayment(transactionID string, amount float64, reason string) (*RefundResult, error) {
	return &RefundResult{RefundID: "REFUND-" + transactionID, Status: "success", Message: fmt.Sprintf("Mock refund of ₹%.2f issued", amount)}, nil
}

// ---- Razorpay Gateway (production stub — wire in Razorpay SDK when ready) ----

type RazorpayGateway struct{ KeyID, KeySecret string }

func NewRazorpayGateway(keyID, keySecret string) PaymentGateway {
	return &RazorpayGateway{KeyID: keyID, KeySecret: keySecret}
}
func (g *RazorpayGateway) Name() string { return "razorpay" }

func (g *RazorpayGateway) InitiatePayment(bookingID uint, amount float64, method string) (*PaymentResult, error) {
	return nil, fmt.Errorf("razorpay not yet configured — set PAYMENT_GATEWAY=mock for dev")
}
func (g *RazorpayGateway) VerifyPayment(transactionID string) (*PaymentResult, error) {
	return nil, fmt.Errorf("razorpay verification not yet configured")
}
func (g *RazorpayGateway) RefundPayment(transactionID string, amount float64, reason string) (*RefundResult, error) {
	return nil, fmt.Errorf("razorpay refund not yet configured")
}

// ---- Cash Gateway (offline; no network call) ----

type CashGateway struct{}

func NewCashGateway() PaymentGateway { return &CashGateway{} }
func (g *CashGateway) Name() string  { return "cash" }

func (g *CashGateway) InitiatePayment(bookingID uint, amount float64, method string) (*PaymentResult, error) {
	txID := fmt.Sprintf("CASH%d", bookingID)
	return &PaymentResult{TransactionID: txID, Status: "success", GatewayRef: txID, Message: "Cash payment recorded offline"}, nil
}
func (g *CashGateway) VerifyPayment(transactionID string) (*PaymentResult, error) {
	return &PaymentResult{TransactionID: transactionID, Status: "success", Message: "Cash — no verification needed"}, nil
}
func (g *CashGateway) RefundPayment(transactionID string, amount float64, reason string) (*RefundResult, error) {
	return &RefundResult{RefundID: "CASH-REFUND-" + transactionID, Status: "success", Message: "Cash refund to be handled manually"}, nil
}

// ---- Factory ----

// NewGateway selects the gateway by name.
// Set PAYMENT_GATEWAY env var to "razorpay" or "cash"; defaults to "mock".
func NewGateway(gatewayName string) PaymentGateway {
	switch gatewayName {
	case "razorpay":
		return NewRazorpayGateway(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))
	case "cash":
		return NewCashGateway()
	default:
		return NewMockGateway()
	}
}
