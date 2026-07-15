package search

import (
	"encoding/json"
	"fmt"
	"time"

	"genrent/internal/models"
	"gorm.io/gorm"
)

// SavedSearchService manages saved searches and search history
type SavedSearchService struct {
	db              *gorm.DB
	elasticService  *ElasticsearchService
}

// NewSavedSearchService creates a new saved search service
func NewSavedSearchService(db *gorm.DB, elasticService *ElasticsearchService) *SavedSearchService {
	return &SavedSearchService{
		db:             db,
		elasticService: elasticService,
	}
}

// CreateSavedSearch creates a new saved search
func (s *SavedSearchService) CreateSavedSearch(userID uint64, search *models.SavedSearch) error {
	// Verify user exists
	var userCount int64
	if err := s.db.Table("users").Where("id = ?", userID).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if userCount == 0 {
		return fmt.Errorf("user not found")
	}

	search.UserID = userID
	search.CreatedAt = time.Now()
	search.UpdatedAt = time.Now()

	// Calculate current result count
	if searchResult, err := s.elasticService.Search(s.convertToQuery(search)); err == nil {
		search.ResultCount = int(searchResult.Total)
	}

	return s.db.Create(search).Error
}

// GetSavedSearches retrieves all saved searches for a user
func (s *SavedSearchService) GetSavedSearches(userID uint64, includeInactive bool) ([]models.SavedSearch, error) {
	var searches []models.SavedSearch

	query := s.db.Where("user_id = ?", userID)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("created_at DESC").Find(&searches).Error
	return searches, err
}

// GetSavedSearch retrieves a specific saved search
func (s *SavedSearchService) GetSavedSearch(userID, searchID uint64) (*models.SavedSearch, error) {
	var search models.SavedSearch
	err := s.db.Where("id = ? AND user_id = ?", searchID, userID).First(&search).Error
	if err != nil {
		return nil, err
	}
	return &search, nil
}

// UpdateSavedSearch updates an existing saved search
func (s *SavedSearchService) UpdateSavedSearch(userID, searchID uint64, updates map[string]interface{}) error {
	// Verify ownership
	var count int64
	if err := s.db.Model(&models.SavedSearch{}).
		Where("id = ? AND user_id = ?", searchID, userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("saved search not found")
	}

	// Update
	updates["updated_at"] = time.Now()
	return s.db.Model(&models.SavedSearch{}).
		Where("id = ? AND user_id = ?", searchID, userID).
		Updates(updates).Error
}

// DeleteSavedSearch soft deletes a saved search
func (s *SavedSearchService) DeleteSavedSearch(userID, searchID uint64) error {
	return s.db.Model(&models.SavedSearch{}).
		Where("id = ? AND user_id = ?", searchID, userID).
		Update("is_active", false).Error
}

// ExecuteSavedSearch executes a saved search and returns results
func (s *SavedSearchService) ExecuteSavedSearch(userID, searchID uint64, page, perPage int) (*SearchResult, error) {
	search, err := s.GetSavedSearch(userID, searchID)
	if err != nil {
		return nil, err
	}

	query := s.convertToQuery(search)
	query.Page = page
	query.PerPage = perPage

	result, err := s.elasticService.Search(query)
	if err != nil {
		return nil, err
	}

	// Update result count
	s.db.Model(&models.SavedSearch{}).
		Where("id = ?", searchID).
		Update("result_count", result.Total)

	return result, nil
}

// RecordSearchHistory records a search query in history
func (s *SavedSearchService) RecordSearchHistory(userID uint64, query *SearchQuery, resultCount int) error {
	// Serialize filter arrays
	categoryIDsJSON, _ := json.Marshal(query.CategoryIDs)
	citiesJSON, _ := json.Marshal(query.Cities)
	tagsJSON, _ := json.Marshal(query.Tags)

	history := models.SearchHistory{
		UserID:      userID,
		Query:       query.Query,
		FilterCount: len(query.CategoryIDs) + len(query.Cities) + len(query.Tags),
		ResultCount: resultCount,
		CreatedAt:   time.Now(),
	}

	// Store filter details in Query field as JSON
	filterDetails := map[string]interface{}{
		"query":        query.Query,
		"category_ids": string(categoryIDsJSON),
		"cities":       string(citiesJSON),
		"min_price":    query.MinDailyPrice,
		"max_price":    query.MaxDailyPrice,
		"min_rating":   query.MinVendorRating,
		"tags":         string(tagsJSON),
		"brand":        query.Brand,
	}
	detailsJSON, _ := json.Marshal(filterDetails)
	history.Query = string(detailsJSON)

	return s.db.Create(&history).Error
}

