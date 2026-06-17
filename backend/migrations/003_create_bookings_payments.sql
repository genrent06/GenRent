-- +goose Up
CREATE TABLE bookings (
    id               BIGSERIAL PRIMARY KEY,
    customer_id      BIGINT         NOT NULL REFERENCES users(id),
    generator_id     BIGINT         NOT NULL REFERENCES generators(id),
    start_date       DATE           NOT NULL,
    end_date         DATE           NOT NULL,
    total_price      FLOAT          NOT NULL,
    advance_amount   FLOAT          NOT NULL,
    advance_paid     BOOLEAN        NOT NULL DEFAULT FALSE,
    status           VARCHAR(30)    NOT NULL DEFAULT 'requested',
    address          TEXT           NOT NULL,
    notes            TEXT,
    delivery_otp     VARCHAR(10),
    otp_verified     BOOLEAN        NOT NULL DEFAULT FALSE,
    cancel_reason    TEXT,
    customer_rating  INT,
    customer_review  TEXT,
    accepted_at      TIMESTAMPTZ,
    dispatched_at    TIMESTAMPTZ,
    delivered_at     TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookings_customer_id ON bookings (customer_id);
CREATE INDEX idx_bookings_status      ON bookings (status);
CREATE INDEX idx_bookings_generator   ON bookings (generator_id);

CREATE TABLE payments (
    id             BIGSERIAL PRIMARY KEY,
    booking_id     BIGINT         NOT NULL REFERENCES bookings(id),
    total_amount   FLOAT          NOT NULL,
    advance_amount FLOAT          NOT NULL,
    vendor_amount  FLOAT          NOT NULL,
    platform_fee   FLOAT          NOT NULL,
    method         VARCHAR(30)    NOT NULL,
    status         VARCHAR(20)    NOT NULL DEFAULT 'pending',
    transaction_id VARCHAR(100)   NOT NULL UNIQUE,
    paid_at        TIMESTAMPTZ,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE vendor_wallets (
    id           BIGSERIAL PRIMARY KEY,
    vendor_id    BIGINT   NOT NULL UNIQUE REFERENCES vendors(id),
    balance      FLOAT    NOT NULL DEFAULT 0,
    hold_balance FLOAT    NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wallet_transactions (
    id          BIGSERIAL PRIMARY KEY,
    wallet_id   BIGINT         NOT NULL REFERENCES vendor_wallets(id),
    booking_id  BIGINT         REFERENCES bookings(id),
    amount      FLOAT          NOT NULL,
    type        VARCHAR(20)    NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions (wallet_id);

CREATE TABLE notifications (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT         NOT NULL REFERENCES users(id),
    booking_id BIGINT,
    type       VARCHAR(50)    NOT NULL,
    title      VARCHAR(255)   NOT NULL,
    message    TEXT           NOT NULL,
    read       BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_read ON notifications (user_id, read);

CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT,
    action      VARCHAR(100)   NOT NULL,
    entity_type VARCHAR(50)    NOT NULL,
    entity_id   BIGINT         NOT NULL,
    old_value   TEXT,
    new_value   TEXT,
    ip_address  VARCHAR(45),
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_entity ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_user   ON audit_logs (user_id);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS vendor_wallets;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS bookings;

-- +goose Down
-- Downgrade not implemented
