-- Payment Gateway Seed Data
-- Inserts default payment methods, settings, and test data

-- ====================================================================
-- TABLE FIXES (for tables created in earlier migrations)
-- ====================================================================

-- Add missing columns to vendor_wallets if they don't exist
ALTER TABLE vendor_wallets ADD COLUMN IF NOT EXISTS total_earned FLOAT DEFAULT 0;
ALTER TABLE vendor_wallets ADD COLUMN IF NOT EXISTS last_credited_at TIMESTAMP;

-- ====================================================================
-- PAYMENT METHODS SEED DATA
-- ====================================================================

-- Clear existing payment methods (for clean re-seeding)
DELETE FROM payment_methods WHERE true;

-- Insert Razorpay payment methods
INSERT INTO payment_methods (name, gateway, method_type, display_name, icon_url, is_enabled, display_order, min_amount, max_amount) VALUES
-- UPI Methods
('UPI', 'razorpay', 'upi', 'UPI Payment', '/static/icons/payment/upi.png', true, 1, 1.0, 100000.0),
('Google Pay', 'razorpay', 'upi', 'Google Pay (UPI)', '/static/icons/payment/gpay.png', true, 2, 1.0, 100000.0),
('PhonePe', 'razorpay', 'upi', 'PhonePe (UPI)', '/static/icons/payment/phonepe.png', true, 3, 1.0, 100000.0),
('Paytm', 'razorpay', 'upi', 'Paytm (UPI)', '/static/icons/payment/paytm.png', true, 4, 1.0, 100000.0),

-- Card Methods
('Visa Card', 'razorpay', 'card', 'Visa Credit/Debit Card', '/static/icons/payment/visa.png', true, 10, 10.0, 200000.0),
('Mastercard', 'razorpay', 'card', 'Mastercard Credit/Debit', '/static/icons/payment/mastercard.png', true, 11, 10.0, 200000.0),
('RuPay Card', 'razorpay', 'card', 'RuPay Debit Card', '/static/icons/payment/rupay.png', true, 12, 10.0, 100000.0),
('American Express', 'razorpay', 'card', 'American Express Card', '/static/icons/payment/amex.png', false, 13, 50.0, 500000.0),
('Diners Club', 'razorpay', 'card', 'Diners Club Card', '/static/icons/payment/diners.png', false, 14, 50.0, 500000.0),

-- Netbanking Methods
('Netbanking - HDFC', 'razorpay', 'netbanking', 'HDFC Bank Netbanking', '/static/icons/payment/hdfc.png', true, 20, 100.0, 1000000.0),
('Netbanking - ICICI', 'razorpay', 'netbanking', 'ICICI Bank Netbanking', '/static/icons/payment/icici.png', true, 21, 100.0, 1000000.0),
('Netbanking - SBI', 'razorpay', 'netbanking', 'State Bank of India', '/static/icons/payment/sbi.png', true, 22, 100.0, 1000000.0),
('Netbanking - Axis', 'razorpay', 'netbanking', 'Axis Bank Netbanking', '/static/icons/payment/axis.png', true, 23, 100.0, 1000000.0),
('Netbanking - Kotak', 'razorpay', 'netbanking', 'Kotak Mahindra Bank', '/static/icons/payment/kotak.png', true, 24, 100.0, 1000000.0),

-- Wallet Methods
('Paytm Wallet', 'razorpay', 'wallet', 'Paytm Wallet', '/static/icons/payment/paytm-wallet.png', true, 30, 1.0, 10000.0),
('PhonePe Wallet', 'razorpay', 'wallet', 'PhonePe Wallet', '/static/icons/payment/phonepe-wallet.png', true, 31, 1.0, 10000.0),
('Amazon Pay', 'razorpay', 'wallet', 'Amazon Pay Balance', '/static/icons/payment/amazon-pay.png', true, 32, 1.0, 10000.0),
('Mobikwik', 'razorpay', 'wallet', 'Mobikwik Wallet', '/static/icons/payment/mobikwik.png', false, 33, 1.0, 10000.0),
('Freecharge', 'razorpay', 'wallet', 'Freecharge Wallet', '/static/icons/payment/freecharge.png', false, 34, 1.0, 10000.0),
('Airtel Money', 'razorpay', 'wallet', 'Airtel Money Wallet', '/static/icons/payment/airtel.png', false, 35, 1.0, 10000.0),

