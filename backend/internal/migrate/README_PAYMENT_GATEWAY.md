# Payment Gateway Migration Guide

## Overview
This migration adds comprehensive payment gateway support to the GenRent application, supporting both Razorpay and Stripe payment gateways with full escrow management and refund automation.

## Migration Files

### 002_payment_gateway.sql
**Purpose**: Sets up the core database schema for payment gateway integration

**Tables Created**:
- `payment_settings` - Gateway configuration and settings
- `payment_methods` - Available payment methods per gateway
- `payment_transactions` - Audit trail for all payment transactions
- `refund_requests` - Manual refund request tracking
- `vendor_wallets` - Vendor wallet balances for escrow management
- `webhook_events` - Incoming webhook event logging
- `payment_logs` - Audit logs for payment operations
- `payment_settlements` - Settlement records for vendor payments
- `payment_disputes` - Payment disputes and chargebacks handling

**Tables Modified**:
- `payments` - Added gateway-specific columns, escrow fields, and refund tracking

**Features**:
- Multi-gateway support (Razorpay & Stripe)
- Escrow system with automatic fund holding and release
- Vendor wallet management
- Comprehensive audit logging
- Webhook event tracking and replay capability
- Dispute management

### 003_payment_gateway_seed.sql
**Purpose**: Seeds initial data for payment gateway operations

**Data Included**:
- Complete payment methods for both Razorpay and Stripe
  - Razorpay: UPI, Cards, Netbanking, Wallets, EMI, Pay Later
  - Stripe: Cards, SEPA, International methods, Buy Now Pay Later
- Payment gateway settings with fee structures
- Platform fee configuration
- Cancellation policy rules
- Test data for development
- Reporting views for analytics

## How to Run Migrations

### Using Go Application
```go
import "genrent/internal/migrate"

// In your main.go or initialization
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err := migrate.AutoMigrate(db); err != nil {
    log.Fatal("Migration failed:", err)
}
```

### Manual Execution
```bash
# Using psql
psql -U genrent -d genrent_db -f internal/migrate/002_payment_gateway.sql
psql -U genrent -d genrent_db -f internal/migrate/003_payment_gateway_seed.sql
```

### Using Docker
```bash
docker exec -i genrent_db psql -U genrent -d genrent_db < internal/migrate/002_payment_gateway.sql
docker exec -i genrent_db psql -U genrent -d genrent_db < internal/migrate/003_payment_gateway_seed.sql
```

## Post-Migration Verification

### Check Table Creation
```sql
-- Verify all tables were created
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN (
    'payment_settings', 
    'payment_methods', 
    'payment_transactions',
    'refund_requests',
    'vendor_wallets',
    'webhook_events',
    'payment_logs',
    'payment_settlements',
    'payment_disputes'
);
```

### Check Indexes
```sql
-- Verify payment method indexes
SELECT indexname, tablename 
FROM pg_indexes 
WHERE tablename = 'payment_methods';
```

### Check Seed Data
```sql
-- Verify payment methods were seeded
SELECT gateway, COUNT(*) as method_count
FROM payment_methods
GROUP BY gateway;

-- Verify payment settings
SELECT gateway, is_enabled 
FROM payment_settings;
```

## Database Schema

### Key Relationships

```
payments (1) ──────< (N) payment_transactions
payments (1) ──────< (N) refund_requests
payments (1) ──────< (N) payment_logs
vendors (1) ───────< (N) vendor_wallets
vendors (1) ───────< (N) payment_settlements
```

### Escrow Flow

1. **Payment Created**: `payments.status = 'pending'`
2. **Payment Captured**: `payments.status = 'paid'`, funds move to escrow
3. **Escrow Hold**: `payments.escrow_held_at` set, `payments.status = 'escrow'`
4. **Delivery Confirmation**: Trigger fund release
5. **Escrow Release**: 
   - Calculate platform fee (10% default)
   - Credit vendor wallet
   - Create settlement record
   - `payments.status = 'completed'`

### Refund Flow

1. **Cancellation Request**: Create `refund_requests` record
2. **Calculate Refund Amount**: Based on cancellation policy
3. **Process Refund**: Call gateway API
4. **Update Status**: Mark `payments.status = 'refunded'`
5. **Audit Trail**: Create `payment_transactions` record

## Configuration

### Environment Variables Required

```env
# Razorpay Configuration
RAZORPAY_KEY_ID=your_key_id
RAZORPAY_KEY_SECRET=your_key_secret
RAZORPAY_WEBHOOK_SECRET=your_webhook_secret

# Stripe Configuration
STRIPE_PUBLISHABLE_KEY=your_publishable_key
STRIPE_SECRET_KEY=your_secret_key
STRIPE_WEBHOOK_SECRET=your_webhook_signing_secret

# Payment Gateway Settings
PAYMENT_GATEWAY=razorpay  # or stripe, or both
PAYMENT_TIMEOUT=900       # 15 minutes default
PLATFORM_FEE_RATE=10.0    # 10% platform fee
```

### Gateway Selection

