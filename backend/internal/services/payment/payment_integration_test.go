package payment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockDB is a mock database for testing
type MockDB struct {
	*gorm.DB
}

// TestStripeService_CreateOrder tests creating a payment order with Stripe
func TestStripeService_CreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		testMode    bool
		req         CreateOrderRequest
		expectError bool
	}{
		{
			name:     "Valid order in test mode",
			testMode: true,
			req: CreateOrderRequest{
				Amount:        1000.00,
				Currency:      "INR",
				CustomerID:    12345,
				CustomerName:  "Test User",
				CustomerEmail: "test@example.com",
				CustomerPhone: "+919876543210",
				BookingID:     100,
				Description:   "Test booking",
			},
			expectError: false,
		},
		{
			name:     "Valid order in production mode (will fail without keys)",
			testMode: false,
			req: CreateOrderRequest{
				Amount:        1000.00,
				Currency:      "INR",
				CustomerID:    12345,
				CustomerName:  "Test User",
				CustomerEmail: "test@example.com",
				CustomerPhone: "+919876543210",
				BookingID:     100,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewStripeService("test_key", "test_secret", tt.testMode, 900)

			resp, err := service.CreateOrder(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.OrderID)
				assert.Equal(t, tt.req.Amount, resp.Amount)
				assert.Equal(t, "INR", resp.Currency)
				assert.Greater(t, resp.ExpiresAt, time.Now().Unix())
			}
		})
	}
}

// TestRazorpayService_CreateOrder tests creating a payment order with Razorpay
func TestRazorpayService_CreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		testMode    bool
		req         CreateOrderRequest
		expectError bool
	}{
		{
			name:     "Valid order in test mode",
			testMode: true,
			req: CreateOrderRequest{
				Amount:        1000.00,
				Currency:      "INR",
				CustomerID:    12345,
				CustomerName:  "Test User",
				CustomerEmail: "test@example.com",
				CustomerPhone: "+919876543210",
				BookingID:     100,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewRazorpayService("test_key", "test_secret", "webhook_secret", tt.testMode, 900)

			resp, err := service.CreateOrder(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.OrderID)
				assert.Equal(t, tt.req.Amount, resp.Amount)
				assert.Equal(t, "INR", resp.Currency)
				assert.Greater(t, resp.ExpiresAt, time.Now().Unix())
			}
		})
	}
}

// TestPaymentMethods tests getting payment methods from both gateways
func TestPaymentMethods(t *testing.T) {
	t.Run("Stripe payment methods", func(t *testing.T) {
		service := NewStripeService("test_key", "test_secret", true, 900)

		methods, err := service.GetPaymentMethods(context.Background())

		assert.NoError(t, err)
		assert.NotEmpty(t, methods)

		// Check for card method
		cardFound := false
		for _, method := range methods {
			if method.Name == "Card" {
				cardFound = true
				assert.Equal(t, "stripe", method.Gateway)
				assert.True(t, method.IsEnabled)
				break
			}
		}
		assert.True(t, cardFound, "Card payment method should be available")
	})

	t.Run("Razorpay payment methods", func(t *testing.T) {
		service := NewRazorpayService("test_key", "test_secret", "webhook_secret", true, 900)

		methods, err := service.GetPaymentMethods(context.Background())

		assert.NoError(t, err)
		assert.NotEmpty(t, methods)

		// Check for UPI method
		upiFound := false
		for _, method := range methods {
			if method.Name == "UPI" {
				upiFound = true
				assert.Equal(t, "razorpay", method.Gateway)
				assert.True(t, method.IsEnabled)
				assert.Equal(t, 1, method.DisplayOrder) // UPI should be first
				break
			}
		}
		assert.True(t, upiFound, "UPI payment method should be available")
	})
}

