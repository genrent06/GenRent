-- User Profile Enhancements Migration
-- Adds profile completion tracking, preferences, and verification features

-- User profiles table
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Basic Information
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    display_name VARCHAR(100),
    bio TEXT,
    profile_image_url VARCHAR(500),
    cover_image_url VARCHAR(500),

    -- Location Information
    country VARCHAR(2),    -- ISO country code
    state VARCHAR(100),
    city VARCHAR(100),
    zip_code VARCHAR(20),
    address TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),

    -- Business Information (for vendors)
    company_name VARCHAR(200),
    business_type VARCHAR(50),    -- individual, company, partnership
    tax_id VARCHAR(50),
    website_url VARCHAR(255),
    established_year INT,

    -- Social Media
    linkedin_url VARCHAR(255),
    twitter_url VARCHAR(255),
    instagram_url VARCHAR(255),
    facebook_url VARCHAR(255),

    -- Contact Preferences
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    preferred_contact VARCHAR(20), -- email, phone, both

    -- Settings
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    currency VARCHAR(3) DEFAULT 'USD',
    date_format VARCHAR(20) DEFAULT 'MM/DD/YYYY',
    time_format VARCHAR(10) DEFAULT '12h',

    -- Profile Completion
    completion_percentage INT DEFAULT 0,
    last_completed_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Notification Preferences
    email_notifications BOOLEAN DEFAULT true,
    push_notifications BOOLEAN DEFAULT true,
    sms_notifications BOOLEAN DEFAULT false,
    marketing_emails BOOLEAN DEFAULT false,
    product_updates BOOLEAN DEFAULT true,
    booking_reminders BOOLEAN DEFAULT true,
    promo_offers BOOLEAN DEFAULT false,
    review_notifications BOOLEAN DEFAULT true,
    message_notifications BOOLEAN DEFAULT true,
    payment_notifications BOOLEAN DEFAULT true,

    -- Privacy Settings
    profile_visibility VARCHAR(20) DEFAULT 'public', -- public, private, connections_only
    show_contact_info BOOLEAN DEFAULT false,
    show_activity_status BOOLEAN DEFAULT true,
    allow_direct_messages BOOLEAN DEFAULT true,
    show_online_status BOOLEAN DEFAULT true,

    -- Search & Discovery
    save_search_history BOOLEAN DEFAULT true,
    personalized_recommendations BOOLEAN DEFAULT true,
    location_based_services BOOLEAN DEFAULT false,

    -- Booking Preferences
    auto_accept_bookings BOOLEAN DEFAULT false,
    require_verification BOOLEAN DEFAULT true,
    minimum_booking_notice INT DEFAULT 24, -- hours
    maximum_booking_duration INT DEFAULT 30, -- days
    cancellation_policy VARCHAR(50) DEFAULT 'flexible', -- flexible, moderate, strict

    -- Payment Preferences
    default_payment_method VARCHAR(100),
    currency_preference VARCHAR(3) DEFAULT 'USD',
    require_deposit BOOLEAN DEFAULT false,
    deposit_percentage INT DEFAULT 0,

    -- Display Preferences
    theme VARCHAR(20) DEFAULT 'light', -- light, dark, auto
    font_size VARCHAR(20) DEFAULT 'medium', -- small, medium, large
    enable_animations BOOLEAN DEFAULT true,
    high_contrast_mode BOOLEAN DEFAULT false,

    -- Communication Preferences
    preferred_language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User verification table
CREATE TABLE IF NOT EXISTS user_verifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Verification Status
    is_verified BOOLEAN DEFAULT false,
    verification_status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, review
    verification_level VARCHAR(20) DEFAULT 'basic', -- basic, standard, premium
    verified_at TIMESTAMP,
    expires_at TIMESTAMP,
    next_review_date TIMESTAMP,

    -- Identity Verification
    identity_verified BOOLEAN DEFAULT false,
    identity_document_type VARCHAR(50), -- passport, driver_license, national_id
    identity_document_number VARCHAR(100),
    identity_document_url VARCHAR(500),
    identity_selfie_url VARCHAR(500),

    -- Business Verification (for vendors)
    business_verified BOOLEAN DEFAULT false,
    business_license_number VARCHAR(100),
    business_license_url VARCHAR(500),
    tax_document_url VARCHAR(500),
    insurance_document_url VARCHAR(500),

    -- Address Verification
    address_verified BOOLEAN DEFAULT false,
    address_proof_type VARCHAR(50), -- utility_bill, bank_statement
    address_proof_url VARCHAR(500),

    -- Additional Verifications
    email_verified BOOLEAN DEFAULT false,
    phone_verified BOOLEAN DEFAULT false,
    social_verified BOOLEAN DEFAULT false,
    bank_account_verified BOOLEAN DEFAULT false,

    -- Verification Metadata
    verification_attempts INT DEFAULT 0,
    last_attempt_at TIMESTAMP,
    rejection_reason TEXT,
    rejection_details TEXT,
    notes TEXT,

    -- Admin Review
    reviewed_by BIGINT REFERENCES users(id),
    reviewed_at TIMESTAMP,
    admin_notes TEXT,

    -- Trust Score
    trust_score DECIMAL(5, 2) DEFAULT 0, -- 0-100

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User activity tracking table
CREATE TABLE IF NOT EXISTS user_activities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL, -- login, booking, review, search, etc.
    activity_data TEXT, -- JSON data
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    device_type VARCHAR(50), -- mobile, desktop, tablet
    referrer VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(500) NOT NULL UNIQUE,
    refresh_token VARCHAR(500),
    device_type VARCHAR(50),
    device_fingerprint VARCHAR(255),
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User achievements and badges table
CREATE TABLE IF NOT EXISTS user_achievements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_type VARCHAR(50) NOT NULL, -- first_booking, top_reviewer, trusted_vendor, etc.
    achievement_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url VARCHAR(500),
    badge_level VARCHAR(20) DEFAULT 'bronze', -- bronze, silver, gold, platinum
    progress DECIMAL(5, 2) DEFAULT 0, -- 0-100 for tiered achievements
    max_progress DECIMAL(5, 2) DEFAULT 100,
    unlocked_at TIMESTAMP,
    is_displayed BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, achievement_type)
);

