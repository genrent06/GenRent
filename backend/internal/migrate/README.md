# Database Migrations

This directory contains SQL migrations for the GenRent platform - a construction equipment rental marketplace.

## Overview

The migration system is embedded in the Go binary using `embed` and runs automatically on application startup. It tracks applied migrations in the `schema_migrations` table to ensure each migration runs only once.

## Architecture

### Migration System

- **File Pattern**: `<version>.sql` (e.g., `001_schema.sql`)
- **Execution**: Migrations run in order by filename
- **Tracking**: Applied migrations stored in `schema_migrations` table
- **Safety**: Each migration runs in a database transaction

### How It Works

1. On startup, `AutoMigrate(db)` is called
2. System reads all `.sql` files from the embedded filesystem
3. Compares against applied migrations in `schema_migrations`
4. Executes any new migrations in a transaction
5. Records successful migrations in `schema_migrations`

## Database Schema

### Entity Relationship Diagram

```
users (customers, vendors, admins)
  ├──> vendors (vendor profiles)
  │     ├──> equipment (equipment listings)
  │     │     └──> equipment_categories
  │     ├──> generators (legacy)
  │     ├──> vendor_wallets
  │     ├──> vendor_bank_accounts
  │     └──> withdrawal_requests
  ├──> bookings (rental orders)
  │     ├──> equipment
  │     ├──> generators
  │     ├──> equipment_categories
  │     ├──> payments
  │     ├──> booking_handovers
  │     └──> damage_disputes
  ├──> payments
  ├──> platform_revenues
  ├──> notifications
  ├──> audit_logs
  └──> password_resets
```

---

## Table Descriptions

### 1. `users`
Stores all user accounts including customers, vendors, and administrators.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `name` | VARCHAR | Full name |
| `email` | VARCHAR | Unique email address |
| `phone` | VARCHAR | Phone number |
| `password` | VARCHAR | Hashed password |
| `role` | VARCHAR(20) | Role: 'customer', 'vendor', 'admin' |
| `risk_score` | FLOAT | Fraud/risk assessment score |
| `deleted_at` | TIMESTAMP | Soft delete timestamp |
| `created_at` | TIMESTAMP | Account creation time |
| `updated_at` | TIMESTAMP | Last update time |

**Indexes**: `email`, `risk_score`, `deleted_at`

---

### 2. `vendors`
Extended profile information for vendor users.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `user_id` | BIGINT | FK to users (CASCADE delete) |
| `company_name` | VARCHAR | Registered company name |
| `address` | TEXT | Full address |
| `city` | VARCHAR | City for location search |
| `latitude` | FLOAT | Geographic coordinate |
| `longitude` | FLOAT | Geographic coordinate |
| `description` | TEXT | Business description |
| `phone` | VARCHAR | Contact phone |
| `verified` | BOOLEAN | Admin verification status |
| `security_deposit` | FLOAT | Required deposit amount |
| `security_deposit_paid` | BOOLEAN | Deposit payment status |
| `reliability_score` | FLOAT | Performance score (0-5) |
| `total_bookings` | INT | Total bookings received |
| `successful_deliveries` | INT | Successful delivery count |
| `cancelled_bookings` | INT | Cancellation count |
| `average_rating` | FLOAT | Average customer rating |
| `total_ratings` | INT | Number of ratings |
| `avg_response_minutes` | FLOAT | Average response time |
| `risk_score` | FLOAT | Vendor risk assessment |
| `deleted_at` | TIMESTAMP | Soft delete timestamp |
| `created_at` | TIMESTAMP | Record creation time |
| `updated_at` | TIMESTAMP | Last update time |

**Indexes**: `user_id`, `city`, `risk_score`, `deleted_at`

---

### 3. `equipment_categories`
Hierarchical categories for equipment classification.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `name` | VARCHAR | Category name |
| `parent_category_id` | BIGINT | FK to equipment_categories (self-referential) |
| `description` | TEXT | Category description |
| `icon_url` | VARCHAR | Category icon URL |
| `display_order` | INT | UI display order |
| `deleted_at` | TIMESTAMP | Soft delete timestamp |
| `created_at` | TIMESTAMP | Record creation time |
| `updated_at` | TIMESTAMP | Last update time |

**Indexes**: `name`, `parent_category_id`, `deleted_at`

---

