-- Seed Equipment Categories
-- Creates standard categories for the GenRent platform

-- Clear any existing data to avoid conflicts
-- TRUNCATE equipment_categories CASCADE;

-- Root Categories (Parent Categories)
INSERT INTO equipment_categories (id, name, description, icon_url, display_order, parent_category_id, created_at, updated_at) VALUES
(1, 'Power Equipment', 'Generators, UPS, power distribution equipment', 'https://images.unsplash.com/photo-1621905252507-b35492cc74b4?w=200', 1, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(6, 'Construction Equipment', 'Heavy construction machinery', 'https://images.unsplash.com/photo-1581094794329-c8112a89af12?w=200', 2, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(12, 'Material Handling', 'Forklifts, cranes, lifts', 'https://images.unsplash.com/photo-1586528116311-ad8dd3c8310d?w=200', 3, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(17, 'Site Equipment', 'Site support equipment', 'https://images.unsplash.com/photo-1504307651254-35680f356dfd?w=200', 4, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Subcategories under Power Equipment
INSERT INTO equipment_categories (id, name, description, icon_url, display_order, parent_category_id, created_at, updated_at) VALUES
(2, 'Generators', 'Diesel, gas, and electric generators', 'https://images.unsplash.com/photo-1621905252507-b35492cc74b4?w=200', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(3, 'Tower Lights', 'Portable lighting towers', 'https://images.unsplash.com/photo-1565514020125-1d2a606072db?w=200', 2, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4, 'Distribution Panels', 'Power distribution boards and panels', 'https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=200', 3, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(5, 'Cables', 'Power cables and extension cords', 'https://images.unsplash.com/photo-1544724569-5f546fd6f2b5?w=200', 4, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Subcategories under Construction Equipment
INSERT INTO equipment_categories (id, name, description, icon_url, display_order, parent_category_id, created_at, updated_at) VALUES
(7, 'Excavators', 'Digging and excavation machinery', 'https://images.unsplash.com/photo-1581094794329-c8112a89af12?w=200', 1, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(8, 'Backhoe Loaders', 'Multipurpose backhoe loaders', 'https://images.unsplash.com/photo-1601585115313-19de85f47bbc?w=200', 2, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(9, 'Concrete Mixers', 'Concrete mixing equipment', 'https://images.unsplash.com/photo-1616401784845-180882ba9ba8?w=200', 3, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(10, 'Compactors', 'Soil and waste compactors', 'https://images.unsplash.com/photo-1509695507497-903c140c43b0?w=200', 4, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(11, 'Road Rollers', 'Road construction rollers', 'https://images.unsplash.com/photo-1568605117036-5fe5e7bab0b7?w=200', 5, 6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Subcategories under Material Handling
INSERT INTO equipment_categories (id, name, description, icon_url, display_order, parent_category_id, created_at, updated_at) VALUES
(13, 'Forklifts', 'Material lifting forklifts', 'https://images.unsplash.com/photo-1586528116311-ad8dd3c8310d?w=200', 1, 12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(14, 'Hydra Cranes', 'Truck mounted cranes', 'https://images.unsplash.com/photo-1565466937379-6848e62d2f2e?w=200', 2, 12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(15, 'Boom Lifts', 'Elevated work platform boom lifts', 'https://images.unsplash.com/photo-1578670398628-e8c3ccfdc5c5?w=200', 3, 12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(16, 'Scissor Lifts', 'Vertical lifting platform', 'https://images.unsplash.com/photo-1596525644297-7f9db629ee5e?w=200', 4, 12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Subcategories under Site Equipment
INSERT INTO equipment_categories (id, name, description, icon_url, display_order, parent_category_id, created_at, updated_at) VALUES
(18, 'Air Compressors', 'Industrial air compressors', 'https://images.unsplash.com/photo-1535131749006-b7f58c99034b?w=200', 1, 17, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(19, 'Water Pumps', 'Water pumping equipment', 'https://images.unsplash.com/photo-1621905251918-48416bd85d5e?w=200', 2, 17, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(20, 'Welding Machines', 'Electric welding equipment', 'https://images.unsplash.com/photo-1504328345606-18bbc8c9d7d1?w=200', 3, 17, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(21, 'Cutting Machines', 'Metal cutting equipment', 'https://images.unsplash.com/photo-1580402437233-9cf84336e5f7?w=200', 4, 17, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Verify inserted data
SELECT
    c.id,
    c.name,
    p.name as parent_name,
    c.display_order
FROM equipment_categories c
LEFT JOIN equipment_categories p ON c.parent_category_id = p.id
ORDER BY COALESCE(c.parent_category_id, 0), c.display_order;