The system supports running multiple gateways simultaneously:

```go
// In your payment service initialization
razorpayService := payment.NewRazorpayService(
    os.Getenv("RAZORPAY_KEY_ID"),
    os.Getenv("RAZORPAY_KEY_SECRET"),
    os.Getenv("RAZORPAY_WEBHOOK_SECRET"),
    testMode,
    900,
)

stripeService := payment.NewStripeService(
    os.Getenv("STRIPE_PUBLISHABLE_KEY"),
    os.Getenv("STRIPE_WEBHOOK_SECRET"),
    testMode,
    900,
)

// Use Razorpay for INR transactions
// Use Stripe for international transactions
```

## Security Considerations

### Webhook Verification
Both gateways provide webhook signatures for verification:

- **Razorpay**: HMAC SHA256 signature
- **Stripe**: Timestamped signature

Always verify webhook signatures before processing events.

### Data Protection
- Never store full card numbers
- Use gateway tokens for recurring payments
- Encrypt sensitive data at rest
- Use TLS for database connections

### PCI Compliance
- The system operates in SAQ A-EP compliance level
- No cardholder data is stored
- All payment processing happens at gateway
- Only payment IDs and metadata are stored

## Monitoring

### Key Metrics to Monitor

1. **Payment Success Rate**: Target > 95%
2. **Webhook Delivery Rate**: Target > 99%
3. **Average Processing Time**: Target < 5 seconds
4. **Refund Rate**: Monitor for anomalies (> 5% warning)
5. **Escrow Release Time**: Monitor for delays
6. **Platform Fee Collection**: Verify accuracy

### Database Views for Monitoring

```sql
-- Daily payment summary
SELECT * FROM payment_summary;

-- Vendor earnings
SELECT * FROM vendor_earnings;

-- Refund summary
SELECT * FROM refund_summary;
```

## Troubleshooting

### Common Issues

#### 1. Migration Fails with "Table Already Exists"
**Solution**: The migration uses `IF NOT EXISTS` clauses. If it still fails, check for conflicting table names:
```sql
SELECT table_name FROM information_schema.tables WHERE table_name LIKE 'payment%';
```

#### 2. Foreign Key Constraints Fail
**Solution**: Ensure referenced tables exist before applying constraints. The migration uses `DO` blocks to conditionally add constraints based on table existence.

#### 3. Seed Data Conflicts
**Solution**: Seed data uses `ON CONFLICT DO NOTHING` to prevent duplicate insertion errors.

#### 4. Webhook Events Not Processing
**Solution**: Check webhook_events table for failed events:
```sql
SELECT * FROM webhook_events WHERE processed = false;
```

## Rollback Procedure

If you need to rollback the payment gateway migration:

```sql
BEGIN;

-- Drop views
DROP VIEW IF EXISTS payment_summary;
DROP VIEW IF EXISTS vendor_earnings;
DROP VIEW IF EXISTS refund_summary;

-- Drop tables (in reverse order of creation)
DROP TABLE IF EXISTS payment_disputes;
DROP TABLE IF EXISTS payment_settlements;
DROP TABLE IF EXISTS payment_logs;
DROP TABLE IF EXISTS webhook_events;
DROP TABLE IF EXISTS vendor_wallets;
DROP TABLE IF EXISTS refund_requests;
DROP TABLE IF EXISTS payment_transactions;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS payment_settings;
DROP TABLE IF EXISTS cancellation_policy_rules;
DROP TABLE IF EXISTS cancellation_policies;
DROP TABLE IF EXISTS platform_fee_settings;

-- Remove payment table columns
ALTER TABLE payments DROP COLUMN IF EXISTS gateway;
ALTER TABLE payments DROP COLUMN IF EXISTS gateway_order_id;
ALTER TABLE payments DROP COLUMN IF EXISTS gateway_payment_id;
-- ... (repeat for all added columns)

COMMIT;
```

## Performance Considerations

### Indexes
The migration creates indexes on frequently queried columns:
- `gateway_payment_id` - Payment lookups
- `booking_id` - Transaction history
- `vendor_id` - Wallet operations
- `status` - Status-based queries

### Partitioning (Optional)
For high-volume deployments, consider partitioning:
- `payment_transactions` by date
- `payment_logs` by date
- `webhook_events` by processed status

## Support

For issues with the migration:
1. Check the application logs for specific error messages
2. Verify database user has necessary permissions
3. Ensure PostgreSQL version is 12+ (for JSONB features)
4. Review migration history in `schema_migrations` table

## Migration History

Track which migrations have been applied:

```sql
SELECT * FROM schema_migrations ORDER BY applied_at DESC;
```

Expected output after completion:
```
version                         | applied_at
--------------------------------+-------------------------
001_schema                      | 2024-07-10 XX:XX:XX
002_payment_gateway             | 2024-07-14 XX:XX:XX
003_payment_gateway_seed        | 2024-07-14 XX:XX:XX
```

---

**Last Updated**: 2026-07-14
**Version**: 1.0
**Status**: Production Ready
