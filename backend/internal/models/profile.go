package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// StringArray is a custom type for handling string arrays in PostgreSQL
type ProfileStringArray []string

// Scan implements the sql.Scanner interface
func (sa *ProfileStringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = ProfileStringArray{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, sa)
	case string:
		return json.Unmarshal([]byte(v), sa)
	default:
		*sa = ProfileStringArray{}
		return nil
	}
}

// Value implements the driver.Valuer interface
func (sa ProfileStringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "[]", nil
	}
	return json.Marshal(sa)
}

// UserProfile represents extended user profile information
type UserProfile struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	UserID    uint64 `json:"user_id" gorm:"unique;not null;index"`

	// Basic Information
	FirstName       string  `json:"first_name,omitempty" gorm:"size:100"`
	LastName        string  `json:"last_name,omitempty" gorm:"size:100"`
	DisplayName     string  `json:"display_name,omitempty" gorm:"size:100"`
	Bio             string  `json:"bio,omitempty" gorm:"type:text"`
	ProfileImageURL string  `json:"profile_image_url,omitempty" gorm:"size:500"`
	CoverImageURL   string  `json:"cover_image_url,omitempty" gorm:"size:500"`

	// Location Information
	Country    string `json:"country,omitempty" gorm:"size:2"`    // ISO country code
	State      string `json:"state,omitempty" gorm:"size:100"`
	City       string `json:"city,omitempty" gorm:"size:100"`
	ZipCode    string `json:"zip_code,omitempty" gorm:"size:20"`
	Address    string `json:"address,omitempty" gorm:"type:text"`
	Latitude   *float64 `json:"latitude,omitempty"`
	Longitude  *float64 `json:"longitude,omitempty"`

	// Business Information (for vendors)
	CompanyName     string `json:"company_name,omitempty" gorm:"size:200"`
	BusinessType    string `json:"business_type,omitempty" gorm:"size:50"` // individual, company, partnership
	TaxID           string `json:"tax_id,omitempty" gorm:"size:50"`
	WebsiteURL      string `json:"website_url,omitempty" gorm:"size:255"`
	EstablishedYear int    `json:"established_year,omitempty"`

	// Social Media
	LinkedInURL    string `json:"linkedin_url,omitempty" gorm:"size:255"`
	TwitterURL     string `json:"twitter_url,omitempty" gorm:"size:255"`
	InstagramURL   string `json:"instagram_url,omitempty" gorm:"size:255"`
	FacebookURL    string `json:"facebook_url,omitempty" gorm:"size:255"`

	// Contact Preferences
	ContactEmail     string `json:"contact_email,omitempty" gorm:"size:255"`
	ContactPhone     string `json:"contact_phone,omitempty" gorm:"size:50"`
	PreferredContact string `json:"preferred_contact,omitempty" gorm:"size:20"` // email, phone, both

	// Settings
	Language      string `json:"language,omitempty" gorm:"size:10;default:en"`
	Timezone      string `json:"timezone,omitempty" gorm:"size:50;default:UTC"`
	Currency      string `json:"currency,omitempty" gorm:"size:3;default:USD"`
	DateFormat    string `json:"date_format,omitempty" gorm:"size:20;default:MM/DD/YYYY"`
	TimeFormat    string `json:"time_format,omitempty" gorm:"size:10;default:12h"`

	// Profile Completion
	CompletionPercentage int `json:"completion_percentage" gorm:"default:0"`
	LastCompletedAt      *time.Time `json:"last_completed_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User     *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserPreferences represents user preferences and settings
type UserPreferences struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	UserID    uint64 `json:"user_id" gorm:"unique;not null;index"`

	// Notification Preferences
	EmailNotifications     bool `json:"email_notifications" gorm:"default:true"`
	PushNotifications      bool `json:"push_notifications" gorm:"default:true"`
	SmsNotifications       bool `json:"sms_notifications" gorm:"default:false"`
	MarketingEmails        bool `json:"marketing_emails" gorm:"default:false"`
	ProductUpdates         bool `json:"product_updates" gorm:"default:true"`
	BookingReminders       bool `json:"booking_reminders" gorm:"default:true"`
	PromoOffers            bool `json:"promo_offers" gorm:"default:false"`
	ReviewNotifications    bool `json:"review_notifications" gorm:"default:true"`
	MessageNotifications   bool `json:"message_notifications" gorm:"default:true"`
	PaymentNotifications   bool `json:"payment_notifications" gorm:"default:true"`

	// Privacy Settings
	ProfileVisibility      string `json:"profile_visibility" gorm:"size:20;default:public"` // public, private, connections_only
	ShowContactInfo        bool   `json:"show_contact_info" gorm:"default:false"`
	ShowActivityStatus     bool   `json:"show_activity_status" gorm:"default:true"`
	AllowDirectMessages    bool   `json:"allow_direct_messages" gorm:"default:true"`
	ShowOnlineStatus       bool   `json:"show_online_status" gorm:"default:true"`

	// Search & Discovery
	SaveSearchHistory      bool `json:"save_search_history" gorm:"default:true"`
	PersonalizedRecommendations bool `json:"personalized_recommendations" gorm:"default:true"`
	LocationBasedServices  bool `json:"location_based_services" gorm:"default:false"`

	// Booking Preferences
	AutoAcceptBookings     bool `json:"auto_accept_bookings" gorm:"default:false"`
	RequireVerification    bool `json:"require_verification" gorm:"default:true"`
	MinimumBookingNotice   int  `json:"minimum_booking_notice" gorm:"default:24"` // hours
	MaximumBookingDuration int  `json:"maximum_booking_duration" gorm:"default:30"` // days
	CancellationPolicy     string `json:"cancellation_policy" gorm:"size:50;default:flexible"` // flexible, moderate, strict

	// Payment Preferences
	DefaultPaymentMethod  string `json:"default_payment_method,omitempty" gorm:"size:100"`
	CurrencyPreference    string `json:"currency_preference,omitempty" gorm:"size:3;default:USD"`
	RequireDeposit        bool   `json:"require_deposit" gorm:"default:false"`
	DepositPercentage     int    `json:"deposit_percentage" gorm:"default:0"`

	// Display Preferences
	Theme                  string `json:"theme" gorm:"size:20;default:light"` // light, dark, auto
	FontSize              string `json:"font_size" gorm:"size:20;default:medium"` // small, medium, large
	EnableAnimations      bool   `json:"enable_animations" gorm:"default:true"`
	HighContrastMode      bool   `json:"high_contrast_mode" gorm:"default:false"`

	// Communication Preferences
	PreferredLanguage     string `json:"preferred_language" gorm:"size:10;default:en"`
	Timezone              string `json:"timezone" gorm:"size:50;default:UTC"`

	// Timestamps
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User                  *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserVerification represents verification status and documents
type UserVerification struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	UserID    uint64 `json:"user_id" gorm:"unique;not null;index"`

	// Verification Status
	IsVerified           bool       `json:"is_verified" gorm:"default:false"`
	VerificationStatus   string     `json:"verification_status" gorm:"size:20;default:pending"` // pending, approved, rejected, review
	VerificationLevel    string     `json:"verification_level" gorm:"size:20;default:basic"` // basic, standard, premium
	VerifiedAt           *time.Time `json:"verified_at,omitempty"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	NextReviewDate       *time.Time `json:"next_review_date,omitempty"`

	// Identity Verification
	IdentityVerified      bool       `json:"identity_verified" gorm:"default:false"`
	IdentityDocumentType  string     `json:"identity_document_type,omitempty" gorm:"size:50"` // passport, driver_license, national_id
	IdentityDocumentNumber string    `json:"identity_document_number,omitempty" gorm:"size:100"`
	IdentityDocumentURL   string     `json:"identity_document_url,omitempty" gorm:"size:500"`
	IdentitySelfieURL     string     `json:"identity_selfie_url,omitempty" gorm:"size:500"`

	// Business Verification (for vendors)
	BusinessVerified      bool       `json:"business_verified" gorm:"default:false"`
	BusinessLicenseNumber string     `json:"business_license_number,omitempty" gorm:"size:100"`
	BusinessLicenseURL     string     `json:"business_license_url,omitempty" gorm:"size:500"`
	TaxDocumentURL         string     `json:"tax_document_url,omitempty" gorm:"size:500"`
	InsuranceDocumentURL   string     `json:"insurance_document_url,omitempty" gorm:"size:500"`

	// Address Verification
	AddressVerified       bool       `json:"address_verified" gorm:"default:false"`
	AddressProofType      string     `json:"address_proof_type,omitempty" gorm:"size:50"` // utility_bill, bank_statement
	AddressProofURL       string     `json:"address_proof_url,omitempty" gorm:"size:500"`

	// Additional Verifications
	EmailVerified         bool       `json:"email_verified" gorm:"default:false"`
	PhoneVerified         bool       `json:"phone_verified" gorm:"default:false"`
	SocialVerified        bool       `json:"social_verified" gorm:"default:false"`
	BankAccountVerified   bool       `json:"bank_account_verified" gorm:"default:false"`

	// Verification Metadata
	VerificationAttempts  int        `json:"verification_attempts" gorm:"default:0"`
	LastAttemptAt         *time.Time `json:"last_attempt_at,omitempty"`
	RejectionReason       string     `json:"rejection_reason,omitempty" gorm:"type:text"`
	RejectionDetails      string     `json:"rejection_details,omitempty" gorm:"type:text"`
	Notes                 string     `json:"notes,omitempty" gorm:"type:text"`

	// Admin Review
	ReviewedBy            *uint64    `json:"reviewed_by,omitempty"`
	ReviewedAt            *time.Time `json:"reviewed_at,omitempty"`
	AdminNotes            string     `json:"admin_notes,omitempty" gorm:"type:text"`

	// Trust Score
	TrustScore            float64    `json:"trust_score" gorm:"default:0"` // 0-100

	// Timestamps
	CreatedAt             time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User                  *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Reviewer              *User      `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

// UserActivity represents user activity tracking
type UserActivity struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	UserID    uint64 `json:"user_id" gorm:"not null;index"`
	ActivityType string `json:"activity_type" gorm:"size:50;not null;index"` // login, booking, review, search, etc.
	ActivityData  string `json:"activity_data,omitempty" gorm:"type:text"` // JSON data
	IPAddress     string `json:"ip_address,omitempty" gorm:"size:50"`
	UserAgent     string `json:"user_agent,omitempty" gorm:"size:500"`
	DeviceType    string `json:"device_type,omitempty" gorm:"size:50"` // mobile, desktop, tablet
	Referrer      string `json:"referrer,omitempty" gorm:"size:500"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime;index"`

	// Relationships
	User          *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserSession represents active user sessions
