-- +goose Up
CREATE TABLE withdrawal_requests (
    id           BIGSERIAL PRIMARY KEY,
    vendor_id    BIGINT         NOT NULL REFERENCES vendors(id),
    amount       FLOAT          NOT NULL,
    status       VARCHAR(20)    NOT NULL DEFAULT 'pending',
    bank_name    VARCHAR(100),
    account_no   VARCHAR(50),
    ifsc         VARCHAR(20),
    account_name VARCHAR(100),
    admin_note   TEXT,
    processed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_vendor_id ON withdrawal_requests (vendor_id);
CREATE INDEX idx_withdrawals_status    ON withdrawal_requests (status);

-- +goose Down
DROP TABLE IF EXISTS withdrawal_requests;

-- +goose Down
-- Downgrade not implemented
