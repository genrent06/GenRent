-- +goose Up
CREATE TABLE vendors (
    id                    BIGSERIAL PRIMARY KEY,
    user_id               BIGINT         NOT NULL REFERENCES users(id),
    company_name          VARCHAR(255)   NOT NULL,
    address               TEXT,
    city                  VARCHAR(100)   NOT NULL DEFAULT '',
    phone                 VARCHAR(20),
    description           TEXT,
    verified              BOOLEAN        NOT NULL DEFAULT FALSE,
    security_deposit      FLOAT          NOT NULL DEFAULT 0,
    security_deposit_paid BOOLEAN        NOT NULL DEFAULT FALSE,
    reliability_score     FLOAT          NOT NULL DEFAULT 0,
    total_bookings        INT            NOT NULL DEFAULT 0,
    successful_deliveries INT            NOT NULL DEFAULT 0,
    cancelled_bookings    INT            NOT NULL DEFAULT 0,
    average_rating        FLOAT          NOT NULL DEFAULT 0,
    total_ratings         INT            NOT NULL DEFAULT 0,
    avg_response_minutes  FLOAT          NOT NULL DEFAULT 0,
    risk_score            FLOAT          NOT NULL DEFAULT 0,
    latitude              FLOAT          NOT NULL DEFAULT 0,
    longitude             FLOAT          NOT NULL DEFAULT 0,
    deleted_at            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vendors_user_id    ON vendors (user_id);
CREATE INDEX idx_vendors_verified   ON vendors (verified);
CREATE INDEX idx_vendors_deleted_at ON vendors (deleted_at);

CREATE TABLE generators (
    id                  BIGSERIAL PRIMARY KEY,
    vendor_id           BIGINT         NOT NULL REFERENCES vendors(id),
    name                VARCHAR(255)   NOT NULL,
    capacity_kva        INT            NOT NULL,
    price_per_day       FLOAT          NOT NULL,
    price_per_month     FLOAT,
    fuel_type           VARCHAR(50)    NOT NULL DEFAULT 'diesel',
    brand               VARCHAR(100),
    location            VARCHAR(255)   NOT NULL,
    city                VARCHAR(100)   NOT NULL,
    latitude            FLOAT          NOT NULL DEFAULT 0,
    longitude           FLOAT          NOT NULL DEFAULT 0,
    availability_status VARCHAR(20)    NOT NULL DEFAULT 'available',
    reservation_expiry  TIMESTAMPTZ,
    description         TEXT,
    image_url           TEXT,
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_generators_vendor_id    ON generators (vendor_id);
CREATE INDEX idx_generators_city         ON generators (city);
CREATE INDEX idx_generators_geo          ON generators (latitude, longitude);
CREATE INDEX idx_generators_status       ON generators (availability_status);
CREATE INDEX idx_generators_deleted_at   ON generators (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS generators;
DROP TABLE IF EXISTS vendors;

-- +goose Down
-- Downgrade not implemented
