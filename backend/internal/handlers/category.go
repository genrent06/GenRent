package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/models"
	"genrent/internal/services/category"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CategoryHandler handles category operations
type CategoryHandler struct {
	categoryService *category.Service
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{
		categoryService: category.NewService(db),
	}
}

// CreateCategory creates a new category (admin only)
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req category.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdCategory, err := h.categoryService.CreateCategory(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdCategory)
}

// GetCategory retrieves a single category by ID
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	category, err := h.categoryService.GetCategory(c, categoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// GetCategoryBySlug retrieves a category by slug
func (h *CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	category, err := h.categoryService.GetCategoryBySlug(c, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// ListCategories retrieves categories with filtering
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	var opts category.CategoryFilterOptions

	// Parse query parameters
	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		if parentID, err := strconv.ParseUint(parentIDStr, 10, 64); err == nil {
			opts.ParentID = &parentID
		}
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			opts.IsActive = &isActive
		}
	}

	opts.SearchQuery = c.Query("search")
	opts.SortBy = c.DefaultQuery("sort_by", "display_order")
	opts.SortOrder = c.DefaultQuery("sort_order", "asc")
	opts.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	opts.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "50"))

	categories, total, err := h.categoryService.ListCategories(c, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"total":      total,
		"page":       opts.Page,
		"page_size":  opts.PageSize,
	})
}

// GetCategoryTree retrieves the full category hierarchy
func (h *CategoryHandler) GetCategoryTree(c *gin.Context) {
	includeInactive := c.DefaultQuery("include_inactive", "false") == "true"

	tree, err := h.categoryService.GetCategoryTree(c, includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": tree})
}

// GetCategoryPath retrieves the full path from root to category
func (h *CategoryHandler) GetCategoryPath(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	path, err := h.categoryService.GetCategoryPath(c, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"path": path})
}

// UpdateCategory updates an existing category (admin only)
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var req category.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedCategory, err := h.categoryService.UpdateCategory(c, categoryID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

// DeleteCategory deletes a category (admin only)
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	if err := h.categoryService.DeleteCategory(c, categoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

// MoveCategory moves a category to a new parent (admin only)
func (h *CategoryHandler) MoveCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	var req struct {
		ParentID *uint64 `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.categoryService.MoveCategory(c, categoryID, req.ParentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category moved"})
}

// GetPopularCategories retrieves categories with most equipment
func (h *CategoryHandler) GetPopularCategories(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	categories, err := h.categoryService.GetPopularCategories(c, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"limit":       limit,
	})
}

// Category Specifications Management

// CreateCategorySpecification creates a new category specification (admin only)
func (h *CategoryHandler) CreateCategorySpecification(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var spec models.CategorySpecification
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.categoryService.CreateCategorySpecification(c, &spec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, spec)
}

// GetCategorySpecifications retrieves specifications for a category
func (h *CategoryHandler) GetCategorySpecifications(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	specs, err := h.categoryService.GetCategorySpecifications(c, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"specifications": specs})
}

// UpdateCategorySpecification updates a category specification (admin only)
func (h *CategoryHandler) UpdateCategorySpecification(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	specIDStr := c.Param("spec_id")
	specID, err := strconv.ParseUint(specIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid specification ID"})
		return
	}

	var spec models.CategorySpecification
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.categoryService.UpdateCategorySpecification(c, specID, &spec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "specification updated"})
}

// DeleteCategorySpecification deletes a category specification (admin only)
func (h *CategoryHandler) DeleteCategorySpecification(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	specIDStr := c.Param("spec_id")
	specID, err := strconv.ParseUint(specIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid specification ID"})
		return
	}

	if err := h.categoryService.DeleteCategorySpecification(c, specID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "specification deleted"})
}

// Category Facets Management

// CreateCategoryFacet creates a new category facet (admin only)
func (h *CategoryHandler) CreateCategoryFacet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var facet models.CategoryFacet
	if err := c.ShouldBindJSON(&facet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.categoryService.CreateCategoryFacet(c, &facet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, facet)
}

// GetCategoryFacets retrieves facets for a category
func (h *CategoryHandler) GetCategoryFacets(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	facets, err := h.categoryService.GetCategoryFacets(c, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"facets": facets})
}

// UpdateCategoryFacet updates a category facet (admin only)
func (h *CategoryHandler) UpdateCategoryFacet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	facetIDStr := c.Param("facet_id")
	facetID, err := strconv.ParseUint(facetIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid facet ID"})
		return
	}

	var facet models.CategoryFacet
	if err := c.ShouldBindJSON(&facet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.categoryService.UpdateCategoryFacet(c, facetID, &facet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "facet updated"})
}

// DeleteCategoryFacet deletes a category facet (admin only)
func (h *CategoryHandler) DeleteCategoryFacet(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	facetIDStr := c.Param("facet_id")
	facetID, err := strconv.ParseUint(facetIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid facet ID"})
		return
	}

	if err := h.categoryService.DeleteCategoryFacet(c, facetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "facet deleted"})
}

// Equipment Specifications Management

// CreateEquipmentSpecification creates equipment specifications
func (h *CategoryHandler) CreateEquipmentSpecification(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var specs []models.EquipmentSpecification
	if err := c.ShouldBindJSON(&specs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.categoryService.CreateEquipmentSpecification(c, specs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "specifications created", "count": len(specs)})
}

// GetEquipmentSpecifications retrieves specifications for equipment
func (h *CategoryHandler) GetEquipmentSpecifications(c *gin.Context) {
	equipmentIDStr := c.Param("equipment_id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid equipment ID"})
		return
	}

	specs, err := h.categoryService.GetEquipmentSpecifications(c, equipmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"specifications": specs})
}

// BulkUpdateDisplayOrder updates display order for multiple categories (admin only)
func (h *CategoryHandler) BulkUpdateDisplayOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Updates []struct {
			ID           uint64 `json:"id"`
			DisplayOrder int    `json:"display_order"`
		} `json:"updates" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.categoryService.BulkUpdateDisplayOrder(c, req.Updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "display orders updated"})
}

// UpdateCategoryStats updates statistics for a category (admin only)
func (h *CategoryHandler) UpdateCategoryStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	if err := h.categoryService.UpdateCategoryStats(c, categoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category stats updated"})
}

// Legacy functions for backward compatibility

// GetCategories returns all equipment categories with optional hierarchy
func GetCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := NewCategoryHandler(db)
		handler.ListCategories(c)
	}
}