// TestEscrowService_HoldFunds tests holding funds in escrow
func TestEscrowService_HoldFunds(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Create required tables
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			booking_id INTEGER,
			total_amount REAL,
			status TEXT,
			escrow_held_at DATETIME,
			escrow_released_at DATETIME,
			vendor_amount REAL,
			platform_fee REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payment_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			payment_id INTEGER,
			booking_id INTEGER,
			transaction_type TEXT,
			amount REAL,
			gateway TEXT,
			status TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(t, err)

	// Insert test payment
	paymentID := uint64(1)
	bookingID := uint64(100)
	err = db.Table("payments").Create(map[string]interface{}{
		"id":          paymentID,
		"booking_id":  bookingID,
		"total_amount": 1000.00,
		"status":      "paid",
		"created_at":  time.Now(),
	}).Error
	assert.NoError(t, err)

	// Create escrow service with mock payment gateway
	mockGateway := &MockPaymentGateway{}
	service := NewEscrowService(db, mockGateway, 10.0)

	// Test holding funds
	ctx := context.Background()
	err = service.HoldFunds(ctx, paymentID, bookingID)
	assert.NoError(t, err)

	// Verify payment status changed to escrow
	var payment struct {
		Status      string
		EscrowHeldAt *time.Time
	}
	err = db.Table("payments").
		Select("status, escrow_held_at").
		Where("id = ?", paymentID).
		First(&payment).Error
	assert.NoError(t, err)
	assert.Equal(t, "escrow", payment.Status)
	assert.NotNil(t, payment.EscrowHeldAt)
}

// TestEscrowService_ReleaseFunds tests releasing funds from escrow
func TestEscrowService_ReleaseFunds(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Create required tables
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			booking_id INTEGER,
			total_amount REAL,
			status TEXT,
			escrow_held_at DATETIME,
			escrow_released_at DATETIME,
			vendor_id INTEGER,
			vendor_amount REAL,
			platform_fee REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS vendor_wallets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vendor_id INTEGER,
			balance REAL DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payment_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			payment_id INTEGER,
			booking_id INTEGER,
			transaction_type TEXT,
			amount REAL,
			gateway TEXT,
			status TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(t, err)

	// Insert test payment in escrow
	paymentID := uint64(1)
	bookingID := uint64(100)
	vendorID := uint64(500)
	err = db.Table("payments").Create(map[string]interface{}{
		"id":          paymentID,
		"booking_id":  bookingID,
		"total_amount": 1000.00,
		"status":      "escrow",
		"vendor_id":   vendorID,
		"escrow_held_at": time.Now(),
		"created_at":  time.Now(),
	}).Error
	assert.NoError(t, err)

	// Create escrow service with mock payment gateway
	mockGateway := &MockPaymentGateway{}
	service := NewEscrowService(db, mockGateway, 10.0)

	// Test releasing funds
	ctx := context.Background()
	err = service.ReleaseFunds(ctx, bookingID)
	assert.NoError(t, err)

	// Verify payment status changed to completed
	var payment struct {
		Status           string
		EscrowReleasedAt *time.Time
		VendorAmount     float64
		PlatformFee      float64
	}
	err = db.Table("payments").
		Select("status, escrow_released_at, vendor_amount, platform_fee").
		Where("id = ?", paymentID).
		First(&payment).Error
	assert.NoError(t, err)
	assert.Equal(t, "completed", payment.Status)
	assert.NotNil(t, payment.EscrowReleasedAt)
	assert.Equal(t, 900.0, payment.VendorAmount)  // 1000 - 10% = 900
	assert.Equal(t, 100.0, payment.PlatformFee)  // 10% of 1000

	// Verify vendor wallet balance updated
	var wallet struct {
		Balance float64
	}
	err = db.Table("vendor_wallets").
		Select("balance").
		Where("vendor_id = ?", vendorID).
		First(&wallet).Error
	assert.NoError(t, err)
	assert.Equal(t, 900.0, wallet.Balance)
}

// TestRefundService tests refund operations
func TestRefundService(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Create required tables
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			booking_id INTEGER,
			total_amount REAL,
			status TEXT,
			gateway TEXT,
			gateway_payment_id TEXT,
			refund_id TEXT,
			refund_amount REAL,
			refund_status TEXT,
			refunded_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS refund_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			payment_id INTEGER,
			booking_id INTEGER,
			requested_by INTEGER,
			amount REAL,
			reason TEXT,
			status TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(t, err)

	// Insert test payment in escrow
	paymentID := uint64(1)
	bookingID := uint64(100)
	err = db.Table("payments").Create(map[string]interface{}{
		"id":                paymentID,
		"booking_id":        bookingID,
		"total_amount":      1000.00,
		"status":           "escrow",
		"gateway":          "razorpay",
		"gateway_payment_id": "pay_test123",
		"created_at":       time.Now(),
	}).Error
	assert.NoError(t, err)

	// Create refund service with mock payment gateway
	mockGateway := &MockPaymentGateway{}
	mockEscrow := NewEscrowService(db, mockGateway, 10.0)
	service := NewRefundService(db, mockGateway, mockEscrow, false)

	// Test initiating refund
	ctx := context.Background()
	resp, err := service.InitiateRefund(ctx, RefundRequest{
		PaymentID: "1",
		Amount:    1000.00,
		Reason:    "Customer cancellation",
	})

	// In test mode, this should succeed
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.RefundID)
	assert.Equal(t, 1000.00, resp.Amount)
}

