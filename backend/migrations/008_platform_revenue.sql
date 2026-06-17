-- +goose Up
CREATE TABLE platform_revenue (
    id          BIGSERIAL PRIMARY KEY,
    payment_id  BIGINT      NOT NULL REFERENCES payments(id),
    booking_id  BIGINT      NOT NULL REFERENCES bookings(id),
    amount      FLOAT       NOT NULL,
    type        VARCHAR(20) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_platform_revenue_payment  ON platform_revenue (payment_id);
CREATE INDEX idx_platform_revenue_booking  ON platform_revenue (booking_id);
CREATE INDEX idx_platform_revenue_type     ON platform_revenue (type);

-- +goose Down
DROP TABLE IF EXISTS platform_revenue;