// RecordSearchClick records when a user clicks on a search result
func (s *SavedSearchService) RecordSearchClick(historyID uint64, equipmentID uint64, position int) error {
	return s.db.Model(&models.SearchHistory{}).
		Where("id = ?", historyID).
		Updates(map[string]interface{}{
			"clicked_id":     equipmentID,
			"click_position": position,
		}).Error
}

// GetSearchHistory retrieves search history for a user
func (s *SavedSearchService) GetSearchHistory(userID uint64, limit int) ([]models.SearchHistory, error) {
	var history []models.SearchHistory
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// GetPopularSearches retrieves most popular searches
func (s *SavedSearchService) GetPopularSearches(limit int) ([]models.SearchHistory, error) {
	var history []models.SearchHistory
	err := s.db.Table("search_history").
		Select("query, COUNT(*) as search_count, AVG(result_count) as avg_results").
		Group("query").
		Order("search_count DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// UpdateSearchSuggestion updates or creates a search suggestion
func (s *SavedSearchService) UpdateSearchSuggestion(sType, text string) error {
	var suggestion models.SearchSuggestion
	err := s.db.Where("type = ? AND text = ?", sType, text).First(&suggestion).Error

	if err == gorm.ErrRecordNotFound {
		// Create new suggestion
		suggestion = models.SearchSuggestion{
			Type:  sType,
			Text:  text,
			Count: 1,
		}
		return s.db.Create(&suggestion).Error
	}

	// Update existing
	suggestion.Count++
	suggestion.Popularity = float32(suggestion.Count) / 100.0 // Normalize
	return s.db.Save(&suggestion).Error
}

// GetSearchSuggestions retrieves search suggestions
func (s *SavedSearchService) GetSearchSuggestions(sType string, limit int) ([]models.SearchSuggestion, error) {
	var suggestions []models.SearchSuggestion

	query := s.db.Order("popularity DESC, count DESC")
	if sType != "" {
		query = query.Where("type = ?", sType)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&suggestions).Error
	return suggestions, err
}

// ProcessSearchAlerts processes alerts for saved searches with new results
func (s *SavedSearchService) ProcessSearchAlerts() error {
	// Get all active saved searches with alerts enabled
	var searches []models.SavedSearch
	err := s.db.Where("is_alert = ? AND is_active = ?", true, true).
		Find(&searches).Error
	if err != nil {
		return err
	}

	for _, search := range searches {
		// Check if it's time to send alert based on frequency
		if !s.shouldSendAlert(&search) {
			continue
		}

		// Execute search
		query := s.convertToQuery(&search)
		result, err := s.elasticService.Search(query)
		if err != nil {
			continue
		}

		// Check if there are new results (more than last time)
		if int(result.Total) > search.ResultCount {
			// TODO: Send notification to user about new results
			// For now, just update last_alert_at
			now := time.Now()
			s.db.Model(&models.SavedSearch{}).
				Where("id = ?", search.ID).
				Updates(map[string]interface{}{
					"last_alert_at":  now,
					"result_count":   int(result.Total),
				})
		}
	}

	return nil
}

// shouldSendAlert determines if it's time to send an alert based on frequency
func (s *SavedSearchService) shouldSendAlert(search *models.SavedSearch) bool {
	now := time.Now()

	// If never sent, send immediately
	if search.LastAlertAt == nil {
		return true
	}

	// Check frequency
	switch search.AlertFrequency {
	case "instant":
		// Always send for instant alerts
		return true
	case "daily":
		// Send if last alert was more than 24 hours ago
		return now.Sub(*search.LastAlertAt) > 24*time.Hour
	case "weekly":
		// Send if last alert was more than 7 days ago
		return now.Sub(*search.LastAlertAt) > 7*24*time.Hour
	default:
		return false
	}
}

// convertToQuery converts a SavedSearch to a SearchQuery for Elasticsearch
func (s *SavedSearchService) convertToQuery(search *models.SavedSearch) *SearchQuery {
	query := &SearchQuery{
		Query:           search.Query,
		MinDailyPrice:    search.MinPrice,
		MaxDailyPrice:    search.MaxPrice,
		MinVendorRating:  search.MinRating,
		Brand:           search.Brand,
		SortBy:          search.SortBy,
		Page:            1,
		PerPage:         20,
	}

	// Parse JSON arrays
	if search.CategoryIDs != "" {
		json.Unmarshal([]byte(search.CategoryIDs), &query.CategoryIDs)
	}
	if search.Cities != "" {
		json.Unmarshal([]byte(search.Cities), &query.Cities)
	}
	if search.Tags != "" {
		json.Unmarshal([]byte(search.Tags), &query.Tags)
	}

	return query
}
