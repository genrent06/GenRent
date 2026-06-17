package models

import "time"

type PaymentStatus string
type PaymentMethod string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
	PaymentExpired   PaymentStatus = "expired" // customer started payment but never completed

	MethodUPI        PaymentMethod = "upi"
	MethodCard       PaymentMethod = "card"
	MethodNetBanking PaymentMethod = "netbanking"
	MethodWallet     PaymentMethod = "wallet"
	MethodCash       PaymentMethod = "cash"
)

type Payment struct {
	ID            uint          `json:"id" gorm:"primaryKey"`
	BookingID     uint          `json:"booking_id" gorm:"not null;uniqueIndex"`
	Booking       Booking       `json:"booking,omitempty" gorm:"foreignKey:BookingID"`
	TotalAmount   float64       `json:"total_amount"`
	AdvanceAmount float64       `json:"advance_amount"` // 30% of total
	VendorAmount  float64       `json:"vendor_amount"`  // 15% → vendor wallet
	PlatformFee   float64       `json:"platform_fee"`   // 15% → company
	Method        PaymentMethod `json:"method" gorm:"type:varchar(20)"`
	Status        PaymentStatus `json:"status" gorm:"type:varchar(20);default:pending"`
	TransactionID string        `json:"transaction_id"`
	PaidAt        *time.Time    `json:"paid_at"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type VendorWallet struct {
	ID                   uint               `json:"id" gorm:"primaryKey"`
	VendorID             uint               `json:"vendor_id" gorm:"uniqueIndex;not null"`
	Vendor               Vendor             `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	Balance              float64            `json:"balance" gorm:"default:0"`               // Available to withdraw
	HoldBalance          float64            `json:"hold_balance" gorm:"default:0"`          // Escrow — released after delivery
	WithdrawalHoldBalance float64           `json:"withdrawal_hold_balance" gorm:"default:0"` // In-flight withdrawal (pending admin approval)
	Transactions         []WalletTransaction `json:"transactions,omitempty" gorm:"foreignKey:WalletID"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

type WalletTransactionType string

const (
	WalletCredit              WalletTransactionType = "credit"               // Available balance credited
	WalletDebit               WalletTransactionType = "debit"                // Available balance debited
	WalletEscrowHold          WalletTransactionType = "escrow_hold"          // Held in escrow after customer pays
	WalletEscrowRelease       WalletTransactionType = "escrow_release"       // Released from escrow to balance
	WalletWithdrawalHold      WalletTransactionType = "withdrawal_hold"      // Moved to withdrawal hold on request
	WalletWithdrawalCompleted WalletTransactionType = "withdrawal_completed" // Payout confirmed by admin
	WalletWithdrawalRefund    WalletTransactionType = "withdrawal_refund"    // Rejected — returned to balance
)

type WithdrawalStatus string

const (
	WithdrawalOTPPending WithdrawalStatus = "otp_pending" // waiting for vendor OTP confirmation
	WithdrawalExpired    WithdrawalStatus = "expired"     // OTP expired without confirmation
	WithdrawalPending    WithdrawalStatus = "pending"     // confirmed by vendor, waiting for admin
	WithdrawalApproved   WithdrawalStatus = "approved"    // admin approved, payout in progress
	WithdrawalRejected   WithdrawalStatus = "rejected"    // admin rejected
	WithdrawalPaid       WithdrawalStatus = "paid"        // bank transfer confirmed
)

// VendorBankAccount stores verified bank accounts for a vendor.
// Vendors add accounts once; withdrawals reference saved accounts.
type VendorBankAccount struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	VendorID    uint      `json:"vendor_id" gorm:"not null;index"`
	BankName    string    `json:"bank_name"`
	AccountNo   string    `json:"account_no"`
	IFSC        string    `json:"ifsc"`
	AccountName string    `json:"account_name"`
	IsPrimary   bool      `json:"is_primary" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WithdrawalRequest struct {
	ID            uint               `json:"id" gorm:"primaryKey"`
	VendorID      uint               `json:"vendor_id" gorm:"not null;index"`
	Vendor        Vendor             `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	BankAccountID *uint              `json:"bank_account_id"`
	Amount        float64            `json:"amount" gorm:"not null"`
	Status        WithdrawalStatus   `json:"status" gorm:"type:varchar(20);default:pending"`
	// Bank details copied at request time (immutable audit record)
	BankName    string     `json:"bank_name"`
	AccountNo   string     `json:"account_no"`
	IFSC        string     `json:"ifsc"`
	AccountName string     `json:"account_name"`
	AdminNote    string     `json:"admin_note,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	OTPCode      string     `json:"-" gorm:"column:otp_code"`        // never expose to client
	OTPExpiresAt *time.Time `json:"otp_expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type WalletTransaction struct {
	ID          uint                  `json:"id" gorm:"primaryKey"`
	WalletID    uint                  `json:"wallet_id" gorm:"not null"`
	BookingID   *uint                 `json:"booking_id"`
	Amount      float64               `json:"amount"`
	Type        WalletTransactionType `json:"type" gorm:"type:varchar(30)"`
	Description string                `json:"description"`
	CreatedAt   time.Time             `json:"created_at"`
}

// PlatformRevenue records each platform fee collected — separate from vendor wallets.
// Total platform earnings = SUM(amount) WHERE type = 'commission'.
type PlatformRevenueType string

const (
	PlatformCommission PlatformRevenueType = "commission" // 15% platform fee on advance
	PlatformRefund     PlatformRevenueType = "refund"     // fee returned on admin refund
	PlatformAdjustment PlatformRevenueType = "adjustment" // manual correction, penalty compensation, dispute resolution
)

type PlatformRevenue struct {
	ID          uint                `json:"id" gorm:"primaryKey"`
	PaymentID   uint                `json:"payment_id" gorm:"not null;index"`
	BookingID   uint                `json:"booking_id" gorm:"not null;index"`
	Amount      float64             `json:"amount"`
	Type        PlatformRevenueType `json:"type" gorm:"type:varchar(20)"`
	Description string              `json:"description"`
	CreatedAt   time.Time           `json:"created_at"`
}
