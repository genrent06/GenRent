-- Reviews and Ratings System Migration
-- Supports customer reviews, vendor ratings, and equipment feedback

-- Equipment reviews table
CREATE TABLE IF NOT EXISTS equipment_reviews (
    id BIGSERIAL PRIMARY KEY,
    equipment_id BIGINT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    booking_id BIGINT REFERENCES bookings(id) ON DELETE SET NULL,
    customer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vendor_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Rating fields (1-5 scale)
    overall_rating SMALLINT NOT NULL CHECK (overall_rating BETWEEN 1 AND 5),
    equipment_quality_rating SMALLINT CHECK (equipment_quality_rating BETWEEN 1 AND 5),
    communication_rating SMALLINT CHECK (communication_rating BETWEEN 1 AND 5),
    value_rating SMALLINT CHECK (value_rating BETWEEN 1 AND 5),
    accuracy_rating SMALLINT CHECK (accuracy_rating BETWEEN 1 AND 5),

    -- Review content
    title VARCHAR(200),
    comment TEXT,
    pros TEXT[] CHECK (array_length(pros, 1) <= 10),
    cons TEXT[] CHECK (array_length(cons, 1) <= 10),

    -- Media attachments
    image_urls TEXT[],

    -- Vendor response
    vendor_response TEXT,
    vendor_response_at TIMESTAMP,

    -- Moderation
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'flagged')),
    flagged_reason TEXT,
    moderated_by BIGINT REFERENCES users(id),
    moderated_at TIMESTAMP,

    -- Helpful votes
    helpful_count INT DEFAULT 0,
    not_helpful_count INT DEFAULT 0,

    -- Verified purchase indicator
    verified_booking BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vendor ratings aggregation table
