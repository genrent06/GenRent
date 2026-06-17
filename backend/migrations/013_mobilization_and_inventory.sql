-- +goose Up
-- Feature 3: Mobilization / Demobilization transport fees
ALTER TABLE equipment ADD COLUMN mobilization_fee   FLOAT NOT NULL DEFAULT 0;
ALTER TABLE equipment ADD COLUMN demobilization_fee FLOAT NOT NULL DEFAULT 0;

-- Feature 4: Multi-unit inventory management
ALTER TABLE equipment ADD COLUMN total_quantity     INT NOT NULL DEFAULT 1;
ALTER TABLE equipment ADD COLUMN available_quantity INT NOT NULL DEFAULT 1;

-- Sync available_quantity with current status for existing rows
UPDATE equipment SET available_quantity = 0 WHERE availability_status IN ('booked', 'reserved');

-- Booking table: store transport fees at booking time so price history is immutable
ALTER TABLE bookings ADD COLUMN mobilization_fee   FLOAT NOT NULL DEFAULT 0;
ALTER TABLE bookings ADD COLUMN demobilization_fee FLOAT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE bookings DROP COLUMN IF EXISTS demobilization_fee;
ALTER TABLE bookings DROP COLUMN IF EXISTS mobilization_fee;
ALTER TABLE equipment DROP COLUMN IF EXISTS available_quantity;
ALTER TABLE equipment DROP COLUMN IF EXISTS total_quantity;
ALTER TABLE equipment DROP COLUMN IF EXISTS demobilization_fee;
ALTER TABLE equipment DROP COLUMN IF EXISTS mobilization_fee;
