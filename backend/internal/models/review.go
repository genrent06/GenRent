package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringArray is a custom type for handling string arrays in PostgreSQL
type StringArray []string

// Scan implements the sql.Scanner interface
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = StringArray{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, sa)
	case string:
		return json.Unmarshal([]byte(v), sa)
	default:
		*sa = StringArray{}
		return nil
	}
}

// Value implements the driver.Valuer interface
func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "[]", nil
	}
	return json.Marshal(sa)
}

// EquipmentReview represents a customer review for equipment
type EquipmentReview struct {
	ID        uint64      `json:"id" gorm:"primaryKey"`
	EquipmentID uint64     `json:"equipment_id" gorm:"not null;index:idx_equipment_reviews_equipment_id"`
	BookingID   uint64     `json:"booking_id,omitempty" gorm:"index"`
	CustomerID  uint64     `json:"customer_id" gorm:"not null;index:idx_equipment_reviews_customer_id"`
	VendorID    uint64     `json:"vendor_id" gorm:"not null;index:idx_equipment_reviews_vendor_id"`

	// Rating fields (1-5 scale)
	OverallRating         int16  `json:"overall_rating" gorm:"not null;check:overall_rating >= 1 AND overall_rating <= 5"`
	EquipmentQualityRating int16 `json:"equipment_quality_rating,omitempty" gorm:"check:equipment_quality_rating >= 1 AND equipment_quality_rating <= 5"`
	CommunicationRating    int16 `json:"communication_rating,omitempty" gorm:"check:communication_rating >= 1 AND communication_rating <= 5"`
	ValueRating           int16 `json:"value_rating,omitempty" gorm:"check:value_rating >= 1 AND value_rating <= 5"`
	AccuracyRating        int16 `json:"accuracy_rating,omitempty" gorm:"check:accuracy_rating >= 1 AND accuracy_rating <= 5"`

	// Review content
	Title   string     `json:"title,omitempty" gorm:"size:200"`
	Comment string     `json:"comment,omitempty" gorm:"type:text"`
	Pros    StringArray `json:"pros,omitempty" gorm:"type:text[]"`
	Cons    StringArray `json:"cons,omitempty" gorm:"type:text[]"`

	// Media attachments
	ImageURLs StringArray `json:"image_urls,omitempty" gorm:"type:text[]"`

	// Vendor response
	VendorResponse   string     `json:"vendor_response,omitempty" gorm:"type:text"`
	VendorResponseAt *time.Time `json:"vendor_response_at,omitempty"`

	// Moderation
	Status        string  `json:"status" gorm:"default:pending;size:20;index:idx_equipment_reviews_status;check:status IN ('pending', 'approved', 'rejected', 'flagged')"`
	FlaggedReason  string  `json:"flagged_reason,omitempty" gorm:"type:text"`
	ModeratedBy   *uint64 `json:"moderated_by,omitempty"`
	ModeratedAt   *time.Time `json:"moderated_at,omitempty"`

	// Helpful votes
	HelpfulCount    int `json:"helpful_count" gorm:"default:0"`
	NotHelpfulCount int `json:"not_helpful_count" gorm:"default:0"`

	// Verified purchase indicator
	VerifiedBooking bool `json:"verified_booking" gorm:"default:false;index:idx_equipment_reviews_verified_booking"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;index:idx_equipment_reviews_created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Equipment *Equipment      `json:"equipment,omitempty" gorm:"foreignKey:EquipmentID"`
	Customer  *User           `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Vendor    *User           `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	Booking   *Booking        `json:"booking,omitempty" gorm:"foreignKey:BookingID"`
	Moderator *User           `json:"moderator,omitempty" gorm:"foreignKey:ModeratedBy"`
}

// Review represents a user review for equipment, vendors, or bookings (legacy compatibility)
type Review struct {
	ID           uint64         `json:"id" gorm:"primaryKey"`
	BookingID    uint64         `json:"booking_id" gorm:"not null;index"`
	EquipmentID  uint64         `json:"equipment_id" gorm:"index"`
	VendorID     uint64         `json:"vendor_id" gorm:"index"`
	ReviewerID   uint64         `json:"reviewer_id" gorm:"not null;index"`
	ReviewerName string         `json:"reviewer_name" gorm:"size:100"`
	ReviewerAvatar string       `json:"reviewer_avatar" gorm:"size:255"`
	Rating       float32        `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Title        string         `json:"title" gorm:"size:200"`
	Comment      string         `json:"comment" gorm:"type:text"`
	Pros         string         `json:"pros" gorm:"type:text"`     // Positive feedback points
	Cons         string         `json:"cons" gorm:"type:text"`     // Negative feedback points
	// Rating breakdowns
	QualityRating      float32 `json:"quality_rating" gorm:"check:quality_rating >= 1 AND quality_rating <= 5"`
	ServiceRating      float32 `json:"service_rating" gorm:"check:service_rating >= 1 AND service_rating <= 5"`
	ValueRating        float32 `json:"value_rating" gorm:"check:value_rating >= 1 AND value_rating <= 5"`
	// Equipment condition
	EquipmentCondition string `json:"equipment_condition" gorm:"size:50"` // "excellent", "good", "fair", "poor"
	// Verification status
	IsVerified        bool       `json:"is_verified" gorm:"default:false"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	// Response from vendor
	VendorResponse     string     `json:"vendor_response" gorm:"type:text"`
	VendorResponseAt   *time.Time `json:"vendor_response_at,omitempty"`
	// Flagging
	IsFlagged          bool       `json:"is_flagged" gorm:"default:false"`
	FlaggedReason      string     `json:"flagged_reason" gorm:"type:text"`
	// Moderation
	Status            string     `json:"status" gorm:"default:pending;size:20"` // pending, approved, rejected, hidden
	ModeratedBy       *uint64    `json:"moderated_by,omitempty"`
	ModeratedAt       *time.Time `json:"moderated_at,omitempty"`
	ModerationNotes    string     `json:"moderation_notes" gorm:"type:text"`
	// Helpfulness
	HelpfulCount       int        `json:"helpful_count" gorm:"default:0"`
	NotHelpfulCount    int        `json:"not_helpful_count" gorm:"default:0"`
	// Timestamps
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// ReviewVote represents user votes on review helpfulness
type ReviewVote struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	ReviewID  uint64    `json:"review_id" gorm:"not null;index:idx_review_votes_review_id"`
	UserID    uint64    `json:"user_id" gorm:"not null;index:idx_review_votes_user_id"`
	VoteType  string    `json:"vote_type" gorm:"not null;size:10;check:vote_type IN ('helpful', 'not_helpful')"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Review *EquipmentReview `json:"review,omitempty" gorm:"foreignKey:ReviewID"`
	User   *User            `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ReviewImage represents images attached to reviews
type ReviewImage struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	ReviewID  uint64 `json:"review_id" gorm:"not null;index"`
	ImageURL  string `json:"image_url" gorm:"size:500;not null"`
	Caption   string `json:"caption" gorm:"size:255"`
	DisplayOrder int  `json:"display_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ReviewHelpful tracks user votes on review helpfulness (legacy compatibility)
type ReviewHelpful struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	ReviewID  uint64    `json:"review_id" gorm:"not null;index"`
	UserID    uint64    `json:"user_id" gorm:"not null;index"`
	IsHelpful bool      `json:"is_helpful" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// VendorRating represents aggregated vendor ratings
type VendorRating struct {
	ID              uint64    `json:"id" gorm:"primaryKey"`
	VendorID        uint64    `json:"vendor_id" gorm:"unique;not null;index"`

	// Rating statistics
	TotalReviews    int       `json:"total_reviews" gorm:"default:0"`
	AverageRating   float64   `json:"average_rating" gorm:"type:decimal(3,2);default:0"`

	// Category-specific ratings
	EquipmentQualityAvg float64 `json:"equipment_quality_avg" gorm:"type:decimal(3,2);default:0"`
	CommunicationAvg     float64 `json:"communication_avg" gorm:"type:decimal(3,2);default:0"`
	ValueAvg            float64 `json:"value_avg" gorm:"type:decimal(3,2);default:0"`
	AccuracyAvg         float64 `json:"accuracy_avg" gorm:"type:decimal(3,2);default:0"`

	// Rating distribution
	Rating1Count  int `json:"rating_1_count" gorm:"default:0"`
	Rating2Count  int `json:"rating_2_count" gorm:"default:0"`
	Rating3Count  int `json:"rating_3_count" gorm:"default:0"`
	Rating4Count  int `json:"rating_4_count" gorm:"default:0"`
	Rating5Count  int `json:"rating_5_count" gorm:"default:0"`

	// Trust indicators
	VerifiedReviewCount int `json:"verified_review_count" gorm:"default:0"`
	RepeatCustomerCount int `json:"repeat_customer_count" gorm:"default:0"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Vendor *User `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
}

// EquipmentRating represents aggregated equipment ratings
type EquipmentRating struct {
	ID              uint64    `json:"id" gorm:"primaryKey"`
	EquipmentID     uint64    `json:"equipment_id" gorm:"unique;not null;index"`

	// Rating statistics
	TotalReviews    int       `json:"total_reviews" gorm:"default:0"`
	AverageRating   float64   `json:"average_rating" gorm:"type:decimal(3,2);default:0"`

	// Category-specific averages
	QualityAvg float64 `json:"quality_avg" gorm:"type:decimal(3,2);default:0"`
	ValueAvg   float64 `json:"value_avg" gorm:"type:decimal(3,2);default:0"`
	AccuracyAvg float64 `json:"accuracy_avg" gorm:"type:decimal(3,2);default:0"`

	// Rating distribution
	Rating1Count  int `json:"rating_1_count" gorm:"default:0"`
	Rating2Count  int `json:"rating_2_count" gorm:"default:0"`
	Rating3Count  int `json:"rating_3_count" gorm:"default:0"`
	Rating4Count  int `json:"rating_4_count" gorm:"default:0"`
	Rating5Count  int `json:"rating_5_count" gorm:"default:0"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Equipment *Equipment `json:"equipment,omitempty" gorm:"foreignKey:EquipmentID"`
}