// GetCategoryHierarchy returns all categories with subcategories nested
func GetCategoryHierarchy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rootCategories []models.EquipmentCategory
		
		// Get all root categories
		if result := db.
			Where("parent_category_id IS NULL").
			Order("display_order ASC").
			Find(&rootCategories); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
			return
		}

		// For each root category, load subcategories
		response := make([]gin.H, 0, len(rootCategories))
		for _, cat := range rootCategories {
			var subcategories []models.EquipmentCategory
			db.Where("parent_category_id = ?", cat.ID).
				Order("display_order ASC").
				Find(&subcategories)

			catData := gin.H{
				"id":             cat.ID,
				"name":           cat.Name,
				"description":    cat.Description,
				"icon_url":       cat.ImageURL,
				"display_order":  cat.DisplayOrder,
				"subcategories": subcategories,
			}
			response = append(response, catData)
		}

		c.JSON(http.StatusOK, gin.H{"categories": response})
	}
}

// GetCategoryEquipment returns equipment in a specific category
func GetCategoryEquipment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Param("id")
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "12")

		var equipment []models.Equipment
		var total int64

		query := db.
			Preload("Vendor").
			Preload("Vendor.User").
			Preload("Category").
			Where("equipment.category_id = ?", categoryID).
			Where("equipment.availability_status = ?", models.StatusAvailable).
			Joins("JOIN vendors ON vendors.id = equipment.vendor_id").
			Where("vendors.verified = ?", true)

		query.Model(&models.Equipment{}).Count(&total)
		
		pageNum, _ := strconv.Atoi(page)
		limitNum, _ := strconv.Atoi(limit)
		offset := (pageNum - 1) * limitNum

		if result := query.
			Order("(vendors.reliability_score * 0.5 + vendors.average_rating * 0.3) DESC").
			Limit(limitNum).
			Offset(offset).
			Find(&equipment); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch equipment"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"equipment": equipmentToGinH(equipment),
			"total":     total,
			"page":      pageNum,
			"limit":     limitNum,
		})
	}
}

// GetCategory returns a single category with all its equipment
func GetCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Param("id")
		var category models.EquipmentCategory

		if result := db.
			Preload("ParentCategory").
			First(&category, categoryID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}

		// Get subcategories if this is a parent
		var subcategories []models.EquipmentCategory
		db.Where("parent_category_id = ?", category.ID).
			Order("display_order ASC").
			Find(&subcategories)

		// Get equipment count
		var equipmentCount int64
		db.Model(&models.Equipment{}).
			Where("category_id = ?", category.ID).
			Count(&equipmentCount)

		c.JSON(http.StatusOK, gin.H{
			"category":         category,
			"subcategories":    subcategories,
			"equipment_count":  equipmentCount,
		})
	}
}

// PopularCategories returns the most popular categories (with most equipment)
func PopularCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "6")

		limitInt, _ := strconv.Atoi(limit)
		if limitInt <= 0 {
			limitInt = 6
		}

		type PopularCategory struct {
			ID             uint   `json:"id"`
			Name           string `json:"name"`
			EquipmentCount int64  `json:"equipment_count"`
		}

		var categories []PopularCategory
		rows := db.
			Select("ec.id, ec.name, COUNT(e.id) as equipment_count").
			Table("equipment_categories ec").
			Joins("LEFT JOIN equipment e ON e.category_id = ec.id AND e.availability_status = ?", models.StatusAvailable).
			Where("ec.parent_category_id IS NOT NULL").
			Group("ec.id, ec.name").
			Order("equipment_count DESC, ec.id ASC").
			Limit(limitInt).
			Scan(&categories)

		if rows.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch popular categories"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}
