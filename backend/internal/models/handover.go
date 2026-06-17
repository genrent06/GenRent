package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONBArray is a JSONB-backed string-array for photo URLs
type JSONBArray []string

func (j JSONBArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONBArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// JSONBMap is a JSONB-backed map for checklist key-value pairs
type JSONBMap map[string]interface{}

func (j JSONBMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONBMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// BookingHandover holds evidence for equipment delivery or return
type BookingHandover struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	BookingID  uint      `json:"booking_id" gorm:"not null;index"`
	Booking    *Booking  `json:"booking,omitempty" gorm:"foreignKey:BookingID"`
	Type       string    `json:"type" gorm:"type:varchar(20);not null"` // "delivery" | "return"
	PhotoURLs  JSONBArray `json:"photo_urls" gorm:"type:jsonb;default:'[]'"`
	Checklist  JSONBMap  `json:"checklist" gorm:"type:jsonb;default:'{}'"`
	Notes      string    `json:"notes"`
	UploadedBy *uint     `json:"uploaded_by"`
	VerifiedAt *time.Time `json:"verified_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func (BookingHandover) TableName() string { return "booking_handovers" }

// DisputeStatus is the lifecycle status of a damage dispute
type DisputeStatus string

const (
	DisputeOpen     DisputeStatus = "open"
	DisputeResolved DisputeStatus = "resolved"
	DisputeRejected DisputeStatus = "rejected"
)

// DamageDispute represents a customer-raised equipment damage claim
type DamageDispute struct {
	ID            uint          `json:"id" gorm:"primaryKey"`
	BookingID     uint          `json:"booking_id" gorm:"not null;index"`
	Booking       *Booking      `json:"booking,omitempty" gorm:"foreignKey:BookingID"`
	RaisedBy      uint          `json:"raised_by" gorm:"not null"`
	User          *User         `json:"user,omitempty" gorm:"foreignKey:RaisedBy"`
	Description   string        `json:"description" gorm:"not null"`
	ClaimedAmount float64       `json:"claimed_amount" gorm:"default:0"`
	PhotoURLs     JSONBArray    `json:"photo_urls" gorm:"type:jsonb;default:'[]'"`
	Status        DisputeStatus `json:"status" gorm:"type:varchar(20);default:open"`
	AdminNotes    string        `json:"admin_notes"`
	ResolvedAt    *time.Time    `json:"resolved_at"`
	CreatedAt     time.Time     `json:"created_at"`
}

func (DamageDispute) TableName() string { return "damage_disputes" }
