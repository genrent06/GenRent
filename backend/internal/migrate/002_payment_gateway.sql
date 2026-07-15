-- Payment Gateway Migration
-- Adds support for Razorpay/Stripe integration, escrow system, and refunds

-- Update payments table with gateway fields
ALTER TABLE payments ADD COLUMN IF NOT EXISTS gateway VARCHAR(20) DEFAULT 'razorpay';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS gateway_order_id VARCHAR;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS gateway_payment_id VARCHAR UNIQUE;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS gateway_status VARCHAR(30);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS payment_method VARCHAR(30);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS payment_metadata JSONB DEFAULT '{}'::jsonb;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS escrow_held_at TIMESTAMP;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS escrow_released_at TIMESTAMP;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS vendor_id BIGINT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS vendor_amount FLOAT DEFAULT 0;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS platform_fee FLOAT DEFAULT 0;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_id VARCHAR;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_amount FLOAT DEFAULT 0;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_status VARCHAR(20);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMP;

-- Add payment settings table
CREATE TABLE IF NOT EXISTS payment_settings (
    id BIGSERIAL PRIMARY KEY,
    gateway VARCHAR(20) NOT NULL UNIQUE,
    is_enabled BOOLEAN DEFAULT true,
    config JSONB DEFAULT '{}'::jsonb,
    min_amount FLOAT DEFAULT 1,
    max_amount FLOAT DEFAULT 1000000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default payment settings
INSERT INTO payment_settings (gateway, is_enabled, config) VALUES
('razorpay', true, '{"supports_upi":true, "supports_card":true, "supports_netbanking":true, "supports_wallet":true}'),
('stripe', false, '{"supports_card":true, "supports_sepa":true}')
ON CONFLICT (gateway) DO NOTHING;

-- Add payment methods table
CREATE TABLE IF NOT EXISTS payment_methods (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    gateway VARCHAR(20) NOT NULL,
    method_type VARCHAR(30) NOT NULL,
    display_name VARCHAR(100),
    icon_url VARCHAR,
    is_enabled BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    min_amount FLOAT DEFAULT 0,
    max_amount FLOAT DEFAULT 1000000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(gateway, method_type)
);

-- Insert default payment methods
INSERT INTO payment_methods (name, gateway, method_type, display_name, display_order) VALUES
-- Razorpay methods
('UPI', 'razorpay', 'upi', 'UPI Payment', 1),
('Card', 'razorpay', 'card', 'Credit/Debit Card', 2),
('Netbanking', 'razorpay', 'netbanking', 'Net Banking', 3),
('Wallet', 'razorpay', 'wallet', 'Mobile Wallet', 4),
-- Stripe methods
('Card', 'stripe', 'card', 'Credit/Debit Card', 10)
ON CONFLICT (gateway, method_type) DO NOTHING;

-- Create indexes for payments
CREATE INDEX IF NOT EXISTS idx_payments_gateway_payment ON payments(gateway_payment_id) WHERE gateway_payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payments_gateway_order ON payments(gateway_order_id) WHERE gateway_order_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_refund ON payments(refund_id) WHERE refund_id IS NOT NULL;

-- Create payment_transactions table for audit trail
CREATE TABLE IF NOT EXISTS payment_transactions (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT REFERENCES payments(id) ON DELETE SET NULL,
    booking_id BIGINT REFERENCES bookings(id) ON DELETE SET NULL,
    transaction_type VARCHAR(30) NOT NULL,
    amount FLOAT NOT NULL,
    gateway VARCHAR(20) NOT NULL,
    gateway_transaction_id VARCHAR,
    status VARCHAR(20) NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_trans_payment ON payment_transactions(payment_id) WHERE payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_trans_booking ON payment_transactions(booking_id) WHERE booking_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_trans_gateway ON payment_transactions(gateway_transaction_id) WHERE gateway_transaction_id IS NOT NULL;

-- Create refund_requests table for manual refund tracking
CREATE TABLE IF NOT EXISTS refund_requests (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT NOT NULL REFERENCES payments(id),
    booking_id BIGINT NOT NULL REFERENCES bookings(id),
    requested_by BIGINT NOT NULL REFERENCES users(id),
    amount FLOAT NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    processed_by BIGINT REFERENCES users(id),
    processed_at TIMESTAMP,
    admin_notes TEXT,
    gateway_refund_id VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_refund_payment ON refund_requests(payment_id);
CREATE INDEX IF NOT EXISTS idx_refund_booking ON refund_requests(booking_id);
CREATE INDEX IF NOT EXISTS idx_refund_status ON refund_requests(status);
CREATE INDEX IF NOT EXISTS idx_refund_requested ON refund_requests(requested_by);

-- Create vendor_wallets table for escrow management
CREATE TABLE IF NOT EXISTS vendor_wallets (
    id BIGSERIAL PRIMARY KEY,
    vendor_id BIGINT NOT NULL UNIQUE,
    balance FLOAT DEFAULT 0,
    total_earned FLOAT DEFAULT 0,
    last_credited_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vendor_wallet_vendor ON vendor_wallets(vendor_id);

-- Add vendor foreign key constraint if vendors table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vendors') THEN
        ALTER TABLE vendor_wallets
        ADD CONSTRAINT fk_vendor_wallet_vendor
        FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add payment foreign key constraint to vendor_wallets (users table as vendors)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        ALTER TABLE vendor_wallets
        ADD CONSTRAINT fk_vendor_wallet_user
        FOREIGN KEY (vendor_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Create webhook_events table for logging
CREATE TABLE IF NOT EXISTS webhook_events (
    id BIGSERIAL PRIMARY KEY,
    gateway VARCHAR(20) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_id VARCHAR UNIQUE,
    payload JSONB NOT NULL,
    signature VARCHAR,
    processed BOOLEAN DEFAULT false,
    processing_attempts INT DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_webhook_gateway ON webhook_events(gateway);
CREATE INDEX IF NOT EXISTS idx_webhook_type ON webhook_events(event_type);
CREATE INDEX IF NOT EXISTS idx_webhook_processed ON webhook_events(processed);
CREATE INDEX IF NOT EXISTS idx_webhook_event_id ON webhook_events(event_id) WHERE event_id IS NOT NULL;

-- Create payment_logs table for audit trail
CREATE TABLE IF NOT EXISTS payment_logs (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT REFERENCES payments(id) ON DELETE SET NULL,
    level VARCHAR(20) DEFAULT 'INFO',
    action VARCHAR(100),
    message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_logs_payment ON payment_logs(payment_id) WHERE payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_logs_level ON payment_logs(level);
CREATE INDEX IF NOT EXISTS idx_payment_logs_created ON payment_logs(created_at);

-- Create payment_settlements table for tracking settlements to vendors
CREATE TABLE IF NOT EXISTS payment_settlements (
    id BIGSERIAL PRIMARY KEY,
    vendor_id BIGINT NOT NULL,
    payment_id BIGINT REFERENCES payments(id),
    booking_id BIGINT REFERENCES bookings(id),
    amount FLOAT NOT NULL,
    platform_fee FLOAT NOT NULL,
    settlement_type VARCHAR(20) DEFAULT 'escrow_release',
    status VARCHAR(20) DEFAULT 'pending',
    processed_at TIMESTAMP,
    gateway_settlement_id VARCHAR,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_settlement_vendor ON payment_settlements(vendor_id);
CREATE INDEX IF NOT EXISTS idx_settlement_payment ON payment_settlements(payment_id);
CREATE INDEX IF NOT EXISTS idx_settlement_status ON payment_settlements(status);

-- Create payment_disputes table for handling chargebacks/disputes
CREATE TABLE IF NOT EXISTS payment_disputes (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT NOT NULL REFERENCES payments(id),
    gateway_dispute_id VARCHAR UNIQUE,
    dispute_type VARCHAR(50),
    amount FLOAT NOT NULL,
    reason TEXT,
    status VARCHAR(20) DEFAULT 'open',
    due_date TIMESTAMP,
    resolved_at TIMESTAMP,
    evidence JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_dispute_payment ON payment_disputes(payment_id);
CREATE INDEX IF NOT EXISTS idx_dispute_status ON payment_disputes(status);
CREATE INDEX IF NOT EXISTS idx_dispute_gateway ON payment_disputes(gateway_dispute_id) WHERE gateway_dispute_id IS NOT NULL;

-- Create functions for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers to relevant tables
DROP TRIGGER IF EXISTS update_payment_settings_updated_at ON payment_settings;
CREATE TRIGGER update_payment_settings_updated_at
    BEFORE UPDATE ON payment_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_payment_methods_updated_at ON payment_methods;
CREATE TRIGGER update_payment_methods_updated_at
    BEFORE UPDATE ON payment_methods
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_vendor_wallets_updated_at ON vendor_wallets;
CREATE TRIGGER update_vendor_wallets_updated_at
    BEFORE UPDATE ON vendor_wallets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_refund_requests_updated_at ON refund_requests;
CREATE TRIGGER update_refund_requests_updated_at
    BEFORE UPDATE ON refund_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_payment_settlements_updated_at ON payment_settlements;
CREATE TRIGGER update_payment_settlements_updated_at
    BEFORE UPDATE ON payment_settlements
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_payment_disputes_updated_at ON payment_disputes;
CREATE TRIGGER update_payment_disputes_updated_at
    BEFORE UPDATE ON payment_disputes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add helpful comments
COMMENT ON TABLE payment_settings IS 'Configuration settings for payment gateways';
COMMENT ON TABLE payment_methods IS 'Available payment methods for each gateway';
COMMENT ON TABLE payment_transactions IS 'Audit trail for all payment transactions';
COMMENT ON TABLE refund_requests IS 'Manual refund requests and tracking';
COMMENT ON TABLE vendor_wallets IS 'Vendor wallet balances for escrow management';
COMMENT ON TABLE webhook_events IS 'Incoming webhook events from payment gateways';
COMMENT ON TABLE payment_logs IS 'Audit logs for payment operations';
COMMENT ON TABLE payment_settlements IS 'Settlement records for vendor payments';
COMMENT ON TABLE payment_disputes IS 'Payment disputes and chargebacks';
