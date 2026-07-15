package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/models"
	"genrent/internal/services/review"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReviewHandler handles review operations
type ReviewHandler struct {
	reviewService *review.ReviewService
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(db *gorm.DB) *ReviewHandler {
	return &ReviewHandler{
		reviewService: review.NewReviewService(db, nil),
	}
}

// CreateReview creates a new review
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		BookingID    uint64  `json:"booking_id" binding:"required"`
		Rating       float32 `json:"rating" binding:"required,min=1,max=5"`
		Title        string  `json:"title" binding:"required,max=200"`
		Comment      string  `json:"comment"`
		Pros         string  `json:"pros"`
		Cons         string  `json:"cons"`
		QualityRating float32 `json:"quality_rating" binding:"omitempty,min=1,max=5"`
		ServiceRating float32 `json:"service_rating" binding:"omitempty,min=1,max=5"`
		ValueRating   float32 `json:"value_rating" binding:"omitempty,min=1,max=5"`
		EquipmentCondition string `json:"equipment_condition" binding:"omitempty,oneof=excellent good fair poor"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get reviewer info from context (or fetch from DB)
	review := &models.Review{
		BookingID:          req.BookingID,
		ReviewerID:          uint64(userID),
		ReviewerName:        "", // Will be filled from user data
		Rating:              req.Rating,
		Title:               req.Title,
		Comment:             req.Comment,
		Pros:                req.Pros,
		Cons:                req.Cons,
		QualityRating:       req.QualityRating,
		ServiceRating:       req.ServiceRating,
		ValueRating:         req.ValueRating,
		EquipmentCondition:  req.EquipmentCondition,
	}

	if err := h.reviewService.CreateReview(review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// GetReview retrieves a specific review
func (h *ReviewHandler) GetReview(c *gin.Context) {
	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	review, err := h.reviewService.GetReview(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetEquipmentReviews retrieves reviews for specific equipment
func (h *ReviewHandler) GetEquipmentReviews(c *gin.Context) {
	equipmentIDStr := c.Param("equipment_id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid equipment ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	reviews, total, err := h.reviewService.GetReviewsByEquipment(equipmentID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get rating summary
	rating, _ := h.reviewService.GetEquipmentRating(equipmentID)

	c.JSON(http.StatusOK, gin.H{
		"reviews":     reviews,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"rating":      rating,
	})
}

// GetVendorReviews retrieves reviews for a vendor
func (h *ReviewHandler) GetVendorReviews(c *gin.Context) {
	vendorIDStr := c.Param("vendor_id")
	vendorID, err := strconv.ParseUint(vendorIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	reviews, total, err := h.reviewService.GetReviewsByVendor(vendorID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get rating summary
	rating, _ := h.reviewService.GetVendorRating(vendorID)

	c.JSON(http.StatusOK, gin.H{
		"reviews":  reviews,
		"total":     total,
		"page":      page,
		"per_page":  perPage,
		"rating":    rating,
	})
}

// GetUserReviews retrieves current user's reviews
func (h *ReviewHandler) GetUserReviews(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	reviews, total, err := h.reviewService.GetUserReviews(uint64(userID), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews":  reviews,
		"total":     total,
		"page":      page,
		"per_page":  perPage,
	})
}

// UpdateReview updates an existing review
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req struct {
		Rating       *float32 `json:"rating" binding:"omitempty,min=1,max=5"`
		Title        *string  `json:"title" binding:"omitempty,max=200"`
		Comment      *string  `json:"comment"`
		Pros         *string  `json:"pros"`
		Cons         *string  `json:"cons"`
		QualityRating *float32 `json:"quality_rating" binding:"omitempty,min=1,max=5"`
		ServiceRating *float32 `json:"service_rating" binding:"omitempty,min=1,max=5"`
		ValueRating   *float32 `json:"value_rating" binding:"omitempty,min=1,max=5"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Rating != nil {
		updates["rating"] = *req.Rating
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Comment != nil {
		updates["comment"] = *req.Comment
	}
	if req.Pros != nil {
		updates["pros"] = *req.Pros
	}
	if req.Cons != nil {
		updates["cons"] = *req.Cons
	}
	if req.QualityRating != nil {
		updates["quality_rating"] = *req.QualityRating
	}
	if req.ServiceRating != nil {
		updates["service_rating"] = *req.ServiceRating
	}
	if req.ValueRating != nil {
		updates["value_rating"] = *req.ValueRating
	}

	if err := h.reviewService.UpdateReview(reviewID, uint64(userID), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review updated"})
}

// DeleteReview deletes a review
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	if err := h.reviewService.DeleteReview(reviewID, uint64(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review deleted"})
}

// RespondToReview allows vendor to respond to a review
func (h *ReviewHandler) RespondToReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req struct {
		Response string `json:"response" binding:"required,max=2000"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reviewService.RespondToReview(reviewID, uint64(userID), req.Response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "response added"})
}

// VoteOnReview marks review as helpful/not helpful
func (h *ReviewHandler) VoteOnReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req struct {
		IsHelpful bool `json:"is_helpful" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reviewService.VoteOnReview(reviewID, uint64(userID), req.IsHelpful); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vote recorded"})
}

// ReportReview reports a review for inappropriate content
func (h *ReviewHandler) ReportReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req struct {
		Reason      string `json:"reason" binding:"required,max=100"`
		Description string `json:"description" binding:"max=500"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reviewService.ReportReview(reviewID, uint64(userID), req.Reason, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "review reported"})
}

// ModerateReview moderates a review (admin only)
func (h *ReviewHandler) ModerateReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=approved rejected hidden"`
		Notes  string `json:"notes" binding:"max=500"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reviewService.ModerateReview(reviewID, uint64(userID), req.Status, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review moderated"})
}

// GetPendingModeration returns reviews pending moderation
func (h *ReviewHandler) GetPendingModeration(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	reviews, err := h.reviewService.GetPendingModeration(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_reviews": reviews,
		"count":           len(reviews),
	})
}

// GetTopRatedVendors returns top-rated vendors
func (h *ReviewHandler) GetTopRatedVendors(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	minRatings, _ := strconv.Atoi(c.DefaultQuery("min_ratings", "5"))

	ratings, err := h.reviewService.GetTopRatedVendors(limit, minRatings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"top_vendors": ratings,
	})
}

// GetTopRatedEquipment returns top-rated equipment
func (h *ReviewHandler) GetTopRatedEquipment(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	minRatings, _ := strconv.Atoi(c.DefaultQuery("min_ratings", "3"))

	ratings, err := h.reviewService.GetTopRatedEquipment(limit, minRatings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"top_equipment": ratings,
	})
}

// GetReviewSummary returns review statistics
func (h *ReviewHandler) GetReviewSummary(c *gin.Context) {
	summary, err := h.reviewService.GetReviewSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetReviewTrends returns review trends over time
func (h *ReviewHandler) GetReviewTrends(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	trends, err := h.reviewService.GetReviewTrends(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"days":   days,
	})
}

// AddReviewImage adds an image to a review
func (h *ReviewHandler) AddReviewImage(c *gin.Context) {
	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req struct {
		ImageURL     string `json:"image_url" binding:"required"`
		Caption      string `json:"caption" binding:"max=255"`
		DisplayOrder int    `json:"display_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reviewService.AddReviewImage(reviewID, req.ImageURL, req.Caption, req.DisplayOrder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "image added"})
}

// GetReviewImages retrieves images for a review
func (h *ReviewHandler) GetReviewImages(c *gin.Context) {
	reviewIDStr := c.Param("id")
	reviewID, err := strconv.ParseUint(reviewIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	images, err := h.reviewService.GetReviewImages(reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images": images,
	})
}
