package review

import (
	"fmt"
	"time"

	"genrent/internal/models"
	"genrent/internal/services/websocket"

	"gorm.io/gorm"
)

// ReviewService handles all review operations
type ReviewService struct {
	db    *gorm.DB
	wsHub *websocket.Hub
}

// NewReviewService creates a new review service
func NewReviewService(db *gorm.DB, wsHub *websocket.Hub) *ReviewService {
	return &ReviewService{
		db:    db,
		wsHub: wsHub,
	}
}

// CreateReview creates a new review
func (s *ReviewService) CreateReview(review *models.Review) error {
	// Validate booking exists and is completed
	var booking struct {
		ID           uint64
		Status       string
		CustomerID   uint64
		VendorID     uint64
		EquipmentID  uint64
		CompletedAt  *time.Time
	}

	err := s.db.Table("bookings").
		Select("id, status, customer_id, vendor_id, equipment_id, completed_at").
		Where("id = ?", review.BookingID).
		First(&booking).Error
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	if booking.Status != "completed" {
		return fmt.Errorf("can only review completed bookings")
	}

	if booking.CustomerID != review.ReviewerID {
		return fmt.Errorf("only customers can review their bookings")
	}

	// Check if review already exists
	var existingReview models.Review
	err = s.db.Where("booking_id = ? AND reviewer_id = ?", review.BookingID, review.ReviewerID).
		First(&existingReview).Error
	if err == nil {
		return fmt.Errorf("review already exists for this booking")
	}

	// Set related IDs
	review.EquipmentID = booking.EquipmentID
	review.VendorID = booking.VendorID
	review.Status = "pending"
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()

	// Create review
	if err := s.db.Create(review).Error; err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	// Update aggregated ratings
	go s.updateVendorRatings(review.VendorID)
	go s.updateEquipmentRatings(review.EquipmentID)

	// Notify vendor
	if s.wsHub != nil {
		s.wsHub.SendToUser(review.VendorID, "new_review", map[string]interface{}{
			"review_id":   review.ID,
			"booking_id":  review.BookingID,
			"rating":      review.Rating,
			"title":       review.Title,
			"comment":     review.Comment,
			"reviewer_id": review.ReviewerID,
			"created_at":  review.CreatedAt.Unix(),
		})
	}

	return nil
}

// GetReview retrieves a review by ID
func (s *ReviewService) GetReview(reviewID uint64) (*models.Review, error) {
	var review models.Review
	err := s.db.Preload("Images").
		Where("id = ?", reviewID).
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// GetReviewsByEquipment retrieves reviews for specific equipment
func (s *ReviewService) GetReviewsByEquipment(equipmentID uint64, page, perPage int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	offset := (page - 1) * perPage

	s.db.Model(&models.Review{}).
		Where("equipment_id = ? AND status = ?", equipmentID, "approved").
		Count(&total)

	err := s.db.Where("equipment_id = ? AND status = ?", equipmentID, "approved").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&reviews).Error

	return reviews, total, err
}

// GetReviewsByVendor retrieves reviews for a vendor
func (s *ReviewService) GetReviewsByVendor(vendorID uint64, page, perPage int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	offset := (page - 1) * perPage

	s.db.Model(&models.Review{}).
		Where("vendor_id = ? AND status = ?", vendorID, "approved").
		Count(&total)

	err := s.db.Where("vendor_id = ? AND status = ?", vendorID, "approved").
		Preload("Booking").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&reviews).Error

	return reviews, total, err
}

// GetUserReviews retrieves reviews by a user
func (s *ReviewService) GetUserReviews(userID uint64, page, perPage int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	offset := (page - 1) * perPage

	s.db.Model(&models.Review{}).
		Where("reviewer_id = ?", userID).
		Count(&total)

	err := s.db.Where("reviewer_id = ?", userID).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&reviews).Error

	return reviews, total, err
}

