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
