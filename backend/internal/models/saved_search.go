package models

import (
	"time"

	"gorm.io/gorm"
)

// SavedSearch represents a user's saved search query
type SavedSearch struct {
	ID          uint64         `json:"id" gorm:"primaryKey"`
	UserID      uint64         `json:"user_id" gorm:"not null;index"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Query       string         `json:"query" gorm:"type:text"`
	CategoryIDs  string         `json:"category_ids" gorm:"type:text"` // JSON array
	Cities       string         `json:"cities" gorm:"type:text"`        // JSON array
	MinPrice     float64        `json:"min_price"`
	MaxPrice     float64        `json:"max_price"`
	MinRating    float32        `json:"min_rating"`
	Brand        string         `json:"brand" gorm:"size:50"`
	Tags         string         `json:"tags" gorm:"type:text"` // JSON array
	SortBy       string         `json:"sort_by" gorm:"size:20;default:'relevance'"`
	IsAlert      bool           `json:"is_alert" gorm:"default:false"`
	AlertFrequency string       `json:"alert_frequency" gorm:"size:20;default:'daily'"` // 'instant', 'daily', 'weekly'
	LastAlertAt  *time.Time     `json:"last_alert_at"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	ResultCount  int            `json:"result_count" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// SearchHistory represents a user's search history
type SearchHistory struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	UserID      uint64    `json:"user_id" gorm:"not null;index"`
	Query       string    `json:"query" gorm:"type:text"`
	FilterCount int       `json:"filter_count" gorm:"default:0"`
	ResultCount int       `json:"result_count" gorm:"default:0"`
	ClickedID   *uint64   `json:"clicked_id"` // Equipment ID that was clicked
	ClickPosition int     `json:"click_position"` // Position in results
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SearchSuggestion represents a search suggestion/autocomplete entry
type SearchSuggestion struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	Type      string    `json:"type" gorm:"size:20;not null"` // 'name', 'category', 'city', 'brand'
	Text      string    `json:"text" gorm:"size:200;not null"`
	Count     int       `json:"count" gorm:"default:0"`
	Popularity float32   `json:"popularity" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for SavedSearch
func (SavedSearch) TableName() string {
	return "saved_searches"
}

// TableName specifies the table name for SearchHistory
func (SearchHistory) TableName() string {
	return "search_history"
}

// TableName specifies the table name for SearchSuggestion
func (SearchSuggestion) TableName() string {
	return "search_suggestions"
}

// BeforeCreate hook
func (s *SavedSearch) BeforeCreate(tx *gorm.DB) error {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	s.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook
func (s *SavedSearch) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}