// TestCalculatePartialRefund tests partial refund calculation
func TestCalculatePartialRefund(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Create payment table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			booking_id INTEGER,
			total_amount REAL,
			status TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	assert.NoError(t, err)

	// Insert test payment
	bookingID := uint64(100)
	err = db.Table("payments").Create(map[string]interface{}{
		"booking_id":    bookingID,
		"total_amount":  1000.00,
		"status":       "paid",
		"created_at":   time.Now(),
	}).Error
	assert.NoError(t, err)

	// Create refund service
	mockGateway := &MockPaymentGateway{}
	mockEscrow := NewEscrowService(db, mockGateway, 10.0)
	service := NewRefundService(db, mockGateway, mockEscrow, false)

	tests := []struct {
		name           string
		hoursBefore    int
		expectedAmount float64
		expectedPolicy string
	}{
		{
			name:           "Full refund - 48+ hours",
			hoursBefore:    48,
			expectedAmount: 1000.00,
			expectedPolicy: "Full refund (cancelled 48+ hours before start)",
		},
		{
			name:           "75% refund - 24-48 hours",
			hoursBefore:    36,
			expectedAmount: 750.00,
			expectedPolicy: "75% refund (cancelled 24-48 hours before start)",
		},
		{
			name:           "50% refund - 12-24 hours",
			hoursBefore:    18,
			expectedAmount: 500.00,
			expectedPolicy: "50% refund (cancelled 12-24 hours before start)",
		},
		{
			name:           "25% refund - 6-12 hours",
			hoursBefore:    8,
			expectedAmount: 250.00,
			expectedPolicy: "25% refund (cancelled 6-12 hours before start)",
		},
		{
			name:           "No refund - less than 6 hours",
			hoursBefore:    4,
			expectedAmount: 0.00,
			expectedPolicy: "No refund (cancelled less than 6 hours before start)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, policy := service.CalculatePartialRefund(bookingID, tt.hoursBefore)
			assert.Equal(t, tt.expectedAmount, amount)
			assert.Equal(t, tt.expectedPolicy, policy)
		})
	}
}

// TestWebhookSignature tests webhook signature verification
func TestWebhookSignature(t *testing.T) {
	t.Run("Razorpay webhook signature", func(t *testing.T) {
		service := NewRazorpayService("test_key", "test_secret", "webhook_secret", true, 900)

		// Test with empty webhook secret (test mode)
		payload := []byte(`{"event":"payment.captured","payload":{"payment":{"entity":{"id":"pay_test123"}}}}`)
		err := service.HandleWebhook(context.Background(), payload, "")

		// Should succeed in test mode
		assert.NoError(t, err)
	})

	t.Run("Stripe webhook signature", func(t *testing.T) {
		service := NewStripeService("test_key", "test_secret", true, 900)

		// Test with empty webhook secret (test mode)
		payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_test123"}}}`)
		err := service.HandleWebhook(context.Background(), payload, "")

		// Should succeed in test mode
		assert.NoError(t, err)
	})
}

// TestGetTestCards tests retrieving test card numbers
func TestGetTestCards(t *testing.T) {
	t.Run("Stripe test cards", func(t *testing.T) {
		service := NewStripeService("test_key", "test_secret", true, 900)

		cards := service.GetTestCards()
		assert.NotEmpty(t, cards)
		assert.Contains(t, cards, "visa_success")
		assert.Equal(t, "4242424242424242", cards["visa_success"])
	})

	t.Run("Razorpay test cards", func(t *testing.T) {
		service := NewRazorpayService("test_key", "test_secret", "webhook_secret", true, 900)

		cards := service.GetTestCards()
		assert.NotEmpty(t, cards)
		assert.Contains(t, cards, "success")
		assert.Contains(t, cards, "upi_success")
	})
}

