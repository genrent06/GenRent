package models

import (
	"time"

	"gorm.io/gorm"
)

// EquipmentCategory represents equipment categories with hierarchy
type EquipmentCategory struct {
	ID          uint64         `json:"id" gorm:"primaryKey"`
	ParentID    *uint64        `json:"parent_id,omitempty" gorm:"index"`
	Name        string         `json:"name" gorm:"size:100;not null;unique"`
	Slug        string         `json:"slug" gorm:"size:100;not null;unique"`
	Description string         `json:"description" gorm:"type:text"`
	Icon        string         `json:"icon" gorm:"size:50"`
	ImageURL    string         `json:"image_url" gorm:"size:255"`
	DisplayOrder int           `json:"display_order" gorm:"default:0"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	SpecsTemplate string       `json:"specs_template" gorm:"type:text"` // JSON template for required specs
	// SEO fields
	MetaTitle       string     `json:"meta_title" gorm:"size:100"`
	MetaDescription string     `json:"meta_description" gorm:"size:255"`
	MetaKeywords    string     `json:"meta_keywords" gorm:"size:255"`
	// Category settings
	RequiresVerification bool    `json:"requires_verification" gorm:"default:true"`
	MinVendorRating      float32 `json:"min_vendor_rating" gorm:"default:0"`
	// Statistics
	EquipmentCount  int       `json:"equipment_count" gorm:"default:0"`
	TotalRentals    int64     `json:"total_rentals" gorm:"default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	// Associations
	Children  []EquipmentCategory `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Parent    *EquipmentCategory   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
}

// CategorySpecification defines required/specifications for a category
type CategorySpecification struct {
	ID           uint64 `json:"id" gorm:"primaryKey"`
	CategoryID   uint64 `json:"category_id" gorm:"not null;index"`
	SpecKey      string `json:"spec_key" gorm:"size:50;not null"`
	SpecLabel    string `json:"spec_label" gorm:"size:100;not null"`
	SpecType     string `json:"spec_type" gorm:"size:20;not null"` // text, number, select, boolean, range
	IsRequired   bool   `json:"is_required" gorm:"default:false"`
	DisplayOrder int    `json:"display_order" gorm:"default:0"`
	Options      string `json:"options" gorm:"type:text"` // JSON array for select type
	Unit         string `json:"unit" gorm:"size:20"` // For number type: kg, hp, liters, etc.
	MinValue     *float64 `json:"min_value"`
	MaxValue     *float64 `json:"max_value"`
	DefaultValue string `json:"default_value" gorm:"size:100"`
	HelpText     string `json:"help_text" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// EquipmentSpecification represents equipment specifications
type EquipmentSpecification struct {
	ID          uint64 `json:"id" gorm:"primaryKey"`
	EquipmentID uint64 `json:"equipment_id" gorm:"not null;index"`
	SpecKey     string `json:"spec_key" gorm:"size:50;not null"`
	SpecValue   string `json:"spec_value" gorm:"type:text;not null"`
	SpecLabel   string `json:"spec_label" gorm:"size:100"`
	DisplayOrder int    `json:"display_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// EquipmentComparison stores user comparison lists
type EquipmentComparison struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	UserID      uint64    `json:"user_id" gorm:"not null;index"`
	EquipmentIDs string   `json:"equipment_ids" gorm:"type:text;not null"` // JSON array
	Name        string    `json:"name" gorm:"size:100"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// EquipmentRecommendation represents AI-powered recommendations
type EquipmentRecommendation struct {
	ID            uint64 `json:"id" gorm:"primaryKey"`
	UserID        uint64 `json:"user_id" gorm:"index"`
	EquipmentID   uint64 `json:"equipment_id" gorm:"index"`
	Reason        string `json:"reason" gorm:"type:text"`
	RelevanceScore float32 `json:"relevance_score" gorm:"default:0"`
	IsViewed      bool   `json:"is_viewed" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// CategoryFacet represents searchable facets for categories
type CategoryFacet struct {
	ID          uint64 `json:"id" gorm:"primaryKey"`
	CategoryID  uint64 `json:"category_id" gorm:"not null;index"`
	FacetKey    string `json:"facet_key" gorm:"size:50;not null"`
	FacetLabel  string `json:"facet_label" gorm:"size:100;not null"`
	FacetType   string `json:"facet_type" gorm:"size:20;not null"` // range, select, multiselect
	Options     string `json:"options" gorm:"type:text"` // JSON array of options
	DisplayOrder int   `json:"display_order" gorm:"default:0"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifications
func (EquipmentCategory) TableName() string {
	return "equipment_categories"
}

func (CategorySpecification) TableName() string {
	return "category_specifications"
}

func (EquipmentSpecification) TableName() string {
	return "equipment_specifications"
}

func (EquipmentComparison) TableName() string {
	return "equipment_comparisons"
}

func (EquipmentRecommendation) TableName() string {
	return "equipment_recommendations"
}

func (CategoryFacet) TableName() string {
	return "category_facets"
}

// BeforeCreate hook
func (c *EquipmentCategory) BeforeCreate(tx *gorm.DB) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook
func (c *EquipmentCategory) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// GetPath returns the full category path from root
func (c *EquipmentCategory) GetPath() []string {
	if c.Parent == nil {
		return []string{c.Name}
	}

	parentPath := c.Parent.GetPath()
	return append(parentPath, c.Name)
}

// GetLevel returns the depth level in the hierarchy (0 = root)
func (c *EquipmentCategory) GetLevel() int {
	if c.Parent == nil {
		return 0
	}
	return c.Parent.GetLevel() + 1
}

// IsRoot checks if this is a root category
func (c *EquipmentCategory) IsRoot() bool {
	return c.ParentID == nil
}

// HasChildren checks if category has subcategories
func (c *EquipmentCategory) HasChildren() bool {
	return len(c.Children) > 0
}

// GetActiveChildren returns active children
func (c *EquipmentCategory) GetActiveChildren() []EquipmentCategory {
	var activeChildren []EquipmentCategory
	for _, child := range c.Children {
		if child.IsActive {
			activeChildren = append(activeChildren, child)
		}
	}
	return activeChildren
}

// GetSpecsForCategory returns specifications for this category
func (c *EquipmentCategory) GetSpecsForCategory(db *gorm.DB) ([]CategorySpecification, error) {
	var specs []CategorySpecification
	err := db.Where("category_id = ?", c.ID).Order("display_order ASC").Find(&specs).Error
	return specs, err
}

// GetFacetsForCategory returns searchable facets for this category
func (c *EquipmentCategory) GetFacetsForCategory(db *gorm.DB) ([]CategoryFacet, error) {
	var facets []CategoryFacet
	err := db.Where("category_id = ? AND is_active = ?", c.ID, true).Order("display_order ASC").Find(&facets).Error
	return facets, err
}
