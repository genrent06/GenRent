package models

import (
	"time"

	"gorm.io/gorm"
)

type AvailabilityStatus string

const (
	StatusAvailable   AvailabilityStatus = "available"
	StatusReserved    AvailabilityStatus = "reserved"    // Temporarily locked after booking created
	StatusBooked      AvailabilityStatus = "booked"
	StatusMaintenance AvailabilityStatus = "maintenance"
)

type Generator struct {
	ID                 uint               `json:"id" gorm:"primaryKey"`
	VendorID           uint               `json:"vendor_id" gorm:"not null;index"`
	Vendor             Vendor             `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	Name               string             `json:"name" gorm:"not null"`
	CapacityKVA        int                `json:"capacity_kva" gorm:"not null"`
	PricePerDay        float64            `json:"price_per_day" gorm:"not null"`
	PricePerMonth      float64            `json:"price_per_month"`
	FuelType           string             `json:"fuel_type" gorm:"default:diesel"`
	Brand              string             `json:"brand"`
	Location           string             `json:"location" gorm:"not null"`
	City               string             `json:"city" gorm:"not null;index"`
	Latitude           float64            `json:"latitude" gorm:"default:0;index"`
	Longitude          float64            `json:"longitude" gorm:"default:0;index"`
	AvailabilityStatus AvailabilityStatus `json:"availability_status" gorm:"type:varchar(20);default:available;index"`
	ReservationExpiry  *time.Time         `json:"reservation_expiry,omitempty"`
	Description        string             `json:"description"`
	ImageURL           string             `json:"image_url"`
	DeletedAt          gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}
