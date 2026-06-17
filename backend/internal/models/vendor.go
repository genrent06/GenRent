package models

import (
	"time"

	"gorm.io/gorm"
)

type Vendor struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	UserID      uint        `json:"user_id" gorm:"uniqueIndex;not null"`
	User        User        `json:"user" gorm:"foreignKey:UserID"`
	CompanyName string      `json:"company_name" gorm:"not null"`
	Address     string      `json:"-"`      // internal only
	City        string      `json:"city" gorm:"not null;index"`
	Latitude    float64     `json:"-" gorm:"default:0"`   // used for geo search, not exposed
	Longitude   float64     `json:"-" gorm:"default:0"`
	Description string      `json:"description"`
	Phone       string      `json:"-"`      // never expose to customers via API
	Verified    bool        `json:"verified" gorm:"default:false"`

	// Security deposit — admin sets, vendor must pay before going live
	SecurityDeposit     float64 `json:"security_deposit" gorm:"default:0"`
	SecurityDepositPaid bool    `json:"security_deposit_paid" gorm:"default:false"`

	// Reliability metrics (auto-updated on each booking)
	ReliabilityScore     float64 `json:"reliability_score" gorm:"default:5.0"` // 0–5
	TotalBookings        int     `json:"total_bookings" gorm:"default:0"`
	SuccessfulDeliveries int     `json:"successful_deliveries" gorm:"default:0"`
	CancelledBookings    int     `json:"cancelled_bookings" gorm:"default:0"`
	AverageRating        float64 `json:"average_rating" gorm:"default:0"`
	TotalRatings         int     `json:"total_ratings" gorm:"default:0"`
	AvgResponseMinutes   float64 `json:"avg_response_minutes" gorm:"default:0"`   // avg time to accept bookings
	RiskScore            float64 `json:"risk_score" gorm:"default:0;index"`        // fraud risk (auto-incremented on bad events)

	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Generators []Generator    `json:"generators,omitempty" gorm:"foreignKey:VendorID"`
}