type UserSession struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	UserID    uint64 `json:"user_id" gorm:"not null;index"`
	SessionToken string `json:"-" gorm:"size:500;not null;unique"`
	RefreshToken string `json:"-" gorm:"size:500"`
	DeviceType   string `json:"device_type" gorm:"size:50"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty" gorm:"size:255"`
	IPAddress    string `json:"ip_address" gorm:"size:50"`
	UserAgent    string `json:"user_agent" gorm:"size:500"`
	LastActivity time.Time `json:"last_activity" gorm:"default:CURRENT_TIMESTAMP"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	User         *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserAchievement represents user achievements and badges
type UserAchievement struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	UserID    uint64 `json:"user_id" gorm:"not null;index"`
	AchievementType string `json:"achievement_type" gorm:"size:50;not null"` // first_booking, top_reviewer, trusted_vendor, etc.
	AchievementName string `json:"achievement_name" gorm:"size:100;not null"`
	Description     string `json:"description" gorm:"type:text"`
	IconURL         string `json:"icon_url,omitempty" gorm:"size:500"`
	BadgeLevel      string `json:"badge_level" gorm:"size:20;default:bronze"` // bronze, silver, gold, platinum
	Progress        float32 `json:"progress" gorm:"default:0"` // 0-100 for tiered achievements
	MaxProgress     float32 `json:"max_progress" gorm:"default:100"`
	UnlockedAt      *time.Time `json:"unlocked_at,omitempty"`
	IsDisplayed     bool   `json:"is_displayed" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	User            *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserProfileField represents custom profile fields