-- EMI Methods
('Card EMI', 'razorpay', 'emi', 'Credit Card EMI (3/6/9/12 months)', '/static/icons/payment/emi.png', true, 40, 1000.0, 500000.0),

-- Payment Link
('Payment Link', 'razorpay', 'payment_link', 'Share Payment Link via SMS/Email', '/static/icons/payment/link.png', true, 50, 1.0, 100000.0),

-- Pay Later
('Pay Later - Simpl', 'razorpay', 'paylater', 'Simpl Pay Later', '/static/icons/payment/simpl.png', false, 60, 100.0, 50000.0),
('Pay Later - LazyPay', 'razorpay', 'paylater', 'LazyPay Pay Later', '/static/icons/payment/lazypay.png', false, 61, 100.0, 50000.0)
ON CONFLICT (gateway, method_type) DO NOTHING;

-- Insert Stripe payment methods
INSERT INTO payment_methods (name, gateway, method_type, display_name, icon_url, is_enabled, display_order, min_amount, max_amount) VALUES
-- Card Methods
('Visa Card', 'stripe', 'card', 'Visa Credit/Debit Card', '/static/icons/payment/visa.png', true, 10, 0.50, 999999.99),
('Mastercard', 'stripe', 'card', 'Mastercard Credit/Debit', '/static/icons/payment/mastercard.png', true, 11, 0.50, 999999.99),
('American Express', 'stripe', 'card', 'American Express Card', '/static/icons/payment/amex.png', true, 12, 0.50, 999999.99),
('Discover', 'stripe', 'card', 'Discover Card', '/static/icons/payment/discover.png', false, 13, 0.50, 999999.99),
('JCB', 'stripe', 'card', 'JCB Card', '/static/icons/payment/jcb.png', false, 14, 0.50, 999999.99),
('Diners Club', 'stripe', 'card', 'Diners Club Card', '/static/icons/payment/diners.png', false, 15, 0.50, 999999.99),

-- International Methods
('SEPA Debit', 'stripe', 'sepa_debit', 'SEPA Direct Debit (Europe)', '/static/icons/payment/sepa.png', false, 20, 0.50, 999999.99),
('Bancontact', 'stripe', 'bancontact', 'Bancontact (Belgium)', '/static/icons/payment/bancontact.png', false, 21, 0.50, 999999.99),
('iDEAL', 'stripe', 'ideal', 'iDEAL (Netherlands)', '/static/icons/payment/ideal.png', false, 22, 0.50, 999999.99),
('Sofort', 'stripe', 'sofort', 'Sofort (Germany/Austria)', '/static/icons/payment/sofort.png', false, 23, 0.50, 999999.99),
('Giropay', 'stripe', 'giropay', 'Giropay (Germany)', '/static/icons/payment/giropay.png', false, 24, 0.50, 999999.99),
('EPS', 'stripe', 'eps', 'EPS (Austria)', '/static/icons/payment/eps.png', false, 25, 0.50, 999999.99),

-- Wallet Methods (Stripe)
('Apple Pay', 'stripe', 'apple_pay', 'Apple Pay', '/static/icons/payment/apple-pay.png', false, 30, 0.50, 999999.99),
('Google Pay', 'stripe', 'google_pay', 'Google Pay', '/static/icons/payment/google-pay.png', false, 31, 0.50, 999999.99),

-- Buy Now Pay Later
('Affirm', 'stripe', 'affirm', 'Affirm - Buy Now Pay Later', '/static/icons/payment/affirm.png', false, 40, 50.0, 30000.0),
('Klarna', 'stripe', 'klarna', 'Klarna - Buy Now Pay Later', '/static/icons/payment/klarna.png', false, 41, 50.0, 30000.0),
('Afterpay', 'stripe', 'afterpay', 'Afterpay - Buy Now Pay Later', '/static/icons/payment/afterpay.png', false, 42, 50.0, 30000.0)
ON CONFLICT (gateway, method_type) DO NOTHING;

-- ====================================================================
-- PAYMENT SETTINGS SEED DATA
-- ====================================================================

-- Clear existing payment settings
DELETE FROM payment_settings WHERE true;

