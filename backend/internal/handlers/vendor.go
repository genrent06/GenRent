package handlers

import (
	"fmt"
	"genrent/internal/middleware"
	"genrent/internal/models"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Server-side deduplication cache to prevent rapid duplicate requests
type RequestCache struct {
	mu    sync.RWMutex
	cache map[string]time.Time
}

func (r *RequestCache) Check(key string, ttl time.Duration) bool {
	r.mu.RLock()
	timestamp, exists := r.cache[key]
	r.mu.RUnlock()

	if !exists {
		return false // Not a duplicate
	}

	// Check if the TTL has expired
	if time.Since(timestamp) > ttl {
		r.mu.Lock()
		delete(r.cache, key)
		r.mu.Unlock()
		return false // Expired, not a duplicate
	}

	return true // Duplicate within TTL
}

func (r *RequestCache) Add(key string) {
	r.mu.Lock()
	r.cache[key] = time.Now()
	r.mu.Unlock()
}

// Global request cache for vendor profile updates
var vendorUpdateCache = &RequestCache{
	cache: make(map[string]time.Time),
}

// Phone validation pattern for Indian phone numbers
var phonePattern = regexp.MustCompile(`^(\+91[-\s]?)?[6-9]\d{9}$`)

// validatePhone validates phone number format (Indian format)
func validatePhone(phone string) bool {
	if phone == "" {
		return true // Phone is optional
	}
	// Remove spaces for validation
	normalizedPhone := regexp.MustCompile(`\s`).ReplaceAllString(phone, "")
	return phonePattern.MatchString(normalizedPhone)
}

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

		// Validate phone number if provided
		if req.Phone != "" && !validatePhone(req.Phone) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid phone number",
				"errors": map[string]interface{}{
					"phone": "Phone number must be 10 digits (e.g., 9876543210)",
				},
			})
			return
		}

		// Validate company name length
		if len(req.CompanyName) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid company name",
				"errors": map[string]interface{}{
					"company_name": "Company name must be at least 3 characters",
				},
			})
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

		// SERVER-SIDE DEDUPLICATION: Create a unique key based on user ID and request data
		var req CreateVendorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		// Create cache key from user ID + request content
		cacheKey := fmt.Sprintf("vendor_update_%d_%s", userID,
			fmt.Sprintf("%s|%s|%s|%s|%f|%f|%s",
				req.CompanyName, req.City, req.Phone, req.Address,
				req.Latitude, req.Longitude, req.Description))

		// Check for duplicate request within 10 seconds
		if vendorUpdateCache.Check(cacheKey, 10*time.Second) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Please wait before saving again",
				"message": "Duplicate request detected",
			})
			return
		}

		// Add this request to cache
		vendorUpdateCache.Add(cacheKey)

		var vendor models.Vendor
		result := db.Where("user_id = ?", userID).First(&vendor)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
			return
		}

		// Validate phone number if provided
		if req.Phone != "" && !validatePhone(req.Phone) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid phone number",
				"errors": map[string]interface{}{
					"phone": "Phone number must be 10 digits (e.g., 9876543210)",
				},
			})
			return
		}

		// Validate company name length
		if len(req.CompanyName) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid company name",
				"errors": map[string]interface{}{
					"company_name": "Company name must be at least 3 characters",
				},
			})
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
