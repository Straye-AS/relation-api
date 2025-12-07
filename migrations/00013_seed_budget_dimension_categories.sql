-- +goose Up
-- +goose StatementBegin

-- Seed budget dimension categories with construction-specific categories
-- Uses UPSERT pattern to safely handle re-runs and updates

INSERT INTO budget_dimension_categories (id, name, description, display_order, is_active, created_at, updated_at) VALUES
    ('steel_structure', 'Steel Structure', 'Primary steel framework and structural elements', 1, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('hybrid_structure', 'Hybrid Structure', 'Combined steel and other material structures', 2, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('roofing', 'Roofing', 'Roof installation, materials and labor', 3, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cladding', 'Cladding', 'Wall cladding and facade materials', 4, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('foundation', 'Foundation', 'Concrete foundation and groundwork', 5, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('assembly', 'Assembly', 'On-site assembly and installation labor', 6, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('transport', 'Transport', 'Delivery and logistics costs', 7, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('engineering', 'Engineering', 'Design and engineering services', 8, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('project_management', 'Project Management', 'PM overhead and coordination', 9, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('crane_rigging', 'Crane & Rigging', 'Crane rental and rigging services', 10, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('miscellaneous', 'Miscellaneous', 'Other uncategorized costs', 11, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('contingency', 'Contingency', 'Risk buffer and unforeseen costs', 12, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the seed data (only the specific IDs we inserted)
DELETE FROM budget_dimension_categories
WHERE id IN (
    'steel_structure',
    'hybrid_structure',
    'roofing',
    'cladding',
    'foundation',
    'assembly',
    'transport',
    'engineering',
    'project_management',
    'crane_rigging',
    'miscellaneous',
    'contingency'
);

-- +goose StatementEnd