// UpdateReview updates an existing review
func (s *ReviewService) UpdateReview(reviewID, userID uint64, updates map[string]interface{}) error {
	var review models.Review
	err := s.db.Where("id = ? AND reviewer_id = ?", reviewID, userID).
		First(&review).Error
	if err != nil {
		return fmt.Errorf("review not found: %w", err)
	}

	// Check if review is editable
	if !review.IsEditable() {
		return fmt.Errorf("review can only be edited within 7 days of creation")
	}

	// Update review
	updates["updated_at"] = time.Now()
	err = s.db.Model(&review).Updates(updates).Error
	if err != nil {
		return err
	}

	// Update aggregated ratings
	go s.updateVendorRatings(review.VendorID)
	go s.updateEquipmentRatings(review.EquipmentID)

	return nil
}

// DeleteReview soft deletes a review
func (s *ReviewService) DeleteReview(reviewID, userID uint64) error {
	var review models.Review
	err := s.db.Where("id = ? AND reviewer_id = ?", reviewID, userID).
		First(&review).Error
	if err != nil {
		return fmt.Errorf("review not found: %w", err)
	}

	// Soft delete by changing status
	err = s.db.Model(&review).Update("status", "hidden").Error
	if err != nil {
		return err
	}

	// Update aggregated ratings
	go s.updateVendorRatings(review.VendorID)
	go s.updateEquipmentRatings(review.EquipmentID)

	return nil
}

// RespondToReview allows vendor to respond to a review
func (s *ReviewService) RespondToReview(reviewID, vendorID uint64, response string) error {
	var review models.Review
	err := s.db.Where("id = ? AND vendor_id = ?", reviewID, vendorID).
		First(&review).Error
	if err != nil {
		return fmt.Errorf("review not found: %w", err)
	}

	if !review.CanRespond(vendorID) {
		return fmt.Errorf("cannot respond to this review")
	}

	now := time.Now()
	err = s.db.Model(&review).Updates(map[string]interface{}{
		"vendor_response":   response,
		"vendor_response_at": now,
	}).Error
	if err != nil {
		return err
	}

	// Update vendor response rate
	go s.updateVendorRatings(vendorID)

	// Notify reviewer
	if s.wsHub != nil {
		s.wsHub.SendToUser(review.ReviewerID, "review_response", map[string]interface{}{
			"review_id": reviewID,
			"vendor_id": vendorID,
			"response":  response,
			"responded_at": now.Unix(),
		})
	}

	return nil
}

// VoteOnReview marks review as helpful or not
func (s *ReviewService) VoteOnReview(reviewID, userID uint64, isHelpful bool) error {
	// Check if user already voted
	var existingVote models.ReviewHelpful
	err := s.db.Where("review_id = ? AND user_id = ?", reviewID, userID).
		First(&existingVote).Error

	if err == nil {
		// Update existing vote
		if existingVote.IsHelpful != isHelpful {
			s.db.Model(&existingVote).Update("is_helpful", isHelpful)
			s.updateHelpfulCounts(reviewID)
		}
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	// Create new vote
	vote := &models.ReviewHelpful{
		ReviewID:  reviewID,
		UserID:    userID,
		IsHelpful: isHelpful,
	}

	if err := s.db.Create(vote).Error; err != nil {
		return err
	}

	s.updateHelpfulCounts(reviewID)
	return nil
}

// updateHelpfulCounts updates helpful/not helpful counts for a review
func (s *ReviewService) updateHelpfulCounts(reviewID uint64) {
	var helpfulCount, notHelpfulCount int64

	s.db.Model(&models.ReviewHelpful{}).
		Where("review_id = ? AND is_helpful = ?", reviewID, true).
		Count(&helpfulCount)

	s.db.Model(&models.ReviewHelpful{}).
		Where("review_id = ? AND is_helpful = ?", reviewID, false).
		Count(&notHelpfulCount)

	s.db.Model(&models.Review{}).
		Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"helpful_count":    helpfulCount,
			"not_helpful_count": notHelpfulCount,
		})
}

// ReportReview reports a review for inappropriate content
func (s *ReviewService) ReportReview(reviewID, reporterID uint64, reason, description string) error {
	report := &models.ReviewReport{
		ReviewID:   reviewID,
		ReporterID:  reporterID,
		Reason:     reason,
		Description: description,
		Status:     "pending",
	}

	return s.db.Create(report).Error
}

