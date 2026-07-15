package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"genrent/internal/models"
	"gorm.io/gorm"
)

// Service handles profile-related business logic
type Service struct {
	db *gorm.DB
}

// NewService creates a new profile service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// UpdateProfileRequest represents the request to update user profile
type UpdateProfileRequest struct {
	FirstName       *string  `json:"first_name,omitempty"`
	LastName        *string  `json:"last_name,omitempty"`
	DisplayName     *string  `json:"display_name,omitempty"`
	Bio             *string  `json:"bio,omitempty"`
	ProfileImageURL *string  `json:"profile_image_url,omitempty"`
	CoverImageURL   *string  `json:"cover_image_url,omitempty"`
	Country         *string  `json:"country,omitempty"`
	State           *string  `json:"state,omitempty"`
	City            *string  `json:"city,omitempty"`
	ZipCode         *string  `json:"zip_code,omitempty"`
	Address         *string  `json:"address,omitempty"`
	Latitude        *float64 `json:"latitude,omitempty"`
	Longitude       *float64 `json:"longitude,omitempty"`
	CompanyName     *string  `json:"company_name,omitempty"`
	BusinessType    *string  `json:"business_type,omitempty"`
	TaxID           *string  `json:"tax_id,omitempty"`
	WebsiteURL      *string  `json:"website_url,omitempty"`
	EstablishedYear *int     `json:"established_year,omitempty"`
	LinkedInURL     *string  `json:"linkedin_url,omitempty"`
	TwitterURL      *string  `json:"twitter_url,omitempty"`
	InstagramURL    *string  `json:"instagram_url,omitempty"`
	FacebookURL     *string  `json:"facebook_url,omitempty"`
	ContactEmail    *string  `json:"contact_email,omitempty"`
	ContactPhone    *string  `json:"contact_phone,omitempty"`
	PreferredContact *string `json:"preferred_contact,omitempty"`
	Language        *string  `json:"language,omitempty"`
	Timezone        *string  `json:"timezone,omitempty"`
	Currency        *string  `json:"currency,omitempty"`
}

// UpdatePreferencesRequest represents the request to update user preferences
type UpdatePreferencesRequest struct {
	EmailNotifications          *bool   `json:"email_notifications,omitempty"`
	PushNotifications           *bool   `json:"push_notifications,omitempty"`
	SmsNotifications            *bool   `json:"sms_notifications,omitempty"`
	MarketingEmails             *bool   `json:"marketing_emails,omitempty"`
	ProductUpdates              *bool   `json:"product_updates,omitempty"`
	BookingReminders            *bool   `json:"booking_reminders,omitempty"`
	PromoOffers                 *bool   `json:"promo_offers,omitempty"`
	ReviewNotifications         *bool   `json:"review_notifications,omitempty"`
	MessageNotifications        *bool   `json:"message_notifications,omitempty"`
	PaymentNotifications        *bool   `json:"payment_notifications,omitempty"`
	ProfileVisibility           *string `json:"profile_visibility,omitempty"`
	ShowContactInfo             *bool   `json:"show_contact_info,omitempty"`
	ShowActivityStatus          *bool   `json:"show_activity_status,omitempty"`
	AllowDirectMessages         *bool   `json:"allow_direct_messages,omitempty"`
	ShowOnlineStatus            *bool   `json:"show_online_status,omitempty"`
	SaveSearchHistory           *bool   `json:"save_search_history,omitempty"`
	PersonalizedRecommendations *bool   `json:"personalized_recommendations,omitempty"`
	LocationBasedServices       *bool   `json:"location_based_services,omitempty"`
	AutoAcceptBookings          *bool   `json:"auto_accept_bookings,omitempty"`
	RequireVerification         *bool   `json:"require_verification,omitempty"`
	MinimumBookingNotice        *int    `json:"minimum_booking_notice,omitempty"`
	MaximumBookingDuration      *int    `json:"maximum_booking_duration,omitempty"`
	CancellationPolicy          *string `json:"cancellation_policy,omitempty"`
	DefaultPaymentMethod        *string `json:"default_payment_method,omitempty"`
	CurrencyPreference          *string `json:"currency_preference,omitempty"`
	RequireDeposit              *bool   `json:"require_deposit,omitempty"`
	DepositPercentage           *int    `json:"deposit_percentage,omitempty"`
	Theme                       *string `json:"theme,omitempty"`
	FontSize                    *string `json:"font_size,omitempty"`
	EnableAnimations            *bool   `json:"enable_animations,omitempty"`
	HighContrastMode            *bool   `json:"high_contrast_mode,omitempty"`
	PreferredLanguage           *string `json:"preferred_language,omitempty"`
	Timezone                    *string `json:"timezone,omitempty"`
}