type UserProfileField struct {
	ID          uint64 `json:"id" gorm:"primaryKey"`
	FieldKey    string `json:"field_key" gorm:"size:50;not null;unique"`
	FieldLabel  string `json:"field_label" gorm:"size:100;not null"`
	FieldType   string `json:"field_type" gorm:"size:20;not null"` // text, select, multiselect, boolean, date, number
	Options     string `json:"options,omitempty" gorm:"type:text"` // JSON array for select/multiselect
	IsRequired  bool   `json:"is_required" gorm:"default:false"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	DisplayOrder int   `json:"display_order" gorm:"default:0"`
	Description string `json:"description,omitempty" gorm:"type:text"`
	HelpText    string `json:"help_text,omitempty" gorm:"type:text"`
	MinValue    *float64 `json:"min_value,omitempty"`
	MaxValue    *float64 `json:"max_value,omitempty"`
	Validation  string `json:"validation,omitempty" gorm:"type:text"` // JSON validation rules
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserProfileValue represents custom values for user profile fields
type UserProfileValue struct {
	ID          uint64 `json:"id" gorm:"primaryKey"`
	UserID      uint64 `json:"user_id" gorm:"not null;index"`
	FieldID     uint64 `json:"field_id" gorm:"not null;index"`
	Value       string `json:"value" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User        *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Field       *UserProfileField `json:"field,omitempty" gorm:"foreignKey:FieldID"`
}

