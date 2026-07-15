package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateEquipmentRequest struct {
	CategoryID        uint                   `json:"category_id" binding:"required"`
	Name              string                 `json:"name" binding:"required"`
	Brand             string                 `json:"brand"`
	Model             string                 `json:"model"`
	DailyPrice        float64                `json:"daily_price" binding:"required"`
	WeeklyPrice       float64                `json:"weekly_price"`
	MonthlyPrice      float64                `json:"monthly_price"`
	MobilizationFee   float64                `json:"mobilization_fee"`
	DemobilizationFee float64                `json:"demobilization_fee"`
	TotalQuantity     int                    `json:"total_quantity"`
	Location          string                 `json:"location" binding:"required"`
	City              string                 `json:"city" binding:"required"`
	Latitude          float64                `json:"latitude"`
	Longitude         float64                `json:"longitude"`
	Description       string                 `json:"description"`
	ImageURL          string                 `json:"image_url"`
	Specs             map[string]interface{} `json:"specs"`
}

type EquipmentSearchResponse struct {
	Equipment []gin.H `json:"equipment"`
	Total     int64   `json:"total"`
	Page      int     `json:"page"`
	Limit     int     `json:"limit"`
}

// SearchEquipment searches for equipment by category, location, price range, spec filters, and optionally by radius
func SearchEquipment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		categoryStr := c.Query("category")
		minPriceStr := c.Query("min_price")
		maxPriceStr := c.Query("max_price")
		latStr := c.Query("lat")
		lngStr := c.Query("lng")
		radiusStr := c.Query("radius")
		// Feature 1: Category-specific spec filters
		specKey := c.Query("spec_key")   // e.g. "capacity_kva"
		specMin := c.Query("spec_min")   // e.g. "50"
		specMax := c.Query("spec_max")   // e.g. "200"
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
		offset := (page - 1) * limit

		query := db.
			Preload("Vendor").
			Preload("Vendor.User").
			Preload("Category").
			// Feature 4: show listings that are available OR still have stock (not fully booked)
			Where("equipment.availability_status != ? AND equipment.available_quantity > 0", models.StatusMaintenance).
			Joins("JOIN vendors ON vendors.id = equipment.vendor_id").
			Where("vendors.verified = ?", true)

		// Filter by category if provided (support both name and ID)
		if categoryStr != "" {
			// Check if it's a numeric ID or category name
			if _, err := strconv.Atoi(categoryStr); err == nil {
				// Numeric: filter by category_id directly
				query = query.Where("equipment.category_id = ?", categoryStr)
			} else {
				// String: join with equipment_categories and filter by name
				query = query.
					Joins("JOIN equipment_categories ON equipment_categories.id = equipment.category_id").
					Where("equipment_categories.name ILIKE ?", categoryStr)
			}
		}

		// Location-based radius search or city search
		lat, latOK := parseFloat(latStr)
		lng, lngOK := parseFloat(lngStr)
		if latOK && lngOK {
			radius, _ := strconv.ParseFloat(radiusStr, 64)
			if radius <= 0 {
				radius = defaultRadiusKM
			}
			if radius > maxRadiusKM {
				radius = maxRadiusKM
			}
			// Haversine formula in SQL (PostgreSQL)
			haversine := `(6371 * acos(LEAST(1.0, cos(radians(?)) * cos(radians(equipment.latitude)) * cos(radians(equipment.longitude) - radians(?)) + sin(radians(?)) * sin(radians(equipment.latitude)))))`
			query = query.Where(haversine+" <= ?", lat, lng, lat, radius)
		} else if city != "" {
			query = query.Where("equipment.city ILIKE ?", "%"+city+"%")
		}

		// Price filtering
		if minPriceStr != "" {
			if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
				query = query.Where("equipment.daily_price >= ?", minPrice)
			}
		}
		if maxPriceStr != "" {
			if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
				query = query.Where("equipment.daily_price <= ?", maxPrice)
			}
		}

		// Feature 1: JSONB spec filtering via generic spec_key/spec_min/spec_max params
		if specKey != "" && specMin != "" {
			if _, err := strconv.ParseFloat(specMin, 64); err == nil {
				query = query.Where("(equipment.specs->>?)::float >= ?", specKey, specMin)
			}
		}
		if specKey != "" && specMax != "" {
			if _, err := strconv.ParseFloat(specMax, 64); err == nil {
				query = query.Where("(equipment.specs->>?)::float <= ?", specKey, specMax)
			}
		}

		// Also support frontend-style spec params: spec_capacity_kva_min, spec_capacity_kva_max, etc.
		// These are parsed from URL params prefixed with "spec_" and suffixed with "_min"/"_max"
		for key, vals := range c.Request.URL.Query() {
			if len(vals) == 0 {
				continue
			}
			val := vals[0]
			if len(key) > 9 && key[:5] == "spec_" {
				if len(key) > 4 && key[len(key)-4:] == "_min" {
					specFieldKey := key[5 : len(key)-4] // strip "spec_" prefix and "_min" suffix
					if _, err := strconv.ParseFloat(val, 64); err == nil {
						query = query.Where("(equipment.specs->>?)::float >= ?", specFieldKey, val)
					}
				} else if len(key) > 4 && key[len(key)-4:] == "_max" {
					specFieldKey := key[5 : len(key)-4] // strip "spec_" prefix and "_max" suffix
					if _, err := strconv.ParseFloat(val, 64); err == nil {
						query = query.Where("(equipment.specs->>?)::float <= ?", specFieldKey, val)
					}
				}
			}
		}


		// Rank by vendor reliability and rating
		query = query.Order("(vendors.reliability_score * 0.5 + vendors.average_rating * 0.3) DESC")

		var equipment []models.Equipment
		var total int64
		query.Model(&models.Equipment{}).Count(&total)
		if result := query.Limit(limit).Offset(offset).Find(&equipment); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search equipment"})
			return
		}

		c.JSON(http.StatusOK, EquipmentSearchResponse{
			Equipment: equipmentToGinH(equipment),
			Total:     total,
			Page:      page,
			Limit:     limit,
		})
	}
}