// ModerateReview moderates a review (admin action)
func (s *ReviewService) ModerateReview(reviewID, moderatorID uint64, status, notes string) error {
	var review models.Review
	err := s.db.Where("id = ?", reviewID).First(&review).Error
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"moderated_by":  moderatorID,
		"moderated_at":  now,
		"moderation_notes": notes,
	}

	err = s.db.Model(&review).Updates(updates).Error
	if err != nil {
		return err
	}

	// If review is rejected, update aggregated ratings
	if status == "rejected" {
		go s.updateVendorRatings(review.VendorID)
		go s.updateEquipmentRatings(review.EquipmentID)
	}

	return nil
}

// GetPendingModeration returns reviews pending moderation
func (s *ReviewService) GetPendingModeration(limit int) ([]models.Review, error) {
	var reviews []models.Review
	err := s.db.Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Find(&reviews).Error
	return reviews, err
}

// updateVendorRatings updates aggregated vendor ratings
func (s *ReviewService) updateVendorRatings(vendorID uint64) error {
	var stats struct {
		TotalRatings   int
		AverageRating  float32
		QualityRating  float32
		ServiceRating  float32
		ValueRating    float32
		FiveStarCount  int
		FourStarCount  int
		ThreeStarCount int
		TwoStarCount   int
		OneStarCount   int
		ResponseCount  int
		TotalReviews   int
	}

	// Get ratings from approved reviews
	err := s.db.Table("reviews").
		Select(`
			COUNT(*) as total_ratings,
			AVG(rating) as average_rating,
			AVG(COALESCE(quality_rating, rating)) as quality_rating,
			AVG(COALESCE(service_rating, rating)) as service_rating,
			AVG(COALESCE(value_rating, rating)) as value_rating,
			SUM(CASE WHEN rating >= 4.5 THEN 1 ELSE 0 END) as five_star_count,
			SUM(CASE WHEN rating >= 3.5 AND rating < 4.5 THEN 1 ELSE 0 END) as four_star_count,
			SUM(CASE WHEN rating >= 2.5 AND rating < 3.5 THEN 1 ELSE 0 END) as three_star_count,
			SUM(CASE WHEN rating >= 1.5 AND rating < 2.5 THEN 1 ELSE 0 END) as two_star_count,
			SUM(CASE WHEN rating < 1.5 THEN 1 ELSE 0 END) as one_star_count,
			COUNT(CASE WHEN vendor_response != '' THEN 1 END) as response_count
		`).
		Where("vendor_id = ? AND status = ?", vendorID, "approved").
		Scan(&stats).Error

	if err != nil {
		return err
	}

	// Update or create vendor rating
	var rating models.VendorRating
	err = s.db.Where("vendor_id = ?", vendorID).First(&rating).Error

	if err == gorm.ErrRecordNotFound {
		rating = models.VendorRating{
			VendorID:             vendorID,
			TotalReviews:         stats.TotalRatings,
			AverageRating:        float64(stats.AverageRating),
			EquipmentQualityAvg:  float64(stats.QualityRating),
			CommunicationAvg:     float64(stats.ServiceRating),
			ValueAvg:             float64(stats.ValueRating),
			AccuracyAvg:          0,
			Rating5Count:         stats.FiveStarCount,
			Rating4Count:         stats.FourStarCount,
			Rating3Count:         stats.ThreeStarCount,
			Rating2Count:         stats.TwoStarCount,
			Rating1Count:         stats.OneStarCount,
			VerifiedReviewCount:  0,
			RepeatCustomerCount:  0,
		}
		return s.db.Create(&rating).Error
	}

	updates := map[string]interface{}{
		"total_reviews":         stats.TotalRatings,
		"average_rating":        float64(stats.AverageRating),
		"equipment_quality_avg": float64(stats.QualityRating),
		"communication_avg":     float64(stats.ServiceRating),
		"value_avg":             float64(stats.ValueRating),
		"rating_5_count":        stats.FiveStarCount,
		"rating_4_count":        stats.FourStarCount,
		"rating_3_count":        stats.ThreeStarCount,
		"rating_2_count":        stats.TwoStarCount,
		"rating_1_count":        stats.OneStarCount,
	}

	return s.db.Model(&rating).Updates(updates).Error
}