-- Insert Razorpay payment settings
INSERT INTO payment_settings (gateway, is_enabled, config, min_amount, max_amount) VALUES
('razorpay', true, '{
    "api_endpoint": "https://api.razorpay.com/v1",
    "webhook_events": [
        "payment.captured",
        "payment.authorized",
        "payment.failed",
        "refund.processed",
        "refund.failed",
        "order.paid"
    ],
    "supports_upi": true,
    "supports_card": true,
    "supports_netbanking": true,
    "supports_wallet": true,
    "supports_emi": true,
    "supports_payment_link": true,
    "supports_paylater": true,
    "capture_method": "automatic",
    "timeout": 900,
    "currency": "INR",
    "settlement_cycle": "daily",
    "fees": {
        "upi": 0,
        "card": 2.0,
        "netbanking": 0,
        "wallet": 0,
        "emi": 3.0
    }
}'::jsonb, 1.0, 1000000.0);

-- Insert Stripe payment settings
INSERT INTO payment_settings (gateway, is_enabled, config, min_amount, max_amount) VALUES
('stripe', false, '{
    "api_endpoint": "https://api.stripe.com/v1",
    "webhook_events": [
        "payment_intent.succeeded",
        "payment_intent.payment_failed",
        "charge.refunded",
        "charge.refund.updated",
        "charge.dispute.created"
    ],
    "supports_card": true,
    "supports_sepa": true,
    "supports_apple_pay": true,
    "supports_google_pay": true,
    "supports_bancontact": true,
    "supports_ideal": true,
    "supports_sofort": true,
    "supports_giropay": true,
    "supports_eps": true,
    "supports_affirm": true,
    "supports_klarna": true,
    "supports_afterpay": true,
    "capture_method": "automatic",
    "timeout": 900,
    "currency": "INR",
    "supported_currencies": ["INR", "USD", "EUR", "GBP", "AUD", "CAD", "SGD"],
    "settlement_cycle": "daily",
    "fees": {
        "card": 2.9,
        "international_card": 3.9,
        "sepa_debit": 0.8
    }
}'::jsonb, 0.50, 999999.99);

-- ====================================================================
-- PLATFORM FEE CONFIGURATION
-- ====================================================================

