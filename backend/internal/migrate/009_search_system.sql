-- Advanced Search System Migration
-- Adds support for Elasticsearch integration, saved searches, and search analytics

-- Create saved_searches table
CREATE TABLE IF NOT EXISTS saved_searches (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    query TEXT,
    category_ids TEXT, -- JSON array
    cities TEXT, -- JSON array
    min_price FLOAT DEFAULT 0,
    max_price FLOAT DEFAULT 0,
    min_rating FLOAT DEFAULT 0,
    brand VARCHAR(50),
    tags TEXT, -- JSON array
    sort_by VARCHAR(20) DEFAULT 'relevance',
    is_alert BOOLEAN DEFAULT false,
    alert_frequency VARCHAR(20) DEFAULT 'daily',
    last_alert_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    result_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_saved_searches_user ON saved_searches(user_id);
CREATE INDEX IF NOT EXISTS idx_saved_searches_active ON saved_searches(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_saved_searches_alerts ON saved_searches(is_alert, is_active, last_alert_at);

-- Create search_history table
CREATE TABLE IF NOT EXISTS search_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    filter_count INT DEFAULT 0,
    result_count INT DEFAULT 0,
    clicked_id BIGINT REFERENCES equipment(id),
    click_position INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_search_history_user ON search_history(user_id);
CREATE INDEX IF NOT EXISTS idx_search_history_created ON search_history(created_at DESC);

-- Create search_suggestions table for autocomplete
CREATE TABLE IF NOT EXISTS search_suggestions (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL, -- 'name', 'category', 'city', 'brand'
    text VARCHAR(200) NOT NULL,
    count INT DEFAULT 0,
    popularity FLOAT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(type, text)
);

CREATE INDEX IF NOT EXISTS idx_search_suggestions_type ON search_suggestions(type);
CREATE INDEX IF NOT EXISTS idx_search_suggestions_popularity ON search_suggestions(popularity DESC);

-- Create search_analytics table for tracking search performance
CREATE TABLE IF NOT EXISTS search_analytics (
    id BIGSERIAL PRIMARY KEY,
    date DATE NOT NULL,
    total_searches INT DEFAULT 0,
    unique_users INT DEFAULT 0,
    avg_result_count FLOAT DEFAULT 0,
    zero_result_searches INT DEFAULT 0,
    click_through_rate FLOAT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date)
);

CREATE INDEX IF NOT EXISTS idx_search_analytics_date ON search_analytics(date);

-- Create popular_searches aggregated view
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_popular_searches AS
SELECT
    query,
    COUNT(*) as search_count,
    AVG(result_count) as avg_results,
    MAX(created_at) as last_searched
FROM search_history
WHERE created_at > CURRENT_DATE - INTERVAL '30 days'
GROUP BY query
ORDER BY search_count DESC
LIMIT 1000;

CREATE UNIQUE INDEX ON mv_popular_searches (query);

-- Create function to refresh materialized view
CREATE OR REPLACE FUNCTION refresh_popular_searches()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_popular_searches;
END;
$$ LANGUAGE plpgsql;

-- Schedule the refresh (run this via cron or pg_cron)
-- SELECT cron.schedule('refresh-popular-searches', '0 * * * *', 'SELECT refresh_popular_searches()');

-- Add triggers for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_search_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_saved_searches_updated_at ON saved_searches;
CREATE TRIGGER update_saved_searches_updated_at
    BEFORE UPDATE ON saved_searches
    FOR EACH ROW EXECUTE FUNCTION update_search_updated_at();

DROP TRIGGER IF EXISTS update_search_suggestions_updated_at ON search_suggestions;
CREATE TRIGGER update_search_suggestions_updated_at
    BEFORE UPDATE ON search_suggestions
    FOR EACH ROW EXECUTE FUNCTION update_search_updated_at();

-- Add helpful comments
COMMENT ON TABLE saved_searches IS 'User-saved search queries with optional alerts for new results';
COMMENT ON TABLE search_history IS 'History of all searches performed by users for analytics';
COMMENT ON TABLE search_suggestions IS 'Search suggestions and autocomplete data';
COMMENT ON TABLE search_analytics IS 'Daily search performance analytics';
COMMENT ON MATERIALIZED VIEW mv_popular_searches IS 'Most popular search queries in the last 30 days';

-- Insert initial search suggestions from existing equipment data
INSERT INTO search_suggestions (type, text, count, popularity)
SELECT
    'category',
    ec.name,
    COUNT(e.id),
    COUNT(e.id)::float / 100.0
FROM equipment e
JOIN equipment_categories ec ON e.category_id = ec.id
GROUP BY ec.name
ON CONFLICT (type, text) DO NOTHING;

-- Insert city suggestions
INSERT INTO search_suggestions (type, text, count, popularity)
SELECT
    'city',
    city,
    COUNT(*),
    COUNT(*)::float / 100.0
FROM equipment
WHERE city IS NOT NULL AND city != ''
GROUP BY city
ON CONFLICT (type, text) DO NOTHING;

-- Insert brand suggestions
INSERT INTO search_suggestions (type, text, count, popularity)
SELECT
    'brand',
    brand,
    COUNT(*),
    COUNT(*)::float / 100.0
FROM equipment
WHERE brand IS NOT NULL AND brand != ''
GROUP BY brand
ON CONFLICT (type, text) DO NOTHING;

-- Insert equipment name suggestions (top 100)
INSERT INTO search_suggestions (type, text, count, popularity)
SELECT
    'name',
    name,
    1,
    0.01
FROM (
    SELECT DISTINCT name
    FROM equipment
    WHERE name IS NOT NULL AND name != ''
    LIMIT 100
) sub
ON CONFLICT (type, text) DO NOTHING;

-- Migration completion marker

-- Comment removed - invalid SQL