### 4. `equipment`
Individual equipment items available for rent.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `vendor_id` | BIGINT | FK to vendors (CASCADE delete) |
| `category_id` | BIGINT | FK to equipment_categories |
| `name` | VARCHAR | Equipment name |
| `brand` | VARCHAR | Manufacturer brand |
| `model` | VARCHAR | Model identifier |
| `description` | TEXT | Equipment description |
| `daily_price` | FLOAT | Daily rental rate |
| `weekly_price` | FLOAT | Weekly rental rate |
| `monthly_price` | FLOAT | Monthly rental rate |
| `mobilization_fee` | FLOAT | Delivery/setup fee |
| `demobilization_fee` | FLOAT | Pickup/reset fee |
| `total_quantity` | INT | Total owned quantity |
| `available_quantity` | INT | Currently available |
| `location` | VARCHAR | Storage location |
| `city` | VARCHAR | City for search |
| `latitude` | FLOAT | Geographic coordinate |
| `longitude` | FLOAT | Geographic coordinate |
| `availability_status` | VARCHAR(20) | 'available', 'rented', 'maintenance' |
| `reservation_expiry` | TIMESTAMP | When current booking ends |
| `image_url` | VARCHAR | Primary image URL |
| `specs` | JSONB | Equipment specifications |
| `deleted_at` | TIMESTAMP | Soft delete timestamp |
| `created_at` | TIMESTAMP | Record creation time |
| `updated_at` | TIMESTAMP | Last update time |

**Indexes**: `vendor_id`, `category_id`, `city`, `availability_status`, `deleted_at`

---

### 5. `generators`
**[LEGACY]** Specific table for generators (being migrated to equipment).

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `vendor_id` | BIGINT | FK to vendors (CASCADE delete) |
| `name` | VARCHAR | Generator name |
| `capacity_kva` | INT | Power capacity in KVA |
| `price_per_day` | FLOAT | Daily rental rate |
| `price_per_month` | FLOAT | Monthly rental rate |
| `fuel_type` | VARCHAR | 'diesel', 'petrol', 'gas' |
| `brand` | VARCHAR | Manufacturer |
| `location` | VARCHAR | Storage location |
| `city` | VARCHAR | City for search |
| `latitude` | FLOAT | Geographic coordinate |
| `longitude` | FLOAT | Geographic coordinate |
| `availability_status` | VARCHAR(20) | Current status |
| `reservation_expiry` | TIMESTAMP | Booking end time |
| `description` | TEXT | Description |
| `image_url` | VARCHAR | Image URL |
| `deleted_at` | TIMESTAMP | Soft delete |
| `created_at` | TIMESTAMP | Creation time |
| `updated_at` | TIMESTAMP | Update time |

**Indexes**: `vendor_id`, `city`, `availability_status`, `deleted_at`

---

### 6. `bookings`
Rental booking orders.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `customer_id` | BIGINT | FK to users |
| `generator_id` | BIGINT | FK to generators (nullable) |
| `equipment_id` | BIGINT | FK to equipment (nullable) |
| `category_id` | BIGINT | FK to equipment_categories |
| `start_date` | TIMESTAMP | Rental start time |
| `end_date` | TIMESTAMP | Rental end time |
| `total_price` | FLOAT | Final total cost |
| `rental_price` | FLOAT | Base rental cost |
| `mobilization_fee` | FLOAT | Delivery fee charged |
| `demobilization_fee` | FLOAT | Return fee charged |
| `advance_amount` | FLOAT | Required advance payment |
| `advance_paid` | BOOLEAN | Advance payment status |
| `status` | VARCHAR(30) | 'requested', 'confirmed', 'dispatched', 'delivered', 'completed', 'cancelled' |
| `address` | TEXT | Delivery address |
| `notes` | TEXT | Customer notes |
| `delivery_otp` | VARCHAR(6) | Delivery verification OTP |
| `otp_verified` | BOOLEAN | OTP verification status |
| `accepted_at` | TIMESTAMP | Vendor accept time |
| `dispatched_at` | TIMESTAMP | Dispatch time |
| `delivered_at` | TIMESTAMP | Delivery confirmation |
| `completed_at` | TIMESTAMP | Booking completion |
| `customer_rating` | INT | Rating given (1-5) |
| `customer_review` | TEXT | Customer review text |
| `cancel_reason` | TEXT | Cancellation reason |
| `return_initiated_at` | TIMESTAMP | Return process start |
| `return_otp` | VARCHAR(6) | Return verification OTP |
| `return_otp_verified` | BOOLEAN | Return OTP status |
| `created_at` | TIMESTAMP | Creation time |
| `updated_at` | TIMESTAMP | Update time |

**Indexes**: `customer_id`, `generator_id`, `equipment_id`, `status`

---

