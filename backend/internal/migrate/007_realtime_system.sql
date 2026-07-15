-- Real-time Notification System Migration
-- Adds support for WebSocket chat, push notifications, and real-time features

-- Create conversations table for chat system
CREATE TABLE IF NOT EXISTS conversations (
    id BIGSERIAL PRIMARY KEY,
    booking_id BIGINT REFERENCES bookings(id) ON DELETE SET NULL,
    vendor_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    customer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_message TEXT,
    last_message_at TIMESTAMP,
    vendor_read BOOLEAN DEFAULT false,
    customer_read BOOLEAN DEFAULT false,
    status VARCHAR(20) DEFAULT 'active', -- active, archived, blocked
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (vendor_id != customer_id)
);

CREATE INDEX IF NOT EXISTS idx_conversations_booking ON conversations(booking_id);
CREATE INDEX IF NOT EXISTS idx_conversations_vendor ON conversations(vendor_id);
CREATE INDEX IF NOT EXISTS idx_conversations_customer ON conversations(customer_id);
CREATE INDEX IF NOT EXISTS idx_conversations_status ON conversations(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_conversations_unique ON conversations(vendor_id, customer_id, COALESCE(booking_id, 0));

-- Create messages table for chat messages
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text', -- text, image, file, system
    attachment_url VARCHAR,
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages(receiver_id);
CREATE INDEX IF NOT EXISTS idx_messages_read ON messages(is_read, receiver_id);
CREATE INDEX IF NOT EXISTS idx_messages_sent ON messages(sent_at DESC);

-- Create notification_preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT true,
    sms_enabled BOOLEAN DEFAULT false,
    push_enabled BOOLEAN DEFAULT true,
    in_app_enabled BOOLEAN DEFAULT true,
    booking_updates BOOLEAN DEFAULT true,
    message_alerts BOOLEAN DEFAULT true,
    promotions BOOLEAN DEFAULT false,
    reviews BOOLEAN DEFAULT true,
    availability BOOLEAN DEFAULT true,
    quiet_hours_start VARCHAR(5) DEFAULT '22:00', -- HH:MM format
    quiet_hours_end VARCHAR(5) DEFAULT '08:00', -- HH:MM format
    timezone VARCHAR(50) DEFAULT 'Asia/Kolkata',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notif_prefs_user ON notification_preferences(user_id);

-- Create device_registrations table for push notifications
CREATE TABLE IF NOT EXISTS device_registrations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token VARCHAR(500) NOT NULL UNIQUE,
    platform VARCHAR(20) NOT NULL, -- ios, android, web
    device_name VARCHAR(100),
    app_version VARCHAR(20),
    os_version VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_device_reg_user ON device_registrations(user_id);
CREATE INDEX IF NOT EXISTS idx_device_reg_platform ON device_registrations(platform);
CREATE INDEX IF NOT EXISTS idx_device_reg_active ON device_registrations(user_id, is_active);

-- Create real_time_inventory table for real-time availability tracking
CREATE TABLE IF NOT EXISTS real_time_inventory (
    id BIGSERIAL PRIMARY KEY,
    equipment_id BIGINT NOT NULL UNIQUE REFERENCES equipment(id) ON DELETE CASCADE,
    available_qty INT NOT NULL DEFAULT 0,
    reserved_qty INT NOT NULL DEFAULT 0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_realtime_inventory_equipment ON real_time_inventory(equipment_id);
CREATE INDEX IF NOT EXISTS idx_realtime_inventory_updated ON real_time_inventory(last_updated DESC);

-- Create typing_status table for real-time typing indicators
CREATE TABLE IF NOT EXISTS typing_status (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_typing BOOLEAN DEFAULT false,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(conversation_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_typing_status_conversation ON typing_status(conversation_id);
CREATE INDEX IF NOT EXISTS idx_typing_status_user ON typing_status(user_id);

-- Create presence table for online user tracking
CREATE TABLE IF NOT EXISTS user_presence (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_online BOOLEAN DEFAULT false,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    device_info JSONB DEFAULT '{}'::jsonb,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_presence_online ON user_presence(is_online, last_seen DESC);

-- Create notification_queue table for reliable notification delivery
CREATE TABLE IF NOT EXISTS notification_queue (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB DEFAULT '{}'::jsonb,
    priority INT DEFAULT 0, -- 0=normal, 1=high, 2=urgent
    status VARCHAR(20) DEFAULT 'pending', -- pending, sent, failed
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    last_error TEXT,
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notif_queue_user ON notification_queue(user_id);
CREATE INDEX IF NOT EXISTS idx_notif_queue_status ON notification_queue(status, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_notif_queue_priority ON notification_queue(priority, created_at);

-- Create chat_analytics table for chat metrics
CREATE TABLE IF NOT EXISTS chat_analytics (
    id BIGSERIAL PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    total_messages INT DEFAULT 0,
    unique_conversations INT DEFAULT 0,
    active_users INT DEFAULT 0,
    avg_response_time INT DEFAULT 0, -- in seconds
    messages_per_conversation FLOAT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chat_analytics_date ON chat_analytics(date);

-- Create function to update conversation updated_at
CREATE OR REPLACE FUNCTION update_conversation_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for conversation timestamp updates
DROP TRIGGER IF EXISTS update_conversations_timestamp ON conversations;
CREATE TRIGGER update_conversations_timestamp
    BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_conversation_updated_at();

-- Create function to handle new message
CREATE OR REPLACE FUNCTION handle_new_message()
RETURNS TRIGGER AS $$
BEGIN
    -- Update conversation
    UPDATE conversations
    SET last_message = NEW.message,
        last_message_at = NEW.sent_at,
        customer_read = CASE WHEN NEW.sender_id = customer_id THEN false ELSE customer_read END,
        vendor_read = CASE WHEN NEW.sender_id = vendor_id THEN false ELSE vendor_read END,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.conversation_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for new messages
DROP TRIGGER IF EXISTS trigger_handle_new_message ON messages;
CREATE TRIGGER trigger_handle_new_message
    AFTER INSERT ON messages
    FOR EACH ROW EXECUTE FUNCTION handle_new_message();

-- Create function to update typing status timestamp
CREATE OR REPLACE FUNCTION update_typing_status_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for typing status
DROP TRIGGER IF EXISTS update_typing_status_timestamp ON typing_status;
CREATE TRIGGER update_typing_status_timestamp
    BEFORE INSERT OR UPDATE ON typing_status
    FOR EACH ROW EXECUTE FUNCTION update_typing_status_timestamp();

-- Create function to update presence timestamp
CREATE OR REPLACE FUNCTION update_presence_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for user presence
DROP TRIGGER IF EXISTS update_presence_timestamp ON user_presence;
CREATE TRIGGER update_presence_timestamp
    BEFORE INSERT OR UPDATE ON user_presence
    FOR EACH ROW EXECUTE FUNCTION update_presence_timestamp();

-- Insert default notification preferences for existing users
INSERT INTO notification_preferences (user_id)
SELECT id FROM users
WHERE id NOT IN (SELECT user_id FROM notification_preferences)
ON CONFLICT (user_id) DO NOTHING;

-- Create view for active conversations with latest message
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_active_conversations AS
SELECT
    c.id,
    c.booking_id,
    c.vendor_id,
    c.customer_id,
    c.last_message,
    c.last_message_at,
    c.vendor_read,
    c.customer_read,
    c.status,
    u1.name as vendor_name,
    u2.name as customer_name,
    (SELECT COUNT(*) FROM messages WHERE conversation_id = c.id AND is_read = false AND receiver_id = c.vendor_id) as vendor_unread,
    (SELECT COUNT(*) FROM messages WHERE conversation_id = c.id AND is_read = false AND receiver_id = c.customer_id) as customer_unread,
    (SELECT COUNT(*) FROM messages WHERE conversation_id = c.id) as message_count
FROM conversations c
JOIN users u1 ON c.vendor_id = u1.id
JOIN users u2 ON c.customer_id = u2.id
WHERE c.status = 'active';

CREATE UNIQUE INDEX ON mv_active_conversations (id);

-- Create function to refresh materialized view
CREATE OR REPLACE FUNCTION refresh_active_conversations()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_active_conversations;
END;
$$ LANGUAGE plpgsql;

-- Create view for online users
CREATE OR REPLACE VIEW v_online_users AS
SELECT
    u.id as user_id,
    u.name,
    u.email,
    u.role,
    p.is_online,
    p.last_seen,
    p.device_info
FROM users u
LEFT JOIN user_presence p ON u.id = p.user_id
WHERE p.is_online = true;

-- Create view for unread message counts
CREATE OR REPLACE VIEW v_unread_counts AS
SELECT
    receiver_id as user_id,
    COUNT(*) as unread_count
FROM messages
WHERE is_read = false
GROUP BY receiver_id;

-- Helpful comments
COMMENT ON TABLE conversations IS 'Chat conversations between vendors and customers';
COMMENT ON TABLE messages IS 'Individual chat messages';
COMMENT ON TABLE notification_preferences IS 'User notification preferences and settings';
COMMENT ON TABLE device_registrations IS 'Mobile devices for push notifications';
COMMENT ON TABLE real_time_inventory IS 'Real-time equipment availability tracking';
COMMENT ON TABLE typing_status IS 'Real-time typing indicators';
COMMENT ON TABLE user_presence IS 'User online presence and status';
COMMENT ON TABLE notification_queue IS 'Queue for reliable notification delivery';
COMMENT ON TABLE chat_analytics IS 'Daily chat metrics and analytics';

-- Migration completion marker

-- Comment removed - invalid SQL