// ReviewReport represents user reports for inappropriate reviews
type ReviewReport struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	ReviewID   uint64    `json:"review_id" gorm:"not null;index"`
	ReporterID uint64    `json:"reporter_id" gorm:"not null;index"`
	Reason     string    `json:"reason" gorm:"size:100;not null"`
	Description string   `json:"description" gorm:"type:text"`
	Status     string    `json:"status" gorm:"default:pending;size:20"` // pending, investigating, resolved, dismissed
	ReviewedBy *uint64   `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	Notes      string    `json:"notes" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RatingQuestion represents custom rating questions for specific categories
type RatingQuestion struct {
	ID          uint64 `json:"id" gorm:"primaryKey"`
	CategoryID  uint64 `json:"category_id" gorm:"index"`
	Question    string `json:"question" gorm:"size:255;not null"`
	QuestionKey string `json:"question_key" gorm:"size:100;not null;unique"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	DisplayOrder int  `json:"display_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RatingAnswer represents answers to custom rating questions
type RatingAnswer struct {
	ID              uint64 `json:"id" gorm:"primaryKey"`
	ReviewID        uint64 `json:"review_id" gorm:"not null;index"`
	QuestionID      uint64 `json:"question_id" gorm:"not null;index"`
	Answer          float32 `json:"answer" gorm:"not null"` // 1-5 scale
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ReviewTemplate represents predefined response templates for vendors
type ReviewTemplate struct {
	ID          uint64 `json:"id" gorm:"primaryKey"`
	VendorID    uint64 `json:"vendor_id" gorm:"not null;index"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Template    string `json:"template" gorm:"type:text;not null"`
	IsDefault   bool   `json:"is_default" gorm:"default:false"`
	ResponseCount int  `json:"response_count" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifications
func (EquipmentReview) TableName() string {
	return "equipment_reviews"
}

func (Review) TableName() string {
	return "reviews"
}

func (ReviewImage) TableName() string {
	return "review_images"
}

func (ReviewVote) TableName() string {
	return "review_votes"
}

func (ReviewHelpful) TableName() string {
	return "review_helpful"
}

func (VendorRating) TableName() string {
	return "vendor_ratings"
}

func (EquipmentRating) TableName() string {
	return "equipment_ratings"
}

func (ReviewReport) TableName() string {
	return "review_reports"
}

func (RatingQuestion) TableName() string {
	return "rating_questions"
}

func (RatingAnswer) TableName() string {
	return "rating_answers"
}

func (ReviewTemplate) TableName() string {
	return "review_templates"
}

// BeforeCreate hook
func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}
	r.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook
func (r *Review) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

// IsEditable checks if a review can be edited
func (r *Review) IsEditable() bool {
	// Reviews can be edited within 7 days of creation
	return time.Since(r.CreatedAt) < 7*24*time.Hour
}

// CanRespond checks if vendor can respond to review
func (r *Review) CanRespond(vendorID uint64) bool {
	return r.VendorID == vendorID && r.VendorResponse == ""
}

// GetOverallRating calculates overall rating from sub-ratings
func (r *Review) GetOverallRating() float32 {
	if r.QualityRating == 0 && r.ServiceRating == 0 && r.ValueRating == 0 {
		return r.Rating
	}

	sum := r.QualityRating + r.ServiceRating + r.ValueRating
	return (sum + r.Rating) / 4.0
}

// EquipmentReview methods

// IsEditable checks if an equipment review can be edited (within 7 days of creation)
func (r *EquipmentReview) IsEditable() bool {
	return time.Since(r.CreatedAt) < 7*24*time.Hour
}

// CanRespond checks if vendor can respond to review
func (r *EquipmentReview) CanRespond(vendorID uint64) bool {
	return r.VendorID == vendorID && r.VendorResponse == ""
}

// CanVote checks if user can vote on review (not own review and hasn't voted yet)
func (r *EquipmentReview) CanVote(userID uint64) bool {
	return r.CustomerID != userID
}

// CalculateOverallRating calculates overall rating from sub-ratings
func (r *EquipmentReview) CalculateOverallRating() float32 {
	if r.EquipmentQualityRating == 0 && r.CommunicationRating == 0 &&
		r.ValueRating == 0 && r.AccuracyRating == 0 {
		return float32(r.OverallRating)
	}

	// Calculate average of all ratings including overall (convert int16 to float32)
	sum := float32(r.OverallRating) + float32(r.EquipmentQualityRating) +
		float32(r.CommunicationRating) + float32(r.ValueRating) + float32(r.AccuracyRating)

	count := 1
	if r.EquipmentQualityRating > 0 {
		count++
	}
	if r.CommunicationRating > 0 {
		count++
	}
	if r.ValueRating > 0 {
		count++
	}
	if r.AccuracyRating > 0 {
		count++
	}

	return sum / float32(count)
}

// GetHelpfulScore calculates the helpfulness score (-1 to 1)
func (r *EquipmentReview) GetHelpfulScore() float64 {
	total := r.HelpfulCount + r.NotHelpfulCount
	if total == 0 {
		return 0
	}
	return float64(r.HelpfulCount-r.NotHelpfulCount) / float64(total)
}

// VendorRating methods

// GetRatingPercentage returns percentage for given star rating
func (vr *VendorRating) GetRatingPercentage(stars int) float64 {
	if vr.TotalReviews == 0 {
		return 0
	}

	var count int
	switch stars {
	case 1:
		count = vr.Rating1Count
	case 2:
		count = vr.Rating2Count
	case 3:
		count = vr.Rating3Count
	case 4:
		count = vr.Rating4Count
	case 5:
		count = vr.Rating5Count
	default:
		return 0
	}

	return float64(count) / float64(vr.TotalReviews) * 100
}

// GetVerifiedPercentage returns percentage of verified reviews
func (vr *VendorRating) GetVerifiedPercentage() float64 {
	if vr.TotalReviews == 0 {
		return 0
	}
	return float64(vr.VerifiedReviewCount) / float64(vr.TotalReviews) * 100
}

// EquipmentRating methods

// GetRatingPercentage returns percentage for given star rating
func (er *EquipmentRating) GetRatingPercentage(stars int) float64 {
	if er.TotalReviews == 0 {
		return 0
	}

	var count int
	switch stars {
	case 1:
		count = er.Rating1Count
	case 2:
		count = er.Rating2Count
	case 3:
		count = er.Rating3Count
	case 4:
		count = er.Rating4Count
	case 5:
		count = er.Rating5Count
	default:
		return 0
	}

	return float64(count) / float64(er.TotalReviews) * 100
}

// ReviewVote methods

// IsHelpful checks if vote is helpful
func (rv *ReviewVote) IsHelpful() bool {
	return rv.VoteType == "helpful"
}

// ReviewReport methods

// IsPending checks if report is pending review
func (rr *ReviewReport) IsPending() bool {
	return rr.Status == "pending"
}

// IsResolved checks if report has been resolved
func (rr *ReviewReport) IsResolved() bool {
	return rr.Status == "resolved" || rr.Status == "dismissed"
}