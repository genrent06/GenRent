package category

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"genrent/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles category-related business logic
type Service struct {
	db *gorm.DB
}

// NewService creates a new category service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateCategoryRequest represents the request to create a category
type CreateCategoryRequest struct {
	ParentID             uint64                    `json:"parent_id,omitempty"`
	Name                 string                    `json:"name" binding:"required"`
	Slug                 string                    `json:"slug"`
	Description          string                    `json:"description"`
	Icon                 string                    `json:"icon"`
	ImageURL             string                    `json:"image_url"`
	DisplayOrder         int                       `json:"display_order"`
	IsActive             bool                      `json:"is_active"`
	SpecsTemplate        map[string]interface{}    `json:"specs_template"`
	MetaTitle            string                    `json:"meta_title"`
	MetaDescription      string                    `json:"meta_description"`
	MetaKeywords         string                    `json:"meta_keywords"`
	RequiresVerification bool                      `json:"requires_verification"`
	MinVendorRating      float32                   `json:"min_vendor_rating"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	ParentID             *uint64                  `json:"parent_id,omitempty"`
	Name                 *string                  `json:"name,omitempty"`
	Slug                 *string                  `json:"slug,omitempty"`
	Description          *string                  `json:"description,omitempty"`
	Icon                 *string                  `json:"icon,omitempty"`
	ImageURL             *string                  `json:"image_url,omitempty"`
	DisplayOrder         *int                     `json:"display_order,omitempty"`
	IsActive             *bool                    `json:"is_active,omitempty"`
	SpecsTemplate        map[string]interface{}    `json:"specs_template,omitempty"`
	MetaTitle            *string                  `json:"meta_title,omitempty"`
	MetaDescription      *string                  `json:"meta_description,omitempty"`
	MetaKeywords         *string                  `json:"meta_keywords,omitempty"`
	RequiresVerification *bool                    `json:"requires_verification,omitempty"`
	MinVendorRating      *float32                 `json:"min_vendor_rating,omitempty"`
}

// CategoryFilterOptions represents filtering options for listing categories
type CategoryFilterOptions struct {
	ParentID    *uint64 `json:"parent_id,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
	Level       *int    `json:"level,omitempty"`           // Filter by hierarchy level
	IncludeInactive bool `json:"include_inactive,omitempty"` // Include inactive categories
	SortBy       string  `json:"sort_by,omitempty"`         // name, display_order, equipment_count
	SortOrder    string  `json:"sort_order,omitempty"`      // asc, desc
	Page         int     `json:"page,omitempty"`
	PageSize     int     `json:"page_size,omitempty"`
}

// CreateCategory creates a new equipment category
func (s *Service) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*models.EquipmentCategory, error) {
	// Validate parent category exists if provided
	if req.ParentID > 0 {
		var parent models.EquipmentCategory
		if err := s.db.WithContext(ctx).First(&parent, req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("parent category not found")
			}
			return nil, fmt.Errorf("failed to validate parent category: %w", err)
		}

		// Check for circular reference
		if err := s.checkCircularReference(ctx, req.ParentID, 0); err != nil {
			return nil, err
		}
	}

	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	// Check for unique slug
	var existingCount int64
	if err := s.db.WithContext(ctx).Model(&models.EquipmentCategory{}).
		Where("slug = ?", slug).Count(&existingCount).Error; err != nil {
		return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
	}
	if existingCount > 0 {
		slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
	}

	// Convert specs template to JSON
	var specsTemplateJSON string
	if req.SpecsTemplate != nil {
		specsBytes, err := json.Marshal(req.SpecsTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal specs template: %w", err)
		}
		specsTemplateJSON = string(specsBytes)
	}

	category := &models.EquipmentCategory{
		ParentID:             nilIfZero(req.ParentID),
		Name:                 req.Name,
		Slug:                 slug,
		Description:          req.Description,
		Icon:                 req.Icon,
		ImageURL:             req.ImageURL,
		DisplayOrder:         req.DisplayOrder,
		IsActive:             req.IsActive,
		SpecsTemplate:        specsTemplateJSON,
		MetaTitle:            req.MetaTitle,
		MetaDescription:      req.MetaDescription,
		MetaKeywords:         req.MetaKeywords,
		RequiresVerification: req.RequiresVerification,
		MinVendorRating:       req.MinVendorRating,
	}

	if err := s.db.WithContext(ctx).Create(category).Error; err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// UpdateCategory updates an existing category
