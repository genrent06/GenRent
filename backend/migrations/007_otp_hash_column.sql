-- +goose Up
-- Expand otp_code column to hold SHA256 hex (64 chars)
ALTER TABLE withdrawal_requests ALTER COLUMN otp_code TYPE VARCHAR(64);

-- +goose Down
ALTER TABLE withdrawal_requests ALTER COLUMN otp_code TYPE VARCHAR(6);

-- +goose Down
-- Downgrade not implemented