### 7. `payments`
Payment records for bookings.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `booking_id` | BIGINT | FK to bookings (unique) |
| `total_amount` | FLOAT | Total charged |
| `advance_amount` | FLOAT | Advance paid |
| `vendor_amount` | FLOAT | Vendor share |
| `platform_fee` | FLOAT | Platform commission |
| `method` | VARCHAR(20) | Payment method |
| `status` | VARCHAR(20) | 'pending', 'completed', 'failed', 'refunded' |
| `transaction_id` | VARCHAR | Gateway transaction ID |
| `paid_at` | TIMESTAMP | Payment completion time |
| `created_at` | TIMESTAMP | Creation time |
| `updated_at` | TIMESTAMP | Update time |

**Indexes**: `booking_id`

---

### 8. `vendor_wallets`
Wallet accounts for vendors to receive payments.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `vendor_id` | BIGINT | FK to vendors (unique, CASCADE) |
| `balance` | FLOAT | Available balance |
| `hold_balance` | FLOAT | On hold for active bookings |
| `withdrawal_hold_balance` | FLOAT | On hold for withdrawals |
| `created_at` | TIMESTAMP | Creation time |
| `updated_at` | TIMESTAMP | Update time |

---

### 9. `wallet_transactions`
Transaction history for vendor wallets.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `wallet_id` | BIGINT | FK to vendor_wallets (CASCADE) |
| `booking_id` | BIGINT | Related booking |
| `amount` | FLOAT | Transaction amount |
| `type` | VARCHAR(30) | 'credit', 'debit', 'withdrawal', 'refund' |
| `description` | TEXT | Transaction description |
| `created_at` | TIMESTAMP | Creation time |

**Indexes**: `wallet_id`, `booking_id`

---

### 10. `vendor_bank_accounts`
Bank account details for vendor withdrawals.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `vendor_id` | BIGINT | FK to vendors (CASCADE) |
| `bank_name` | VARCHAR | Bank name |
| `account_no` | VARCHAR | Account number |
| `ifsc` | VARCHAR | IFSC code |
| `account_name` | VARCHAR | Account holder name |
| `is_primary` | BOOLEAN | Primary account flag |
| `created_at` | TIMESTAMP | Creation time |
| `updated_at` | TIMESTAMP | Update time |

---

### 11. `withdrawal_requests`
Vendor withdrawal requests.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `vendor_id` | BIGINT | FK to vendors (CASCADE) |
| `bank_account_id` | BIGINT | FK to vendor_bank_accounts |
| `amount` | FLOAT | Withdrawal amount |
| `status` | VARCHAR(20) | 'pending', 'approved', 'rejected', 'processing', 'completed' |
| `bank_name` | VARCHAR | Snapshot bank name |
| `account_no` | VARCHAR | Snapshot account |
| `ifsc` | VARCHAR | Snapshot IFSC |
| `account_name` | VARCHAR | Snapshot account name |
| `admin_note` | TEXT | Admin notes |
| `processed_at` | TIMESTAMP | Processing time |
| `otp_code` | VARCHAR | OTP for confirmation |
| `otp_expires_at` | TIMESTAMP | OTP expiry |
| `created_at` | TIMESTAMP | Creation time |
| `updated_at` | TIMESTAMP | Update time |

---

### 12. `platform_revenues`
Platform fee tracking.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `payment_id` | BIGINT | FK to payments |
| `booking_id` | BIGINT | FK to bookings |
| `amount` | FLOAT | Platform fee amount |
| `type` | VARCHAR(20) | Fee type |
| `description` | TEXT | Description |
| `created_at` | TIMESTAMP | Creation time |

---

### 13. `booking_handovers`
Handover documentation for equipment delivery/return.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `booking_id` | BIGINT | FK to bookings (CASCADE) |
| `type` | VARCHAR(20) | 'delivery', 'return' |
| `photo_urls` | JSONB | Array of photo URLs |
| `checklist` | JSONB | Checklist items |
| `notes` | TEXT | Handover notes |
| `uploaded_by` | BIGINT | User who uploaded |
| `verified_at` | TIMESTAMP | Verification time |
| `created_at` | TIMESTAMP | Creation time |

---

### 14. `damage_disputes`
Dispute records for equipment damage claims.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `booking_id` | BIGINT | FK to bookings (CASCADE) |
| `raised_by` | BIGINT | FK to users |
| `description` | TEXT | Dispute details |
| `claimed_amount` | FLOAT | Claim amount |
| `photo_urls` | JSONB | Evidence photos |
| `status` | VARCHAR(20) | 'open', 'investigating', 'resolved', 'rejected' |
| `admin_notes` | TEXT | Admin notes |
| `resolved_at` | TIMESTAMP | Resolution time |
| `created_at` | TIMESTAMP | Creation time |

---