// TableName specifications
func (UserProfile) TableName() string {
	return "user_profiles"
}

func (UserPreferences) TableName() string {
	return "user_preferences"
}

func (UserVerification) TableName() string {
	return "user_verifications"
}

func (UserActivity) TableName() string {
	return "user_activities"
}

func (UserSession) TableName() string {
	return "user_sessions"
}

func (UserAchievement) TableName() string {
	return "user_achievements"
}

func (UserProfileField) TableName() string {
	return "user_profile_fields"
}

func (UserProfileValue) TableName() string {
	return "user_profile_values"
}

// UserProfile methods

// IsComplete checks if profile is complete (90%+)
func (p *UserProfile) IsComplete() bool {
	return p.CompletionPercentage >= 90
}

// CalculateCompletion calculates profile completion percentage
func (p *UserProfile) CalculateCompletion() int {
	completed := 0
	total := 9

	// Check essential fields
	if p.FirstName != "" && p.LastName != "" {
		completed++
	}
	if p.Bio != "" {
		completed++
	}
	if p.ProfileImageURL != "" {
		completed++
	}
	if p.Country != "" && p.City != "" {
		completed++
	}
	if p.PhoneVerified() {
		completed++
	}
	if p.EmailVerified() {
		completed++
	}
	if p.PreferredContact != "" {
		completed++
	}
	if p.Language != "" {
		completed++
	}
	if p.Timezone != "" {
		completed++
	}

	percentage := (completed * 100) / total
	return percentage
}

// PhoneVerified checks if phone is verified
func (p *UserProfile) PhoneVerified() bool {
	// This would check against UserVerification
	return false
}