CREATE TABLE IF NOT EXISTS vendor_ratings (
    id BIGSERIAL PRIMARY KEY,
    vendor_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Rating statistics
    total_reviews INT DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0,

    -- Category-specific ratings
    equipment_quality_avg DECIMAL(3,2) DEFAULT 0,
    communication_avg DECIMAL(3,2) DEFAULT 0,
    value_avg DECIMAL(3,2) DEFAULT 0,
    accuracy_avg DECIMAL(3,2) DEFAULT 0,

    -- Rating distribution
    rating_1_count INT DEFAULT 0,
    rating_2_count INT DEFAULT 0,
    rating_3_count INT DEFAULT 0,
    rating_4_count INT DEFAULT 0,
    rating_5_count INT DEFAULT 0,

    -- Trust indicators
    verified_review_count INT DEFAULT 0,
    repeat_customer_count INT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Review helpful votes
CREATE TABLE IF NOT EXISTS review_votes (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL REFERENCES equipment_reviews(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type VARCHAR(10) NOT NULL CHECK (vote_type IN ('helpful', 'not_helpful')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(review_id, user_id)
);

-- Equipment ratings aggregation table
CREATE TABLE IF NOT EXISTS equipment_ratings (
    id BIGSERIAL PRIMARY KEY,
    equipment_id BIGINT NOT NULL UNIQUE REFERENCES equipment(id) ON DELETE CASCADE,

    -- Rating statistics
    total_reviews INT DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0,

    -- Category-specific averages
    quality_avg DECIMAL(3,2) DEFAULT 0,
    value_avg DECIMAL(3,2) DEFAULT 0,
    accuracy_avg DECIMAL(3,2) DEFAULT 0,

    -- Rating distribution
    rating_1_count INT DEFAULT 0,
    rating_2_count INT DEFAULT 0,
    rating_3_count INT DEFAULT 0,
    rating_4_count INT DEFAULT 0,
    rating_5_count INT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_equipment_reviews_equipment_id ON equipment_reviews(equipment_id);
CREATE INDEX IF NOT EXISTS idx_equipment_reviews_customer_id ON equipment_reviews(customer_id);
CREATE INDEX IF NOT EXISTS idx_equipment_reviews_vendor_id ON equipment_reviews(vendor_id);
CREATE INDEX IF NOT EXISTS idx_equipment_reviews_status ON equipment_reviews(status);
CREATE INDEX IF NOT EXISTS idx_equipment_reviews_rating ON equipment_reviews(overall_rating);
CREATE INDEX IF NOT EXISTS idx_equipment_reviews_created_at ON equipment_reviews(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_equipment_reviews_verified_booking ON equipment_reviews(verified_booking);

CREATE INDEX IF NOT EXISTS idx_review_votes_review_id ON review_votes(review_id);
CREATE INDEX IF NOT EXISTS idx_review_votes_user_id ON review_votes(user_id);

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_equipment_reviews_updated_at BEFORE UPDATE ON equipment_reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vendor_ratings_updated_at BEFORE UPDATE ON vendor_ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_equipment_ratings_updated_at BEFORE UPDATE ON equipment_ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to update vendor rating aggregations
CREATE OR REPLACE FUNCTION update_vendor_rating_aggregations()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        INSERT INTO vendor_ratings (vendor_id, total_reviews, average_rating,
            equipment_quality_avg, communication_avg, value_avg, accuracy_avg,
            rating_1_count, rating_2_count, rating_3_count, rating_4_count, rating_5_count,
            verified_review_count)
        SELECT
            NEW.vendor_id,
            COUNT(*),
            ROUND(AVG(overall_rating)::NUMERIC, 2),
            ROUND(AVG(equipment_quality_rating)::NUMERIC, 2),
            ROUND(AVG(communication_rating)::NUMERIC, 2),
            ROUND(AVG(value_rating)::NUMERIC, 2),
            ROUND(AVG(accuracy_rating)::NUMERIC, 2),
            COUNT(*) FILTER (WHERE overall_rating = 1),
            COUNT(*) FILTER (WHERE overall_rating = 2),
            COUNT(*) FILTER (WHERE overall_rating = 3),
            COUNT(*) FILTER (WHERE overall_rating = 4),
            COUNT(*) FILTER (WHERE overall_rating = 5),
            COUNT(*) FILTER (WHERE verified_booking = TRUE)
        FROM equipment_reviews
        WHERE vendor_id = NEW.vendor_id AND status = 'approved'
        ON CONFLICT (vendor_id) DO UPDATE SET
            total_reviews = EXCLUDED.total_reviews,
            average_rating = EXCLUDED.average_rating,
            equipment_quality_avg = EXCLUDED.equipment_quality_avg,
            communication_avg = EXCLUDED.communication_avg,
            value_avg = EXCLUDED.value_avg,
            accuracy_avg = EXCLUDED.accuracy_avg,
            rating_1_count = EXCLUDED.rating_1_count,
            rating_2_count = EXCLUDED.rating_2_count,
            rating_3_count = EXCLUDED.rating_3_count,
            rating_4_count = EXCLUDED.rating_4_count,
            rating_5_count = EXCLUDED.rating_5_count,
            verified_review_count = EXCLUDED.verified_review_count,
            updated_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for vendor rating updates
CREATE TRIGGER update_vendor_ratings_trigger
AFTER INSERT OR UPDATE ON equipment_reviews
FOR EACH ROW
WHEN (NEW.status = 'approved')
EXECUTE FUNCTION update_vendor_rating_aggregations();

-- Create function to update equipment rating aggregations
CREATE OR REPLACE FUNCTION update_equipment_rating_aggregations()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        INSERT INTO equipment_ratings (equipment_id, total_reviews, average_rating,
            quality_avg, value_avg, accuracy_avg,
            rating_1_count, rating_2_count, rating_3_count, rating_4_count, rating_5_count)
        SELECT
            NEW.equipment_id,
            COUNT(*),
            ROUND(AVG(overall_rating)::NUMERIC, 2),
            ROUND(AVG(equipment_quality_rating)::NUMERIC, 2),
            ROUND(AVG(value_rating)::NUMERIC, 2),
            ROUND(AVG(accuracy_rating)::NUMERIC, 2),
            COUNT(*) FILTER (WHERE overall_rating = 1),
            COUNT(*) FILTER (WHERE overall_rating = 2),
            COUNT(*) FILTER (WHERE overall_rating = 3),
            COUNT(*) FILTER (WHERE overall_rating = 4),
            COUNT(*) FILTER (WHERE overall_rating = 5)
        FROM equipment_reviews
        WHERE equipment_id = NEW.equipment_id AND status = 'approved'
        ON CONFLICT (equipment_id) DO UPDATE SET
            total_reviews = EXCLUDED.total_reviews,
            average_rating = EXCLUDED.average_rating,
            quality_avg = EXCLUDED.quality_avg,
            value_avg = EXCLUDED.value_avg,
            accuracy_avg = EXCLUDED.accuracy_avg,
            rating_1_count = EXCLUDED.rating_1_count,
            rating_2_count = EXCLUDED.rating_2_count,
            rating_3_count = EXCLUDED.rating_3_count,
            rating_4_count = EXCLUDED.rating_4_count,
            rating_5_count = EXCLUDED.rating_5_count,
            updated_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for equipment rating updates
CREATE TRIGGER update_equipment_ratings_trigger
AFTER INSERT OR UPDATE ON equipment_reviews
FOR EACH ROW
WHEN (NEW.status = 'approved')
EXECUTE FUNCTION update_equipment_rating_aggregations();

-- Create view for review summaries
CREATE OR REPLACE VIEW equipment_review_summary AS
SELECT
    e.id AS equipment_id,
    e.name AS equipment_name,
    er.total_reviews,
    er.average_rating,
    er.rating_1_count,
    er.rating_2_count,
    er.rating_3_count,
    er.rating_4_count,
    er.rating_5_count,
    (er.rating_5_count * 5 + er.rating_4_count * 4 + er.rating_3_count * 3 + er.rating_2_count * 2 + er.rating_1_count * 1)::FLOAT / NULLIF(er.total_reviews, 0) AS weighted_average
FROM equipment e
LEFT JOIN equipment_ratings er ON e.id = er.equipment_id;

-- Create view for vendor performance summary
CREATE OR REPLACE VIEW vendor_performance_summary AS
SELECT
    u.id AS vendor_id,
    u.name AS vendor_name,
    vr.total_reviews,
    vr.average_rating,
    vr.verified_review_count,
    vr.equipment_quality_avg,
    vr.communication_avg,
    vr.value_avg,
    vr.accuracy_avg
FROM users u
LEFT JOIN vendor_ratings vr ON u.id = vr.vendor_id
WHERE u.role = 'vendor';

-- Add comments for documentation
COMMENT ON TABLE equipment_reviews IS 'Reviews for equipment rentals by customers';
COMMENT ON TABLE vendor_ratings IS 'Aggregated rating statistics for vendors';
COMMENT ON TABLE equipment_ratings IS 'Aggregated rating statistics for equipment';
COMMENT ON TABLE review_votes IS 'User votes on review helpfulness';