// UserProfile methods

// GetOrCreateProfile retrieves or creates a user profile
func (s *Service) GetOrCreateProfile(ctx context.Context, userID uint64) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error

	if err == nil {
		return &profile, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new profile with default values
		profile = models.UserProfile{
			UserID:              userID,
			Language:            "en",
			Timezone:            "UTC",
			Currency:            "USD",
			DateFormat:          "MM/DD/YYYY",
			TimeFormat:          "12h",
			CompletionPercentage: 0,
		}

		if err := s.db.WithContext(ctx).Create(&profile).Error; err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}

		return &profile, nil
	}

	return nil, fmt.Errorf("failed to get profile: %w", err)
}

// UpdateProfile updates user profile information
func (s *Service) UpdateProfile(ctx context.Context, userID uint64, req UpdateProfileRequest) (*models.UserProfile, error) {
	profile, err := s.GetOrCreateProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.ProfileImageURL != nil {
		updates["profile_image_url"] = *req.ProfileImageURL
	}
	if req.CoverImageURL != nil {
		updates["cover_image_url"] = *req.CoverImageURL
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.State != nil {
		updates["state"] = *req.State
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.ZipCode != nil {
		updates["zip_code"] = *req.ZipCode
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Latitude != nil {
		updates["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		updates["longitude"] = *req.Longitude
	}
	if req.CompanyName != nil {
		updates["company_name"] = *req.CompanyName
	}
	if req.BusinessType != nil {
		updates["business_type"] = *req.BusinessType
	}
	if req.TaxID != nil {
		updates["tax_id"] = *req.TaxID
	}
	if req.WebsiteURL != nil {
		updates["website_url"] = *req.WebsiteURL
	}
	if req.EstablishedYear != nil {
		updates["established_year"] = *req.EstablishedYear
	}
	if req.LinkedInURL != nil {
		updates["linkedin_url"] = *req.LinkedInURL
	}
	if req.TwitterURL != nil {
		updates["twitter_url"] = *req.TwitterURL
	}
	if req.InstagramURL != nil {
		updates["instagram_url"] = *req.InstagramURL
	}
	if req.FacebookURL != nil {
		updates["facebook_url"] = *req.FacebookURL
	}
	if req.ContactEmail != nil {
		updates["contact_email"] = *req.ContactEmail
	}
	if req.ContactPhone != nil {
		updates["contact_phone"] = *req.ContactPhone
	}
	if req.PreferredContact != nil {
		updates["preferred_contact"] = *req.PreferredContact
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(profile).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update profile: %w", err)
		}

		// Recalculate completion percentage
		if err := s.UpdateProfileCompletion(ctx, userID); err != nil {
			// Log error but don't fail the update
			fmt.Printf("Failed to update profile completion: %v", err)
		}
	}

	// Reload profile
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, fmt.Errorf("failed to reload profile: %w", err)
	}

	return profile, nil
}

// UpdateProfileCompletion calculates and updates profile completion percentage
func (s *Service) UpdateProfileCompletion(ctx context.Context, userID uint64) error {
	profile, err := s.GetOrCreateProfile(ctx, userID)
	if err != nil {
		return err
	}

	completion := profile.CalculateCompletion()
	now := time.Now()

	return s.db.WithContext(ctx).Model(profile).Updates(map[string]interface{}{
		"completion_percentage": completion,
		"last_completed_at":    &now,
	}).Error
}

// GetProfile retrieves a user profile by user ID
func (s *Service) GetProfile(ctx context.Context, userID uint64) (*models.UserProfile, error) {
	profile, err := s.GetOrCreateProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

// GetPublicProfile retrieves a user's public profile information
func (s *Service) GetPublicProfile(ctx context.Context, userID uint64) (map[string]interface{}, error) {
	var profile models.UserProfile
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("profile not found")
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Get user preferences to check privacy settings
	var preferences models.UserPreferences
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&preferences).Error; err != nil {
		// If no preferences, use defaults
		preferences = models.UserPreferences{
			ProfileVisibility: "public",
		}
	}

	// Build public profile based on privacy settings
	publicProfile := make(map[string]interface{})

	// Always include basic info
	publicProfile["user_id"] = profile.UserID
	publicProfile["display_name"] = profile.DisplayName
	if publicProfile["display_name"] == "" {
		publicProfile["display_name"] = profile.FirstName + " " + profile.LastName
	}
	publicProfile["profile_image_url"] = profile.ProfileImageURL
	publicProfile["bio"] = profile.Bio

	// Include location if public
	if preferences.ProfileVisibility != "private" {
		publicProfile["location"] = profile.GetLocation()
	}

	// Include contact info if allowed
	if preferences.ShowContactInfo {
		publicProfile["contact_email"] = profile.ContactEmail
		publicProfile["contact_phone"] = profile.ContactPhone
	}

	// Include business info for vendors
	if profile.CompanyName != "" {
		publicProfile["company_name"] = profile.CompanyName
		publicProfile["business_type"] = profile.BusinessType
	}

	// Include verification status
	var verification models.UserVerification
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&verification).Error; err == nil {
		publicProfile["is_verified"] = verification.IsVerified
		publicProfile["verification_level"] = verification.VerificationLevel
	}

	// Include completion status
	publicProfile["profile_complete"] = profile.IsComplete()

	return publicProfile, nil
}

// UserPreferences methods

// GetOrCreatePreferences retrieves or creates user preferences
func (s *Service) GetOrCreatePreferences(ctx context.Context, userID uint64) (*models.UserPreferences, error) {
	var preferences models.UserPreferences
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&preferences).Error

	if err == nil {
		return &preferences, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new preferences with default values
		preferences = models.UserPreferences{
			UserID:                      userID,
			EmailNotifications:          true,
			PushNotifications:           true,
			SmsNotifications:            false,
			MarketingEmails:             false,
			ProductUpdates:              true,
			BookingReminders:            true,
			PromoOffers:                 false,
			ReviewNotifications:         true,
			MessageNotifications:        true,
			PaymentNotifications:        true,
			ProfileVisibility:           "public",
			ShowContactInfo:             false,
			ShowActivityStatus:          true,
			AllowDirectMessages:         true,
			ShowOnlineStatus:            true,
			SaveSearchHistory:           true,
			PersonalizedRecommendations:  true,
			LocationBasedServices:       false,
			AutoAcceptBookings:          false,
			RequireVerification:         true,
			MinimumBookingNotice:        24,
			MaximumBookingDuration:      30,
			CancellationPolicy:          "flexible",
			CurrencyPreference:          "USD",
			RequireDeposit:              false,
			DepositPercentage:           0,
			Theme:                       "light",
			FontSize:                    "medium",
			EnableAnimations:            true,
			HighContrastMode:            false,
			PreferredLanguage:           "en",
			Timezone:                    "UTC",
		}

		if err := s.db.WithContext(ctx).Create(&preferences).Error; err != nil {
			return nil, fmt.Errorf("failed to create preferences: %w", err)
		}

		return &preferences, nil
	}

	return nil, fmt.Errorf("failed to get preferences: %w", err)
}