// TestAmountFormatting tests amount conversion between gateways
func TestAmountFormatting(t *testing.T) {
	t.Run("Stripe amount formatting", func(t *testing.T) {
		service := NewStripeService("test_key", "test_secret", true, 900)

		// Test formatting (to cents)
		cents := service.FormatAmountForGateway(1000.50)
		assert.Equal(t, int64(100050), cents)

		// Test parsing (from cents)
		amount := service.ParseAmountFromGateway(100050)
		assert.Equal(t, 1000.50, amount)
	})

	t.Run("Razorpay amount formatting", func(t *testing.T) {
		service := NewRazorpayService("test_key", "test_secret", "webhook_secret", true, 900)

		// Test formatting (to paise)
		paise := service.FormatAmountForGateway(1000.50)
		assert.Equal(t, 100050, paise)

		// Test parsing (from paise)
		amount := service.ParseAmountFromGateway(100050)
		assert.Equal(t, 1000.50, amount)
	})
}

// TestPaymentStatusText tests human-readable status conversion
func TestPaymentStatusText(t *testing.T) {
	t.Run("Razorpay status text", func(t *testing.T) {
		service := NewRazorpayService("test_key", "test_secret", "webhook_secret", true, 900)

		assert.Equal(t, "Payment successful", service.GetPaymentStatusText("captured"))
		assert.Equal(t, "Payment failed", service.GetPaymentStatusText("failed"))
		assert.Equal(t, "Unknown status: unknown", service.GetPaymentStatusText("unknown"))
	})

	t.Run("Stripe status text", func(t *testing.T) {
		service := NewStripeService("test_key", "test_secret", true, 900)

		assert.Equal(t, "Payment successful", service.GetPaymentStatusText("succeeded"))
		assert.Equal(t, "Requires customer action", service.GetPaymentStatusText("requires_action"))
	})
}

// TestValidatePaymentAmount tests payment amount validation
func TestValidatePaymentAmount(t *testing.T) {
	t.Run("Razorpay amount validation", func(t *testing.T) {
		service := NewRazorpayService("test_key", "test_secret", "webhook_secret", true, 900)

		assert.NoError(t, service.ValidatePaymentAmount(100.00))
		assert.Error(t, service.ValidatePaymentAmount(0.50))  // Below minimum
		assert.Error(t, service.ValidatePaymentAmount(2000000.00))  // Above maximum
	})

	t.Run("Stripe amount validation", func(t *testing.T) {
		service := NewStripeService("test_key", "test_secret", true, 900)

		assert.NoError(t, service.ValidatePaymentAmount(100.00))
		assert.Error(t, service.ValidatePaymentAmount(0.25))  // Below minimum
		assert.Error(t, service.ValidatePaymentAmount(2000000.00))  // Above maximum
	})
}

// MockPaymentGateway is a mock implementation of PaymentGateway for testing
type MockPaymentGateway struct{}

func (m *MockPaymentGateway) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
	return &OrderResponse{
		OrderID:   "mock_order_123",
		Amount:    req.Amount,
		Currency:  req.Currency,
		ExpiresAt: time.Now().Add(900 * time.Second).Unix(),
	}, nil
}

func (m *MockPaymentGateway) VerifyPayment(ctx context.Context, paymentID string) (*PaymentDetails, error) {
	return &PaymentDetails{
		PaymentID:  paymentID,
		OrderID:    "mock_order_123",
		Amount:     1000.00,
		Currency:   "INR",
		Status:     "captured",
		Method:     "upi",
		CapturedAt: time.Now().Unix(),
		Metadata:   make(map[string]interface{}),
	}, nil
}

func (m *MockPaymentGateway) ProcessRefund(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	return &RefundResponse{
		RefundID: "mock_refund_123",
		Amount:   amount,
		Status:   "processed",
	}, nil
}

func (m *MockPaymentGateway) GetPaymentMethods(ctx context.Context) ([]PaymentMethod, error) {
	return []PaymentMethod{
		{
			Name:        "UPI",
			Gateway:     "mock",
			MethodType:  "upi",
			DisplayName: "UPI Payment",
			IconURL:     "/static/icons/upi.png",
			IsEnabled:   true,
			DisplayOrder: 1,
		},
	}, nil
}

func (m *MockPaymentGateway) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	// Mock webhook handler - always succeeds in test mode
	return nil
}