// EmailVerified checks if email is verified
func (p *UserProfile) EmailVerified() bool {
	// This would check against UserVerification
	return false
}

// GetFullName returns the user's full name
func (p *UserProfile) GetFullName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.FirstName + " " + p.LastName
}

// GetLocation returns formatted location string
func (p *UserProfile) GetLocation() string {
	location := ""
	if p.City != "" {
		location = p.City
	}
	if p.State != "" {
		if location != "" {
			location += ", "
		}
		location += p.State
	}
	if p.Country != "" {
		if location != "" {
			location += ", "
		}
		location += p.Country
	}
	return location
}

// UserPreferences methods

// GetNotificationSettings returns notification settings as map
func (p *UserPreferences) GetNotificationSettings() map[string]bool {
	return map[string]bool{
		"email":            p.EmailNotifications,
		"push":             p.PushNotifications,
		"sms":              p.SmsNotifications,
		"marketing":        p.MarketingEmails,
		"product_updates":  p.ProductUpdates,
		"booking_reminders": p.BookingReminders,
		"promo_offers":     p.PromoOffers,
		"reviews":          p.ReviewNotifications,
		"messages":         p.MessageNotifications,
		"payments":         p.PaymentNotifications,
	}
}

// IsPrivate checks if profile is private
func (p *UserPreferences) IsPrivate() bool {
	return p.ProfileVisibility == "private"
}

// IsPublic checks if profile is public
func (p *UserPreferences) IsPublic() bool {
	return p.ProfileVisibility == "public"
}

// UserVerification methods

// IsFullyVerified checks if user is fully verified
func (v *UserVerification) IsFullyVerified() bool {
	return v.IsVerified && v.IdentityVerified && v.EmailVerified && v.PhoneVerified
}

// IsBusinessVerified checks if business is verified
func (v *UserVerification) IsBusinessVerified() bool {
	return v.BusinessVerified && v.BusinessLicenseURL != ""
}

// GetVerificationProgress returns verification progress
func (v *UserVerification) GetVerificationProgress() int {
	steps := 6
	completed := 0

	if v.EmailVerified {
		completed++
	}
	if v.PhoneVerified {
		completed++
	}
	if v.IdentityVerified {
		completed++
	}
	if v.AddressVerified {
		completed++
	}
	if v.BusinessVerified {
		completed++
	}
	if v.IsVerified {
		completed++
	}

	return (completed * 100) / steps
}

// NeedsRenewal checks if verification needs renewal
func (v *UserVerification) NeedsRenewal() bool {
	if v.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*v.ExpiresAt)
}

// UserActivity methods

// IsRecent checks if activity is recent (within 24 hours)
func (a *UserActivity) IsRecent() bool {
	return time.Since(a.CreatedAt) < 24*time.Hour
}

// GetActivityData parses activity data JSON
func (a *UserActivity) GetActivityData() map[string]interface{} {
	var data map[string]interface{}
	if a.ActivityData != "" {
		json.Unmarshal([]byte(a.ActivityData), &data)
	}
	return data
}

// UserSession methods

// IsExpired checks if session is expired
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if session is valid and active
func (s *UserSession) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// UpdateActivity updates last activity timestamp
func (s *UserSession) UpdateActivity() {
	s.LastActivity = time.Now()
}

// UserAchievement methods

// IsUnlocked checks if achievement is unlocked
func (a *UserAchievement) IsUnlocked() bool {
	return a.UnlockedAt != nil
}

// GetProgressPercentage returns progress as percentage
func (a *UserAchievement) GetProgressPercentage() float32 {
	if a.MaxProgress == 0 {
		return 0
	}
	return (a.Progress / a.MaxProgress) * 100
}

// UserProfileValue methods

// GetTypedValue returns value as appropriate type
func (v *UserProfileValue) GetTypedValue() interface{} {
	// Implementation would depend on field type
	return v.Value
}