// updateEquipmentRatings updates aggregated equipment ratings
func (s *ReviewService) updateEquipmentRatings(equipmentID uint64) error {
	var stats struct {
		TotalRatings   int
		AverageRating  float32
		QualityRating  float32
		ServiceRating  float32
		ValueRating    float32
		FiveStarCount  int
		FourStarCount  int
		ThreeStarCount int
		TwoStarCount   int
		OneStarCount   int
	}

	// Get ratings from approved reviews
	err := s.db.Table("reviews").
		Select(`
			COUNT(*) as total_ratings,
			AVG(rating) as average_rating,
			AVG(COALESCE(quality_rating, rating)) as quality_rating,
			AVG(COALESCE(service_rating, rating)) as service_rating,
			AVG(COALESCE(value_rating, rating)) as value_rating,
			SUM(CASE WHEN rating >= 4.5 THEN 1 ELSE 0 END) as five_star_count,
			SUM(CASE WHEN rating >= 3.5 AND rating < 4.5 THEN 1 ELSE 0 END) as four_star_count,
			SUM(CASE WHEN rating >= 2.5 AND rating < 3.5 THEN 1 ELSE 0 END) as three_star_count,
			SUM(CASE WHEN rating >= 1.5 AND rating < 2.5 THEN 1 ELSE 0 END) as two_star_count,
			SUM(CASE WHEN rating < 1.5 THEN 1 ELSE 0 END) as one_star_count
		`).
		Where("equipment_id = ? AND status = ?", equipmentID, "approved").
		Scan(&stats).Error

	if err != nil {
		return err
	}

	// Update or create equipment rating
	var rating models.EquipmentRating
	err = s.db.Where("equipment_id = ?", equipmentID).First(&rating).Error

	if err == gorm.ErrRecordNotFound {
		rating = models.EquipmentRating{
			EquipmentID:    equipmentID,
			TotalReviews:   stats.TotalRatings,
			AverageRating:  float64(stats.AverageRating),
			QualityAvg:     float64(stats.QualityRating),
			ValueAvg:       float64(stats.ValueRating),
			AccuracyAvg:    0,
			Rating5Count:   stats.FiveStarCount,
			Rating4Count:   stats.FourStarCount,
			Rating3Count:   stats.ThreeStarCount,
			Rating2Count:   stats.TwoStarCount,
			Rating1Count:   stats.OneStarCount,
		}
		return s.db.Create(&rating).Error
	}

	updates := map[string]interface{}{
		"total_reviews":   stats.TotalRatings,
		"average_rating":  float64(stats.AverageRating),
		"quality_avg":     float64(stats.QualityRating),
		"value_avg":       float64(stats.ValueRating),
		"rating_5_count":  stats.FiveStarCount,
		"rating_4_count":  stats.FourStarCount,
		"rating_3_count":  stats.ThreeStarCount,
		"rating_2_count":  stats.TwoStarCount,
		"rating_1_count":  stats.OneStarCount,
	}

	return s.db.Model(&rating).Updates(updates).Error
}

// GetVendorRating retrieves vendor rating summary
func (s *ReviewService) GetVendorRating(vendorID uint64) (*models.VendorRating, error) {
	var rating models.VendorRating
	err := s.db.Where("vendor_id = ?", vendorID).First(&rating).Error
	if err != nil {
		// Return default rating if not found
		return &models.VendorRating{
			VendorID:      vendorID,
			TotalReviews:  0,
			AverageRating: 0.0,
		}, nil
	}
	return &rating, nil
}

// GetEquipmentRating retrieves equipment rating summary
func (s *ReviewService) GetEquipmentRating(equipmentID uint64) (*models.EquipmentRating, error) {
	var rating models.EquipmentRating
	err := s.db.Where("equipment_id = ?", equipmentID).First(&rating).Error
	if err != nil {
		// Return default rating if not found
		return &models.EquipmentRating{
			EquipmentID:   equipmentID,
			TotalReviews:  0,
			AverageRating: 0.0,
		}, nil
	}
	return &rating, nil
}

// AddReviewImage adds an image to a review
func (s *ReviewService) AddReviewImage(reviewID uint64, imageURL, caption string, displayOrder int) error {
	image := &models.ReviewImage{
		ReviewID:     reviewID,
		ImageURL:     imageURL,
		Caption:      caption,
		DisplayOrder: displayOrder,
	}
	return s.db.Create(image).Error
}

