package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateVendorRequest struct {
	CompanyName string  `json:"company_name" binding:"required"`
	Address     string  `json:"address"`
	City        string  `json:"city" binding:"required"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Phone       string  `json:"phone"`
	Description string  `json:"description"`
}

func CreateVendor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var existing models.Vendor
		if result := db.Where("user_id = ?", userID).First(&existing); result.Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "vendor profile already exists"})
			return
		}

		var req CreateVendorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		vendor := models.Vendor{
			UserID:      userID,
			CompanyName: req.CompanyName,
			Address:     req.Address,
			City:        req.City,
			Latitude:    req.Latitude,
			Longitude:   req.Longitude,
			Phone:       req.Phone,
			Description: req.Description,
			Verified:    true, // auto-verify; admin can revoke if needed
		}

		if result := db.Create(&vendor); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create vendor profile"})
			return
		}

		// Update user role to vendor
		db.Model(&models.User{}).Where("id = ?", userID).Update("role", models.RoleVendor)

		c.JSON(http.StatusCreated, vendor)
	}
}

func GetMyVendorProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var vendor models.Vendor
		if result := db.Preload("Generators").Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}
		c.JSON(http.StatusOK, vendor)
	}
}

func UpdateVendorProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		var req CreateVendorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		db.Model(&vendor).Updates(map[string]interface{}{
			"company_name": req.CompanyName,
			"address":      req.Address,
			"city":         req.City,
			"latitude":     req.Latitude,
			"longitude":    req.Longitude,
			"phone":        req.Phone,
			"description":  req.Description,
		})

		c.JSON(http.StatusOK, vendor)
	}
}

func ListVendors(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		latStr := c.Query("lat")
		lngStr := c.Query("lng")
		radiusStr := c.Query("radius")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		query := db.Preload("User").Where("verified = ?", true)

		lat, latOK := parseVendorFloat(latStr)
		lng, lngOK := parseVendorFloat(lngStr)
		if latOK && lngOK {
			radius, _ := strconv.ParseFloat(radiusStr, 64)
			if radius <= 0 {
				radius = 5.0
			}
			if radius > 25.0 {
				radius = 25.0
			}
			haversine := `(6371 * acos(LEAST(1.0, cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) + sin(radians(?)) * sin(radians(latitude)))))`
			query = query.Where(haversine+" <= ?", lat, lng, lat, radius)
		} else if city != "" {
			query = query.Where("city ILIKE ?", "%"+city+"%")
		}

		var vendors []models.Vendor
		var total int64
		query.Model(&models.Vendor{}).Count(&total)
		query.Limit(limit).Offset(offset).Find(&vendors)

		c.JSON(http.StatusOK, gin.H{
			"vendors": vendors,
			"total":   total,
			"page":    page,
			"limit":   limit,
		})
	}
}

func parseVendorFloat(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func GetVendorByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var vendor models.Vendor
		if result := db.Preload("User").Preload("Generators").First(&vendor, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
			return
		}
		c.JSON(http.StatusOK, vendor)
	}
}
