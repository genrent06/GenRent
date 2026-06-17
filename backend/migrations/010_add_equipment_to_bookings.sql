-- +goose Up

-- Add equipment_id column to bookings table (nullable for backward compatibility)
ALTER TABLE bookings ADD COLUMN equipment_id BIGINT REFERENCES equipment(id);

-- Migrate existing bookings: link to equipment migrated from generators
UPDATE bookings b 
SET equipment_id = (
    SELECT e.id 
    FROM equipment e 
    JOIN generators g ON e.model = CONCAT(g.capacity_kva, 'KVA') AND e.vendor_id = g.vendor_id
    WHERE g.id = b.generator_id
    LIMIT 1
);

-- Create index for new column
CREATE INDEX idx_bookings_equipment_id ON bookings (equipment_id);

-- Add equipment_category_id to vendor_wallets for easier category-based revenue tracking (optional)
ALTER TABLE bookings ADD COLUMN category_id BIGINT REFERENCES equipment_categories(id);

-- Auto-populate category_id from equipment
UPDATE bookings b
SET category_id = (
    SELECT e.category_id FROM equipment e WHERE e.id = b.equipment_id
);

-- +goose Down

ALTER TABLE bookings DROP COLUMN IF EXISTS category_id;
DROP INDEX IF EXISTS idx_bookings_equipment_id;
ALTER TABLE bookings DROP COLUMN IF EXISTS equipment_id;