// GetReviewImages retrieves images for a review
func (s *ReviewService) GetReviewImages(reviewID uint64) ([]models.ReviewImage, error) {
	var images []models.ReviewImage
	err := s.db.Where("review_id = ?", reviewID).
		Order("display_order ASC").
		Find(&images).Error
	return images, err
}

// GetTopRatedVendors retrieves top-rated vendors
func (s *ReviewService) GetTopRatedVendors(limit int, minRatings int) ([]models.VendorRating, error) {
	var ratings []models.VendorRating
	err := s.db.Where("total_ratings >= ?", minRatings).
		Order("average_rating DESC").
		Limit(limit).
		Find(&ratings).Error
	return ratings, err
}

// GetTopRatedEquipment retrieves top-rated equipment
func (s *ReviewService) GetTopRatedEquipment(limit int, minRatings int) ([]models.EquipmentRating, error) {
	var ratings []models.EquipmentRating
	err := s.db.Where("total_ratings >= ?", minRatings).
		Order("average_rating DESC").
		Limit(limit).
		Find(&ratings).Error
	return ratings, err
}

// VerifyReview marks a review as verified (confirmed by admin)
func (s *ReviewService) VerifyReview(reviewID uint64) error {
	now := time.Now()
	return s.db.Model(&models.Review{}).
		Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"is_verified": true,
			"verified_at": now,
		}).Error
}

// GetReviewSummary returns summary statistics for reviews
func (s *ReviewService) GetReviewSummary() (map[string]interface{}, error) {
	var summary struct {
		TotalReviews      int64
		PendingReviews   int64
		AverageRating     float64
		FiveStarPercent   float64
		VerifiedPercent   float64
		WithResponse      int64
	}

	err := s.db.Table("reviews").
		Select(`
			COUNT(*) as total_reviews,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_reviews,
			AVG(rating) as average_rating,
			(SUM(CASE WHEN rating >= 4.5 THEN 1 ELSE 0 END)::float / COUNT(*)) * 100 as five_star_percent,
			(SUM(CASE WHEN is_verified = true THEN 1 ELSE 0 END)::float / COUNT(*)) * 100 as verified_percent,
			SUM(CASE WHEN vendor_response != '' THEN 1 ELSE 0 END) as with_response
		`).
		Where("status = ?", "approved").
		Scan(&summary).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_reviews":     summary.TotalReviews,
		"pending_reviews":   summary.PendingReviews,
		"average_rating":     summary.AverageRating,
		"five_star_percent":  summary.FiveStarPercent,
		"verified_percent":   summary.VerifiedPercent,
		"with_response":      summary.WithResponse,
	}, nil
}

// GetReviewTrends returns review trends over time
func (s *ReviewService) GetReviewTrends(days int) ([]map[string]interface{}, error) {
	var trends []map[string]interface{}

	query := `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as review_count,
			AVG(rating) as avg_rating,
			COUNT(CASE WHEN rating >= 4 THEN 1 END) as positive_count,
			COUNT(CASE WHEN rating <= 2 THEN 1 END) as negative_count
		FROM reviews
		WHERE created_at > CURRENT_DATE - INTERVAL '? days'
			AND status = 'approved'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`

	err := s.db.Raw(query, days).Scan(&trends).Error
	return trends, err
}

// CreateRatingQuestion creates a custom rating question for a category
func (s *ReviewService) CreateRatingQuestion(question *models.RatingQuestion) error {
	return s.db.Create(question).Error
}

// GetRatingQuestions retrieves custom rating questions for a category
func (s *ReviewService) GetRatingQuestions(categoryID uint64) ([]models.RatingQuestion, error) {
	var questions []models.RatingQuestion
	err := s.db.Where("category_id = ? AND is_active = ?", categoryID, true).
		Order("display_order ASC").
		Find(&questions).Error
	return questions, err
}

// SaveRatingAnswer saves answers to custom rating questions
func (s *ReviewService) SaveRatingAnswer(reviewID, questionID uint64, answer float32) error {
	ratingAnswer := &models.RatingAnswer{
		ReviewID:   reviewID,
		QuestionID: questionID,
		Answer:     answer,
	}
	return s.db.Create(ratingAnswer).Error
}
