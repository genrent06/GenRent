-- +goose Up
-- Allow generator_id to be NULL so equipment-only bookings work without a generator reference
ALTER TABLE bookings ALTER COLUMN generator_id DROP NOT NULL;

-- +goose Down
-- Revert: this could break data if there are bookings without a generator_id
-- Only do this carefully if no such rows exist
ALTER TABLE bookings ALTER COLUMN generator_id SET NOT NULL;
