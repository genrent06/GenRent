package handlers

import (
	"encoding/json"
	"genrent/internal/middleware"
	"genrent/internal/models"
	"genrent/internal/services/search"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SearchHandler handles search-related requests
type SearchHandler struct {
	db                *gorm.DB
	elasticService    *search.ElasticsearchService
	savedSearchSvc    *search.SavedSearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(db *gorm.DB, elasticURL, indexName string) (*SearchHandler, error) {
	elasticSvc, err := search.NewElasticsearchService(elasticURL, indexName)
	if err != nil {
		return nil, err
	}

	savedSearchSvc := search.NewSavedSearchService(db, elasticSvc)

	return &SearchHandler{
		db:             db,
		elasticService: elasticSvc,
		savedSearchSvc: savedSearchSvc,
	}, nil
}

// AdvancedSearchRequest represents an advanced search request
type AdvancedSearchRequest struct {
	Query              string   `json:"query"`
	CategoryIDs        []uint64 `json:"category_ids"`
	Cities             []string `json:"cities"`
	MinDailyPrice      float64  `json:"min_daily_price"`
	MaxDailyPrice      float64  `json:"max_daily_price"`
	MinVendorRating    float32  `json:"min_vendor_rating"`
	AvailabilityStatus string   `json:"availability_status"`
	Latitude           float64  `json:"latitude"`
	Longitude          float64  `json:"longitude"`
	Radius             float64  `json:"radius"`
	Brand              string   `json:"brand"`
	Tags               []string `json:"tags"`
	StartDate          string   `json:"start_date"` // YYYY-MM-DD format
	EndDate            string   `json:"end_date"`   // YYYY-MM-DD format
	SortBy             string   `json:"sort_by"`
	Page               int      `json:"page"`
	PerPage            int      `json:"per_page"`
}

// AdvancedSearch performs advanced full-text search with filters
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	var req AdvancedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Also try query parameters
		req.Query = c.Query("q")
		if citiesParam := c.Query("cities"); citiesParam != "" {
			req.Cities = strings.Split(citiesParam, ",")
		}
		if categoryIDsParam := c.Query("category_ids"); categoryIDsParam != "" {
			for _, idStr := range strings.Split(categoryIDsParam, ",") {
				if id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64); err == nil {
					req.CategoryIDs = append(req.CategoryIDs, id)
				}
			}
		}
		req.StartDate = c.Query("start_date")
		req.EndDate = c.Query("end_date")
		req.SortBy = c.DefaultQuery("sort_by", "relevance")
		req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
		req.PerPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}

	// Build Elasticsearch query
	query := &search.SearchQuery{
		Query:              req.Query,
		CategoryIDs:        req.CategoryIDs,
		Cities:             req.Cities,
		MinDailyPrice:      req.MinDailyPrice,
		MaxDailyPrice:      req.MaxDailyPrice,
		MinVendorRating:    req.MinVendorRating,
		AvailabilityStatus: req.AvailabilityStatus,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
		Radius:             req.Radius,
		Brand:              req.Brand,
		Tags:               req.Tags,
		SortBy:             req.SortBy,
		Page:               req.Page,
		PerPage:            req.PerPage,
	}

	// Execute search
	result, err := h.elasticService.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed: " + err.Error()})
		return
	}

	// Record search history if user is authenticated
	if userID := middleware.GetUserID(c); userID > 0 {
		go h.savedSearchSvc.RecordSearchHistory(uint64(userID), query, int(result.Total))
	}

	// Apply date availability filtering if dates provided
	if req.StartDate != "" && req.EndDate != "" {
		result.Hits = h.filterByAvailability(result.Hits, req.StartDate, req.EndDate)
		result.Total = int64(len(result.Hits))
	}

	c.JSON(http.StatusOK, gin.H{
		"results":   result.Hits,
		"total":     result.Total,
		"page":      result.Page,
		"per_page":  result.PerPage,
		"max_score": result.MaxScore,
	})
}

// SearchAutocomplete provides search suggestions
func (h *SearchHandler) SearchAutocomplete(c *gin.Context) {
	query := c.Query("q")
	field := c.DefaultQuery("field", "")
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	autocompleteQuery := &search.AutocompleteQuery{
		Query: query,
		Field: field,
		Size:  size,
	}

	suggestions, err := h.elasticService.Autocomplete(autocompleteQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "autocomplete failed: " + err.Error()})
		return
	}

	// Also get database suggestions for recent searches
	var dbSuggestions []models.SearchSuggestion
	if userID := middleware.GetUserID(c); userID > 0 {
		dbSuggestions, _ = h.savedSearchSvc.GetSearchSuggestions("", 5)
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestions":     suggestions,
		"recent_searches": dbSuggestions,
	})
}

// GetSearchAggregations returns filter aggregations for search
func (h *SearchHandler) GetSearchAggregations(c *gin.Context) {
	var req AdvancedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Query = c.DefaultQuery("q", "")
	}

	query := &search.SearchQuery{
		Query:       req.Query,
		CategoryIDs: req.CategoryIDs,
		Cities:      req.Cities,
		Brand:       req.Brand,
	}

	aggregations, err := h.elasticService.GetAggregations(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "aggregations failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"aggregations": aggregations,
	})
}