func (s *Service) UpdateCategory(ctx context.Context, categoryID uint64, req UpdateCategoryRequest) (*models.EquipmentCategory, error) {
	var category models.EquipmentCategory
	if err := s.db.WithContext(ctx).First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to find category: %w", err)
	}

	// Validate parent category if being changed
	if req.ParentID != nil {
		if *req.ParentID == categoryID {
			return nil, fmt.Errorf("category cannot be its own parent")
		}

		var parent models.EquipmentCategory
		if err := s.db.WithContext(ctx).First(&parent, *req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("parent category not found")
			}
			return nil, fmt.Errorf("failed to validate parent category: %w", err)
		}

		// Check for circular reference
		if err := s.checkCircularReference(ctx, *req.ParentID, categoryID); err != nil {
			return nil, err
		}
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Slug != nil {
		// Check for unique slug
		var existingCount int64
		if err := s.db.WithContext(ctx).Model(&models.EquipmentCategory{}).
			Where("slug = ? AND id != ?", *req.Slug, categoryID).Count(&existingCount).Error; err != nil {
			return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
		}
		if existingCount > 0 {
			return nil, fmt.Errorf("slug already exists")
		}
		updates["slug"] = *req.Slug
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.ImageURL != nil {
		updates["image_url"] = *req.ImageURL
	}
	if req.DisplayOrder != nil {
		updates["display_order"] = *req.DisplayOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SpecsTemplate != nil {
		specsBytes, err := json.Marshal(req.SpecsTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal specs template: %w", err)
		}
		updates["specs_template"] = string(specsBytes)
	}
	if req.MetaTitle != nil {
		updates["meta_title"] = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		updates["meta_description"] = *req.MetaDescription
	}
	if req.MetaKeywords != nil {
		updates["meta_keywords"] = *req.MetaKeywords
	}
	if req.RequiresVerification != nil {
		updates["requires_verification"] = *req.RequiresVerification
	}
	if req.MinVendorRating != nil {
		updates["min_vendor_rating"] = *req.MinVendorRating
	}
	if req.ParentID != nil {
		updates["parent_id"] = nilIfZero(*req.ParentID)
	}

	if err := s.db.WithContext(ctx).Model(&category).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	// Reload category
	if err := s.db.WithContext(ctx).First(&category, categoryID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload category: %w", err)
	}

	return &category, nil
}

// GetCategory retrieves a single category by ID
func (s *Service) GetCategory(ctx context.Context, categoryID uint64) (*models.EquipmentCategory, error) {
	var category models.EquipmentCategory
	if err := s.db.WithContext(ctx).First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// GetCategoryBySlug retrieves a category by slug
func (s *Service) GetCategoryBySlug(ctx context.Context, slug string) (*models.EquipmentCategory, error) {
	var category models.EquipmentCategory
	if err := s.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// ListCategories retrieves categories with filtering and pagination
func (s *Service) ListCategories(ctx context.Context, opts CategoryFilterOptions) ([]*models.EquipmentCategory, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.EquipmentCategory{})

	// Apply filters
	if opts.ParentID != nil {
		query = query.Where("parent_id = ?", *opts.ParentID)
	}
	if opts.IsActive != nil && !opts.IncludeInactive {
		query = query.Where("is_active = ?", *opts.IsActive)
	}
	if opts.SearchQuery != "" {
		searchTerm := "%" + opts.SearchQuery + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	// Apply sorting
	sortBy := "display_order"
	if opts.SortBy != "" {
		switch opts.SortBy {
		case "name":
			sortBy = "name"
		case "display_order":
			sortBy = "display_order"
		case "equipment_count":
			sortBy = "equipment_count"
		case "created_at":
			sortBy = "created_at"
		}
	}

	sortOrder := "asc"
	if opts.SortOrder == "desc" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	// Execute query
	var categories []*models.EquipmentCategory
	if err := query.Find(&categories).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, total, nil
}

// GetCategoryTree retrieves the full category tree
func (s *Service) GetCategoryTree(ctx context.Context, includeInactive bool) ([]*models.EquipmentCategory, error) {
	query := s.db.WithContext(ctx).Where("parent_id IS NULL")

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	var rootCategories []*models.EquipmentCategory
	if err := query.Order("display_order ASC").Find(&rootCategories).Error; err != nil {
		return nil, fmt.Errorf("failed to get root categories: %w", err)
	}

	// Load children for each root category
	for _, category := range rootCategories {
		if err := s.loadChildren(ctx, category, includeInactive); err != nil {
			return nil, fmt.Errorf("failed to load children: %w", err)
		}
	}

	return rootCategories, nil
}

// GetCategoryPath retrieves the full path from root to category
func (s *Service) GetCategoryPath(ctx context.Context, categoryID uint64) ([]*models.EquipmentCategory, error) {
	var path []*models.EquipmentCategory
	currentID := categoryID

	for {
		var category models.EquipmentCategory
		if err := s.db.WithContext(ctx).First(&category, currentID).Error; err != nil {
			return nil, fmt.Errorf("failed to get category: %w", err)
		}

		path = append([]*models.EquipmentCategory{&category}, path...)

		if category.ParentID == nil {
			break
		}

		currentID = *category.ParentID
	}

	return path, nil
}

// DeleteCategory deletes a category
func (s *Service) DeleteCategory(ctx context.Context, categoryID uint64) error {
	var category models.EquipmentCategory
	if err := s.db.WithContext(ctx).First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("category not found")
		}
		return fmt.Errorf("failed to find category: %w", err)
	}

	// Check if category has children
	var childCount int64
	if err := s.db.WithContext(ctx).Model(&models.EquipmentCategory{}).
		Where("parent_id = ?", categoryID).Count(&childCount).Error; err != nil {
		return fmt.Errorf("failed to check for children: %w", err)
	}
	if childCount > 0 {
		return fmt.Errorf("cannot delete category with children. Move or delete children first")
	}

	// Check if category has equipment
	var equipmentCount int64
	if err := s.db.WithContext(ctx).Table("equipment").
		Where("category_id = ?", categoryID).Count(&equipmentCount).Error; err != nil {
		return fmt.Errorf("failed to check for equipment: %w", err)
	}
	if equipmentCount > 0 {
		return fmt.Errorf("cannot delete category with equipment. Move or delete equipment first")
	}

	if err := s.db.WithContext(ctx).Delete(&category).Error; err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// Category Specifications Management

// CreateCategorySpecification creates a new category specification
func (s *Service) CreateCategorySpecification(ctx context.Context, spec *models.CategorySpecification) error {
	// Validate category exists
	var category models.EquipmentCategory
	if err := s.db.WithContext(ctx).First(&category, spec.CategoryID).Error; err != nil {
		return fmt.Errorf("category not found")
	}

	// Check for duplicate spec key
	var existingCount int64
	if err := s.db.WithContext(ctx).Model(&models.CategorySpecification{}).
		Where("category_id = ? AND spec_key = ?", spec.CategoryID, spec.SpecKey).
		Count(&existingCount).Error; err != nil {
		return fmt.Errorf("failed to check for duplicate spec: %w", err)
	}
	if existingCount > 0 {
		return fmt.Errorf("spec_key already exists for this category")
	}

	if err := s.db.WithContext(ctx).Create(spec).Error; err != nil {
		return fmt.Errorf("failed to create specification: %w", err)
	}

	return nil
}

// UpdateCategorySpecification updates an existing category specification
func (s *Service) UpdateCategorySpecification(ctx context.Context, specID uint64, spec *models.CategorySpecification) error {
	var existing models.CategorySpecification
	if err := s.db.WithContext(ctx).First(&existing, specID).Error; err != nil {
		return fmt.Errorf("specification not found")
	}

	if err := s.db.WithContext(ctx).Model(&existing).Updates(spec).Error; err != nil {
		return fmt.Errorf("failed to update specification: %w", err)
	}

	return nil
}

// GetCategorySpecifications retrieves specifications for a category
func (s *Service) GetCategorySpecifications(ctx context.Context, categoryID uint64) ([]models.CategorySpecification, error) {
	var specs []models.CategorySpecification
	if err := s.db.WithContext(ctx).Where("category_id = ?", categoryID).
		Order("display_order ASC").Find(&specs).Error; err != nil {
		return nil, fmt.Errorf("failed to get specifications: %w", err)
	}

	return specs, nil
}

// DeleteCategorySpecification deletes a category specification
func (s *Service) DeleteCategorySpecification(ctx context.Context, specID uint64) error {
	if err := s.db.WithContext(ctx).Delete(&models.CategorySpecification{}, specID).Error; err != nil {
		return fmt.Errorf("failed to delete specification: %w", err)
	}
	return nil
}

// Category Facets Management

// CreateCategoryFacet creates a new category facet
func (s *Service) CreateCategoryFacet(ctx context.Context, facet *models.CategoryFacet) error {
	// Validate category exists
	var category models.EquipmentCategory
	if err := s.db.WithContext(ctx).First(&category, facet.CategoryID).Error; err != nil {
		return fmt.Errorf("category not found")
	}

	if err := s.db.WithContext(ctx).Create(facet).Error; err != nil {
		return fmt.Errorf("failed to create facet: %w", err)
	}

	return nil
}

// GetCategoryFacets retrieves facets for a category
func (s *Service) GetCategoryFacets(ctx context.Context, categoryID uint64) ([]models.CategoryFacet, error) {
	var facets []models.CategoryFacet
	if err := s.db.WithContext(ctx).Where("category_id = ? AND is_active = ?", categoryID, true).
		Order("display_order ASC").Find(&facets).Error; err != nil {
		return nil, fmt.Errorf("failed to get facets: %w", err)
	}

	return facets, nil
}

// UpdateCategoryFacet updates an existing category facet
func (s *Service) UpdateCategoryFacet(ctx context.Context, facetID uint64, facet *models.CategoryFacet) error {
	var existing models.CategoryFacet
	if err := s.db.WithContext(ctx).First(&existing, facetID).Error; err != nil {
		return fmt.Errorf("facet not found")
	}

	if err := s.db.WithContext(ctx).Model(&existing).Updates(facet).Error; err != nil {
		return fmt.Errorf("failed to update facet: %w", err)
	}

	return nil
}

// DeleteCategoryFacet deletes a category facet
func (s *Service) DeleteCategoryFacet(ctx context.Context, facetID uint64) error {
	if err := s.db.WithContext(ctx).Delete(&models.CategoryFacet{}, facetID).Error; err != nil {
		return fmt.Errorf("failed to delete facet: %w", err)
	}
	return nil
}

// Equipment Specifications Management

// CreateEquipmentSpecification creates equipment specifications
func (s *Service) CreateEquipmentSpecification(ctx context.Context, specs []models.EquipmentSpecification) error {
	if len(specs) == 0 {
		return nil
	}

	// Validate equipment exists
	equipmentID := specs[0].EquipmentID
	var equipment models.Equipment
	if err := s.db.WithContext(ctx).First(&equipment, equipmentID).Error; err != nil {
		return fmt.Errorf("equipment not found")
	}

	// Delete existing specs for this equipment
	if err := s.db.WithContext(ctx).Where("equipment_id = ?", equipmentID).
		Delete(&models.EquipmentSpecification{}).Error; err != nil {
		return fmt.Errorf("failed to delete existing specs: %w", err)
	}

	// Create new specs
	if err := s.db.WithContext(ctx).Create(&specs).Error; err != nil {
		return fmt.Errorf("failed to create specifications: %w", err)
	}

	return nil
}

// GetEquipmentSpecifications retrieves specifications for equipment
func (s *Service) GetEquipmentSpecifications(ctx context.Context, equipmentID uint64) ([]models.EquipmentSpecification, error) {
	var specs []models.EquipmentSpecification
	if err := s.db.WithContext(ctx).Where("equipment_id = ?", equipmentID).
		Order("display_order ASC").Find(&specs).Error; err != nil {
		return nil, fmt.Errorf("failed to get specifications: %w", err)
	}

	return specs, nil
}

// UpdateCategoryStats updates statistics for a category
func (s *Service) UpdateCategoryStats(ctx context.Context, categoryID uint64) error {
	var stats struct {
		EquipmentCount int64
		TotalRentals   int64
	}

	// Get equipment count
	if err := s.db.WithContext(ctx).Table("equipment").
		Where("category_id = ?", categoryID).Count(&stats.EquipmentCount).Error; err != nil {
		return fmt.Errorf("failed to count equipment: %w", err)
	}

	// Get total rentals
	if err := s.db.WithContext(ctx).Table("bookings").
		Where("equipment_id IN (SELECT id FROM equipment WHERE category_id = ?)", categoryID).
		Count(&stats.TotalRentals).Error; err != nil {
		return fmt.Errorf("failed to count rentals: %w", err)
	}

	// Update category
	if err := s.db.WithContext(ctx).Model(&models.EquipmentCategory{}).
		Where("id = ?", categoryID).Updates(map[string]interface{}{
			"equipment_count": stats.EquipmentCount,
			"total_rentals":   stats.TotalRentals,
		}).Error; err != nil {
		return fmt.Errorf("failed to update category stats: %w", err)
	}

	return nil
}

// Helper functions

// loadChildren recursively loads children for a category
func (s *Service) loadChildren(ctx context.Context, category *models.EquipmentCategory, includeInactive bool) error {
	query := s.db.WithContext(ctx).Where("parent_id = ?", category.ID)

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("display_order ASC").Find(&category.Children).Error; err != nil {
		return err
	}

	// Recursively load children
	for i := range category.Children {
		if err := s.loadChildren(ctx, &category.Children[i], includeInactive); err != nil {
			return err
		}
	}

	return nil
}

// checkCircularReference checks if setting parent would create a circular reference
func (s *Service) checkCircularReference(ctx context.Context, parentID uint64, excludeID uint64) error {
	visited := make(map[uint64]bool)
	currentID := parentID

	for {
		if visited[currentID] {
			return fmt.Errorf("circular reference detected in category hierarchy")
		}
		visited[currentID] = true

		if currentID == excludeID {
			return fmt.Errorf("cannot set category as its own descendant")
		}

		var category models.EquipmentCategory
		if err := s.db.WithContext(ctx).Select("parent_id").First(&category, currentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return err
		}

		if category.ParentID == nil {
			break
		}

		currentID = *category.ParentID
	}

	return nil
}

// generateSlug creates a URL-friendly slug from a string
func generateSlug(text string) string {
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove special characters
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()

	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "category"
	}

	return slug
}

// nilIfZero returns nil if value is 0, otherwise returns pointer to value
func nilIfZero(value uint64) *uint64 {
	if value == 0 {
		return nil
	}
	return &value
}

// GetPopularCategories retrieves categories with most equipment
func (s *Service) GetPopularCategories(ctx context.Context, limit int) ([]*models.EquipmentCategory, error) {
	var categories []*models.EquipmentCategory
	if err := s.db.WithContext(ctx).Where("is_active = ? AND equipment_count > 0", true).
		Order("equipment_count DESC, total_rentals DESC").
		Limit(limit).
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular categories: %w", err)
	}

	return categories, nil
}

