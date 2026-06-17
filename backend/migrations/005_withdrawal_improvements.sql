-- +goose Up

-- Add withdrawal_hold_balance to track money pending payout separately from available balance
ALTER TABLE vendor_wallets ADD COLUMN withdrawal_hold_balance FLOAT NOT NULL DEFAULT 0;

-- Saved bank accounts per vendor (verified before use)
CREATE TABLE vendor_bank_accounts (
    id           BIGSERIAL PRIMARY KEY,
    vendor_id    BIGINT         NOT NULL REFERENCES vendors(id),
    bank_name    VARCHAR(100)   NOT NULL,
    account_no   VARCHAR(50)    NOT NULL,
    ifsc         VARCHAR(20)    NOT NULL,
    account_name VARCHAR(100)   NOT NULL,
    is_primary   BOOLEAN        NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bank_accounts_vendor ON vendor_bank_accounts (vendor_id);

-- Add bank_account_id reference to withdrawal_requests (nullable — historical records kept inline)
ALTER TABLE withdrawal_requests ADD COLUMN bank_account_id BIGINT REFERENCES vendor_bank_accounts(id);

-- +goose Down
ALTER TABLE withdrawal_requests DROP COLUMN IF EXISTS bank_account_id;
DROP TABLE IF EXISTS vendor_bank_accounts;
ALTER TABLE vendor_wallets DROP COLUMN IF EXISTS withdrawal_hold_balance;

-- +goose Down
-- Downgrade not implemented