// UpdatePreferences updates user preferences
func (s *Service) UpdatePreferences(ctx context.Context, userID uint64, req UpdatePreferencesRequest) (*models.UserPreferences, error) {
	preferences, err := s.GetOrCreatePreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.EmailNotifications != nil {
		updates["email_notifications"] = *req.EmailNotifications
	}
	if req.PushNotifications != nil {
		updates["push_notifications"] = *req.PushNotifications
	}
	if req.SmsNotifications != nil {
		updates["sms_notifications"] = *req.SmsNotifications
	}
	if req.MarketingEmails != nil {
		updates["marketing_emails"] = *req.MarketingEmails
	}
	if req.ProductUpdates != nil {
		updates["product_updates"] = *req.ProductUpdates
	}
	if req.BookingReminders != nil {
		updates["booking_reminders"] = *req.BookingReminders
	}
	if req.PromoOffers != nil {
		updates["promo_offers"] = *req.PromoOffers
	}
	if req.ReviewNotifications != nil {
		updates["review_notifications"] = *req.ReviewNotifications
	}
	if req.MessageNotifications != nil {
		updates["message_notifications"] = *req.MessageNotifications
	}
	if req.PaymentNotifications != nil {
		updates["payment_notifications"] = *req.PaymentNotifications
	}
	if req.ProfileVisibility != nil {
		updates["profile_visibility"] = *req.ProfileVisibility
	}
	if req.ShowContactInfo != nil {
		updates["show_contact_info"] = *req.ShowContactInfo
	}
	if req.ShowActivityStatus != nil {
		updates["show_activity_status"] = *req.ShowActivityStatus
	}
	if req.AllowDirectMessages != nil {
		updates["allow_direct_messages"] = *req.AllowDirectMessages
	}
	if req.ShowOnlineStatus != nil {
		updates["show_online_status"] = *req.ShowOnlineStatus
	}
	if req.SaveSearchHistory != nil {
		updates["save_search_history"] = *req.SaveSearchHistory
	}
	if req.PersonalizedRecommendations != nil {
		updates["personalized_recommendations"] = *req.PersonalizedRecommendations
	}
	if req.LocationBasedServices != nil {
		updates["location_based_services"] = *req.LocationBasedServices
	}
	if req.AutoAcceptBookings != nil {
		updates["auto_accept_bookings"] = *req.AutoAcceptBookings
	}
	if req.RequireVerification != nil {
		updates["require_verification"] = *req.RequireVerification
	}
	if req.MinimumBookingNotice != nil {
		updates["minimum_booking_notice"] = *req.MinimumBookingNotice
	}
	if req.MaximumBookingDuration != nil {
		updates["maximum_booking_duration"] = *req.MaximumBookingDuration
	}
	if req.CancellationPolicy != nil {
		updates["cancellation_policy"] = *req.CancellationPolicy
	}
	if req.DefaultPaymentMethod != nil {
		updates["default_payment_method"] = *req.DefaultPaymentMethod
	}
	if req.CurrencyPreference != nil {
		updates["currency_preference"] = *req.CurrencyPreference
	}
	if req.RequireDeposit != nil {
		updates["require_deposit"] = *req.RequireDeposit
	}
	if req.DepositPercentage != nil {
		updates["deposit_percentage"] = *req.DepositPercentage
	}
	if req.Theme != nil {
		updates["theme"] = *req.Theme
	}
	if req.FontSize != nil {
		updates["font_size"] = *req.FontSize
	}
	if req.EnableAnimations != nil {
		updates["enable_animations"] = *req.EnableAnimations
	}
	if req.HighContrastMode != nil {
		updates["high_contrast_mode"] = *req.HighContrastMode
	}
	if req.PreferredLanguage != nil {
		updates["preferred_language"] = *req.PreferredLanguage
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(preferences).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update preferences: %w", err)
		}
	}

	// Reload preferences
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&preferences).Error; err != nil {
		return nil, fmt.Errorf("failed to reload preferences: %w", err)
	}

	return preferences, nil
}

