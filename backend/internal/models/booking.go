package models

import "time"

type BookingStatus string

const (
	// Step 1: Customer creates booking
	BookingRequested BookingStatus = "requested"
	// Step 2: Vendor accepts
	BookingAccepted BookingStatus = "accepted"
	// Step 3: Customer pays 30% advance (held in escrow)
	BookingAdvancePaid BookingStatus = "advance_paid"
	// Step 4: Vendor dispatches generator
	BookingDispatched BookingStatus = "dispatched"
	// Step 5: Customer confirms via OTP
	BookingDelivered BookingStatus = "delivered"
	// Step 6: Service completed, escrow released to vendor
	BookingCompleted BookingStatus = "completed"
	// Cancelled at any stage
	BookingCancelled BookingStatus = "cancelled"

	// Legacy aliases kept for compatibility
	BookingPending   BookingStatus = "requested"
	BookingConfirmed BookingStatus = "accepted"
)

// allowedTransitions defines valid status transitions
var allowedTransitions = map[BookingStatus][]BookingStatus{
	BookingRequested:   {BookingAccepted, BookingCancelled},
	BookingAccepted:    {BookingAdvancePaid, BookingCancelled},
	BookingAdvancePaid: {BookingDispatched, BookingCancelled},
	BookingDispatched:  {BookingDelivered},
	BookingDelivered:   {BookingCompleted},
	BookingCompleted:   {},
	BookingCancelled:   {},
}

// CanTransitionTo returns true if the booking can move to the given status
func (b *Booking) CanTransitionTo(next BookingStatus) bool {
	allowed := allowedTransitions[b.Status]
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

type Booking struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	CustomerID    uint              `json:"customer_id" gorm:"not null"`
	Customer      User              `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	GeneratorID   *uint             `json:"generator_id"` // nullable for backward compatibility
	Generator     *Generator        `json:"generator,omitempty" gorm:"foreignKey:GeneratorID"`
	EquipmentID   *uint             `json:"equipment_id"` // new: equipment support
	Equipment     *Equipment        `json:"equipment,omitempty" gorm:"foreignKey:EquipmentID"`
	CategoryID    *uint             `json:"category_id"` // equipment category for revenue tracking
	Category      *EquipmentCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	StartDate          time.Time         `json:"start_date" gorm:"not null"`
	EndDate            time.Time         `json:"end_date" gorm:"not null"`
	TotalPrice         float64           `json:"total_price"`
	RentalPrice        float64           `json:"rental_price"`         // price before transport fees
	MobilizationFee    float64           `json:"mobilization_fee" gorm:"default:0"`
	DemobilizationFee  float64           `json:"demobilization_fee" gorm:"default:0"`
	AdvanceAmount      float64           `json:"advance_amount"` // 30% of total
	AdvancePaid        bool              `json:"advance_paid" gorm:"default:false"`
	Status             BookingStatus     `json:"status" gorm:"type:varchar(30);default:requested"`
	Address            string            `json:"address" gorm:"not null"`
	Notes              string            `json:"notes"`

	// OTP for delivery confirmation
	DeliveryOTP string `json:"-" gorm:"size:6"` // hidden from JSON
	OTPVerified bool   `json:"otp_verified" gorm:"default:false"`

	// Timestamps for each stage
	AcceptedAt   *time.Time `json:"accepted_at"`
	DispatchedAt *time.Time `json:"dispatched_at"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	CompletedAt  *time.Time `json:"completed_at"`

	// Customer rating after completion
	CustomerRating int    `json:"customer_rating" gorm:"default:0"` // 1-5
	CustomerReview string `json:"customer_review"`

	// Cancellation reason
	CancelReason string `json:"cancel_reason"`

	// Return / handover flow (Feature 5)
	ReturnInitiatedAt *time.Time `json:"return_initiated_at"`
	ReturnOTP         string     `json:"-" gorm:"size:6"`
	ReturnOTPVerified bool       `json:"return_otp_verified" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