-- Create platform_fee_settings table if not exists
CREATE TABLE IF NOT EXISTS platform_fee_settings (
    id BIGSERIAL PRIMARY KEY,
    category VARCHAR(50) DEFAULT 'equipment_rental',
    fee_type VARCHAR(20) DEFAULT 'percentage',
    fee_value FLOAT DEFAULT 10.0,
    min_fee FLOAT DEFAULT 0,
    max_fee FLOAT DEFAULT 10000.0,
    is_active BOOLEAN DEFAULT true,
    effective_from TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    effective_until TIMESTAMP,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO platform_fee_settings (category, fee_type, fee_value, min_fee, max_fee, description) VALUES
('equipment_rental', 'percentage', 10.0, 50.0, 5000.0, 'Standard platform fee for equipment rentals (10%)'),
('equipment_rental_premium', 'percentage', 15.0, 100.0, 10000.0, 'Premium category equipment (15%)'),
('equipment_rental_basic', 'percentage', 8.0, 30.0, 3000.0, 'Basic category equipment (8%)'),
('accessories_rental', 'percentage', 12.0, 40.0, 4000.0, 'Accessories and add-ons (12%)'),
('delivery_service', 'percentage', 5.0, 20.0, 2000.0, 'Equipment delivery service fee (5%)')
ON CONFLICT DO NOTHING;

-- ====================================================================
-- CANCELLATION POLICY SEED DATA
-- ====================================================================

CREATE TABLE IF NOT EXISTS cancellation_policies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cancellation_policy_rules (
    id BIGSERIAL PRIMARY KEY,
    policy_id BIGINT NOT NULL REFERENCES cancellation_policies(id) ON DELETE CASCADE,
    hours_before_start INT NOT NULL,
    refund_percent FLOAT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert standard cancellation policy
INSERT INTO cancellation_policies (name, description, is_default) VALUES
('Standard Cancellation Policy', 'Refund percentage decreases as booking start time approaches', true)
ON CONFLICT DO NOTHING;

-- Insert cancellation policy rules
INSERT INTO cancellation_policy_rules (policy_id, hours_before_start, refund_percent, description) VALUES
(1, 48, 100.0, 'Full refund if cancelled 48+ hours before start'),
(1, 24, 75.0, '75% refund if cancelled 24-48 hours before start'),
(1, 12, 50.0, '50% refund if cancelled 12-24 hours before start'),
(1, 6, 25.0, '25% refund if cancelled 6-12 hours before start'),
(1, 0, 0.0, 'No refund if cancelled less than 6 hours before start')
ON CONFLICT DO NOTHING;

-- ====================================================================
-- TEST DATA (FOR DEVELOPMENT/STAGING ONLY)
-- ====================================================================

-- Note: Remove this section before production deployment

-- Insert test vendor wallets
-- COMMENTED OUT: Test vendor IDs don't exist in database
-- INSERT INTO vendor_wallets (vendor_id, balance, total_earned) VALUES
-- (1001, 5000.00, 25000.00),
-- (1002, 3200.50, 18500.00),
-- (1003, 8750.25, 42000.75)
-- ON CONFLICT (vendor_id) DO NOTHING;

-- Insert test payment transactions
-- COMMENTED OUT: Test payment IDs don't exist in database
-- INSERT INTO payment_transactions (payment_id, booking_id, transaction_type, amount, gateway, status, metadata) VALUES
-- (1, 1001, 'credit', 4500.00, 'razorpay', 'completed', '{"type":"escrow_release","platform_fee":450.00,"vendor_id":1001}'::jsonb),
-- (2, 1002, 'credit', 3200.50, 'razorpay', 'completed', '{"type":"escrow_release","platform_fee":320.05,"vendor_id":1002}'::jsonb),
-- (3, 1003, 'debit', 500.00, 'razorpay', 'completed', '{"type":"refund","reason":"customer_cancellation"}'::jsonb)
-- ON CONFLICT DO NOTHING;

-- ====================================================================
-- VIEWS FOR REPORTING
-- ====================================================================

-- Create view for payment summary
CREATE OR REPLACE VIEW payment_summary AS
SELECT
    DATE(p.created_at) as payment_date,
    p.gateway,
    COUNT(*) as total_payments,
    SUM(p.total_amount) as total_amount,
    SUM(p.platform_fee) as total_platform_fee,
    SUM(p.vendor_amount) as total_vendor_amount,
    COUNT(CASE WHEN p.status = 'paid' THEN 1 END) as successful_payments,
    COUNT(CASE WHEN p.status = 'failed' THEN 1 END) as failed_payments,
    COUNT(CASE WHEN p.status = 'refunded' THEN 1 END) as refunded_payments
FROM payments p
GROUP BY DATE(p.created_at), p.gateway
ORDER BY payment_date DESC;

-- Create view for vendor earnings
CREATE OR REPLACE VIEW vendor_earnings AS
SELECT
    v.vendor_id,
    v.balance as current_balance,
    v.total_earned,
    v.last_credited_at,
    COUNT(DISTINCT p.id) as total_bookings,
    SUM(p.vendor_amount) as total_released,
    AVG(p.vendor_amount) as avg_earning_per_booking
FROM vendor_wallets v
LEFT JOIN payments p ON v.vendor_id = p.vendor_id AND p.status = 'completed'
GROUP BY v.vendor_id, v.balance, v.total_earned, v.last_credited_at
ORDER BY v.total_earned DESC;

-- Create view for refund summary
CREATE OR REPLACE VIEW refund_summary AS
SELECT
    DATE(p.refunded_at) as refund_date,
    p.gateway,
    COUNT(*) as total_refunds,
    SUM(p.refund_amount) as total_refunded,
    AVG(p.refund_amount) as avg_refund_amount,
    COUNT(CASE WHEN p.refund_status = 'processed' THEN 1 END) as processed_refunds,
    COUNT(CASE WHEN p.refund_status = 'pending' THEN 1 END) as pending_refunds
FROM payments p
WHERE p.status = 'refunded'
GROUP BY DATE(p.refunded_at), p.gateway
ORDER BY refund_date DESC;

-- ====================================================================
-- MIGRATION COMPLETION MARKER
-- ====================================================================

-- Insert migration record

-- Grant permissions (adjust based on your database user)
-- GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO genrent_app;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO genrent_app;

-- Comment removed - invalid SQL
