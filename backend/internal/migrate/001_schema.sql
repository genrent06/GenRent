-- Users table
CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR NOT NULL,
	email VARCHAR NOT NULL UNIQUE,
	phone VARCHAR NOT NULL,
	password VARCHAR NOT NULL,
	role VARCHAR(20) DEFAULT 'customer',
	risk_score FLOAT DEFAULT 0,
	deleted_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vendors table
CREATE TABLE IF NOT EXISTS vendors (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	company_name VARCHAR NOT NULL,
	address TEXT,
	city VARCHAR NOT NULL,
	latitude FLOAT DEFAULT 0,
	longitude FLOAT DEFAULT 0,
	description TEXT,
	phone VARCHAR,
	verified BOOLEAN DEFAULT false,
	security_deposit FLOAT DEFAULT 0,
	security_deposit_paid BOOLEAN DEFAULT false,
	reliability_score FLOAT DEFAULT 5.0,
	total_bookings INT DEFAULT 0,
	successful_deliveries INT DEFAULT 0,
	cancelled_bookings INT DEFAULT 0,
	average_rating FLOAT DEFAULT 0,
	total_ratings INT DEFAULT 0,
	avg_response_minutes FLOAT DEFAULT 0,
	risk_score FLOAT DEFAULT 0,
	deleted_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Equipment categories table
CREATE TABLE IF NOT EXISTS equipment_categories (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR NOT NULL,
	parent_category_id BIGINT REFERENCES equipment_categories(id),
	description TEXT,
	icon_url VARCHAR,
	display_order INT DEFAULT 0,
	deleted_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Equipment table
CREATE TABLE IF NOT EXISTS equipment (
	id BIGSERIAL PRIMARY KEY,
	vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
	category_id BIGINT NOT NULL REFERENCES equipment_categories(id),
	name VARCHAR NOT NULL,
	brand VARCHAR,
	model VARCHAR,
	description TEXT,
	daily_price FLOAT NOT NULL,
	weekly_price FLOAT,
	monthly_price FLOAT,
	mobilization_fee FLOAT DEFAULT 0,
	demobilization_fee FLOAT DEFAULT 0,
	total_quantity INT DEFAULT 1,
	available_quantity INT DEFAULT 1,
	location VARCHAR NOT NULL,
	city VARCHAR NOT NULL,
	latitude FLOAT DEFAULT 0,
	longitude FLOAT DEFAULT 0,
	availability_status VARCHAR(20) DEFAULT 'available',
	reservation_expiry TIMESTAMP,
	image_url VARCHAR,
	specs JSONB DEFAULT '{}'::jsonb,
	deleted_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Generators table (legacy)
CREATE TABLE IF NOT EXISTS generators (
	id BIGSERIAL PRIMARY KEY,
	vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
	name VARCHAR NOT NULL,
	capacity_kva INT NOT NULL,
	price_per_day FLOAT NOT NULL,
	price_per_month FLOAT,
	fuel_type VARCHAR DEFAULT 'diesel',
	brand VARCHAR,
	location VARCHAR NOT NULL,
	city VARCHAR NOT NULL,
	latitude FLOAT DEFAULT 0,
	longitude FLOAT DEFAULT 0,
	availability_status VARCHAR(20) DEFAULT 'available',
	reservation_expiry TIMESTAMP,
	description TEXT,
	image_url VARCHAR,
	deleted_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bookings table
CREATE TABLE IF NOT EXISTS bookings (
	id BIGSERIAL PRIMARY KEY,
	customer_id BIGINT NOT NULL REFERENCES users(id),
	generator_id BIGINT REFERENCES generators(id),
	equipment_id BIGINT REFERENCES equipment(id),
	category_id BIGINT REFERENCES equipment_categories(id),
	start_date TIMESTAMP NOT NULL,
	end_date TIMESTAMP NOT NULL,
	total_price FLOAT,
	rental_price FLOAT,
	mobilization_fee FLOAT DEFAULT 0,
	demobilization_fee FLOAT DEFAULT 0,
	advance_amount FLOAT,
	advance_paid BOOLEAN DEFAULT false,
	status VARCHAR(30) DEFAULT 'requested',
	address TEXT NOT NULL,
	notes TEXT,
	delivery_otp VARCHAR(6),
	otp_verified BOOLEAN DEFAULT false,
	accepted_at TIMESTAMP,
	dispatched_at TIMESTAMP,
	delivered_at TIMESTAMP,
	completed_at TIMESTAMP,
	customer_rating INT DEFAULT 0,
	customer_review TEXT,
	cancel_reason TEXT,
	return_initiated_at TIMESTAMP,
	return_otp VARCHAR(6),
	return_otp_verified BOOLEAN DEFAULT false,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
	id BIGSERIAL PRIMARY KEY,
	booking_id BIGINT NOT NULL UNIQUE REFERENCES bookings(id),
	total_amount FLOAT,
	advance_amount FLOAT,
	vendor_amount FLOAT,
	platform_fee FLOAT,
	method VARCHAR(20),
	status VARCHAR(20) DEFAULT 'pending',
	transaction_id VARCHAR,
	paid_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vendor wallets table
CREATE TABLE IF NOT EXISTS vendor_wallets (
	id BIGSERIAL PRIMARY KEY,
	vendor_id BIGINT NOT NULL UNIQUE REFERENCES vendors(id) ON DELETE CASCADE,
	balance FLOAT DEFAULT 0,
	hold_balance FLOAT DEFAULT 0,
	withdrawal_hold_balance FLOAT DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Wallet transactions table
CREATE TABLE IF NOT EXISTS wallet_transactions (
	id BIGSERIAL PRIMARY KEY,
	wallet_id BIGINT NOT NULL REFERENCES vendor_wallets(id) ON DELETE CASCADE,
	booking_id BIGINT,
	amount FLOAT,
	type VARCHAR(30),
	description TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vendor bank accounts table
CREATE TABLE IF NOT EXISTS vendor_bank_accounts (
	id BIGSERIAL PRIMARY KEY,
	vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
	bank_name VARCHAR,
	account_no VARCHAR,
	ifsc VARCHAR,
	account_name VARCHAR,
	is_primary BOOLEAN DEFAULT false,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Withdrawal requests table
CREATE TABLE IF NOT EXISTS withdrawal_requests (
	id BIGSERIAL PRIMARY KEY,
	vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
	bank_account_id BIGINT,
	amount FLOAT NOT NULL,
	status VARCHAR(20) DEFAULT 'pending',
	bank_name VARCHAR,
	account_no VARCHAR,
	ifsc VARCHAR,
	account_name VARCHAR,
	admin_note TEXT,
	processed_at TIMESTAMP,
	otp_code VARCHAR,
	otp_expires_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Platform revenue table
CREATE TABLE IF NOT EXISTS platform_revenues (
	id BIGSERIAL PRIMARY KEY,
	payment_id BIGINT NOT NULL REFERENCES payments(id),
	booking_id BIGINT NOT NULL REFERENCES bookings(id),
	amount FLOAT,
	type VARCHAR(20),
	description TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Booking handovers table
CREATE TABLE IF NOT EXISTS booking_handovers (
	id BIGSERIAL PRIMARY KEY,
	booking_id BIGINT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
	type VARCHAR(20) NOT NULL,
	photo_urls JSONB DEFAULT '[]'::jsonb,
	checklist JSONB DEFAULT '{}'::jsonb,
	notes TEXT,
	uploaded_by BIGINT,
	verified_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Damage disputes table
CREATE TABLE IF NOT EXISTS damage_disputes (
	id BIGSERIAL PRIMARY KEY,
	booking_id BIGINT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
	raised_by BIGINT NOT NULL REFERENCES users(id),
	description TEXT NOT NULL,
	claimed_amount FLOAT DEFAULT 0,
	photo_urls JSONB DEFAULT '[]'::jsonb,
	status VARCHAR(20) DEFAULT 'open',
	admin_notes TEXT,
	resolved_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	booking_id BIGINT REFERENCES bookings(id),
	type VARCHAR(30),
	title VARCHAR,
	message TEXT,
	read BOOLEAN DEFAULT false,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT REFERENCES users(id),
	action VARCHAR NOT NULL,
	entity_type VARCHAR,
	entity_id BIGINT,
	old_value TEXT,
	new_value TEXT,
	ip_address VARCHAR,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_risk_score ON users(risk_score);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE INDEX IF NOT EXISTS idx_vendors_user_id ON vendors(user_id);
CREATE INDEX IF NOT EXISTS idx_vendors_city ON vendors(city);
CREATE INDEX IF NOT EXISTS idx_vendors_risk_score ON vendors(risk_score);
CREATE INDEX IF NOT EXISTS idx_vendors_deleted_at ON vendors(deleted_at);

CREATE INDEX IF NOT EXISTS idx_equipment_categories_name ON equipment_categories(name);
CREATE INDEX IF NOT EXISTS idx_equipment_categories_parent_id ON equipment_categories(parent_category_id);
CREATE INDEX IF NOT EXISTS idx_equipment_categories_deleted_at ON equipment_categories(deleted_at);

CREATE INDEX IF NOT EXISTS idx_equipment_vendor_id ON equipment(vendor_id);
CREATE INDEX IF NOT EXISTS idx_equipment_category_id ON equipment(category_id);
CREATE INDEX IF NOT EXISTS idx_equipment_city ON equipment(city);
CREATE INDEX IF NOT EXISTS idx_equipment_status ON equipment(availability_status);
CREATE INDEX IF NOT EXISTS idx_equipment_deleted_at ON equipment(deleted_at);

CREATE INDEX IF NOT EXISTS idx_generators_vendor_id ON generators(vendor_id);
CREATE INDEX IF NOT EXISTS idx_generators_city ON generators(city);
CREATE INDEX IF NOT EXISTS idx_generators_status ON generators(availability_status);
CREATE INDEX IF NOT EXISTS idx_generators_deleted_at ON generators(deleted_at);

CREATE INDEX IF NOT EXISTS idx_bookings_customer_id ON bookings(customer_id);
CREATE INDEX IF NOT EXISTS idx_bookings_generator_id ON bookings(generator_id);
CREATE INDEX IF NOT EXISTS idx_bookings_equipment_id ON bookings(equipment_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);

CREATE INDEX IF NOT EXISTS idx_payments_booking_id ON payments(booking_id);

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_booking_id ON wallet_transactions(booking_id);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications(read);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);

-- Password resets table
CREATE TABLE IF NOT EXISTS password_resets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);