// MoveCategory moves a category to a new parent
func (s *Service) MoveCategory(ctx context.Context, categoryID uint64, newParentID *uint64) error {
	var category models.EquipmentCategory
	if err := s.db.WithContext(ctx).First(&category, categoryID).Error; err != nil {
		return fmt.Errorf("category not found")
	}

	// Validate new parent
	if newParentID != nil {
		if *newParentID == categoryID {
			return fmt.Errorf("category cannot be its own parent")
		}

		var parent models.EquipmentCategory
		if err := s.db.WithContext(ctx).First(&parent, *newParentID).Error; err != nil {
			return fmt.Errorf("new parent category not found")
		}

		// Check for circular reference
		if err := s.checkCircularReference(ctx, *newParentID, categoryID); err != nil {
			return err
		}
	}

	// Update parent
	if err := s.db.WithContext(ctx).Model(&category).Update("parent_id", newParentID).Error; err != nil {
		return fmt.Errorf("failed to move category: %w", err)
	}

	return nil
}

// GetCategoriesForEquipment retrieves all categories that could apply to equipment
func (s *Service) GetCategoriesForEquipment(ctx context.Context) ([]*models.EquipmentCategory, error) {
	var categories []*models.EquipmentCategory
	if err := s.db.WithContext(ctx).Where("is_active = ?", true).
		Order("display_order ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

// BulkUpdateDisplayOrder updates display order for multiple categories
func (s *Service) BulkUpdateDisplayOrder(ctx context.Context, updates []struct {
	ID           uint64 `json:"id"`
	DisplayOrder int    `json:"display_order"`
}) error {
	if len(updates) == 0 {
		return nil
	}

	// Use transaction
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			if err := tx.Model(&models.EquipmentCategory{}).
				Where("id = ?", update.ID).
				Update("display_order", update.DisplayOrder).Error; err != nil {
				return fmt.Errorf("failed to update category %d: %w", update.ID, err)
			}
		}
		return nil
	})
}
