-- +goose Up

-- Add rental_price column to bookings (before mobilization fee was separate)
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS rental_price DOUBLE PRECISION NOT NULL DEFAULT 0;

-- Populate rental_price from existing bookings (where mobilization fees are 0, total_price = rental_price)
UPDATE bookings SET rental_price = total_price - COALESCE(mobilization_fee, 0) - COALESCE(demobilization_fee, 0) WHERE rental_price = 0;

-- +goose Down

ALTER TABLE bookings DROP COLUMN IF EXISTS rental_price;
