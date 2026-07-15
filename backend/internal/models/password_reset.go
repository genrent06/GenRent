package models

import "time"

type PasswordReset struct {
	ID        uint          `json:"id" gorm:"primaryKey"`
	UserID    uint          `json:"user_id" gorm:"not null;index"`
	Token     string        `json:"token" gorm:"uniqueIndex;not null;index"`
	ExpiresAt time.Time     `json:"expires_at" gorm:"not null;index"`
	CreatedAt time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UsedAt    *time.Time    `json:"used_at,omitempty"`
}