// GetEquipment retrieves a single equipment by ID
func GetEquipment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var equipment models.Equipment
		if result := db.
			Preload("Vendor").
			Preload("Vendor.User").
			Preload("Category").
			First(&equipment, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "equipment not found"})
			return
		}
		c.JSON(http.StatusOK, equipment)
	}
}

// CreateEquipment creates a new equipment listing (vendor only)
func CreateEquipment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile required to add equipment"})
			return
		}

		// Abuse protection: max 500 equipment per vendor
		var equipCount int64
		db.Model(&models.Equipment{}).Where("vendor_id = ?", vendor.ID).Count(&equipCount)
		if equipCount >= 500 {
			c.JSON(http.StatusForbidden, gin.H{"error": "maximum 500 equipment items allowed per vendor"})
			return
		}

		var req CreateEquipmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		// Verify category exists
		var category models.EquipmentCategory
		if result := db.First(&category, req.CategoryID); result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category not found"})
			return
		}

		qty := req.TotalQuantity
		if qty < 1 {
			qty = 1
		}
		equipment := models.Equipment{
			VendorID:          vendor.ID,
			CategoryID:        req.CategoryID,
			Name:              req.Name,
			Brand:             req.Brand,
			Model:             req.Model,
			DailyPrice:        req.DailyPrice,
			WeeklyPrice:       req.WeeklyPrice,
			MonthlyPrice:      req.MonthlyPrice,
			MobilizationFee:   req.MobilizationFee,
			DemobilizationFee: req.DemobilizationFee,
			TotalQuantity:     qty,
			AvailableQuantity: qty,
			Location:          req.Location,
			City:              req.City,
			Latitude:          req.Latitude,
			Longitude:         req.Longitude,
			Description:       req.Description,
			ImageURL:          req.ImageURL,
			Specs:             models.EquipmentSpecs(req.Specs),
		}

		if result := db.Create(&equipment); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add equipment"})
			return
		}

		c.JSON(http.StatusCreated, equipment)
	}
}

