-- +goose Up

-- Feature 5: Delivery handover photos and checklists
CREATE TABLE booking_handovers (
    id          BIGSERIAL PRIMARY KEY,
    booking_id  BIGINT       NOT NULL REFERENCES bookings(id),
    type        VARCHAR(20)  NOT NULL CHECK (type IN ('delivery', 'return')),
    photo_urls  JSONB        NOT NULL DEFAULT '[]',
    checklist   JSONB        NOT NULL DEFAULT '{}',
    notes       TEXT,
    uploaded_by BIGINT       REFERENCES users(id),
    verified_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_handovers_booking_id ON booking_handovers (booking_id);
CREATE INDEX idx_handovers_type       ON booking_handovers (booking_id, type);

-- Feature 5: Damage disputes
CREATE TABLE damage_disputes (
    id             BIGSERIAL PRIMARY KEY,
    booking_id     BIGINT       NOT NULL REFERENCES bookings(id),
    raised_by      BIGINT       NOT NULL REFERENCES users(id),
    description    TEXT         NOT NULL,
    claimed_amount FLOAT        NOT NULL DEFAULT 0,
    photo_urls     JSONB        NOT NULL DEFAULT '[]',
    status         VARCHAR(20)  NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved', 'rejected')),
    admin_notes    TEXT,
    resolved_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_disputes_booking_id ON damage_disputes (booking_id);
CREATE INDEX idx_disputes_status     ON damage_disputes (status);

-- Feature 5: Return flow columns on bookings
ALTER TABLE bookings ADD COLUMN return_initiated_at TIMESTAMPTZ;
ALTER TABLE bookings ADD COLUMN return_otp          VARCHAR(6);
ALTER TABLE bookings ADD COLUMN return_otp_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE bookings DROP COLUMN IF EXISTS return_otp_verified;
ALTER TABLE bookings DROP COLUMN IF EXISTS return_otp;
ALTER TABLE bookings DROP COLUMN IF EXISTS return_initiated_at;
DROP TABLE IF EXISTS damage_disputes;
DROP TABLE IF EXISTS booking_handovers;
