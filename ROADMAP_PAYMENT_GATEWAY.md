# Payment Gateway Integration - Implementation Plan

## Overview

Integrate payment gateways (Razorpay for India, Stripe for international) to enable real-time payments, UPI support, and automated escrow management.

---

## Table of Contents

- [Business Requirements](#business-requirements)
- [Technical Architecture](#technical-architecture)
- [Phase 1: Foundation](#phase-1-foundation)
- [Phase 2: Razorpay Integration](#phase-2-razorpay-integration)
- [Phase 3: Stripe Integration](#phase-3-stripe-integration)
- [Phase 4: Escrow & Refunds](#phase-4-escrow--refunds)
- [Phase 5: Testing & Launch](#phase-5-testing--launch)
- [Security Considerations](#security-considerations)
- [Database Schema](#database-schema)
- [API Endpoints](#api-endpoints)

---

## Business Requirements

### Core Features

| Feature | Description | Priority |
|---------|-------------|----------|
| **Razorpay Integration** | Indian market payments | P0 |
| **UPI Support** | UPI payment method | P0 |
| **Stripe Integration** | International payments | P1 |
| **Card Payments** | Credit/Debit cards | P0 |
| **Net Banking** | Bank transfer option | P1 |
| **Wallet Payments** | Paytm, PhonePe, GPay | P1 |
| **Escrow System** | Hold funds until delivery | P0 |
| **Auto-Refunds** | Automatic refund processing | P0 |
| **Payment Links** | Shareable payment links | P2 |
| **Subscription Billing** | Recurring payments | P2 |

### User Stories

**As a Customer:**
- I want to pay securely using multiple payment methods
- I want to receive payment confirmation immediately
- I want to get automatic refunds if booking is cancelled
- I want to see payment history for all my bookings

**As a Vendor:**
- I want to receive payments automatically after delivery
- I want to track my earnings and withdrawals
- I want to receive payout notifications
- I want to manage multiple bank accounts

**As an Admin:**
- I want to monitor all payment transactions
- I want to handle payment disputes
- I want to generate payment reports
- I want to reconcile payments

---

## Technical Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend                                │
│  Payment Form, UPI QR, Payment Status, Wallet Balance      │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Backend API                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Payment Service                            │  │
│  │  - Create Order                                        │  │
│  │  - Verify Payment                                      │  │
│  │  - Process Refund                                      │  │
│  │  - Handle Webhooks                                     │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Escrow Service                            │  │
│  │  - Hold Funds                                         │  │
│  │  - Release Funds                                      │  │
│  │  - Calculate Platform Fee                             │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Wallet Service                            │  │
│  │  - Vendor Balance                                     │  │
│  │  - Withdrawal Processing                              │  │
│  │  - Transaction History                                │  │
│  └──────────────────────────────────────────────────────┘  │
└──────────────────┬────────────────────────────────────────┘
                   │
        ┌──────────┴──────────┬──────────────┐
        ▼                     ▼              ▼
┌─────────────┐      ┌─────────────┐  ┌─────────────┐
│  Razorpay   │      │   Stripe    │  │  Database   │
│   API       │      │    API      │  │ (PostgreSQL) │
└─────────────┘      └─────────────┘  └─────────────┘
```

### Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Payment Gateways** | Razorpay SDK, Stripe SDK | Payment processing |
| **Database** | PostgreSQL | Transaction records |
| **Queue** | Redis (optional) | Async webhook processing |
| **Logging** | Structured logs | Payment audit trail |
| **Encryption** | AES-256 | Sensitive data protection |

---

## Phase 1: Foundation

### Duration: 1-2 days

### Tasks

#### 1.1 Create Payment Service Structure
```go
// backend/internal/services/payment/
├── payment.go          // Main payment service interface
├── razorpay.go         // Razorpay implementation
├── stripe.go           // Stripe implementation
├── escrow.go           // Escrow management
├── refund.go           // Refund processing
├── webhook.go          // Webhook handling
└── models.go           // Payment models
```

#### 1.2 Configuration Setup

**backend/.env**
```env
# Razorpay (India)
RAZORPAY_KEY_ID=rzp_live_xxxxx
RAZORPAY_KEY_SECRET=secret
RAZORPAY_WEBHOOK_SECRET=webhook_secret

# Stripe (International)
STRIPE_PUBLISHABLE_KEY=pk_live_xxxxx
STRIPE_SECRET_KEY=sk_live_xxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxx

# Payment Settings
PAYMENT_TIMEOUT=900           # 15 minutes
PAYMENT_CURRENCY_INR=INR
PAYMENT_CURRENCY_USD=USD
PLATFORM_FEE_PERCENTAGE=10   # 10% platform fee
REFUND_AUTO_PROCESS=true
```

#### 1.3 Database Schema Updates

```sql
-- Update payments table
ALTER TABLE payments ADD COLUMN gateway VARCHAR(20) DEFAULT 'razorpay';
ALTER TABLE payments ADD COLUMN gateway_order_id VARCHAR;
ALTER TABLE payments ADD COLUMN gateway_payment_id VARCHAR UNIQUE;
ALTER TABLE payments ADD COLUMN gateway_status VARCHAR(30);
ALTER TABLE payments ADD COLUMN payment_method VARCHAR(30);
ALTER TABLE payments ADD COLUMN payment_metadata JSONB DEFAULT '{}'::jsonb;
ALTER TABLE payments ADD COLUMN refund_id VARCHAR;
ALTER TABLE payments ADD COLUMN refund_amount FLOAT DEFAULT 0;
ALTER TABLE payments ADD COLUMN refund_status VARCHAR(20);
ALTER TABLE payments ADD COLUMN refunded_at TIMESTAMP;

-- Add payment settings table
CREATE TABLE payment_settings (
    id BIGSERIAL PRIMARY KEY,
    gateway VARCHAR(20) NOT NULL,
    is_enabled BOOLEAN DEFAULT true,
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add payment methods table
CREATE TABLE payment_methods (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    gateway VARCHAR(20) NOT NULL,
    method_type VARCHAR(30) NOT NULL,
    is_enabled BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default payment methods
INSERT INTO payment_methods (name, gateway, method_type, display_order) VALUES
('UPI', 'razorpay', 'upi', 1),
('Card', 'razorpay', 'card', 2),
('Netbanking', 'razorpay', 'netbanking', 3),
('Wallet', 'razorpay', 'wallet', 4),
('Card', 'stripe', 'card', 10);
```

#### 1.4 Create Base Payment Interface

```go
// backend/internal/services/payment/payment.go
package payment

import (
    "context"
    "gorm.io/gorm"
)

type PaymentGateway interface {
    CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error)
    VerifyPayment(ctx context.Context, paymentID string) (*PaymentDetails, error)
    ProcessRefund(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error)
    GetPaymentMethods(ctx context.Context) ([]PaymentMethod, error)
    HandleWebhook(ctx context.Context, payload []byte, signature string) error
}

type CreateOrderRequest struct {
    Amount        float64
    Currency      string
    CustomerID    uint64
    CustomerEmail string
    CustomerPhone string
    BookingID     uint64
    Description   string
    Metadata      map[string]interface{}
}

type OrderResponse struct {
    OrderID       string
    Amount        float64
    Currency      string
    PaymentURL    string
    ExpiresAt     int64
}

type PaymentDetails struct {
    PaymentID     string
    OrderID       string
    Amount        float64
    Status        string
    Method        string
    Metadata      map[string]interface{}
}
```

---

## Phase 2: Razorpay Integration

### Duration: 3-4 days

### Tasks

#### 2.1 Install Razorpay SDK

```bash
go get github.com/razorpay/razorpay-go/v1
```

#### 2.2 Implement Razorpay Service

**backend/internal/services/payment/razorpay.go**

```go
package payment

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/razorpay/razorpay-go/v1"
)

type RazorpayService struct {
    client        *razorpay.Client
    keyID         string
    keySecret     string
    webhookSecret string
}

func NewRazorpayService(keyID, keySecret, webhookSecret string) *RazorpayService {
    client := razorpay.NewClient(keyID, keySecret)
    return &RazorpayService{
        client:        client,
        keyID:         keyID,
        keySecret:     keySecret,
        webhookSecret: webhookSecret,
    }
}

func (r *RazorpayService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
    // Create order data
    orderData := map[string]interface{}{
        "amount":          req.Amount * 100, // Razorpay uses paise
        "currency":        "INR",
        "receipt":         fmt.Sprintf("booking_%d", req.BookingID),
        "payment_capture": 1,
        "notes": map[string]string{
            "booking_id":  fmt.Sprintf("%d", req.BookingID),
            "customer_id": fmt.Sprintf("%d", req.CustomerID),
        },
    }

    // Create order
    body, err := json.Marshal(orderData)
    if err != nil {
        return nil, err
    }

    resp, err := r.client.Order.Create(body, nil)
    if err != nil {
        return nil, fmt.Errorf("razorpay order creation failed: %w", err)
    }

    // Parse response
    var result map[string]interface{}
    json.Unmarshal(resp, &result)

    orderID := result["id"].(string)
    amount := result["amount"].(float64) / 100

    return &OrderResponse{
        OrderID:    orderID,
        Amount:     amount,
        Currency:   "INR",
        PaymentURL: fmt.Sprintf("https://api.razorpay.com/v1/checkout/%s", orderID),
        ExpiresAt:  int64(result["created_at"].(float64)) + 900, // 15 min
    }, nil
}

func (r *RazorpayService) VerifyPayment(ctx context.Context, paymentID string) (*PaymentDetails, error) {
    // Fetch payment from Razorpay
    resp, err := r.client.Payment.Fetch(paymentID, nil, nil)
    if err != nil {
        return nil, fmt.Errorf("razorpay payment fetch failed: %w", err)
    }

    var result map[string]interface{}
    json.Unmarshal(resp, &result)

    return &PaymentDetails{
        PaymentID: result["id"].(string),
        OrderID:   result["order_id"].(string),
        Amount:    result["amount"].(float64) / 100,
        Status:    result["status"].(string),
        Method:    result["method"].(string),
        Metadata:  extractMetadata(result),
    }, nil
}

func (r *RazorpayService) ProcessRefund(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
    refundData := map[string]interface{}{
        "amount": amount * 100, // Convert to paise
    }

    body, err := json.Marshal(refundData)
    if err != nil {
        return nil, err
    }

    resp, err := r.client.Payment.Refund(paymentID, body, nil)
    if err != nil {
        return nil, fmt.Errorf("razorpay refund failed: %w", err)
    }

    var result map[string]interface{}
    json.Unmarshal(resp, &result)

    return &RefundResponse{
        RefundID: result["id"].(string),
        Amount:   result["amount"].(float64) / 100,
        Status:   result["status"].(string),
    }, nil
}

func (r *RazorpayService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
    // Verify webhook signature
    if !verifyWebhookSignature(payload, signature, r.webhookSecret) {
        return fmt.Errorf("invalid webhook signature")
    }

    // Parse webhook
    var webhook map[string]interface{}
    if err := json.Unmarshal(payload, &webhook); err != nil {
        return err
    }

    // Handle event
    event := webhook["event"].(string)
    switch event {
    case "payment.captured":
        return r.handlePaymentCaptured(webhook)
    case "payment.failed":
        return r.handlePaymentFailed(webhook)
    case "refund.processed":
        return r.handleRefundProcessed(webhook)
    }

    return nil
}
```

#### 2.3 Add Payment Handlers

**backend/internal/handlers/payment.go**

```go
package handlers

import (
    "genrent/internal/services/payment"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type PaymentHandler struct {
    paymentService payment.PaymentGateway
}

func NewPaymentHandler(ps payment.PaymentGateway) *PaymentHandler {
    return &PaymentHandler{paymentService: ps}
}

// CreatePaymentOrder - Initiates payment
func (h *PaymentHandler) CreatePaymentOrder(c *gin.Context) {
    var req struct {
        BookingID     uint64  `json:"booking_id" binding:"required"`
        Amount        float64 `json:"amount" binding:"required"`
        PaymentMethod string  `json:"payment_method"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Get user from context
    user := c.MustGet("user").(User)

    // Create order
    order, err := h.paymentService.CreateOrder(c.Request.Context(), payment.CreateOrderRequest{
        Amount:        req.Amount,
        Currency:      "INR",
        CustomerID:    user.ID,
        CustomerEmail: user.Email,
        CustomerPhone: user.Phone,
        BookingID:     req.BookingID,
        Description:   "Equipment rental booking",
    })

    if err != nil {
        c.JSON(500, gin.H{"error": "failed to create payment order"})
        return
    }

    c.JSON(200, gin.H{
        "order_id":    order.OrderID,
        "amount":      order.Amount,
        "currency":    order.Currency,
        "payment_url": order.PaymentURL,
        "expires_at":  order.ExpiresAt,
    })
}

// VerifyPayment - Verifies payment after completion
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
    paymentID := c.Query("payment_id")

    details, err := h.paymentService.VerifyPayment(c.Request.Context(), paymentID)
    if err != nil {
        c.JSON(500, gin.H{"error": "payment verification failed"})
        return
    }

    c.JSON(200, gin.H{
        "payment_id": details.PaymentID,
        "status":     details.Status,
        "amount":     details.Amount,
    })
}

// ProcessRefund - Initiates refund
func (h *PaymentHandler) ProcessRefund(c *gin.Context) {
    var req struct {
        PaymentID string  `json:"payment_id" binding:"required"`
        Amount    float64 `json:"amount" binding:"required"`
        Reason    string  `json:"reason"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    refund, err := h.paymentService.ProcessRefund(c.Request.Context(), req.PaymentID, req.Amount)
    if err != nil {
        c.JSON(500, gin.H{"error": "refund processing failed"})
        return
    }

    c.JSON(200, gin.H{
        "refund_id": refund.RefundID,
        "amount":    refund.Amount,
        "status":    refund.Status,
    })
}

// HandlePaymentWebhook - Handles payment gateway webhooks
func (h *PaymentHandler) HandlePaymentWebhook(c *gin.Context) {
    payload, _ := c.GetRawData()
    signature := c.GetHeader("X-Razorpay-Signature")

    if err := h.paymentService.HandleWebhook(c.Request.Context(), payload, signature); err != nil {
        c.JSON(400, gin.H{"error": "webhook processing failed"})
        return
    }

    c.JSON(200, gin.H{"status": "webhook processed"})
}
```

#### 2.4 Update Route Registration

**backend/cmd/main.go**

```go
// Payment routes
paymentHandler := handlers.NewPaymentHandler(razorpayService)

protected := api.Group("/")
protected.Use(middleware.Auth(cfg.JWTSecret))
{
    // Customer payment routes
    payments := protected.Group("/payments")
    {
        payments.POST("/create-order", paymentHandler.CreatePaymentOrder)
        payments.GET("/verify", paymentHandler.VerifyPayment)
        payments.POST("/refund", paymentHandler.ProcessRefund)
    }
}

// Public webhook endpoint
api.POST("/webhooks/payment", paymentHandler.HandlePaymentWebhook)
```

---

## Phase 3: Stripe Integration

### Duration: 2-3 days

### Tasks

#### 3.1 Install Stripe SDK

```bash
go get github.com/stripe/stripe-go/v74
```

#### 3.2 Implement Stripe Service

Similar structure to Razorpay but using Stripe's API:
- Create Payment Intent
- Confirm Payment
- Process Refund
- Handle Webhooks

**Key Differences:**
- Stripe uses cents (not paise)
- Different webhook signature format
- Payment Intent vs Order approach

---

## Phase 4: Escrow & Refunds

### Duration: 2-3 days

### Tasks

#### 4.1 Escrow Service

**backend/internal/services/payment/escrow.go**

```go
package payment

import (
    "context"
    "gorm.io/gorm"
)

type EscrowService struct {
    db              *gorm.DB
    platformFeeRate float64 // e.g., 0.10 for 10%
}

func NewEscrowService(db *gorm.DB, platformFeeRate float64) *EscrowService {
    return &EscrowService{
        db:              db,
        platformFeeRate: platformFeeRate,
    }
}

// HoldFunds - Mark payment as held in escrow
func (e *EscrowService) HoldFunds(ctx context.Context, paymentID uint64, bookingID uint64) error {
    var payment Payment
    if err := e.db.First(&payment, paymentID).Error; err != nil {
        return err
    }

    payment.Status = "escrow"
    payment.EscrowReleasedAt = nil
    return e.db.Save(&payment).Error
}

// ReleaseFunds - Release funds to vendor after delivery
func (e *EscrowService) ReleaseFunds(ctx context.Context, bookingID uint64) error {
    return e.db.Transaction(func(tx *gorm.DB) error {
        var payment Payment
        if err := tx.Where("booking_id = ?", bookingID).First(&payment).Error; err != nil {
            return err
        }

        // Calculate amounts
        platformFee := payment.TotalAmount * e.platformFeeRate
        vendorAmount := payment.TotalAmount - platformFee

        // Update payment
        payment.Status = "completed"
        payment.VendorAmount = vendorAmount
        payment.PlatformFee = platformFee
        payment.EscrowReleasedAt = &now

        if err := tx.Save(&payment).Error; err != nil {
            return err
        }

        // Credit vendor wallet
        var wallet Wallet
        if err := tx.Where("vendor_id = ?", payment.Booking.VendorID).First(&wallet).Error; err != nil {
            return err
        }

        wallet.Balance += vendorAmount
        wallet.TotalEarnings += vendorAmount

        // Create transaction record
        transaction := WalletTransaction{
            WalletID: wallet.ID,
            Amount:   vendorAmount,
            Type:     "credit",
            Description: fmt.Sprintf("Booking #%d completed", bookingID),
        }

        return tx.Create(&transaction).Error
    })
}

// ProcessRefund - Auto refund for cancellations
func (e *EscrowService) ProcessRefund(ctx context.Context, bookingID uint64, reason string) error {
    return e.db.Transaction(func(tx *gorm.DB) error {
        var payment Payment
        if err := tx.Where("booking_id = ?", bookingID).First(&payment).Error; err != nil {
            return err
        }

        if payment.Status != "escrow" && payment.Status != "paid" {
            return fmt.Errorf("payment not refundable")
        }

        // Process refund via gateway
        refund, err := e.gateway.ProcessRefund(ctx, payment.GatewayPaymentID, payment.TotalAmount)
        if err != nil {
            return err
        }

        // Update payment
        payment.RefundID = refund.RefundID
        payment.RefundAmount = refund.Amount
        payment.RefundStatus = refund.Status
        payment.Status = "refunded"

        return tx.Save(&payment).Error
    })
}
```

---

## Phase 5: Testing & Launch

### Duration: 2-3 days

### Tasks

#### 5.1 Testing Checklist

**Unit Tests**
- [ ] Payment order creation
- [ ] Payment verification
- [ ] Refund processing
- [ ] Escrow hold/release
- [ ] Webhook handling
- [ ] Error handling

**Integration Tests**
- [ ] Razorpay test mode
- [ ] Stripe test mode
- [ ] End-to-end payment flow
- [ ] Webhook delivery

**Manual Testing**
- [ ] Test payment with ₹1
- [ ] Test payment failure scenarios
- [ ] Test refund processing
- [ ] Test webhook verification
- [ ] Test concurrent payments

#### 5.2 Test Accounts

**Razorpay Test Mode**
- Key ID: `rzp_test_...`
- Test Cards:
  - Success: `4242 4242 4242 4242`
  - Failure: `4000 0000 0000 0002`
  - 3D Secure: `4000 0000 0000 3220`

**Stripe Test Mode**
- Publishable Key: `pk_test_...`
- Test Cards: Available in Stripe docs

#### 5.3 Pre-Launch Checklist

- [ ] Production API keys configured
- [ ] Webhook endpoints registered
- [ ] Webhook secrets configured
- [ ] Database migrations applied
- [ ] Error monitoring setup
- [ ] Payment notification emails configured
- [ ] Refund policy configured
- [ ] Platform fee percentage set
- [ ] KYC compliance verified
- [ ] PCI DSS compliance checked

---

## Security Considerations

### 1. API Key Security

```go
// Never log payment details
type Payment struct {
    GatewayPaymentID string `json:"-"`
    GatewayOrderID    string `json:"-"`
}
```

### 2. Webhook Verification

```go
func verifyWebhookSignature(payload []byte, signature, secret string) bool {
    // Implement HMAC SHA256 verification
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(payload)
    calculatedSignature := hex.EncodeToString(h.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(calculatedSignature))
}
```

### 3. PCI DSS Compliance

- Never store full card details
- Use tokenization
- Use iframes for card input
- Implement 3D Secure

### 4. Rate Limiting

```go
// Payment endpoints: 10 requests per minute
paymentGroup.Use(middleware.RateLimit(10, 60))
```

---

## API Endpoints

### Payment Endpoints

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/payments/create-order` | POST | Yes | Create payment order |
| `/api/v1/payments/verify` | GET | Yes | Verify payment status |
| `/api/v1/payments/refund` | POST | Admin | Process refund |
| `/api/v1/payments/methods` | GET | No | Get payment methods |
| `/api/v1/webhooks/payment` | POST | No | Payment webhook |

---

## Database Schema

### Payment Tables

```sql
-- Updated payments table structure
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    booking_id BIGINT NOT NULL REFERENCES bookings(id),
    
    -- Order details
    total_amount FLOAT NOT NULL,
    advance_amount FLOAT DEFAULT 0,
    vendor_amount FLOAT,
    platform_fee FLOAT,
    
    -- Gateway details
    gateway VARCHAR(20) DEFAULT 'razorpay',
    gateway_order_id VARCHAR,
    gateway_payment_id VARCHAR UNIQUE,
    gateway_status VARCHAR(30),
    payment_method VARCHAR(30),
    payment_metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Escrow
    status VARCHAR(20) DEFAULT 'pending',
    escrow_held_at TIMESTAMP,
    escrow_released_at TIMESTAMP,
    
    -- Refund
    refund_id VARCHAR,
    refund_amount FLOAT DEFAULT 0,
    refund_status VARCHAR(20),
    refunded_at TIMESTAMP,
    
    -- Timestamps
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_gateway_payment ON payments(gateway_payment_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_booking ON payments(booking_id);
```

---

## Cost Estimates

### Razorpay Fees (India)

| Transaction Type | Fee |
|-----------------|-----|
| UPI | 2% + ₹3 |
| Card (Domestic) | 2% + ₹3 |
| Netbanking | 2% + ₹3 |
| Wallet | 2% + ₹3 |
| International Card | 3% + ₹3 |

### Stripe Fees (International)

| Transaction Type | Fee |
|-----------------|-----|
| Domestic Card | 2.9% + $0.30 |
| International Card | 3.9% + $0.30 |
| SEPA Direct Debit | 0.8% + €0.25 |

---

## Timeline Summary

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Phase 1: Foundation | 1-2 days | Service structure, config, schema |
| Phase 2: Razorpay | 3-4 days | Razorpay integration, handlers |
| Phase 3: Stripe | 2-3 days | Stripe integration |
| Phase 4: Escrow | 2-3 days | Escrow service, refunds |
| Phase 5: Testing | 2-3 days | Testing, bug fixes |
| **Total** | **10-15 days** | Production-ready payment system |

---

## Next Steps

1. Review this plan with the team
2. Set up Razorpay test account
3. Begin Phase 1 implementation
4. Schedule daily progress reviews
5. Create issues in GitHub tracker

---

**Document Version:** 1.0
**Last Updated:** January 2024
**Status:** Awaiting Approval
