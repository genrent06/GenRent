-- +goose Up

-- Create equipment_categories table
CREATE TABLE equipment_categories (
    id                BIGSERIAL PRIMARY KEY,
    name              VARCHAR(255)   NOT NULL,
    parent_category_id BIGINT         REFERENCES equipment_categories(id),
    description       TEXT,
    icon_url          TEXT,
    display_order     INT            NOT NULL DEFAULT 0,
    deleted_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_name       ON equipment_categories (name);
CREATE INDEX idx_categories_parent_id  ON equipment_categories (parent_category_id);
CREATE INDEX idx_categories_deleted_at ON equipment_categories (deleted_at);

-- Create equipment table (generalized from generators)
CREATE TABLE equipment (
    id                  BIGSERIAL PRIMARY KEY,
    vendor_id           BIGINT         NOT NULL REFERENCES vendors(id),
    category_id         BIGINT         NOT NULL REFERENCES equipment_categories(id),
    name                VARCHAR(255)   NOT NULL,
    brand               VARCHAR(100),
    model               VARCHAR(100),
    description         TEXT,
    daily_price         FLOAT          NOT NULL,
    weekly_price        FLOAT,
    monthly_price       FLOAT,
    location            VARCHAR(255)   NOT NULL,
    city                VARCHAR(100)   NOT NULL,
    latitude            FLOAT          NOT NULL DEFAULT 0,
    longitude           FLOAT          NOT NULL DEFAULT 0,
    availability_status VARCHAR(20)    NOT NULL DEFAULT 'available',
    reservation_expiry  TIMESTAMPTZ,
    image_url           TEXT,
    specs               JSONB,
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_equipment_vendor_id    ON equipment (vendor_id);
CREATE INDEX idx_equipment_category_id  ON equipment (category_id);
CREATE INDEX idx_equipment_city         ON equipment (city);
CREATE INDEX idx_equipment_geo          ON equipment (latitude, longitude);
CREATE INDEX idx_equipment_status       ON equipment (availability_status);
CREATE INDEX idx_equipment_deleted_at   ON equipment (deleted_at);

-- Seed default equipment categories
INSERT INTO equipment_categories (name, description, display_order, created_at, updated_at) VALUES
    ('Power Equipment', 'Generators, distribution panels, cables', 1, NOW(), NOW()),
    ('Generators', 'Diesel and portable generators', 1, NOW(), NOW()),
    ('Tower Lights', 'Mobile tower lighting solutions', 2, NOW(), NOW()),
    ('Distribution Panels', 'Electrical distribution equipment', 3, NOW(), NOW()),
    ('Cables', 'Industrial power cables and extensions', 4, NOW(), NOW()),
    ('Construction Equipment', 'Excavators, loaders, compactors', 2, NOW(), NOW()),
    ('Excavators', 'Hydraulic and mini excavators', 1, NOW(), NOW()),
    ('Backhoe Loaders', 'Multi-function backhoe loaders', 2, NOW(), NOW()),
    ('Concrete Mixers', 'Portable and drum concrete mixers', 3, NOW(), NOW()),
    ('Compactors', 'Plate and vibratory compactors', 4, NOW(), NOW()),
    ('Road Rollers', 'Static and vibratory road rollers', 5, NOW(), NOW()),
    ('Material Handling', 'Forklifts, cranes, lifts', 3, NOW(), NOW()),
    ('Forklifts', '2T, 3T, and heavy-duty forklifts', 1, NOW(), NOW()),
    ('Hydra Cranes', 'Mobile and self-propelled cranes', 2, NOW(), NOW()),
    ('Boom Lifts', 'Articulating and telescopic boom lifts', 3, NOW(), NOW()),
    ('Scissor Lifts', 'Vertical and rough-terrain scissor lifts', 4, NOW(), NOW()),
    ('Site Equipment', 'Compressors, pumps, welders', 4, NOW(), NOW()),
    ('Air Compressors', 'Portable and stationary air compressors', 1, NOW(), NOW()),
    ('Water Pumps', 'Submersible and centrifugal pumps', 2, NOW(), NOW()),
    ('Welding Machines', 'Arc and inverter welding machines', 3, NOW(), NOW()),
    ('Cutting Machines', 'Concrete and metal cutting machines', 4, NOW(), NOW());

-- Migrate existing generators to equipment table
-- First, ensure "Generators" category exists
INSERT INTO equipment (vendor_id, category_id, name, brand, model, description, daily_price, weekly_price, monthly_price, location, city, latitude, longitude, availability_status, image_url, created_at, updated_at)
SELECT 
    g.vendor_id,
    (SELECT id FROM equipment_categories WHERE name = 'Generators' LIMIT 1),
    g.name,
    g.brand,
    CONCAT(g.capacity_kva, 'KVA'),
    COALESCE(g.description, ''),
    g.price_per_day,
    NULL,
    g.price_per_month,
    g.location,
    g.city,
    g.latitude,
    g.longitude,
    g.availability_status,
    g.image_url,
    g.created_at,
    g.updated_at
FROM generators g
WHERE g.deleted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS idx_equipment_deleted_at;
DROP INDEX IF EXISTS idx_equipment_status;
DROP INDEX IF EXISTS idx_equipment_geo;
DROP INDEX IF EXISTS idx_equipment_city;
DROP INDEX IF EXISTS idx_equipment_category_id;
DROP INDEX IF EXISTS idx_equipment_vendor_id;
DROP TABLE IF EXISTS equipment;

DROP INDEX IF EXISTS idx_categories_deleted_at;
DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_categories_name;
DROP TABLE IF EXISTS equipment_categories;
