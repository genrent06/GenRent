package handlers

import (
	"genrent/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetCategories returns all equipment categories with optional hierarchy
func GetCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var categories []models.EquipmentCategory
		
		// Get all root categories (no parent)
		if result := db.
			Where("parent_category_id IS NULL").
			Preload("ParentCategory").
			Order("display_order ASC").
			Find(&categories); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"categories": categories})
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
				"icon_url":       cat.IconURL,
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