// GetPreferences retrieves user preferences
func (s *Service) GetPreferences(ctx context.Context, userID uint64) (*models.UserPreferences, error) {
	return s.GetOrCreatePreferences(ctx, userID)
}

// UserVerification methods

// GetOrCreateVerification retrieves or creates user verification record
func (s *Service) GetOrCreateVerification(ctx context.Context, userID uint64) (*models.UserVerification, error) {
	var verification models.UserVerification
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&verification).Error

	if err == nil {
		return &verification, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new verification record with default values
		verification = models.UserVerification{
			UserID:            userID,
			IsVerified:        false,
			VerificationStatus: "pending",
			VerificationLevel: "basic",
			TrustScore:        0,
		}

		if err := s.db.WithContext(ctx).Create(&verification).Error; err != nil {
			return nil, fmt.Errorf("failed to create verification record: %w", err)
		}

		return &verification, nil
	}

	return nil, fmt.Errorf("failed to get verification: %w", err)
}

// SubmitIdentityVerification submits identity verification documents
func (s *Service) SubmitIdentityVerification(ctx context.Context, userID uint64, documentType, documentNumber, documentURL, selfieURL string) error {
	verification, err := s.GetOrCreateVerification(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"identity_document_type":   documentType,
		"identity_document_number":  documentNumber,
		"identity_document_url":     documentURL,
		"identity_selfie_url":       selfieURL,
		"verification_attempts":     verification.VerificationAttempts + 1,
		"last_attempt_at":           &now,
		"verification_status":      "pending",
	}

	if err := s.db.WithContext(ctx).Model(verification).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to submit identity verification: %w", err)
	}

	return nil
}

