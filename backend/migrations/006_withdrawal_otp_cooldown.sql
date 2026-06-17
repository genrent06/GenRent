-- +goose Up
ALTER TABLE withdrawal_requests ADD COLUMN otp_code      VARCHAR(6);
ALTER TABLE withdrawal_requests ADD COLUMN otp_expires_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE withdrawal_requests DROP COLUMN IF EXISTS otp_expires_at;
ALTER TABLE withdrawal_requests DROP COLUMN IF EXISTS otp_code;

-- +goose Down
-- Downgrade not implemented