### 15. `notifications`
User notifications.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `user_id` | BIGINT | FK to users (CASCADE) |
| `booking_id` | BIGINT | FK to bookings |
| `type` | VARCHAR(30) | Notification type |
| `title` | VARCHAR | Notification title |
| `message` | TEXT | Message content |
| `read` | BOOLEAN | Read status |
| `created_at` | TIMESTAMP | Creation time |

**Indexes**: `user_id`, `read`, `type`

---

### 16. `audit_logs`
Audit trail for system actions.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `user_id` | BIGINT | FK to users |
| `action` | VARCHAR | Action performed |
| `entity_type` | VARCHAR | Entity type |
| `entity_id` | BIGINT | Entity ID |
| `old_value` | TEXT | Previous value |
| `new_value` | TEXT | New value |
| `ip_address` | VARCHAR | Request IP |
| `created_at` | TIMESTAMP | Creation time |

**Indexes**: `user_id`, `action`, `entity_type+entity_id`

---

### 17. `password_resets`
Password reset tokens.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `user_id` | BIGINT | FK to users (CASCADE) |
| `token` | VARCHAR | Unique reset token |
| `expires_at` | TIMESTAMPTZ | Token expiration |
| `created_at` | TIMESTAMPTZ | Creation time |
| `used_at` | TIMESTAMPTZ | When token was used |

**Indexes**: `token`, `user_id`, `expires_at`

---

## Adding New Migrations

### Step-by-Step Process

1. **Create a new SQL file**
   ```bash
   # In backend/internal/migrate/
   touch 002_add_new_feature.sql
   ```

2. **Write your migration SQL**
   ```sql
   -- Add your new column/table
   ALTER TABLE users ADD COLUMN avatar_url VARCHAR;
   
   -- Create indexes if needed
   CREATE INDEX idx_users_avatar_url ON users(avatar_url);
   ```

3. **Rebuild the application**
   ```bash
   cd backend
   go build ./...
   ```

4. **Run the application**
   - Migration runs automatically on startup
   - Check logs for: `Applying 002_add_new_feature...`

### Best Practices

- **Always use IF NOT EXISTS / IF EXISTS** for idempotent operations
- **One change per migration file** (use numbered files like 002, 003, 004)
- **Never modify existing migrations** - create a new one instead
- **Test migrations locally** before deploying
- **Use transactions** - automatically handled by the migration system
- **Add indexes** for columns used in WHERE/JOIN clauses
- **Use descriptive comments** with `--` prefix

### Migration Naming Convention

- Format: `<NNN>_<description>.sql`
- NNN: Zero-padded 3-digit number (001, 002, 003...)
- description: Snake_case description of changes

Examples:
- `001_schema.sql` - Initial schema
- `002_add_user_avatar.sql` - Add avatar column
- `003_add_booking_index.sql` - Add performance index

---

## Important Notes

### Cascade Deletes
- `users` → `notifications`, `password_resets`
- `vendors` → `equipment`, `vendor_wallets`, `vendor_bank_accounts`, `withdrawal_requests`
- `bookings` → `booking_handovers`, `damage_disputes`
- `vendor_wallets` → `wallet_transactions`

### Soft Deletes
Tables with soft deletes (use `deleted_at` column):
- `users`
- `vendors`
- `equipment_categories`
- `equipment`
- `generators`

### Status Values

**Booking Status:**
- `requested` - Initial request
- `confirmed` - Vendor accepted
- `dispatched` - Equipment shipped
- `delivered` - Delivered to customer
- `completed` - Rental period ended
- `cancelled` - Booking cancelled

**Payment Status:**
- `pending` - Awaiting payment
- `completed` - Payment successful
- `failed` - Payment failed
- `refunded` - Payment refunded

**Withdrawal Status:**
- `pending` - Awaiting approval
- `approved` - Approved for processing
- `rejected` - Rejected by admin
- `processing` - Being processed
- `completed` - Transfer complete

**Equipment Availability:**
- `available` - Available for rent
- `rented` - Currently rented
- `maintenance` - Under maintenance
- `reserved` - Reserved for booking

---

## Database Support

Currently tested with:
- **PostgreSQL** (recommended for production)
- SQLite for development

---

## Troubleshooting

### Migration fails on startup
1. Check the logs for specific SQL errors
2. Verify database connection in `.env`
3. Ensure user has CREATE TABLE permissions

### Need to rerun a migration
```sql
-- Manual intervention required
DELETE FROM schema_migrations WHERE version = '002_some_migration';
-- Restart application to re-run
```

### Schema drift detected
If you see unexpected behavior, the schema may have drifted from migrations:
```sql
-- Compare current schema vs migrations
SELECT version FROM schema_migrations ORDER BY version;
```

---

## Contact & Support

For questions or issues with migrations:
1. Check existing migrations for patterns
2. Review the `migrate.go` source code
3. Contact the development team
