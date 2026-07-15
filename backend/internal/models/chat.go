package models

import (
	"time"

	"gorm.io/gorm"
)

// Conversation represents a chat conversation between users
type Conversation struct {
	ID           uint64         `json:"id" gorm:"primaryKey"`
	BookingID    *uint64        `json:"booking_id,omitempty" gorm:"index"`
	VendorID     uint64         `json:"vendor_id" gorm:"not null;index"`
	CustomerID   uint64         `json:"customer_id" gorm:"not null;index"`
	LastMessage  string         `json:"last_message" gorm:"type:text"`
	LastMessageAt *time.Time     `json:"last_message_at,omitempty"`
	VendorRead   bool           `json:"vendor_read" gorm:"default:false"`
	CustomerRead  bool           `json:"customer_read" gorm:"default:false"`
	Status       string         `json:"status" gorm:"default:active"` // active, archived, blocked
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Message represents a chat message
type Message struct {
	ID             uint64         `json:"id" gorm:"primaryKey"`
	ConversationID uint64         `json:"conversation_id" gorm:"not null;index"`
	SenderID       uint64         `json:"sender_id" gorm:"not null;index"`
	ReceiverID     uint64         `json:"receiver_id" gorm:"not null;index"`
	Message        string         `json:"message" gorm:"type:text;not null"`
	MessageType    string         `json:"message_type" gorm:"default:text"` // text, image, file, system
	AttachmentURL  string         `json:"attachment_url,omitempty"`
	IsRead         bool           `json:"is_read" gorm:"default:false"`
	ReadAt         *time.Time     `json:"read_at,omitempty"`
	SentAt         time.Time      `json:"sent_at" gorm:"autoCreateTime"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

// NotificationPreference represents user notification settings
type NotificationPreference struct {
	ID              uint64    `json:"id" gorm:"primaryKey"`
	UserID          uint64    `json:"user_id" gorm:"unique;not null;index"`
	EmailEnabled    bool      `json:"email_enabled" gorm:"default:true"`
	SMSEnabled      bool      `json:"sms_enabled" gorm:"default:false"`
	PushEnabled     bool      `json:"push_enabled" gorm:"default:true"`
	InAppEnabled    bool      `json:"in_app_enabled" gorm:"default:true"`
	BookingUpdates  bool      `json:"booking_updates" gorm:"default:true"`
	MessageAlerts   bool      `json:"message_alerts" gorm:"default:true"`
	Promotions      bool      `json:"promotions" gorm:"default:false"`
	Reviews         bool      `json:"reviews" gorm:"default:true"`
	Availability   bool      `json:"availability" gorm:"default:true"`
	QuietHoursStart string    `json:"quiet_hours_start" gorm:"default:22:00"` // HH:MM format
	QuietHoursEnd   string    `json:"quiet_hours_end" gorm:"default:08:00"`   // HH:MM format
	Timezone        string    `json:"timezone" gorm:"default:Asia/Kolkata"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DeviceRegistration represents mobile device for push notifications
type DeviceRegistration struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	UserID     uint64    `json:"user_id" gorm:"not null;index"`
	DeviceToken string   `json:"device_token" gorm:"not null;unique"`
	Platform    string   `json:"platform" gorm:"not null"` // ios, android, web
	DeviceName  string   `json:"device_name"`
	AppVersion  string   `json:"app_version"`
	OSVersion   string   `json:"os_version"`
	IsActive    bool     `json:"is_active" gorm:"default:true"`
	LastUsedAt  time.Time `json:"last_used_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// RealTimeInventory represents real-time equipment availability
type RealTimeInventory struct {
	ID               uint64    `json:"id" gorm:"primaryKey"`
	EquipmentID      uint64    `json:"equipment_id" gorm:"unique;not null;index"`
	AvailableQty     int       `json:"available_qty" gorm:"not null;default:0"`
	ReservedQty      int       `json:"reserved_qty" gorm:"default:0"`
	LastUpdated      time.Time `json:"last_updated"`
	UpdatedBy        uint64    `json:"updated_by"` // User ID who made the update
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for Conversation
func (Conversation) TableName() string {
	return "conversations"
}

// TableName specifies the table name for Message
func (Message) TableName() string {
	return "messages"
}

// TableName specifies the table name for NotificationPreference
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}

// TableName specifies the table name for DeviceRegistration
func (DeviceRegistration) TableName() string {
	return "device_registrations"
}

// TableName specifies the table name for RealTimeInventory
func (RealTimeInventory) TableName() string {
	return "real_time_inventory"
}

// BeforeCreate hook
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook
func (c *Conversation) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate hook
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	m.SentAt = time.Now()
	return nil
}

// BeforeCreate hook
func (n *NotificationPreference) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}
	n.UpdatedAt = now
	return nil
}

// BeforeUpdate hook
func (n *NotificationPreference) BeforeUpdate(tx *gorm.DB) error {
	n.UpdatedAt = time.Now()
	return nil
}
