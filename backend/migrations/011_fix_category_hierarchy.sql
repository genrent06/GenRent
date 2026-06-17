-- +goose Up

-- Update subcategories of 'Power Equipment'
UPDATE equipment_categories
SET parent_category_id = (SELECT id FROM equipment_categories WHERE name = 'Power Equipment' AND parent_category_id IS NULL LIMIT 1)
WHERE name IN ('Generators', 'Tower Lights', 'Distribution Panels', 'Cables');

-- Update subcategories of 'Construction Equipment'
UPDATE equipment_categories
SET parent_category_id = (SELECT id FROM equipment_categories WHERE name = 'Construction Equipment' AND parent_category_id IS NULL LIMIT 1)
WHERE name IN ('Excavators', 'Backhoe Loaders', 'Concrete Mixers', 'Compactors', 'Road Rollers');

-- Update subcategories of 'Material Handling'
UPDATE equipment_categories
SET parent_category_id = (SELECT id FROM equipment_categories WHERE name = 'Material Handling' AND parent_category_id IS NULL LIMIT 1)
WHERE name IN ('Forklifts', 'Hydra Cranes', 'Boom Lifts', 'Scissor Lifts');

-- Update subcategories of 'Site Equipment'
UPDATE equipment_categories
SET parent_category_id = (SELECT id FROM equipment_categories WHERE name = 'Site Equipment' AND parent_category_id IS NULL LIMIT 1)
WHERE name IN ('Air Compressors', 'Water Pumps', 'Welding Machines', 'Cutting Machines');

-- +goose Down
UPDATE equipment_categories
SET parent_category_id = NULL
WHERE name IN ('Generators', 'Tower Lights', 'Distribution Panels', 'Cables', 
               'Excavators', 'Backhoe Loaders', 'Concrete Mixers', 'Compactors', 'Road Rollers',
               'Forklifts', 'Hydra Cranes', 'Boom Lifts', 'Scissor Lifts',
               'Air Compressors', 'Water Pumps', 'Welding Machines', 'Cutting Machines');