// SubmitBusinessVerification submits business verification documents
func (s *Service) SubmitBusinessVerification(ctx context.Context, userID uint64, licenseNumber, licenseURL, taxURL, insuranceURL string) error {
	verification, err := s.GetOrCreateVerification(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"business_license_number": licenseNumber,
		"business_license_url":    licenseURL,
		"tax_document_url":        taxURL,
		"insurance_document_url":  insuranceURL,
		"verification_attempts":   verification.VerificationAttempts + 1,
		"last_attempt_at":         &now,
		"verification_status":    "pending",
	}

	if err := s.db.WithContext(ctx).Model(verification).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to submit business verification: %w", err)
	}

	return nil
}

// VerifyEmail marks email as verified
func (s *Service) VerifyEmail(ctx context.Context, userID uint64) error {
	return s.db.WithContext(ctx).Model(&models.UserVerification{}).
		Where("user_id = ?", userID).
		Update("email_verified", true).Error
}

// VerifyPhone marks phone as verified
func (s *Service) VerifyPhone(ctx context.Context, userID uint64) error {
	return s.db.WithContext(ctx).Model(&models.UserVerification{}).
		Where("user_id = ?", userID).
		Update("phone_verified", true).Error
}

// ApproveVerification approves user verification (admin only)
func (s *Service) ApproveVerification(ctx context.Context, userID uint64, moderatorID uint64, level string) error {
	verification, err := s.GetOrCreateVerification(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	expiresAt := now.AddDate(1, 0, 0) // 1 year expiration

	updates := map[string]interface{}{
		"is_verified":        true,
		"verification_status": "approved",
		"verification_level":  level,
		"verified_at":        &now,
		"expires_at":         &expiresAt,
		"reviewed_by":        moderatorID,
		"reviewed_at":        &now,
		"rejection_reason":   nil,
		"rejection_details":  nil,
	}

	if err := s.db.WithContext(ctx).Model(verification).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to approve verification: %w", err)
	}

	return nil
}

// RejectVerification rejects user verification (admin only)
func (s *Service) RejectVerification(ctx context.Context, userID uint64, moderatorID uint64, reason, details string) error {
	verification, err := s.GetOrCreateVerification(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"is_verified":        false,
		"verification_status": "rejected",
		"reviewed_by":        moderatorID,
		"reviewed_at":        &now,
		"rejection_reason":   reason,
		"rejection_details":  details,
	}

	if err := s.db.WithContext(ctx).Model(verification).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to reject verification: %w", err)
	}

	return nil
}

// GetVerification retrieves user verification status
func (s *Service) GetVerification(ctx context.Context, userID uint64) (*models.UserVerification, error) {
	return s.GetOrCreateVerification(ctx, userID)
}

// UserActivity methods

// LogActivity logs a user activity
func (s *Service) LogActivity(ctx context.Context, userID uint64, activityType string, activityData map[string]interface{}, ipAddress, userAgent, deviceType, referrer string) error {
	var dataJSON string
	if activityData != nil {
		dataBytes, err := json.Marshal(activityData)
		if err != nil {
			return fmt.Errorf("failed to marshal activity data: %w", err)
		}
		dataJSON = string(dataBytes)
	}

	activity := &models.UserActivity{
		UserID:       userID,
		ActivityType: activityType,
		ActivityData: dataJSON,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		DeviceType:   deviceType,
		Referrer:     referrer,
	}

	if err := s.db.WithContext(ctx).Create(activity).Error; err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}

// GetUserActivities retrieves user activities with pagination
func (s *Service) GetUserActivities(ctx context.Context, userID uint64, activityType string, limit int) ([]models.UserActivity, error) {
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)

	if activityType != "" {
		query = query.Where("activity_type = ?", activityType)
	}

	var activities []models.UserActivity
	if err := query.Order("created_at DESC").Limit(limit).Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, nil
}

// CleanupOldActivities removes activities older than specified days
func (s *Service) CleanupOldActivities(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	return s.db.WithContext(ctx).Where("created_at < ?", cutoffDate).Delete(&models.UserActivity{}).Error
}

// UserAchievement methods

