-- Sample seed data for development and testing

-- Insert sample customers
INSERT INTO customers (id, name, industry, description, website, region, created_at, updated_at) VALUES
('11111111-1111-1111-1111-111111111111', 'Acme Corporation', 'Technology', 'Leading provider of innovative solutions', 'https://acme.com', 'North America', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('22222222-2222-2222-2222-222222222222', 'Global Finance Inc', 'Finance', 'International financial services', 'https://globalfinance.com', 'Europe', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('33333333-3333-3333-3333-333333333333', 'Tech Solutions Ltd', 'Technology', 'Enterprise software solutions', 'https://techsolutions.com', 'Asia', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Insert sample contacts
INSERT INTO contacts (id, customer_id, first_name, last_name, email, phone, job_title, created_at, updated_at) VALUES
('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'John', 'Smith', 'john.smith@acme.com', '+1-555-0101', 'CEO', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'Jane', 'Doe', 'jane.doe@acme.com', '+1-555-0102', 'CTO', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('66666666-6666-6666-6666-666666666666', '22222222-2222-2222-2222-222222222222', 'Bob', 'Johnson', 'bob.johnson@globalfinance.com', '+44-20-5555-0101', 'CFO', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Insert sample projects
INSERT INTO projects (id, customer_id, name, summary, budget, spent, status, start_date, end_date, created_at, updated_at) VALUES
('77777777-7777-7777-7777-777777777777', '11111111-1111-1111-1111-111111111111', 'Website Redesign', 'Complete overhaul of corporate website', 50000, 15000, 'Active', '2024-01-01', '2024-06-30', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('88888888-8888-8888-8888-888888888888', '22222222-2222-2222-2222-222222222222', 'Mobile App Development', 'New mobile banking application', 150000, 75000, 'Active', '2024-02-01', '2024-12-31', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('99999999-9999-9999-9999-999999999999', '33333333-3333-3333-3333-333333333333', 'Cloud Migration', 'Migrate infrastructure to cloud', 200000, 50000, 'Planning', '2024-06-01', '2025-06-30', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Insert sample offers
INSERT INTO offers (id, customer_id, project_id, title, total_amount, valid_until, phase, created_at, updated_at) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '77777777-7777-7777-7777-777777777777', 'Website Redesign Proposal', 50000, '2024-03-31', 'Won', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222222', '88888888-8888-8888-8888-888888888888', 'Mobile App Development Phase 1', 80000, '2024-12-31', 'Negotiation', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('cccccccc-cccc-cccc-cccc-cccccccccccc', '33333333-3333-3333-3333-333333333333', NULL, 'Consulting Services Q1', 25000, '2024-04-30', 'Sent', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Insert sample offer items
INSERT INTO offer_items (id, offer_id, name, description, quantity, unit_price, created_at, updated_at) VALUES
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'UI/UX Design', 'Complete design system and wireframes', 1, 15000, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Frontend Development', 'React-based frontend implementation', 200, 150, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('ffffffff-ffff-ffff-ffff-ffffffffffff', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Backend Development', 'API and database implementation', 100, 200, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Insert sample activities
INSERT INTO activities (id, target_type, target_id, title, body, occurred_at, creator_name, created_at, updated_at) VALUES
('10000000-0000-0000-0000-000000000001', 'Customer', '11111111-1111-1111-1111-111111111111', 'Customer created', 'New customer Acme Corporation was created', CURRENT_TIMESTAMP, 'System', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('10000000-0000-0000-0000-000000000002', 'Project', '77777777-7777-7777-7777-777777777777', 'Project started', 'Website Redesign project has been initiated', CURRENT_TIMESTAMP, 'John Smith', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('10000000-0000-0000-0000-000000000003', 'Offer', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Offer won', 'Website Redesign Proposal was accepted by customer', CURRENT_TIMESTAMP, 'Jane Doe', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

