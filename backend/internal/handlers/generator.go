package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	defaultRadiusKM = 5.0
	maxRadiusKM     = 25.0
)

type CreateGeneratorRequest struct {
	Name          string  `json:"name" binding:"required"`
	CapacityKVA   int     `json:"capacity_kva" binding:"required"`
	PricePerDay   float64 `json:"price_per_day" binding:"required"`
	PricePerMonth float64 `json:"price_per_month"`
	FuelType      string  `json:"fuel_type"`
	Brand         string  `json:"brand"`
	Location      string  `json:"location" binding:"required"`
	City          string  `json:"city" binding:"required"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Description   string  `json:"description"`
	ImageURL      string  `json:"image_url"`
}

func haversineKM(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKM = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadiusKM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func SearchGenerators(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		capacityStr := c.Query("capacity")
		minPriceStr := c.Query("min_price")
		maxPriceStr := c.Query("max_price")
		latStr := c.Query("lat")
		lngStr := c.Query("lng")
		radiusStr := c.Query("radius")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
		offset := (page - 1) * limit

		query := db.Preload("Vendor").Preload("Vendor.User").
			Where("availability_status = ?", models.StatusAvailable).
			Joins("JOIN vendors ON vendors.id = generators.vendor_id").
			Where("vendors.verified = ?", true)

		// Location-based radius search
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
			haversine := `(6371 * acos(LEAST(1.0, cos(radians(?)) * cos(radians(generators.latitude)) * cos(radians(generators.longitude) - radians(?)) + sin(radians(?)) * sin(radians(generators.latitude)))))`
			query = query.Where(haversine+" <= ?", lat, lng, lat, radius)
		} else if city != "" {
			query = query.Where("generators.city ILIKE ?", "%"+city+"%")
		}

		if capacityStr != "" {
			if capacity, err := strconv.Atoi(capacityStr); err == nil {
				query = query.Where("generators.capacity_kva >= ?", capacity)
			}
		}
		if minPriceStr != "" {
			if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
				query = query.Where("generators.price_per_day >= ?", minPrice)
			}
		}
		if maxPriceStr != "" {
			if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
				query = query.Where("generators.price_per_day <= ?", maxPrice)
			}
		}

		// Rank by vendor reliability and rating: reliability_score*0.5 + average_rating*0.3
		query = query.Order("(vendors.reliability_score * 0.5 + vendors.average_rating * 0.3) DESC")

		var generators []models.Generator
		var total int64
		query.Model(&models.Generator{}).Count(&total)
		query.Limit(limit).Offset(offset).Find(&generators)

		c.JSON(http.StatusOK, gin.H{
			"generators": generators,
			"total":      total,
			"page":       page,
			"limit":      limit,
		})
	}
}

func parseFloat(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func GetGenerator(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var generator models.Generator
		if result := db.Preload("Vendor").Preload("Vendor.User").First(&generator, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "generator not found"})
			return
		}
		c.JSON(http.StatusOK, generator)
	}
}

func CreateGenerator(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile required to add generators"})
			return
		}

		// Abuse protection: max 50 generators per vendor
		var genCount int64
		db.Model(&models.Generator{}).Where("vendor_id = ?", vendor.ID).Count(&genCount)
		if genCount >= 50 {
			c.JSON(http.StatusForbidden, gin.H{"error": "maximum 50 generators allowed per vendor"})
			return
		}

		var req CreateGeneratorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		fuelType := req.FuelType
		if fuelType == "" {
			fuelType = "diesel"
		}

		generator := models.Generator{
			VendorID:      vendor.ID,
			Name:          req.Name,
			CapacityKVA:   req.CapacityKVA,
			PricePerDay:   req.PricePerDay,
			PricePerMonth: req.PricePerMonth,
			FuelType:      fuelType,
			Brand:         req.Brand,
			Location:      req.Location,
			City:          req.City,
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			Description:   req.Description,
			ImageURL:      req.ImageURL,
		}

		if result := db.Create(&generator); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add generator"})
			return
		}

		c.JSON(http.StatusCreated, generator)
	}
}

func UpdateGenerator(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		genID := c.Param("id")

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		var generator models.Generator
		if result := db.Where("id = ? AND vendor_id = ?", genID, vendor.ID).First(&generator); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "generator not found or not owned by you"})
			return
		}

		var req CreateGeneratorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		db.Model(&generator).Updates(map[string]interface{}{
			"name":            req.Name,
			"capacity_kva":    req.CapacityKVA,
			"price_per_day":   req.PricePerDay,
			"price_per_month": req.PricePerMonth,
			"fuel_type":       req.FuelType,
			"brand":           req.Brand,
			"location":        req.Location,
			"city":            req.City,
			"latitude":        req.Latitude,
			"longitude":       req.Longitude,
			"description":     req.Description,
			"image_url":       req.ImageURL,
		})

		c.JSON(http.StatusOK, generator)
	}
}

func DeleteGenerator(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		genID := c.Param("id")

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		if result := db.Where("id = ? AND vendor_id = ?", genID, vendor.ID).Delete(&models.Generator{}); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete generator"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "generator deleted"})
	}
}

func GetMyGenerators(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var generators []models.Generator
		db.Where("vendor_id = ?", vendor.ID).Find(&generators)
		c.JSON(http.StatusOK, gin.H{"generators": generators})
	}
}