// AwardAchievement awards an achievement to a user
func (s *Service) AwardAchievement(ctx context.Context, userID uint64, achievementType, achievementName, description, iconURL, badgeLevel string) error {
	var achievement models.UserAchievement
	err := s.db.WithContext(ctx).Where("user_id = ? AND achievement_type = ?", userID, achievementType).
		First(&achievement).Error

	now := time.Now()

	if err == nil {
		// Update existing achievement
		if achievement.UnlockedAt == nil {
			// Unlock it
			if err := s.db.WithContext(ctx).Model(&achievement).Updates(map[string]interface{}{
				"unlocked_at": &now,
				"progress":    100,
			}).Error; err != nil {
				return fmt.Errorf("failed to unlock achievement: %w", err)
			}
		}
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new achievement
		achievement = models.UserAchievement{
			UserID:          userID,
			AchievementType: achievementType,
			AchievementName: achievementName,
			Description:     description,
			IconURL:         iconURL,
			BadgeLevel:      badgeLevel,
			Progress:        100,
			MaxProgress:     100,
			UnlockedAt:      &now,
			IsDisplayed:     true,
		}

		if err := s.db.WithContext(ctx).Create(&achievement).Error; err != nil {
			return fmt.Errorf("failed to create achievement: %w", err)
		}
		return nil
	}

	return fmt.Errorf("failed to award achievement: %w", err)
}

// GetUserAchievements retrieves user achievements
func (s *Service) GetUserAchievements(ctx context.Context, userID uint64) ([]models.UserAchievement, error) {
	var achievements []models.UserAchievement
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("unlocked_at DESC").Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}

	return achievements, nil
}

// UpdateAchievementProgress updates progress for a tiered achievement
func (s *Service) UpdateAchievementProgress(ctx context.Context, userID uint64, achievementType string, progress float32) error {
	var achievement models.UserAchievement
	err := s.db.WithContext(ctx).Where("user_id = ? AND achievement_type = ?", userID, achievementType).
		First(&achievement).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Achievement doesn't exist yet, this shouldn't happen normally
			return fmt.Errorf("achievement not found")
		}
		return fmt.Errorf("failed to find achievement: %w", err)
	}

	updates := map[string]interface{}{
		"progress": progress,
	}

	if progress >= 100 && achievement.UnlockedAt == nil {
		now := time.Now()
		updates["unlocked_at"] = &now
	}

	if err := s.db.WithContext(ctx).Model(&achievement).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update achievement progress: %w", err)
	}

	return nil
}

// GetProfileSummary returns a comprehensive summary of user profile data
func (s *Service) GetProfileSummary(ctx context.Context, userID uint64) (map[string]interface{}, error) {
	profile, err := s.GetOrCreateProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	preferences, err := s.GetOrCreatePreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	verification, err := s.GetOrCreateVerification(ctx, userID)
	if err != nil {
		return nil, err
	}

	var achievementCount int64
	s.db.WithContext(ctx).Model(&models.UserAchievement{}).
		Where("user_id = ? AND unlocked_at IS NOT NULL", userID).Count(&achievementCount)

	var recentActivities []models.UserActivity
	s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").Limit(5).Find(&recentActivities)

	return map[string]interface{}{
		"profile":      profile,
		"preferences":  preferences,
		"verification": verification,
		"stats": map[string]interface{}{
			"achievements_count": achievementCount,
			"profile_complete":    profile.IsComplete(),
			"is_verified":        verification.IsVerified,
			"verification_level": verification.VerificationLevel,
		},
		"recent_activities": recentActivities,
	}, nil
}

// CheckProfileComplete checks if user profile is complete
func (s *Service) CheckProfileComplete(ctx context.Context, userID uint64) (bool, error) {
	profile, err := s.GetOrCreateProfile(ctx, userID)
	if err != nil {
		return false, err
	}

	return profile.IsComplete(), nil
}

// GetIncompleteProfileFields returns list of incomplete profile fields
func (s *Service) GetIncompleteProfileFields(ctx context.Context, userID uint64) ([]string, error) {
	profile, err := s.GetOrCreateProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	var incompleteFields []string

	if profile.FirstName == "" || profile.LastName == "" {
		incompleteFields = append(incompleteFields, "name")
	}
	if profile.Bio == "" {
		incompleteFields = append(incompleteFields, "bio")
	}
	if profile.ProfileImageURL == "" {
		incompleteFields = append(incompleteFields, "profile_image")
	}
	if profile.Country == "" || profile.City == "" {
		incompleteFields = append(incompleteFields, "location")
	}

	// Check verification
	verification, err := s.GetOrCreateVerification(ctx, userID)
	if err == nil && !verification.EmailVerified {
		incompleteFields = append(incompleteFields, "email_verification")
	}
	if err == nil && !verification.PhoneVerified {
		incompleteFields = append(incompleteFields, "phone_verification")
	}

	return incompleteFields, nil
}
