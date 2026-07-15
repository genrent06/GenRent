package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type EquipmentSpecs map[string]interface{}

func (es EquipmentSpecs) Value() (driver.Value, error) {
	return json.Marshal(es)
}

func (es *EquipmentSpecs) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(bytes, &es)
}

type Equipment struct {
	ID                 uint               `json:"id" gorm:"primaryKey"`
	VendorID           uint               `json:"vendor_id" gorm:"not null;index"`
	Vendor             Vendor             `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	CategoryID         uint               `json:"category_id" gorm:"not null;index"`
	Category           EquipmentCategory  `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Name               string             `json:"name" gorm:"not null"`
	Brand              string             `json:"brand"`
	Model              string             `json:"model"`
	Description        string             `json:"description"`
	DailyPrice         float64            `json:"daily_price" gorm:"not null"`
	WeeklyPrice        float64            `json:"weekly_price"`
	MonthlyPrice       float64            `json:"monthly_price"`
	MobilizationFee    float64            `json:"mobilization_fee" gorm:"default:0"`
	DemobilizationFee  float64            `json:"demobilization_fee" gorm:"default:0"`
	TotalQuantity      int                `json:"total_quantity" gorm:"default:1"`
	AvailableQuantity  int                `json:"available_quantity" gorm:"default:1"`
	Location           string             `json:"location" gorm:"not null"`
	City               string             `json:"city" gorm:"not null;index"`
	Latitude           float64            `json:"latitude" gorm:"default:0;index"`
	Longitude          float64            `json:"longitude" gorm:"default:0;index"`
	AvailabilityStatus AvailabilityStatus `json:"availability_status" gorm:"type:varchar(20);default:available;index"`
	ReservationExpiry  *time.Time         `json:"reservation_expiry,omitempty"`
	ImageURL           string             `json:"image_url"`
	Specs              EquipmentSpecs     `json:"specs" gorm:"type:jsonb;serializer:json"`
	DeletedAt          gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// TableName specifies table name for Equipment
func (Equipment) TableName() string {
	return "equipment"
}