// CreateSavedSearch creates a new saved search
func (h *SearchHandler) CreateSavedSearch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Name            string   `json:"name" binding:"required"`
		Query           string   `json:"query"`
		CategoryIDs     []uint64 `json:"category_ids"`
		Cities          []string `json:"cities"`
		MinPrice        float64  `json:"min_price"`
		MaxPrice        float64  `json:"max_price"`
		MinRating       float32  `json:"min_rating"`
		Brand           string   `json:"brand"`
		Tags            []string `json:"tags"`
		SortBy          string   `json:"sort_by"`
		IsAlert         bool     `json:"is_alert"`
		AlertFrequency  string   `json:"alert_frequency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert arrays to JSON
	categoryIDsJSON, _ := json.Marshal(req.CategoryIDs)
	citiesJSON, _ := json.Marshal(req.Cities)
	tagsJSON, _ := json.Marshal(req.Tags)

	savedSearch := &models.SavedSearch{
		UserID:          uint64(userID),
		Name:            req.Name,
		Query:           req.Query,
		CategoryIDs:     string(categoryIDsJSON),
		Cities:          string(citiesJSON),
		MinPrice:        req.MinPrice,
		MaxPrice:        req.MaxPrice,
		MinRating:       req.MinRating,
		Brand:           req.Brand,
		Tags:            string(tagsJSON),
		SortBy:          req.SortBy,
		IsAlert:         req.IsAlert,
		AlertFrequency:  req.AlertFrequency,
		IsActive:        true,
	}

	if err := h.savedSearchSvc.CreateSavedSearch(uint64(userID), savedSearch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create saved search: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, savedSearch)
}

// GetSavedSearches retrieves user's saved searches
func (h *SearchHandler) GetSavedSearches(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	includeInactive := c.DefaultQuery("include_inactive", "false") == "true"
	searches, err := h.savedSearchSvc.GetSavedSearches(uint64(userID), includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve saved searches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"saved_searches": searches,
	})
}

// ExecuteSavedSearch executes a saved search
func (h *SearchHandler) ExecuteSavedSearch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	searchIDStr := c.Param("id")
	searchID, err := strconv.ParseUint(searchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid search ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	result, err := h.savedSearchSvc.ExecuteSavedSearch(uint64(userID), searchID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute search: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":   result.Hits,
		"total":     result.Total,
		"page":      result.Page,
		"per_page":  result.PerPage,
	})
}

// UpdateSavedSearch updates a saved search
func (h *SearchHandler) UpdateSavedSearch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	searchIDStr := c.Param("id")
	searchID, err := strconv.ParseUint(searchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid search ID"})
		return
	}

	var req struct {
		Name           *string  `json:"name"`
		IsAlert        *bool    `json:"is_alert"`
		AlertFrequency *string  `json:"alert_frequency"`
		IsActive       *bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.IsAlert != nil {
		updates["is_alert"] = *req.IsAlert
	}
	if req.AlertFrequency != nil {
		updates["alert_frequency"] = *req.AlertFrequency
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.savedSearchSvc.UpdateSavedSearch(uint64(userID), searchID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update saved search"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "saved search updated"})
}

// DeleteSavedSearch deletes a saved search
func (h *SearchHandler) DeleteSavedSearch(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	searchIDStr := c.Param("id")
	searchID, err := strconv.ParseUint(searchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid search ID"})
		return
	}

	if err := h.savedSearchSvc.DeleteSavedSearch(uint64(userID), searchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete saved search"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "saved search deleted"})
}

// GetSearchHistory retrieves user's search history
func (h *SearchHandler) GetSearchHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	history, err := h.savedSearchSvc.GetSearchHistory(uint64(userID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve search history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}

// filterByAvailability filters results by equipment availability for given dates
func (h *SearchHandler) filterByAvailability(hits []*search.EquipmentDocument, startDate, endDate string) []*search.EquipmentDocument {
	var available []*search.EquipmentDocument

	for _, hit := range hits {
		// Check if equipment is available for the given date range
		if h.isAvailable(hit.ID, startDate, endDate) {
			available = append(available, hit)
		}
	}

	return available
}

// isAvailable checks if equipment is available for the given date range
func (h *SearchHandler) isAvailable(equipmentID uint64, startDate, endDate string) bool {
	// Check if equipment has enough available quantity for the dates
	var count int64

	// Count confirmed bookings that overlap with the requested dates
	h.db.Table("bookings").
		Where("equipment_id = ?", equipmentID).
		Where("status IN ?", []string{"confirmed", "pending", "dispatched", "delivered"}).
		Where("NOT (end_date < ? OR start_date > ?)", startDate, endDate).
		Count(&count)

	// Get equipment available quantity
	var availableQty int
	h.db.Table("equipment").
		Select("available_quantity").
		Where("id = ?", equipmentID).
		Pluck("available_quantity", &availableQty)

	return int(count) < availableQty
}