// UpdateEquipment updates equipment details (vendor only)
func UpdateEquipment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		equipID := c.Param("id")

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		var equipment models.Equipment
		if result := db.Where("id = ? AND vendor_id = ?", equipID, vendor.ID).First(&equipment); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "equipment not found or not owned by you"})
			return
		}

		var req CreateEquipmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		// Verify category exists if provided
		if req.CategoryID != 0 {
			var category models.EquipmentCategory
			if result := db.First(&category, req.CategoryID); result.Error != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "category not found"})
				return
			}
		}

		db.Model(&equipment).Updates(map[string]interface{}{
			"category_id":        req.CategoryID,
			"name":               req.Name,
			"brand":              req.Brand,
			"model":              req.Model,
			"daily_price":        req.DailyPrice,
			"weekly_price":       req.WeeklyPrice,
			"monthly_price":      req.MonthlyPrice,
			"mobilization_fee":   req.MobilizationFee,
			"demobilization_fee": req.DemobilizationFee,
			"location":           req.Location,
			"city":               req.City,
			"latitude":           req.Latitude,
			"longitude":          req.Longitude,
			"description":        req.Description,
			"image_url":          req.ImageURL,
			"specs":              models.EquipmentSpecs(req.Specs),
		})

		c.JSON(http.StatusOK, equipment)
	}
}

// DeleteEquipment deletes equipment (vendor only)
func DeleteEquipment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		equipID := c.Param("id")

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		if result := db.Where("id = ? AND vendor_id = ?", equipID, vendor.ID).Delete(&models.Equipment{}); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete equipment"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "equipment deleted"})
	}
}

// GetMyEquipment returns all equipment for the current vendor
func GetMyEquipment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var equipment []models.Equipment
		db.Where("vendor_id = ?", vendor.ID).
			Preload("Category").
			Find(&equipment)
		c.JSON(http.StatusOK, gin.H{"equipment": equipment})
	}
}

// UpdateEquipmentStatus updates availability status (vendor only)
func UpdateEquipmentStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		equipID := c.Param("id")

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		var equipment models.Equipment
		if result := db.Where("id = ? AND vendor_id = ?", equipID, vendor.ID).First(&equipment); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "equipment not found"})
			return
		}

		var req struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err)})
			return
		}

		// Validate status
		validStatuses := map[string]bool{
			"available":   true,
			"reserved":    true,
			"booked":      true,
			"maintenance": true,
		}
		if !validStatuses[req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}

		db.Model(&equipment).Update("availability_status", req.Status)
		c.JSON(http.StatusOK, equipment)
	}
}

// Helper function to convert equipment to gin.H for consistent API response
func equipmentToGinH(equipment []models.Equipment) []gin.H {
	result := make([]gin.H, 0, len(equipment))
	for _, e := range equipment {
		vendorName := ""
		if e.Vendor.ID != 0 {
			vendorName = e.Vendor.CompanyName
		}
		categoryName := ""
		if e.Category.ID != 0 {
			categoryName = e.Category.Name
		}
		result = append(result, gin.H{
			"id":                   e.ID,
			"name":                 e.Name,
			"category":             categoryName,
			"category_id":          e.CategoryID,
			"brand":                e.Brand,
			"model":                e.Model,
			"daily_price":          e.DailyPrice,
			"weekly_price":         e.WeeklyPrice,
			"monthly_price":        e.MonthlyPrice,
			"mobilization_fee":     e.MobilizationFee,
			"demobilization_fee":   e.DemobilizationFee,
			"total_quantity":       e.TotalQuantity,
			"available_quantity":   e.AvailableQuantity,
			"location":             e.Location,
			"city":                 e.City,
			"availability_status":  e.AvailabilityStatus,
			"vendor":               vendorName,
			"image_url":            e.ImageURL,
			"specs":                e.Specs,
		})
	}
	return result
}