-- Custom profile fields table
CREATE TABLE IF NOT EXISTS user_profile_fields (
    id BIGSERIAL PRIMARY KEY,
    field_key VARCHAR(50) NOT NULL UNIQUE,
    field_label VARCHAR(100) NOT NULL,
    field_type VARCHAR(20) NOT NULL, -- text, select, multiselect, boolean, date, number
    options TEXT, -- JSON array for select/multiselect
    is_required BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    description TEXT,
    help_text TEXT,
    min_value DECIMAL(10, 2),
    max_value DECIMAL(10, 2),
    validation TEXT, -- JSON validation rules
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Custom profile values table
CREATE TABLE IF NOT EXISTS user_profile_values (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    field_id BIGINT NOT NULL REFERENCES user_profile_fields(id) ON DELETE CASCADE,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, field_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_profiles_location ON user_profiles(country, city);
CREATE INDEX IF NOT EXISTS idx_user_profiles_completion ON user_profiles(completion_percentage);

CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_preferences(user_id);

CREATE INDEX IF NOT EXISTS idx_user_verifications_user_id ON user_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_user_verifications_status ON user_verifications(verification_status);
CREATE INDEX IF NOT EXISTS idx_user_verifications_level ON user_verifications(verification_level);
CREATE INDEX IF NOT EXISTS idx_user_verifications_trust ON user_verifications(trust_score);

CREATE INDEX IF NOT EXISTS idx_user_activities_user_id ON user_activities(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activities_type ON user_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_user_activities_created_at ON user_activities(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(is_active, expires_at);

CREATE INDEX IF NOT EXISTS idx_user_achievements_user_id ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_type ON user_achievements(achievement_type);
CREATE INDEX IF NOT EXISTS idx_user_achievements_unlocked ON user_achievements(unlocked_at) WHERE unlocked_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_user_profile_values_user_id ON user_profile_values(user_id);
CREATE INDEX IF NOT EXISTS idx_user_profile_values_field_id ON user_profile_values(field_id);

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_user_profiles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER user_profiles_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_user_profiles_updated_at();

CREATE OR REPLACE FUNCTION update_user_preferences_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER user_preferences_updated_at BEFORE UPDATE ON user_preferences
    FOR EACH ROW EXECUTE FUNCTION update_user_preferences_updated_at();

CREATE OR REPLACE FUNCTION update_user_verifications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER user_verifications_updated_at BEFORE UPDATE ON user_verifications
    FOR EACH ROW EXECUTE FUNCTION update_user_verifications_updated_at();

-- Create view for user verification summary
CREATE OR REPLACE VIEW user_verification_summary AS
SELECT
    u.id AS user_id,
    u.name,
    u.email,
    uv.is_verified,
    uv.verification_status,
    uv.verification_level,
    uv.identity_verified,
    uv.email_verified,
    uv.phone_verified,
    uv.business_verified,
    uv.trust_score,
    uv.verified_at,
    uv.expires_at
FROM users u
LEFT JOIN user_verifications uv ON u.id = uv.user_id;

-- Create view for user activity summary
CREATE OR REPLACE VIEW user_activity_summary AS
SELECT
    user_id,
    COUNT(*) AS total_activities,
    COUNT(CASE WHEN created_at > CURRENT_DATE - INTERVAL '7 days' THEN 1 END) AS recent_activities,
    MAX(created_at) AS last_activity,
    COUNT(DISTINCT device_type) AS device_types_used
FROM user_activities
GROUP BY user_id;

-- Create view for user achievement summary
CREATE OR REPLACE VIEW user_achievement_summary AS
SELECT
    user_id,
    COUNT(*) FILTER (WHERE unlocked_at IS NOT NULL) AS total_achievements,
    COUNT(*) FILTER (WHERE badge_level = 'gold') AS gold_badges,
    COUNT(*) FILTER (WHERE badge_level = 'silver') AS silver_badges,
    COUNT(*) FILTER (WHERE badge_level = 'bronze') AS bronze_badges,
    AVG(progress) AS avg_progress
FROM user_achievements
GROUP BY user_id;

-- Add comments for documentation
COMMENT ON TABLE user_profiles IS 'Extended user profile information';
COMMENT ON TABLE user_preferences IS 'User notification and privacy preferences';
COMMENT ON TABLE user_verifications IS 'User verification and identity verification data';
COMMENT ON TABLE user_activities IS 'User activity tracking for analytics';
COMMENT ON TABLE user_sessions IS 'Active user session management';
COMMENT ON TABLE user_achievements IS 'User achievements and badges';
COMMENT ON TABLE user_profile_fields IS 'Custom profile field definitions';
COMMENT ON TABLE user_profile_values IS 'Custom values for user profile fields';
