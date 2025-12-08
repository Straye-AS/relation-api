-- ============================================================================
-- Straye Relation API - Comprehensive Seed Data
-- Norwegian Construction Industry Test Data
-- ============================================================================
-- Run with: make seed (local development only)
-- ============================================================================

-- ============================================================================
-- PART 1: CLEAR EXISTING DATA (FK-safe order)
-- ============================================================================

DELETE FROM notifications;
DELETE FROM audit_logs;
DELETE FROM activities;
DELETE FROM deal_stage_history;
DELETE FROM budget_dimensions;
DELETE FROM project_actual_costs;
DELETE FROM offer_items;
DELETE FROM files;
DELETE FROM contact_relationships;
DELETE FROM projects;
DELETE FROM offers;
DELETE FROM deals;
DELETE FROM contacts;
DELETE FROM user_permissions;
DELETE FROM user_roles;
DELETE FROM users;
DELETE FROM customers;
-- Note: Do NOT delete from companies or budget_dimension_categories (seeded by migrations)

-- ============================================================================
-- PART 2: USERS (12 Norwegian users across all Straye companies)
-- Users table: id is VARCHAR(100), name is single field
-- ============================================================================

INSERT INTO users (id, name, email, first_name, last_name, department, company_id, is_active, roles, created_at, updated_at) VALUES
-- Gruppen (parent company)
('usr-erik-hansen-0001', 'Erik Hansen', 'erik.hansen@straye.no', 'Erik', 'Hansen', 'IT', 'gruppen', true, ARRAY['super_admin'], NOW() - INTERVAL '180 days', NOW()),

-- Stalbygg
('usr-marte-olsen-0002', 'Marte Olsen', 'marte.olsen@straye.no', 'Marte', 'Olsen', 'Ledelse', 'stalbygg', true, ARRAY['company_admin'], NOW() - INTERVAL '170 days', NOW()),
('usr-lars-johansen-0003', 'Lars Johansen', 'lars.johansen@straye.no', 'Lars', 'Johansen', 'Salg', 'stalbygg', true, ARRAY['market'], NOW() - INTERVAL '160 days', NOW()),
('usr-kristine-berg-0004', 'Kristine Berg', 'kristine.berg@straye.no', 'Kristine', 'Berg', 'Prosjekt', 'stalbygg', true, ARRAY['project_manager'], NOW() - INTERVAL '150 days', NOW()),

-- Hybridbygg
('usr-anders-nilsen-0005', 'Anders Nilsen', 'anders.nilsen@straye.no', 'Anders', 'Nilsen', 'Ledelse', 'hybridbygg', true, ARRAY['company_admin'], NOW() - INTERVAL '140 days', NOW()),
('usr-ingrid-bakke-0006', 'Ingrid Bakke', 'ingrid.bakke@straye.no', 'Ingrid', 'Bakke', 'Salg', 'hybridbygg', true, ARRAY['market'], NOW() - INTERVAL '130 days', NOW()),

-- Industri
('usr-thomas-dahl-0007', 'Thomas Dahl', 'thomas.dahl@straye.no', 'Thomas', 'Dahl', 'Ledelse', 'industri', true, ARRAY['company_admin'], NOW() - INTERVAL '120 days', NOW()),
('usr-hanne-lie-0008', 'Hanne Lie', 'hanne.lie@straye.no', 'Hanne', 'Lie', 'Prosjekt', 'industri', true, ARRAY['project_leader'], NOW() - INTERVAL '110 days', NOW()),

-- Tak
('usr-ole-martinsen-0009', 'Ole Martinsen', 'ole.martinsen@straye.no', 'Ole', 'Martinsen', 'Drift', 'tak', true, ARRAY['manager'], NOW() - INTERVAL '100 days', NOW()),
('usr-silje-haugen-0010', 'Silje Haugen', 'silje.haugen@straye.no', 'Silje', 'Haugen', 'Salg', 'tak', true, ARRAY['market'], NOW() - INTERVAL '90 days', NOW()),

-- Montasje
('usr-knut-svendsen-0011', 'Knut Svendsen', 'knut.svendsen@straye.no', 'Knut', 'Svendsen', 'Prosjekt', 'montasje', true, ARRAY['project_manager'], NOW() - INTERVAL '80 days', NOW()),
('usr-maria-holm-0012', 'Maria Holm', 'maria.holm@straye.no', 'Maria', 'Holm', 'Administrasjon', 'montasje', true, ARRAY['viewer'], NOW() - INTERVAL '70 days', NOW());

-- ============================================================================
-- PART 3: USER ROLES
-- ============================================================================

INSERT INTO user_roles (id, user_id, role, company_id, granted_by, granted_at, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), 'usr-erik-hansen-0001', 'super_admin', NULL, 'system', NOW() - INTERVAL '180 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-marte-olsen-0002', 'company_admin', 'stalbygg', 'usr-erik-hansen-0001', NOW() - INTERVAL '170 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-lars-johansen-0003', 'market', 'stalbygg', 'usr-marte-olsen-0002', NOW() - INTERVAL '160 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-kristine-berg-0004', 'project_manager', 'stalbygg', 'usr-marte-olsen-0002', NOW() - INTERVAL '150 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-anders-nilsen-0005', 'company_admin', 'hybridbygg', 'usr-erik-hansen-0001', NOW() - INTERVAL '140 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-ingrid-bakke-0006', 'market', 'hybridbygg', 'usr-anders-nilsen-0005', NOW() - INTERVAL '130 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-thomas-dahl-0007', 'company_admin', 'industri', 'usr-erik-hansen-0001', NOW() - INTERVAL '120 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-hanne-lie-0008', 'project_leader', 'industri', 'usr-thomas-dahl-0007', NOW() - INTERVAL '110 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-ole-martinsen-0009', 'manager', 'tak', 'usr-erik-hansen-0001', NOW() - INTERVAL '100 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-silje-haugen-0010', 'market', 'tak', 'usr-ole-martinsen-0009', NOW() - INTERVAL '90 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-knut-svendsen-0011', 'project_manager', 'montasje', 'usr-erik-hansen-0001', NOW() - INTERVAL '80 days', true, NOW(), NOW()),
(gen_random_uuid(), 'usr-maria-holm-0012', 'viewer', 'montasje', 'usr-knut-svendsen-0011', NOW() - INTERVAL '70 days', true, NOW(), NOW());

-- ============================================================================
-- PART 4: CUSTOMERS (18 Norwegian construction/industrial companies)
-- ============================================================================

INSERT INTO customers (id, name, org_number, email, phone, address, city, postal_code, country, contact_person, contact_email, contact_phone, status, tier, industry, company_id, created_at, updated_at) VALUES
-- Large customers (Platinum/Gold) - Stalbygg
('a0000001-0001-0001-0001-000000000001', 'Veidekke ASA', '987654321', 'veidekke@straye.no', '+47 21 05 50 00', 'Skabos vei 4', 'Oslo', '0278', 'Norway', 'Per Andersen', 'per.andersen@straye.no', '+47 922 33 445', 'active', 'platinum', 'construction', 'stalbygg', NOW() - INTERVAL '150 days', NOW()),
('a0000002-0002-0002-0002-000000000002', 'AF Gruppen ASA', '912345678', 'afgruppen@straye.no', '+47 22 89 11 00', 'Innspurten 15', 'Oslo', '0663', 'Norway', 'Kari Nordmann', 'kari.nordmann@straye.no', '+47 911 22 334', 'active', 'platinum', 'construction', 'stalbygg', NOW() - INTERVAL '145 days', NOW()),
('a0000003-0003-0003-0003-000000000003', 'Skanska Norge AS', '923456789', 'skanska@straye.no', '+47 40 00 64 00', 'Drammensveien 60', 'Oslo', '0271', 'Norway', 'Ole Svendsen', 'ole.svendsen@straye.no', '+47 900 11 223', 'active', 'gold', 'construction', 'stalbygg', NOW() - INTERVAL '140 days', NOW()),

-- Medium customers (Silver) - Hybridbygg
('a0000004-0004-0004-0004-000000000004', 'Statsbygg', '934567890', 'statsbygg@straye.no', '+47 22 95 42 00', 'Biskop Gunnerus gate 6', 'Oslo', '0155', 'Norway', 'Anne Larsen', 'anne.larsen@straye.no', '+47 933 44 556', 'active', 'gold', 'public_sector', 'hybridbygg', NOW() - INTERVAL '135 days', NOW()),
('a0000005-0005-0005-0005-000000000005', 'OBOS Eiendom AS', '945678901', 'obos@straye.no', '+47 22 86 55 00', 'Hammersborg torg 1', 'Oslo', '0179', 'Norway', 'Erik Nilsen', 'erik.nilsen@straye.no', '+47 944 55 667', 'active', 'silver', 'real_estate', 'hybridbygg', NOW() - INTERVAL '130 days', NOW()),
('a0000006-0006-0006-0006-000000000006', 'Selvaag Bolig ASA', '956789012', 'selvaag@straye.no', '+47 23 15 25 00', 'Silurveien 2', 'Oslo', '0380', 'Norway', 'Maria Hansen', 'maria.hansen@straye.no', '+47 955 66 778', 'active', 'silver', 'real_estate', 'hybridbygg', NOW() - INTERVAL '125 days', NOW()),

-- Industri customers
('a0000007-0007-0007-0007-000000000007', 'Norsk Hydro ASA', '967890123', 'hydro@straye.no', '+47 22 53 81 00', 'Drammensveien 260', 'Oslo', '0283', 'Norway', 'Thomas Berg', 'thomas.berg@straye.no', '+47 966 77 889', 'active', 'gold', 'manufacturing', 'industri', NOW() - INTERVAL '120 days', NOW()),
('a0000008-0008-0008-0008-000000000008', 'Elkem ASA', '978901234', 'elkem@straye.no', '+47 22 45 01 00', 'Hoffsveien 65B', 'Oslo', '0377', 'Norway', 'Ingrid Holm', 'ingrid.holm@straye.no', '+47 977 88 990', 'active', 'silver', 'manufacturing', 'industri', NOW() - INTERVAL '115 days', NOW()),
('a0000009-0009-0009-0009-000000000009', 'Aibel AS', '989012345', 'aibel@straye.no', '+47 51 81 80 00', 'Kokstadflaten 5', 'Bergen', '5257', 'Norway', 'Kristian Dahl', 'kristian.dahl@straye.no', '+47 988 99 001', 'active', 'silver', 'energy', 'industri', NOW() - INTERVAL '110 days', NOW()),

-- Tak customers
('a0000010-0010-0010-0010-000000000010', 'Trondheim Kommune', '990123456', 'trondheim@straye.no', '+47 72 54 00 00', 'Munkegata 1', 'Trondheim', '7004', 'Norway', 'Bjorn Olsen', 'bjorn.olsen@straye.no', '+47 999 00 112', 'active', 'gold', 'public_sector', 'tak', NOW() - INTERVAL '105 days', NOW()),
('a0000011-0011-0011-0011-000000000011', 'Bergen Kommune', '901234567', 'bergen@straye.no', '+47 55 56 56 56', 'Radhuset', 'Bergen', '5014', 'Norway', 'Liv Johansen', 'liv.johansen@straye.no', '+47 900 11 223', 'active', 'silver', 'public_sector', 'tak', NOW() - INTERVAL '100 days', NOW()),
('a0000012-0012-0012-0012-000000000012', 'Coop Norge SA', '912345670', 'coop@straye.no', '+47 22 89 89 00', 'Osterhausgate 32', 'Oslo', '0183', 'Norway', 'Knut Eriksen', 'knut.eriksen@straye.no', '+47 911 22 334', 'active', 'silver', 'retail', 'tak', NOW() - INTERVAL '95 days', NOW()),

-- Montasje customers
('a0000013-0013-0013-0013-000000000013', 'IKEA Norge AS', '923456780', 'ikea@straye.no', '+47 22 31 00 00', 'Stovner Senter', 'Oslo', '0985', 'Norway', 'Anna Lindstrom', 'anna.lindstrom@straye.no', '+47 922 33 445', 'active', 'gold', 'retail', 'montasje', NOW() - INTERVAL '90 days', NOW()),
('a0000014-0014-0014-0014-000000000014', 'XXL Sport og Villmark', '934567891', 'xxl@straye.no', '+47 23 00 00 00', 'Alnabru', 'Oslo', '0614', 'Norway', 'Morten Bakken', 'morten.bakken@straye.no', '+47 933 44 556', 'active', 'silver', 'retail', 'montasje', NOW() - INTERVAL '85 days', NOW()),
('a0000015-0015-0015-0015-000000000015', 'Bama Gruppen AS', '945678902', 'bama@straye.no', '+47 23 30 00 00', 'Liertoppen', 'Lier', '3400', 'Norway', 'Hilde Pedersen', 'hilde.pedersen@straye.no', '+47 944 55 667', 'active', 'bronze', 'logistics', 'montasje', NOW() - INTERVAL '80 days', NOW()),

-- Additional diverse customers
('a0000016-0016-0016-0016-000000000016', 'Nortura SA', '956789013', 'nortura@straye.no', '+47 22 09 56 00', 'Lorenveien 37', 'Oslo', '0585', 'Norway', 'Per Haugen', 'per.haugen@straye.no', '+47 955 66 778', 'active', 'silver', 'food_production', 'industri', NOW() - INTERVAL '75 days', NOW()),
('a0000017-0017-0017-0017-000000000017', 'NTNU', '967890124', 'ntnu@straye.no', '+47 73 59 50 00', 'Glosehaugen', 'Trondheim', '7491', 'Norway', 'Professor Olsen', 'professor.olsen@straye.no', '+47 966 77 889', 'active', 'gold', 'education', 'hybridbygg', NOW() - INTERVAL '70 days', NOW()),
('a0000018-0018-0018-0018-000000000018', 'Equinor ASA', '978901235', 'equinor@straye.no', '+47 51 99 00 00', 'Forusbeen 50', 'Stavanger', '4035', 'Norway', 'Bjarne Strand', 'bjarne.strand@straye.no', '+47 977 88 990', 'active', 'platinum', 'energy', 'industri', NOW() - INTERVAL '65 days', NOW());

-- ============================================================================
-- PART 5: CONTACTS (36 contacts, 2 per customer)
-- ============================================================================

INSERT INTO contacts (id, first_name, last_name, email, phone, mobile, title, department, primary_customer_id, city, country, contact_type, is_active, created_at, updated_at) VALUES
-- Veidekke contacts
('b0000001-0001-0001-0001-000000000001', 'Per', 'Andersen', 'per.andersen@straye.no', '+47 922 33 445', '+47 922 33 445', 'Prosjektleder', 'Prosjekt', 'a0000001-0001-0001-0001-000000000001', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '150 days', NOW()),
('b0000002-0002-0002-0002-000000000002', 'Silje', 'Berg', 'silje.berg@straye.no', '+47 933 44 556', '+47 933 44 556', 'Innkjopssjef', 'Innkjop', 'a0000001-0001-0001-0001-000000000001', 'Oslo', 'Norway', 'secondary', true, NOW() - INTERVAL '145 days', NOW()),

-- AF Gruppen contacts
('b0000003-0003-0003-0003-000000000003', 'Kari', 'Nordmann', 'kari.nordmann@straye.no', '+47 911 22 334', '+47 911 22 334', 'Prosjektdirektor', 'Ledelse', 'a0000002-0002-0002-0002-000000000002', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '145 days', NOW()),
('b0000004-0004-0004-0004-000000000004', 'Ola', 'Hansen', 'ola.hansen@straye.no', '+47 922 33 445', '+47 922 33 445', 'Teknisk Leder', 'Teknikk', 'a0000002-0002-0002-0002-000000000002', 'Oslo', 'Norway', 'technical', true, NOW() - INTERVAL '140 days', NOW()),

-- Skanska contacts
('b0000005-0005-0005-0005-000000000005', 'Ole', 'Svendsen', 'ole.svendsen@straye.no', '+47 900 11 223', '+47 900 11 223', 'Anleggsleder', 'Drift', 'a0000003-0003-0003-0003-000000000003', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '140 days', NOW()),
('b0000006-0006-0006-0006-000000000006', 'Nina', 'Karlsen', 'nina.karlsen@straye.no', '+47 911 22 334', '+47 911 22 334', 'Okonomisjef', 'Okonomi', 'a0000003-0003-0003-0003-000000000003', 'Oslo', 'Norway', 'billing', true, NOW() - INTERVAL '135 days', NOW()),

-- Statsbygg contacts
('b0000007-0007-0007-0007-000000000007', 'Anne', 'Larsen', 'anne.larsen@straye.no', '+47 933 44 556', '+47 933 44 556', 'Prosjektansvarlig', 'Bygg', 'a0000004-0004-0004-0004-000000000004', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '135 days', NOW()),
('b0000008-0008-0008-0008-000000000008', 'Henrik', 'Dahl', 'henrik.dahl@straye.no', '+47 944 55 667', '+47 944 55 667', 'Seniorradgiver', 'Plan', 'a0000004-0004-0004-0004-000000000004', 'Oslo', 'Norway', 'technical', true, NOW() - INTERVAL '130 days', NOW()),

-- OBOS contacts
('b0000009-0009-0009-0009-000000000009', 'Erik', 'Nilsen', 'erik.nilsen@straye.no', '+47 944 55 667', '+47 944 55 667', 'Utviklingsdirektor', 'Utvikling', 'a0000005-0005-0005-0005-000000000005', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '130 days', NOW()),
('b0000010-0010-0010-0010-000000000010', 'Mette', 'Olsen', 'mette.olsen@straye.no', '+47 955 66 778', '+47 955 66 778', 'Prosjektleder', 'Prosjekt', 'a0000005-0005-0005-0005-000000000005', 'Oslo', 'Norway', 'secondary', true, NOW() - INTERVAL '125 days', NOW()),

-- Selvaag contacts
('b0000011-0011-0011-0011-000000000011', 'Maria', 'Hansen', 'maria.hansen@straye.no', '+47 955 66 778', '+47 955 66 778', 'Prosjektsjef', 'Prosjekt', 'a0000006-0006-0006-0006-000000000006', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '125 days', NOW()),
('b0000012-0012-0012-0012-000000000012', 'Jonas', 'Eriksen', 'jonas.eriksen@straye.no', '+47 966 77 889', '+47 966 77 889', 'Byggeleder', 'Bygg', 'a0000006-0006-0006-0006-000000000006', 'Oslo', 'Norway', 'technical', true, NOW() - INTERVAL '120 days', NOW()),

-- Hydro contacts
('b0000013-0013-0013-0013-000000000013', 'Thomas', 'Berg', 'thomas.berg@straye.no', '+47 966 77 889', '+47 966 77 889', 'Fabrikkssjef', 'Produksjon', 'a0000007-0007-0007-0007-000000000007', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '120 days', NOW()),
('b0000014-0014-0014-0014-000000000014', 'Kristin', 'Pedersen', 'kristin.pedersen@straye.no', '+47 977 88 990', '+47 977 88 990', 'Vedlikeholdssjef', 'Vedlikehold', 'a0000007-0007-0007-0007-000000000007', 'Oslo', 'Norway', 'technical', true, NOW() - INTERVAL '115 days', NOW()),

-- Elkem contacts
('b0000015-0015-0015-0015-000000000015', 'Ingrid', 'Holm', 'ingrid.holm@straye.no', '+47 977 88 990', '+47 977 88 990', 'Anleggssjef', 'Drift', 'a0000008-0008-0008-0008-000000000008', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '115 days', NOW()),
('b0000016-0016-0016-0016-000000000016', 'Steinar', 'Lund', 'steinar.lund@straye.no', '+47 988 99 001', '+47 988 99 001', 'Teknisk Direktor', 'Teknikk', 'a0000008-0008-0008-0008-000000000008', 'Oslo', 'Norway', 'technical', true, NOW() - INTERVAL '110 days', NOW()),

-- Aibel contacts
('b0000017-0017-0017-0017-000000000017', 'Kristian', 'Dahl', 'kristian.dahl@straye.no', '+47 988 99 001', '+47 988 99 001', 'Prosjektdirektor', 'Offshore', 'a0000009-0009-0009-0009-000000000009', 'Bergen', 'Norway', 'primary', true, NOW() - INTERVAL '110 days', NOW()),
('b0000018-0018-0018-0018-000000000018', 'Lise', 'Haugen', 'lise.haugen@straye.no', '+47 999 00 112', '+47 999 00 112', 'Innkjopsansvarlig', 'Innkjop', 'a0000009-0009-0009-0009-000000000009', 'Bergen', 'Norway', 'billing', true, NOW() - INTERVAL '105 days', NOW()),

-- Trondheim Kommune contacts
('b0000019-0019-0019-0019-000000000019', 'Bjorn', 'Olsen', 'bjorn.olsen@straye.no', '+47 999 00 112', '+47 999 00 112', 'Eiendomssjef', 'Eiendom', 'a0000010-0010-0010-0010-000000000010', 'Trondheim', 'Norway', 'primary', true, NOW() - INTERVAL '105 days', NOW()),
('b0000020-0020-0020-0020-000000000020', 'Grete', 'Strom', 'grete.strom@straye.no', '+47 900 11 223', '+47 900 11 223', 'Prosjektkoordinator', 'Bygg', 'a0000010-0010-0010-0010-000000000010', 'Trondheim', 'Norway', 'secondary', true, NOW() - INTERVAL '100 days', NOW()),

-- Bergen Kommune contacts
('b0000021-0021-0021-0021-000000000021', 'Liv', 'Johansen', 'liv.johansen@straye.no', '+47 900 11 223', '+47 900 11 223', 'Bygningssjef', 'Bygg', 'a0000011-0011-0011-0011-000000000011', 'Bergen', 'Norway', 'primary', true, NOW() - INTERVAL '100 days', NOW()),
('b0000022-0022-0022-0022-000000000022', 'Rune', 'Viken', 'rune.viken@straye.no', '+47 911 22 334', '+47 911 22 334', 'Vedlikeholdsansvarlig', 'Vedlikehold', 'a0000011-0011-0011-0011-000000000011', 'Bergen', 'Norway', 'technical', true, NOW() - INTERVAL '95 days', NOW()),

-- Coop contacts
('b0000023-0023-0023-0023-000000000023', 'Knut', 'Eriksen', 'knut.eriksen@straye.no', '+47 911 22 334', '+47 911 22 334', 'Driftssjef', 'Drift', 'a0000012-0012-0012-0012-000000000012', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '95 days', NOW()),
('b0000024-0024-0024-0024-000000000024', 'Heidi', 'Bakke', 'heidi.bakke@straye.no', '+47 922 33 445', '+47 922 33 445', 'Butikksjef', 'Butikk', 'a0000012-0012-0012-0012-000000000012', 'Oslo', 'Norway', 'secondary', true, NOW() - INTERVAL '90 days', NOW()),

-- IKEA contacts
('b0000025-0025-0025-0025-000000000025', 'Anna', 'Lindstrom', 'anna.lindstrom@straye.no', '+47 922 33 445', '+47 922 33 445', 'Varehussjef', 'Drift', 'a0000013-0013-0013-0013-000000000013', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '90 days', NOW()),
('b0000026-0026-0026-0026-000000000026', 'Magnus', 'Larsen', 'magnus.larsen@straye.no', '+47 933 44 556', '+47 933 44 556', 'Logistikksjef', 'Logistikk', 'a0000013-0013-0013-0013-000000000013', 'Oslo', 'Norway', 'technical', true, NOW() - INTERVAL '85 days', NOW()),

-- XXL contacts
('b0000027-0027-0027-0027-000000000027', 'Morten', 'Bakken', 'morten.bakken@straye.no', '+47 933 44 556', '+47 933 44 556', 'Ekspansjonssjef', 'Utvikling', 'a0000014-0014-0014-0014-000000000014', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '85 days', NOW()),
('b0000028-0028-0028-0028-000000000028', 'Tonje', 'Vik', 'tonje.vik@straye.no', '+47 944 55 667', '+47 944 55 667', 'Butikkutvikler', 'Butikk', 'a0000014-0014-0014-0014-000000000014', 'Oslo', 'Norway', 'secondary', true, NOW() - INTERVAL '80 days', NOW()),

-- Bama contacts
('b0000029-0029-0029-0029-000000000029', 'Hilde', 'Pedersen', 'hilde.pedersen@straye.no', '+47 944 55 667', '+47 944 55 667', 'Logistikkdirektor', 'Logistikk', 'a0000015-0015-0015-0015-000000000015', 'Lier', 'Norway', 'primary', true, NOW() - INTERVAL '80 days', NOW()),
('b0000030-0030-0030-0030-000000000030', 'Anders', 'Fjell', 'anders.fjell@straye.no', '+47 955 66 778', '+47 955 66 778', 'Lagersjef', 'Lager', 'a0000015-0015-0015-0015-000000000015', 'Lier', 'Norway', 'technical', true, NOW() - INTERVAL '75 days', NOW()),

-- Nortura contacts
('b0000031-0031-0031-0031-000000000031', 'Per', 'Haugen', 'per.haugen@straye.no', '+47 955 66 778', '+47 955 66 778', 'Fabrikkssjef', 'Produksjon', 'a0000016-0016-0016-0016-000000000016', 'Oslo', 'Norway', 'primary', true, NOW() - INTERVAL '75 days', NOW()),
('b0000032-0032-0032-0032-000000000032', 'Eva', 'Strand', 'eva.strand@straye.no', '+47 966 77 889', '+47 966 77 889', 'Vedlikeholdssjef', 'Vedlikehold', 'a0000016-0016-0016-0016-000000000016', 'Oslo', 'Norway', 'technical', true, NOW() - INTERVAL '70 days', NOW()),

-- NTNU contacts
('b0000033-0033-0033-0033-000000000033', 'Professor', 'Olsen', 'professor.olsen@straye.no', '+47 966 77 889', '+47 966 77 889', 'Instituttleder', 'Bygg', 'a0000017-0017-0017-0017-000000000017', 'Trondheim', 'Norway', 'primary', true, NOW() - INTERVAL '70 days', NOW()),
('b0000034-0034-0034-0034-000000000034', 'Dr', 'Svendsen', 'dr.svendsen@straye.no', '+47 977 88 990', '+47 977 88 990', 'Forskningsleder', 'Forskning', 'a0000017-0017-0017-0017-000000000017', 'Trondheim', 'Norway', 'technical', true, NOW() - INTERVAL '65 days', NOW()),

-- Equinor contacts
('b0000035-0035-0035-0035-000000000035', 'Bjarne', 'Strand', 'bjarne.strand@straye.no', '+47 977 88 990', '+47 977 88 990', 'Prosjektdirektor', 'Offshore', 'a0000018-0018-0018-0018-000000000018', 'Stavanger', 'Norway', 'primary', true, NOW() - INTERVAL '65 days', NOW()),
('b0000036-0036-0036-0036-000000000036', 'Linda', 'Moe', 'linda.moe@straye.no', '+47 988 99 001', '+47 988 99 001', 'Innkjopsleder', 'Innkjop', 'a0000018-0018-0018-0018-000000000018', 'Stavanger', 'Norway', 'billing', true, NOW() - INTERVAL '60 days', NOW());

-- ============================================================================
-- PART 6: CONTACT RELATIONSHIPS
-- ============================================================================

INSERT INTO contact_relationships (id, contact_id, entity_type, entity_id, role, is_primary, created_at) VALUES
(gen_random_uuid(), 'b0000001-0001-0001-0001-000000000001', 'customer', 'a0000001-0001-0001-0001-000000000001', 'Prosjektleder', true, NOW() - INTERVAL '150 days'),
(gen_random_uuid(), 'b0000002-0002-0002-0002-000000000002', 'customer', 'a0000001-0001-0001-0001-000000000001', 'Innkjopssjef', false, NOW() - INTERVAL '145 days'),
(gen_random_uuid(), 'b0000003-0003-0003-0003-000000000003', 'customer', 'a0000002-0002-0002-0002-000000000002', 'Prosjektdirektor', true, NOW() - INTERVAL '145 days'),
(gen_random_uuid(), 'b0000004-0004-0004-0004-000000000004', 'customer', 'a0000002-0002-0002-0002-000000000002', 'Teknisk Leder', false, NOW() - INTERVAL '140 days'),
(gen_random_uuid(), 'b0000005-0005-0005-0005-000000000005', 'customer', 'a0000003-0003-0003-0003-000000000003', 'Anleggsleder', true, NOW() - INTERVAL '140 days'),
(gen_random_uuid(), 'b0000006-0006-0006-0006-000000000006', 'customer', 'a0000003-0003-0003-0003-000000000003', 'Okonomisjef', false, NOW() - INTERVAL '135 days'),
(gen_random_uuid(), 'b0000007-0007-0007-0007-000000000007', 'customer', 'a0000004-0004-0004-0004-000000000004', 'Prosjektansvarlig', true, NOW() - INTERVAL '135 days'),
(gen_random_uuid(), 'b0000008-0008-0008-0008-000000000008', 'customer', 'a0000004-0004-0004-0004-000000000004', 'Seniorradgiver', false, NOW() - INTERVAL '130 days'),
(gen_random_uuid(), 'b0000009-0009-0009-0009-000000000009', 'customer', 'a0000005-0005-0005-0005-000000000005', 'Utviklingsdirektor', true, NOW() - INTERVAL '130 days'),
(gen_random_uuid(), 'b0000010-0010-0010-0010-000000000010', 'customer', 'a0000005-0005-0005-0005-000000000005', 'Prosjektleder', false, NOW() - INTERVAL '125 days'),
(gen_random_uuid(), 'b0000011-0011-0011-0011-000000000011', 'customer', 'a0000006-0006-0006-0006-000000000006', 'Prosjektsjef', true, NOW() - INTERVAL '125 days'),
(gen_random_uuid(), 'b0000012-0012-0012-0012-000000000012', 'customer', 'a0000006-0006-0006-0006-000000000006', 'Byggeleder', false, NOW() - INTERVAL '120 days'),
(gen_random_uuid(), 'b0000013-0013-0013-0013-000000000013', 'customer', 'a0000007-0007-0007-0007-000000000007', 'Fabrikkssjef', true, NOW() - INTERVAL '120 days'),
(gen_random_uuid(), 'b0000014-0014-0014-0014-000000000014', 'customer', 'a0000007-0007-0007-0007-000000000007', 'Vedlikeholdssjef', false, NOW() - INTERVAL '115 days'),
(gen_random_uuid(), 'b0000015-0015-0015-0015-000000000015', 'customer', 'a0000008-0008-0008-0008-000000000008', 'Anleggssjef', true, NOW() - INTERVAL '115 days'),
(gen_random_uuid(), 'b0000016-0016-0016-0016-000000000016', 'customer', 'a0000008-0008-0008-0008-000000000008', 'Teknisk Direktor', false, NOW() - INTERVAL '110 days'),
(gen_random_uuid(), 'b0000017-0017-0017-0017-000000000017', 'customer', 'a0000009-0009-0009-0009-000000000009', 'Prosjektdirektor', true, NOW() - INTERVAL '110 days'),
(gen_random_uuid(), 'b0000018-0018-0018-0018-000000000018', 'customer', 'a0000009-0009-0009-0009-000000000009', 'Innkjopsansvarlig', false, NOW() - INTERVAL '105 days');

-- ============================================================================
-- PART 7: DEALS (25 deals across pipeline stages)
-- ============================================================================

INSERT INTO deals (id, title, description, customer_id, company_id, customer_name, stage, probability, value, currency, expected_close_date, actual_close_date, owner_id, owner_name, source, notes, lost_reason, loss_reason_category, created_at, updated_at) VALUES
-- Stalbygg deals
('c0000001-0001-0001-0001-000000000001', 'Lagerbygg Alnabru', 'Nytt lagerbygg 5000m2 for Veidekke', 'a0000001-0001-0001-0001-000000000001', 'stalbygg', 'Veidekke ASA', 'lead', 10, 12000000.00, 'NOK', NOW() + INTERVAL '90 days', NULL, 'usr-lars-johansen-0003', 'Lars Johansen', 'Innkommende', NULL, NULL, NULL, NOW() - INTERVAL '10 days', NOW()),
('c0000002-0002-0002-0002-000000000002', 'Verkstedhall AF', 'Stalverksted 2000m2', 'a0000002-0002-0002-0002-000000000002', 'stalbygg', 'AF Gruppen ASA', 'qualified', 30, 8500000.00, 'NOK', NOW() + INTERVAL '60 days', NULL, 'usr-lars-johansen-0003', 'Lars Johansen', 'Anbud', NULL, NULL, NULL, NOW() - INTERVAL '25 days', NOW()),
('c0000003-0003-0003-0003-000000000003', 'Produksjonshall Skanska', 'Produksjonshall med kontor', 'a0000003-0003-0003-0003-000000000003', 'stalbygg', 'Skanska Norge AS', 'proposal', 50, 15000000.00, 'NOK', NOW() + INTERVAL '45 days', NULL, 'usr-marte-olsen-0002', 'Marte Olsen', 'Eksisterende kunde', NULL, NULL, NULL, NOW() - INTERVAL '40 days', NOW()),
('c0000004-0004-0004-0004-000000000004', 'Lagerbygg Gardermoen', 'Stort lagerbygg ved OSL', 'a0000001-0001-0001-0001-000000000001', 'stalbygg', 'Veidekke ASA', 'won', 100, 18500000.00, 'NOK', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days', 'usr-lars-johansen-0003', 'Lars Johansen', 'Anbud', 'Signert kontrakt', NULL, NULL, NOW() - INTERVAL '90 days', NOW()),
('c0000005-0005-0005-0005-000000000005', 'Logistikksenter Vestby', 'Logistikksenter med kjolelager', 'a0000002-0002-0002-0002-000000000002', 'stalbygg', 'AF Gruppen ASA', 'negotiation', 70, 22000000.00, 'NOK', NOW() + INTERVAL '30 days', NULL, 'usr-lars-johansen-0003', 'Lars Johansen', 'Referanse', NULL, NULL, NULL, NOW() - INTERVAL '60 days', NOW()),
('c0000006-0006-0006-0006-000000000006', 'Flerbrukshall Drammen', 'Flerbrukshall tapt til konkurrent', 'a0000003-0003-0003-0003-000000000003', 'stalbygg', 'Skanska Norge AS', 'lost', 0, 9500000.00, 'NOK', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days', 'usr-marte-olsen-0002', 'Marte Olsen', 'Anbud', NULL, 'Pris for hoy', 'price', NOW() - INTERVAL '75 days', NOW()),

-- Hybridbygg deals
('c0000007-0007-0007-0007-000000000007', 'Kontorbygg Statsbygg', 'Nytt kontorbygg for Statsbygg', 'a0000004-0004-0004-0004-000000000004', 'hybridbygg', 'Statsbygg', 'proposal', 60, 35000000.00, 'NOK', NOW() + INTERVAL '60 days', NULL, 'usr-anders-nilsen-0005', 'Anders Nilsen', 'Offentlig anbud', NULL, NULL, NULL, NOW() - INTERVAL '50 days', NOW()),
('c0000008-0008-0008-0008-000000000008', 'Boligprosjekt Loren', 'Hybridbygg boligprosjekt', 'a0000005-0005-0005-0005-000000000005', 'hybridbygg', 'OBOS Eiendom AS', 'qualified', 40, 28000000.00, 'NOK', NOW() + INTERVAL '75 days', NULL, 'usr-ingrid-bakke-0006', 'Ingrid Bakke', 'Nettverksmote', NULL, NULL, NULL, NOW() - INTERVAL '35 days', NOW()),
('c0000009-0009-0009-0009-000000000009', 'Kontorbygg Majorstuen', 'Moderne kontorbygg', 'a0000006-0006-0006-0006-000000000006', 'hybridbygg', 'Selvaag Bolig ASA', 'won', 100, 25000000.00, 'NOK', NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days', 'usr-ingrid-bakke-0006', 'Ingrid Bakke', 'Eksisterende kunde', 'Kontrakt signert', NULL, NULL, NOW() - INTERVAL '100 days', NOW()),
('c0000010-0010-0010-0010-000000000010', 'Universitetsbygg NTNU', 'Nytt undervisningsbygg', 'a0000017-0017-0017-0017-000000000017', 'hybridbygg', 'NTNU', 'negotiation', 75, 42000000.00, 'NOK', NOW() + INTERVAL '45 days', NULL, 'usr-anders-nilsen-0005', 'Anders Nilsen', 'Offentlig anbud', NULL, NULL, NULL, NOW() - INTERVAL '80 days', NOW()),

-- Industri deals
('c0000011-0011-0011-0011-000000000011', 'Verksted Hydro', 'Vedlikeholdsverksted', 'a0000007-0007-0007-0007-000000000007', 'industri', 'Norsk Hydro ASA', 'lead', 15, 6500000.00, 'NOK', NOW() + INTERVAL '120 days', NULL, 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Innkommende', NULL, NULL, NULL, NOW() - INTERVAL '5 days', NOW()),
('c0000012-0012-0012-0012-000000000012', 'Verksted Karmoy', 'Nytt produksjonsverksted', 'a0000008-0008-0008-0008-000000000008', 'industri', 'Elkem ASA', 'won', 100, 11500000.00, 'NOK', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', 'usr-hanne-lie-0008', 'Hanne Lie', 'Eksisterende kunde', 'Prosjekt igangsatt', NULL, NULL, NOW() - INTERVAL '120 days', NOW()),
('c0000013-0013-0013-0013-000000000013', 'Offshore modul Aibel', 'Modul for offshore installasjon', 'a0000009-0009-0009-0009-000000000009', 'industri', 'Aibel AS', 'proposal', 55, 45000000.00, 'NOK', NOW() + INTERVAL '90 days', NULL, 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Anbud', NULL, NULL, NULL, NOW() - INTERVAL '45 days', NOW()),
('c0000014-0014-0014-0014-000000000014', 'Industrihall Bergen', 'Industrihall tapt', 'a0000009-0009-0009-0009-000000000009', 'industri', 'Aibel AS', 'lost', 0, 18000000.00, 'NOK', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Anbud', NULL, 'Valgte lokal leverandor', 'competitor', NOW() - INTERVAL '130 days', NOW()),
('c0000015-0015-0015-0015-000000000015', 'Prosessanlegg Equinor', 'Prosessanlegg utvidelse', 'a0000018-0018-0018-0018-000000000018', 'industri', 'Equinor ASA', 'qualified', 35, 65000000.00, 'NOK', NOW() + INTERVAL '150 days', NULL, 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Strategisk partner', NULL, NULL, NULL, NOW() - INTERVAL '30 days', NOW()),

-- Tak deals
('c0000016-0016-0016-0016-000000000016', 'Takutskifting Moholt', 'Takutskifting studentboliger', 'a0000010-0010-0010-0010-000000000010', 'tak', 'Trondheim Kommune', 'lead', 20, 3200000.00, 'NOK', NOW() + INTERVAL '90 days', NULL, 'usr-silje-haugen-0010', 'Silje Haugen', 'Offentlig anbud', NULL, NULL, NULL, NOW() - INTERVAL '8 days', NOW()),
('c0000017-0017-0017-0017-000000000017', 'Takprosjekt Bryggen', 'Historisk tak renovering', 'a0000011-0011-0011-0011-000000000011', 'tak', 'Bergen Kommune', 'proposal', 50, 4800000.00, 'NOK', NOW() + INTERVAL '60 days', NULL, 'usr-ole-martinsen-0009', 'Ole Martinsen', 'Offentlig anbud', NULL, NULL, NULL, NOW() - INTERVAL '30 days', NOW()),
('c0000018-0018-0018-0018-000000000018', 'Takprosjekt Coop', 'Takutskifting varehus', 'a0000012-0012-0012-0012-000000000012', 'tak', 'Coop Norge SA', 'won', 100, 2800000.00, 'NOK', NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days', 'usr-ole-martinsen-0009', 'Ole Martinsen', 'Rammeavtale', 'Under arbeid', NULL, NULL, NOW() - INTERVAL '60 days', NOW()),
('c0000019-0019-0019-0019-000000000019', 'Skoletak Stavanger', 'Skoletakrenovering', 'a0000011-0011-0011-0011-000000000011', 'tak', 'Bergen Kommune', 'won', 100, 1900000.00, 'NOK', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', 'usr-silje-haugen-0010', 'Silje Haugen', 'Offentlig anbud', 'Ferdigstilt', NULL, NULL, NOW() - INTERVAL '150 days', NOW()),

-- Montasje deals
('c0000020-0020-0020-0020-000000000020', 'IKEA Innredning', 'Varehusinnredning ny butikk', 'a0000013-0013-0013-0013-000000000013', 'montasje', 'IKEA Norge AS', 'negotiation', 80, 8500000.00, 'NOK', NOW() + INTERVAL '30 days', NULL, 'usr-knut-svendsen-0011', 'Knut Svendsen', 'Eksisterende kunde', NULL, NULL, NULL, NOW() - INTERVAL '55 days', NOW()),
('c0000021-0021-0021-0021-000000000021', 'XXL Butikk Oslo', 'Butikkinnredning', 'a0000014-0014-0014-0014-000000000014', 'montasje', 'XXL Sport og Villmark', 'qualified', 40, 4200000.00, 'NOK', NOW() + INTERVAL '75 days', NULL, 'usr-knut-svendsen-0011', 'Knut Svendsen', 'Nettverksmote', NULL, NULL, NULL, NOW() - INTERVAL '20 days', NOW()),
('c0000022-0022-0022-0022-000000000022', 'Bama Lager', 'Lagerhyller og reoler', 'a0000015-0015-0015-0015-000000000015', 'montasje', 'Bama Gruppen AS', 'proposal', 55, 3800000.00, 'NOK', NOW() + INTERVAL '45 days', NULL, 'usr-maria-holm-0012', 'Maria Holm', 'Innkommende', NULL, NULL, NULL, NOW() - INTERVAL '25 days', NOW()),
('c0000023-0023-0023-0023-000000000023', 'Monteringsjobb Bodo', 'Butikkmontering', 'a0000014-0014-0014-0014-000000000014', 'montasje', 'XXL Sport og Villmark', 'won', 100, 2100000.00, 'NOK', NOW() - INTERVAL '40 days', NOW() - INTERVAL '40 days', 'usr-knut-svendsen-0011', 'Knut Svendsen', 'Rammeavtale', 'Ferdigstilt', NULL, NULL, NOW() - INTERVAL '100 days', NOW()),
('c0000024-0024-0024-0024-000000000024', 'Nortura Produksjonslinje', 'Installasjon produksjonslinje', 'a0000016-0016-0016-0016-000000000016', 'montasje', 'Nortura SA', 'lead', 15, 5500000.00, 'NOK', NOW() + INTERVAL '100 days', NULL, 'usr-knut-svendsen-0011', 'Knut Svendsen', 'Innkommende', NULL, NULL, NULL, NOW() - INTERVAL '3 days', NOW()),
('c0000025-0025-0025-0025-000000000025', 'Sportsutstyr montering', 'Tapt til konkurrent', 'a0000014-0014-0014-0014-000000000014', 'montasje', 'XXL Sport og Villmark', 'lost', 0, 1800000.00, 'NOK', NOW() - INTERVAL '50 days', NOW() - INTERVAL '50 days', 'usr-maria-holm-0012', 'Maria Holm', 'Anbud', NULL, 'For lang leveringstid', 'timing', NOW() - INTERVAL '90 days', NOW());

-- ============================================================================
-- PART 8: DEAL STAGE HISTORY
-- ============================================================================

INSERT INTO deal_stage_history (id, deal_id, from_stage, to_stage, changed_by_id, changed_by_name, notes, changed_at) VALUES
-- Won deal histories
(gen_random_uuid(), 'c0000004-0004-0004-0004-000000000004', NULL, 'lead', 'usr-lars-johansen-0003', 'Lars Johansen', 'Ny mulighet registrert', NOW() - INTERVAL '90 days'),
(gen_random_uuid(), 'c0000004-0004-0004-0004-000000000004', 'lead', 'qualified', 'usr-lars-johansen-0003', 'Lars Johansen', 'Kunde interessert', NOW() - INTERVAL '75 days'),
(gen_random_uuid(), 'c0000004-0004-0004-0004-000000000004', 'qualified', 'proposal', 'usr-lars-johansen-0003', 'Lars Johansen', 'Tilbud sendt', NOW() - INTERVAL '60 days'),
(gen_random_uuid(), 'c0000004-0004-0004-0004-000000000004', 'proposal', 'negotiation', 'usr-lars-johansen-0003', 'Lars Johansen', 'Forhandling startet', NOW() - INTERVAL '45 days'),
(gen_random_uuid(), 'c0000004-0004-0004-0004-000000000004', 'negotiation', 'won', 'usr-lars-johansen-0003', 'Lars Johansen', 'Kontrakt signert', NOW() - INTERVAL '30 days'),

(gen_random_uuid(), 'c0000009-0009-0009-0009-000000000009', NULL, 'lead', 'usr-ingrid-bakke-0006', 'Ingrid Bakke', 'Henvendelse mottatt', NOW() - INTERVAL '100 days'),
(gen_random_uuid(), 'c0000009-0009-0009-0009-000000000009', 'lead', 'qualified', 'usr-ingrid-bakke-0006', 'Ingrid Bakke', 'Befaring gjennomfort', NOW() - INTERVAL '85 days'),
(gen_random_uuid(), 'c0000009-0009-0009-0009-000000000009', 'qualified', 'proposal', 'usr-ingrid-bakke-0006', 'Ingrid Bakke', 'Tilbud levert', NOW() - INTERVAL '70 days'),
(gen_random_uuid(), 'c0000009-0009-0009-0009-000000000009', 'proposal', 'won', 'usr-ingrid-bakke-0006', 'Ingrid Bakke', 'Vunnet uten forhandling', NOW() - INTERVAL '45 days'),

-- Lost deal histories
(gen_random_uuid(), 'c0000006-0006-0006-0006-000000000006', NULL, 'lead', 'usr-marte-olsen-0002', 'Marte Olsen', 'Anbud identifisert', NOW() - INTERVAL '75 days'),
(gen_random_uuid(), 'c0000006-0006-0006-0006-000000000006', 'lead', 'qualified', 'usr-marte-olsen-0002', 'Marte Olsen', 'Kvalifisert for anbud', NOW() - INTERVAL '60 days'),
(gen_random_uuid(), 'c0000006-0006-0006-0006-000000000006', 'qualified', 'proposal', 'usr-marte-olsen-0002', 'Marte Olsen', 'Tilbud innlevert', NOW() - INTERVAL '45 days'),
(gen_random_uuid(), 'c0000006-0006-0006-0006-000000000006', 'proposal', 'lost', 'usr-marte-olsen-0002', 'Marte Olsen', 'Tapt pga pris', NOW() - INTERVAL '15 days'),

(gen_random_uuid(), 'c0000014-0014-0014-0014-000000000014', NULL, 'lead', 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Anbud registrert', NOW() - INTERVAL '130 days'),
(gen_random_uuid(), 'c0000014-0014-0014-0014-000000000014', 'lead', 'qualified', 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Prekvalifisert', NOW() - INTERVAL '110 days'),
(gen_random_uuid(), 'c0000014-0014-0014-0014-000000000014', 'qualified', 'proposal', 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Tilbud sendt', NOW() - INTERVAL '90 days'),
(gen_random_uuid(), 'c0000014-0014-0014-0014-000000000014', 'proposal', 'lost', 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Lokal konkurrent valgt', NOW() - INTERVAL '70 days');

-- ============================================================================
-- PART 9: OFFERS (18 offers)
-- ============================================================================

INSERT INTO offers (id, title, customer_id, customer_name, company_id, phase, probability, value, status, responsible_user_id, responsible_user_name, description, notes, created_at, updated_at) VALUES
-- Stalbygg offers
('d0000001-0001-0001-0001-000000000001', 'Lagerbygg Alnabru - Stalkonstruksjon', 'a0000001-0001-0001-0001-000000000001', 'Veidekke ASA', 'stalbygg', 'draft', 10, 12000000.00, 'active', 'usr-lars-johansen-0003', 'Lars Johansen', 'Komplett stalkonstruksjon for lagerbygg', NULL, NOW() - INTERVAL '8 days', NOW()),
('d0000002-0002-0002-0002-000000000002', 'Verkstedhall AF - Tilbud', 'a0000002-0002-0002-0002-000000000002', 'AF Gruppen ASA', 'stalbygg', 'in_progress', 30, 8500000.00, 'active', 'usr-lars-johansen-0003', 'Lars Johansen', 'Stalverksted med porter og kraner', NULL, NOW() - INTERVAL '20 days', NOW()),
('d0000003-0003-0003-0003-000000000003', 'Produksjonshall Skanska', 'a0000003-0003-0003-0003-000000000003', 'Skanska Norge AS', 'stalbygg', 'sent', 50, 15000000.00, 'active', 'usr-marte-olsen-0002', 'Marte Olsen', 'Produksjonshall med kontordel', 'Sendt til kunde for vurdering', NOW() - INTERVAL '35 days', NOW()),
('d0000004-0004-0004-0004-000000000004', 'Lagerbygg Gardermoen', 'a0000001-0001-0001-0001-000000000001', 'Veidekke ASA', 'stalbygg', 'won', 100, 18500000.00, 'won', 'usr-lars-johansen-0003', 'Lars Johansen', 'Stort lagerbygg ved OSL', 'Kontrakt signert', NOW() - INTERVAL '60 days', NOW()),
('d0000005-0005-0005-0005-000000000005', 'Logistikksenter Vestby', 'a0000002-0002-0002-0002-000000000002', 'AF Gruppen ASA', 'stalbygg', 'sent', 70, 22000000.00, 'active', 'usr-lars-johansen-0003', 'Lars Johansen', 'Logistikksenter med kjolelager', 'I forhandling', NOW() - INTERVAL '45 days', NOW()),
('d0000006-0006-0006-0006-000000000006', 'Flerbrukshall Drammen', 'a0000003-0003-0003-0003-000000000003', 'Skanska Norge AS', 'stalbygg', 'lost', 0, 9500000.00, 'lost', 'usr-marte-olsen-0002', 'Marte Olsen', 'Flerbrukshall for kommunen', 'Tapt til konkurrent', NOW() - INTERVAL '50 days', NOW()),

-- Hybridbygg offers
('d0000007-0007-0007-0007-000000000007', 'Kontorbygg Statsbygg', 'a0000004-0004-0004-0004-000000000004', 'Statsbygg', 'hybridbygg', 'sent', 60, 35000000.00, 'active', 'usr-anders-nilsen-0005', 'Anders Nilsen', 'Hybridkonstruksjon kontorbygg', NULL, NOW() - INTERVAL '40 days', NOW()),
('d0000008-0008-0008-0008-000000000008', 'Boligprosjekt Loren', 'a0000005-0005-0005-0005-000000000005', 'OBOS Eiendom AS', 'hybridbygg', 'in_progress', 40, 28000000.00, 'active', 'usr-ingrid-bakke-0006', 'Ingrid Bakke', 'Boligblokk i hybrid', NULL, NOW() - INTERVAL '30 days', NOW()),
('d0000009-0009-0009-0009-000000000009', 'Kontorbygg Majorstuen', 'a0000006-0006-0006-0006-000000000006', 'Selvaag Bolig ASA', 'hybridbygg', 'won', 100, 25000000.00, 'won', 'usr-ingrid-bakke-0006', 'Ingrid Bakke', 'Moderne kontorbygg sentrum', 'Vunnet', NOW() - INTERVAL '75 days', NOW()),
('d0000010-0010-0010-0010-000000000010', 'Universitetsbygg NTNU', 'a0000017-0017-0017-0017-000000000017', 'NTNU', 'hybridbygg', 'sent', 75, 42000000.00, 'active', 'usr-anders-nilsen-0005', 'Anders Nilsen', 'Undervisningsbygg NTNU', 'I forhandling', NOW() - INTERVAL '60 days', NOW()),

-- Industri offers
('d0000011-0011-0011-0011-000000000011', 'Verksted Hydro', 'a0000007-0007-0007-0007-000000000007', 'Norsk Hydro ASA', 'industri', 'draft', 15, 6500000.00, 'active', 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Vedlikeholdsverksted', NULL, NOW() - INTERVAL '4 days', NOW()),
('d0000012-0012-0012-0012-000000000012', 'Verksted Karmoy', 'a0000008-0008-0008-0008-000000000008', 'Elkem ASA', 'industri', 'won', 100, 11500000.00, 'won', 'usr-hanne-lie-0008', 'Hanne Lie', 'Produksjonsverksted', 'Prosjekt startet', NOW() - INTERVAL '90 days', NOW()),
('d0000013-0013-0013-0013-000000000013', 'Offshore modul Aibel', 'a0000009-0009-0009-0009-000000000009', 'Aibel AS', 'industri', 'sent', 55, 45000000.00, 'active', 'usr-thomas-dahl-0007', 'Thomas Dahl', 'Offshore modul for plattform', NULL, NOW() - INTERVAL '35 days', NOW()),

-- Tak offers
('d0000014-0014-0014-0014-000000000014', 'Takprosjekt Bryggen', 'a0000011-0011-0011-0011-000000000011', 'Bergen Kommune', 'tak', 'sent', 50, 4800000.00, 'active', 'usr-ole-martinsen-0009', 'Ole Martinsen', 'Historisk tak renovering', NULL, NOW() - INTERVAL '25 days', NOW()),
('d0000015-0015-0015-0015-000000000015', 'Takprosjekt Coop', 'a0000012-0012-0012-0012-000000000012', 'Coop Norge SA', 'tak', 'won', 100, 2800000.00, 'won', 'usr-ole-martinsen-0009', 'Ole Martinsen', 'Takutskifting varehus', 'Arbeid pagaende', NOW() - INTERVAL '45 days', NOW()),

-- Montasje offers
('d0000016-0016-0016-0016-000000000016', 'IKEA Innredning', 'a0000013-0013-0013-0013-000000000013', 'IKEA Norge AS', 'montasje', 'sent', 80, 8500000.00, 'active', 'usr-knut-svendsen-0011', 'Knut Svendsen', 'Komplett varehusinnredning', 'I forhandling', NOW() - INTERVAL '45 days', NOW()),
('d0000017-0017-0017-0017-000000000017', 'XXL Butikk Oslo', 'a0000014-0014-0014-0014-000000000014', 'XXL Sport og Villmark', 'montasje', 'in_progress', 40, 4200000.00, 'active', 'usr-knut-svendsen-0011', 'Knut Svendsen', 'Butikkinnredning sportsbutikk', NULL, NOW() - INTERVAL '15 days', NOW()),
('d0000018-0018-0018-0018-000000000018', 'Monteringsjobb Bodo', 'a0000014-0014-0014-0014-000000000014', 'XXL Sport og Villmark', 'montasje', 'won', 100, 2100000.00, 'won', 'usr-knut-svendsen-0011', 'Knut Svendsen', 'Butikkmontering nord', 'Ferdigstilt', NOW() - INTERVAL '70 days', NOW());

-- ============================================================================
-- PART 10: BUDGET DIMENSIONS FOR OFFERS
-- ============================================================================

INSERT INTO budget_dimensions (id, parent_type, parent_id, category_id, cost, revenue, target_margin_percent, description, display_order, created_at, updated_at) VALUES
-- Offer d0000004 (Lagerbygg Gardermoen - won)
(gen_random_uuid(), 'offer', 'd0000004-0004-0004-0004-000000000004', 'steel_structure', 9500000.00, 11100000.00, 15.0, 'Hovedkonstruksjon stal', 1, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000004-0004-0004-0004-000000000004', 'engineering', 850000.00, 1020000.00, 16.0, 'Prosjektering og beregninger', 2, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000004-0004-0004-0004-000000000004', 'assembly', 2800000.00, 3360000.00, 16.0, 'Montering pa plass', 3, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000004-0004-0004-0004-000000000004', 'transport', 650000.00, 780000.00, 17.0, 'Transport til byggeplass', 4, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000004-0004-0004-0004-000000000004', 'project_management', 950000.00, 1140000.00, 16.0, 'Prosjektledelse', 5, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000004-0004-0004-0004-000000000004', 'contingency', 850000.00, 1100000.00, 22.0, 'Uforutsett', 6, NOW(), NOW()),

-- Offer d0000009 (Kontorbygg Majorstuen - won)
(gen_random_uuid(), 'offer', 'd0000009-0009-0009-0009-000000000009', 'hybrid_structure', 12500000.00, 15000000.00, 16.7, 'Hybridkonstruksjon', 1, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000009-0009-0009-0009-000000000009', 'engineering', 1200000.00, 1440000.00, 16.7, 'Prosjektering', 2, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000009-0009-0009-0009-000000000009', 'assembly', 4200000.00, 5040000.00, 16.7, 'Montering', 3, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000009-0009-0009-0009-000000000009', 'project_management', 1400000.00, 1680000.00, 16.7, 'Prosjektledelse', 4, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000009-0009-0009-0009-000000000009', 'contingency', 1400000.00, 1840000.00, 24.0, 'Uforutsett', 5, NOW(), NOW()),

-- Offer d0000012 (Verksted Karmoy - won)
(gen_random_uuid(), 'offer', 'd0000012-0012-0012-0012-000000000012', 'steel_structure', 5800000.00, 6960000.00, 16.7, 'Stalkonstruksjon verksted', 1, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000012-0012-0012-0012-000000000012', 'cladding', 1800000.00, 2160000.00, 16.7, 'Kledning og porter', 2, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000012-0012-0012-0012-000000000012', 'assembly', 1500000.00, 1800000.00, 16.7, 'Montering', 3, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000012-0012-0012-0012-000000000012', 'contingency', 480000.00, 580000.00, 17.2, 'Uforutsett', 4, NOW(), NOW()),

-- Offer d0000015 (Takprosjekt Coop - won)
(gen_random_uuid(), 'offer', 'd0000015-0015-0015-0015-000000000015', 'roofing', 1800000.00, 2160000.00, 16.7, 'Taktekning', 1, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000015-0015-0015-0015-000000000015', 'assembly', 450000.00, 540000.00, 16.7, 'Montering', 2, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000015-0015-0015-0015-000000000015', 'contingency', 85000.00, 100000.00, 15.0, 'Uforutsett', 3, NOW(), NOW()),

-- Offer d0000018 (Monteringsjobb Bodo - won)
(gen_random_uuid(), 'offer', 'd0000018-0018-0018-0018-000000000018', 'assembly', 1400000.00, 1680000.00, 16.7, 'Monteringsarbeid', 1, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000018-0018-0018-0018-000000000018', 'transport', 250000.00, 300000.00, 16.7, 'Frakt', 2, NOW(), NOW()),
(gen_random_uuid(), 'offer', 'd0000018-0018-0018-0018-000000000018', 'contingency', 100000.00, 120000.00, 16.7, 'Uforutsett', 3, NOW(), NOW());

-- ============================================================================
-- PART 11: PROJECTS (15 projects)
-- ============================================================================

INSERT INTO projects (id, name, summary, description, customer_id, customer_name, company_id, status, start_date, end_date, budget, spent, manager_id, manager_name, team_members, offer_id, deal_id, has_detailed_budget, health, completion_percent, project_number, created_at, updated_at) VALUES
-- Stalbygg projects
('e0000001-0001-0001-0001-000000000001', 'Lagerbygg Gardermoen', 'Stort lagerbygg ved OSL', 'Komplett stalkonstruksjon lagerbygg 8000m2', 'a0000001-0001-0001-0001-000000000001', 'Veidekke ASA', 'stalbygg', 'active', NOW() - INTERVAL '25 days', NOW() + INTERVAL '120 days', 18500000.00, 4625000.00, 'usr-kristine-berg-0004', 'Kristine Berg', ARRAY['usr-lars-johansen-0003', 'usr-kristine-berg-0004'], 'd0000004-0004-0004-0004-000000000004', 'c0000004-0004-0004-0004-000000000004', true, 'on_track', 25.00, 'SB-2024-001', NOW() - INTERVAL '30 days', NOW()),
('e0000002-0002-0002-0002-000000000002', 'Produksjonshall Lillestrom', 'Produksjonshall for AF', 'Tidligere vunnet prosjekt', 'a0000002-0002-0002-0002-000000000002', 'AF Gruppen ASA', 'stalbygg', 'active', NOW() - INTERVAL '60 days', NOW() + INTERVAL '60 days', 12000000.00, 7200000.00, 'usr-kristine-berg-0004', 'Kristine Berg', ARRAY['usr-kristine-berg-0004'], NULL, NULL, true, 'on_track', 60.00, 'SB-2024-002', NOW() - INTERVAL '65 days', NOW()),
('e0000003-0003-0003-0003-000000000003', 'Lager Drammen', 'Mindre lagerbygg', 'Lagerbygg Skanska', 'a0000003-0003-0003-0003-000000000003', 'Skanska Norge AS', 'stalbygg', 'planning', NOW() + INTERVAL '30 days', NOW() + INTERVAL '150 days', 8500000.00, 0.00, 'usr-kristine-berg-0004', 'Kristine Berg', ARRAY['usr-kristine-berg-0004'], NULL, NULL, false, 'on_track', 0.00, 'SB-2024-003', NOW() - INTERVAL '5 days', NOW()),

-- Hybridbygg projects
('e0000004-0004-0004-0004-000000000004', 'Kontorbygg Majorstuen', 'Kontorbygg Selvaag', 'Moderne hybridkonstruksjon', 'a0000006-0006-0006-0006-000000000006', 'Selvaag Bolig ASA', 'hybridbygg', 'active', NOW() - INTERVAL '40 days', NOW() + INTERVAL '100 days', 25000000.00, 10000000.00, 'usr-anders-nilsen-0005', 'Anders Nilsen', ARRAY['usr-anders-nilsen-0005', 'usr-ingrid-bakke-0006'], 'd0000009-0009-0009-0009-000000000009', 'c0000009-0009-0009-0009-000000000009', true, 'on_track', 40.00, 'HB-2024-001', NOW() - INTERVAL '45 days', NOW()),
('e0000005-0005-0005-0005-000000000005', 'Boligblokk Grunerlokka', 'Boligprosjekt OBOS', 'Tidligere prosjekt', 'a0000005-0005-0005-0005-000000000005', 'OBOS Eiendom AS', 'hybridbygg', 'completed', NOW() - INTERVAL '180 days', NOW() - INTERVAL '30 days', 32000000.00, 31500000.00, 'usr-ingrid-bakke-0006', 'Ingrid Bakke', ARRAY['usr-ingrid-bakke-0006'], NULL, NULL, true, 'on_track', 100.00, 'HB-2023-005', NOW() - INTERVAL '185 days', NOW()),

-- Industri projects
('e0000006-0006-0006-0006-000000000006', 'Verksted Karmoy', 'Produksjonsverksted Elkem', 'Nytt verksted', 'a0000008-0008-0008-0008-000000000008', 'Elkem ASA', 'industri', 'active', NOW() - INTERVAL '55 days', NOW() + INTERVAL '30 days', 11500000.00, 9775000.00, 'usr-hanne-lie-0008', 'Hanne Lie', ARRAY['usr-hanne-lie-0008', 'usr-thomas-dahl-0007'], 'd0000012-0012-0012-0012-000000000012', 'c0000012-0012-0012-0012-000000000012', true, 'on_track', 85.00, 'IN-2024-001', NOW() - INTERVAL '60 days', NOW()),
('e0000007-0007-0007-0007-000000000007', 'Industrihall Bergen', 'Hall pa vent', 'Prosjekt pa vent', 'a0000009-0009-0009-0009-000000000009', 'Aibel AS', 'industri', 'on_hold', NOW() - INTERVAL '90 days', NOW() + INTERVAL '90 days', 18000000.00, 3600000.00, 'usr-thomas-dahl-0007', 'Thomas Dahl', ARRAY['usr-thomas-dahl-0007'], NULL, NULL, false, 'at_risk', 20.00, 'IN-2024-002', NOW() - INTERVAL '95 days', NOW()),
('e0000008-0008-0008-0008-000000000008', 'Prosessanlegg Mongstad', 'Utvidelse prosessanlegg', 'Tidlig fase', 'a0000018-0018-0018-0018-000000000018', 'Equinor ASA', 'industri', 'planning', NOW() + INTERVAL '60 days', NOW() + INTERVAL '240 days', 65000000.00, 0.00, 'usr-thomas-dahl-0007', 'Thomas Dahl', ARRAY['usr-thomas-dahl-0007', 'usr-hanne-lie-0008'], NULL, NULL, false, 'on_track', 0.00, 'IN-2024-003', NOW() - INTERVAL '20 days', NOW()),

-- Tak projects
('e0000009-0009-0009-0009-000000000009', 'Takprosjekt Coop Obs', 'Takutskifting varehus', 'Pagaende takarbeid', 'a0000012-0012-0012-0012-000000000012', 'Coop Norge SA', 'tak', 'active', NOW() - INTERVAL '15 days', NOW() + INTERVAL '45 days', 2800000.00, 1400000.00, 'usr-ole-martinsen-0009', 'Ole Martinsen', ARRAY['usr-ole-martinsen-0009', 'usr-silje-haugen-0010'], 'd0000015-0015-0015-0015-000000000015', 'c0000018-0018-0018-0018-000000000018', true, 'on_track', 50.00, 'TK-2024-001', NOW() - INTERVAL '20 days', NOW()),
('e0000010-0010-0010-0010-000000000010', 'Skoletak Stavanger', 'Skoletakrenovering', 'Ferdigstilt prosjekt', 'a0000011-0011-0011-0011-000000000011', 'Bergen Kommune', 'tak', 'completed', NOW() - INTERVAL '120 days', NOW() - INTERVAL '15 days', 1900000.00, 1850000.00, 'usr-silje-haugen-0010', 'Silje Haugen', ARRAY['usr-silje-haugen-0010'], NULL, 'c0000019-0019-0019-0019-000000000019', true, 'on_track', 100.00, 'TK-2023-008', NOW() - INTERVAL '150 days', NOW()),
('e0000011-0011-0011-0011-000000000011', 'Tak Trondheim Kommune', 'Diverse takarbeid', 'Rammeavtale takarbeid', 'a0000010-0010-0010-0010-000000000010', 'Trondheim Kommune', 'tak', 'active', NOW() - INTERVAL '30 days', NOW() + INTERVAL '90 days', 3500000.00, 1050000.00, 'usr-ole-martinsen-0009', 'Ole Martinsen', ARRAY['usr-ole-martinsen-0009'], NULL, NULL, false, 'on_track', 30.00, 'TK-2024-002', NOW() - INTERVAL '35 days', NOW()),

-- Montasje projects
('e0000012-0012-0012-0012-000000000012', 'Monteringsjobb Bodo', 'XXL butikkmontering', 'Ferdigstilt prosjekt', 'a0000014-0014-0014-0014-000000000014', 'XXL Sport og Villmark', 'montasje', 'completed', NOW() - INTERVAL '70 days', NOW() - INTERVAL '10 days', 2100000.00, 2050000.00, 'usr-knut-svendsen-0011', 'Knut Svendsen', ARRAY['usr-knut-svendsen-0011'], 'd0000018-0018-0018-0018-000000000018', 'c0000023-0023-0023-0023-000000000023', true, 'on_track', 100.00, 'MO-2024-001', NOW() - INTERVAL '100 days', NOW()),
('e0000013-0013-0013-0013-000000000013', 'IKEA Lager Vestby', 'Lagerinnredning IKEA', 'Tidligere prosjekt', 'a0000013-0013-0013-0013-000000000013', 'IKEA Norge AS', 'montasje', 'completed', NOW() - INTERVAL '150 days', NOW() - INTERVAL '60 days', 5500000.00, 5400000.00, 'usr-knut-svendsen-0011', 'Knut Svendsen', ARRAY['usr-knut-svendsen-0011', 'usr-maria-holm-0012'], NULL, NULL, true, 'on_track', 100.00, 'MO-2023-004', NOW() - INTERVAL '155 days', NOW()),
('e0000014-0014-0014-0014-000000000014', 'Bama Lagerinnredning', 'Lagerhyller Bama', 'Planlagt prosjekt', 'a0000015-0015-0015-0015-000000000015', 'Bama Gruppen AS', 'montasje', 'planning', NOW() + INTERVAL '45 days', NOW() + INTERVAL '105 days', 3800000.00, 0.00, 'usr-maria-holm-0012', 'Maria Holm', ARRAY['usr-maria-holm-0012'], NULL, NULL, false, 'on_track', 0.00, 'MO-2024-002', NOW() - INTERVAL '10 days', NOW()),
('e0000015-0015-0015-0015-000000000015', 'XXL Butikk Oslo', 'Butikkinnredning', 'Tidlig planlegging', 'a0000014-0014-0014-0014-000000000014', 'XXL Sport og Villmark', 'montasje', 'planning', NOW() + INTERVAL '75 days', NOW() + INTERVAL '135 days', 4200000.00, 0.00, 'usr-knut-svendsen-0011', 'Knut Svendsen', ARRAY['usr-knut-svendsen-0011'], NULL, 'c0000021-0021-0021-0021-000000000021', false, 'on_track', 0.00, 'MO-2024-003', NOW() - INTERVAL '15 days', NOW());

-- ============================================================================
-- PART 12: BUDGET DIMENSIONS FOR PROJECTS
-- ============================================================================

INSERT INTO budget_dimensions (id, parent_type, parent_id, category_id, cost, revenue, target_margin_percent, description, display_order, created_at, updated_at) VALUES
-- Project e0000001 (Lagerbygg Gardermoen)
(gen_random_uuid(), 'project', 'e0000001-0001-0001-0001-000000000001', 'steel_structure', 9500000.00, 11100000.00, 15.0, 'Hovedkonstruksjon stal', 1, NOW(), NOW()),
(gen_random_uuid(), 'project', 'e0000001-0001-0001-0001-000000000001', 'engineering', 850000.00, 1020000.00, 16.0, 'Prosjektering', 2, NOW(), NOW()),
(gen_random_uuid(), 'project', 'e0000001-0001-0001-0001-000000000001', 'assembly', 2800000.00, 3360000.00, 16.0, 'Montering', 3, NOW(), NOW()),
(gen_random_uuid(), 'project', 'e0000001-0001-0001-0001-000000000001', 'transport', 650000.00, 780000.00, 17.0, 'Transport', 4, NOW(), NOW()),

-- Project e0000004 (Kontorbygg Majorstuen)
(gen_random_uuid(), 'project', 'e0000004-0004-0004-0004-000000000004', 'hybrid_structure', 12500000.00, 15000000.00, 16.7, 'Hybridkonstruksjon', 1, NOW(), NOW()),
(gen_random_uuid(), 'project', 'e0000004-0004-0004-0004-000000000004', 'engineering', 1200000.00, 1440000.00, 16.7, 'Prosjektering', 2, NOW(), NOW()),
(gen_random_uuid(), 'project', 'e0000004-0004-0004-0004-000000000004', 'assembly', 4200000.00, 5040000.00, 16.7, 'Montering', 3, NOW(), NOW()),

-- Project e0000006 (Verksted Karmoy)
(gen_random_uuid(), 'project', 'e0000006-0006-0006-0006-000000000006', 'steel_structure', 5800000.00, 6960000.00, 16.7, 'Stalkonstruksjon', 1, NOW(), NOW()),
(gen_random_uuid(), 'project', 'e0000006-0006-0006-0006-000000000006', 'cladding', 1800000.00, 2160000.00, 16.7, 'Kledning', 2, NOW(), NOW()),
(gen_random_uuid(), 'project', 'e0000006-0006-0006-0006-000000000006', 'assembly', 1500000.00, 1800000.00, 16.7, 'Montering', 3, NOW(), NOW()),

-- Project e0000009 (Takprosjekt Coop)
(gen_random_uuid(), 'project', 'e0000009-0009-0009-0009-000000000009', 'roofing', 1800000.00, 2160000.00, 16.7, 'Taktekning', 1, NOW(), NOW()),
(gen_random_uuid(), 'project', 'e0000009-0009-0009-0009-000000000009', 'assembly', 450000.00, 540000.00, 16.7, 'Montering', 2, NOW(), NOW());

-- ============================================================================
-- PART 13: ACTIVITIES (40 activities across all entities)
-- ============================================================================

INSERT INTO activities (id, target_type, target_id, title, body, occurred_at, creator_name, activity_type, status, scheduled_at, due_date, completed_at, duration_minutes, priority, is_private, creator_id, assigned_to_id, company_id, attendees, created_at, updated_at) VALUES
-- Customer activities
(gen_random_uuid(), 'Customer', 'a0000001-0001-0001-0001-000000000001', 'Introduksjonsmote Veidekke', 'Forste mote med ny kontaktperson', NOW() - INTERVAL '140 days', 'Lars Johansen', 'meeting', 'completed', NOW() - INTERVAL '140 days', NULL, NOW() - INTERVAL '140 days', 60, 2, false, 'usr-lars-johansen-0003', NULL, 'stalbygg', ARRAY['usr-lars-johansen-0003'], NOW() - INTERVAL '145 days', NOW()),
(gen_random_uuid(), 'Customer', 'a0000002-0002-0002-0002-000000000002', 'Befaring AF Gruppen', 'Befaring av eksisterende anlegg', NOW() - INTERVAL '100 days', 'Marte Olsen', 'meeting', 'completed', NOW() - INTERVAL '100 days', NULL, NOW() - INTERVAL '100 days', 120, 2, false, 'usr-marte-olsen-0002', NULL, 'stalbygg', ARRAY['usr-marte-olsen-0002', 'usr-lars-johansen-0003'], NOW() - INTERVAL '105 days', NOW()),
(gen_random_uuid(), 'Customer', 'a0000004-0004-0004-0004-000000000004', 'Statsbygg presentasjon', 'Presentasjon av vare losninger', NOW() - INTERVAL '80 days', 'Anders Nilsen', 'meeting', 'completed', NOW() - INTERVAL '80 days', NULL, NOW() - INTERVAL '80 days', 90, 2, false, 'usr-anders-nilsen-0005', NULL, 'hybridbygg', ARRAY['usr-anders-nilsen-0005', 'usr-ingrid-bakke-0006'], NOW() - INTERVAL '85 days', NOW()),
(gen_random_uuid(), 'Customer', 'a0000007-0007-0007-0007-000000000007', 'Strategimote Hydro', 'Langsiktig samarbeidsplanlegging', NOW() - INTERVAL '70 days', 'Thomas Dahl', 'meeting', 'completed', NOW() - INTERVAL '70 days', NULL, NOW() - INTERVAL '70 days', 90, 2, false, 'usr-thomas-dahl-0007', NULL, 'industri', ARRAY['usr-thomas-dahl-0007', 'usr-hanne-lie-0008'], NOW() - INTERVAL '75 days', NOW()),

-- Deal activities
(gen_random_uuid(), 'Deal', 'c0000004-0004-0004-0004-000000000004', 'Tilbudsgjennomgang Gardermoen', 'Gjennomgang med teknisk team', NOW() - INTERVAL '65 days', 'Lars Johansen', 'meeting', 'completed', NOW() - INTERVAL '65 days', NULL, NOW() - INTERVAL '65 days', 120, 3, false, 'usr-lars-johansen-0003', NULL, 'stalbygg', ARRAY['usr-lars-johansen-0003', 'usr-kristine-berg-0004'], NOW() - INTERVAL '70 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000004-0004-0004-0004-000000000004', 'Kontraktsforhandling', 'Forhandling av endelige vilkar', NOW() - INTERVAL '35 days', 'Marte Olsen', 'meeting', 'completed', NOW() - INTERVAL '35 days', NULL, NOW() - INTERVAL '35 days', 180, 3, false, 'usr-marte-olsen-0002', NULL, 'stalbygg', ARRAY['usr-marte-olsen-0002', 'usr-lars-johansen-0003'], NOW() - INTERVAL '40 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000005-0005-0005-0005-000000000005', 'Befaring Vestby', 'Befaring pa tomten', NOW() - INTERVAL '45 days', 'Lars Johansen', 'meeting', 'completed', NOW() - INTERVAL '45 days', NULL, NOW() - INTERVAL '45 days', 90, 2, false, 'usr-lars-johansen-0003', NULL, 'stalbygg', ARRAY['usr-lars-johansen-0003'], NOW() - INTERVAL '50 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000009-0009-0009-0009-000000000009', 'Oppstartsmote Majorstuen', 'Kickoff for prosjekt', NOW() - INTERVAL '40 days', 'Ingrid Bakke', 'meeting', 'completed', NOW() - INTERVAL '40 days', NULL, NOW() - INTERVAL '40 days', 60, 2, false, 'usr-ingrid-bakke-0006', NULL, 'hybridbygg', ARRAY['usr-ingrid-bakke-0006', 'usr-anders-nilsen-0005'], NOW() - INTERVAL '45 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000013-0013-0013-0013-000000000013', 'Teknisk spesifikasjon offshore', 'Gjennomgang av krav', NOW() - INTERVAL '18 days', 'Thomas Dahl', 'meeting', 'completed', NOW() - INTERVAL '18 days', NULL, NOW() - INTERVAL '18 days', 180, 3, false, 'usr-thomas-dahl-0007', NULL, 'industri', ARRAY['usr-thomas-dahl-0007', 'usr-hanne-lie-0008'], NOW() - INTERVAL '20 days', NOW()),

-- Project activities
(gen_random_uuid(), 'Project', 'e0000001-0001-0001-0001-000000000001', 'Oppstartsmoete Gardermoen', 'Kickoff med alle parter', NOW() - INTERVAL '25 days', 'Kristine Berg', 'meeting', 'completed', NOW() - INTERVAL '25 days', NULL, NOW() - INTERVAL '25 days', 90, 2, false, 'usr-kristine-berg-0004', NULL, 'stalbygg', ARRAY['usr-kristine-berg-0004', 'usr-lars-johansen-0003'], NOW() - INTERVAL '28 days', NOW()),
(gen_random_uuid(), 'Project', 'e0000001-0001-0001-0001-000000000001', 'Ukentlig statusmote', 'Gjennomgang av fremdrift', NOW() - INTERVAL '7 days', 'Kristine Berg', 'meeting', 'completed', NOW() - INTERVAL '7 days', NULL, NOW() - INTERVAL '7 days', 45, 1, false, 'usr-kristine-berg-0004', NULL, 'stalbygg', ARRAY['usr-kristine-berg-0004'], NOW() - INTERVAL '8 days', NOW()),
(gen_random_uuid(), 'Project', 'e0000004-0004-0004-0004-000000000004', 'Byggeplassmote Majorstuen', 'Statusgjennomgang pa plass', NOW() - INTERVAL '15 days', 'Anders Nilsen', 'meeting', 'completed', NOW() - INTERVAL '15 days', NULL, NOW() - INTERVAL '15 days', 60, 2, false, 'usr-anders-nilsen-0005', NULL, 'hybridbygg', ARRAY['usr-anders-nilsen-0005', 'usr-ingrid-bakke-0006'], NOW() - INTERVAL '18 days', NOW()),
(gen_random_uuid(), 'Project', 'e0000006-0006-0006-0006-000000000006', 'Sluttbefaring Karmoy', 'Befaring for overlevering', NOW() - INTERVAL '5 days', 'Hanne Lie', 'meeting', 'completed', NOW() - INTERVAL '5 days', NULL, NOW() - INTERVAL '5 days', 120, 2, false, 'usr-hanne-lie-0008', NULL, 'industri', ARRAY['usr-hanne-lie-0008', 'usr-thomas-dahl-0007'], NOW() - INTERVAL '7 days', NOW()),
(gen_random_uuid(), 'Project', 'e0000009-0009-0009-0009-000000000009', 'Fremdriftsmote Coop tak', 'Ukentlig oppfolging', NOW() - INTERVAL '3 days', 'Ole Martinsen', 'meeting', 'completed', NOW() - INTERVAL '3 days', NULL, NOW() - INTERVAL '3 days', 30, 1, false, 'usr-ole-martinsen-0009', NULL, 'tak', ARRAY['usr-ole-martinsen-0009', 'usr-silje-haugen-0010'], NOW() - INTERVAL '4 days', NOW()),

-- Call activities
(gen_random_uuid(), 'Customer', 'a0000003-0003-0003-0003-000000000003', 'Telefonsamtale Skanska', 'Oppfolging av tilbud', NOW() - INTERVAL '30 days', 'Lars Johansen', 'call', 'completed', NULL, NULL, NOW() - INTERVAL '30 days', 15, 1, false, 'usr-lars-johansen-0003', NULL, 'stalbygg', NULL, NOW() - INTERVAL '30 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000008-0008-0008-0008-000000000008', 'Samtale med OBOS', 'Oppfolging av interesse', NOW() - INTERVAL '25 days', 'Ingrid Bakke', 'call', 'completed', NULL, NULL, NOW() - INTERVAL '25 days', 20, 1, false, 'usr-ingrid-bakke-0006', NULL, 'hybridbygg', NULL, NOW() - INTERVAL '25 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000016-0016-0016-0016-000000000016', 'Samtale Trondheim Kommune', 'Avklaring av behov', NOW() - INTERVAL '6 days', 'Silje Haugen', 'call', 'completed', NULL, NULL, NOW() - INTERVAL '6 days', 25, 1, false, 'usr-silje-haugen-0010', NULL, 'tak', NULL, NOW() - INTERVAL '6 days', NOW()),

-- Email activities
(gen_random_uuid(), 'Offer', 'd0000003-0003-0003-0003-000000000003', 'Tilbud sendt Skanska', 'Tilbud for produksjonshall sendt', NOW() - INTERVAL '35 days', 'Marte Olsen', 'email', 'completed', NULL, NULL, NOW() - INTERVAL '35 days', NULL, 2, false, 'usr-marte-olsen-0002', NULL, 'stalbygg', NULL, NOW() - INTERVAL '35 days', NOW()),
(gen_random_uuid(), 'Offer', 'd0000007-0007-0007-0007-000000000007', 'Tilbud sendt Statsbygg', 'Komplett tilbudspakke sendt', NOW() - INTERVAL '40 days', 'Anders Nilsen', 'email', 'completed', NULL, NULL, NOW() - INTERVAL '40 days', NULL, 2, false, 'usr-anders-nilsen-0005', NULL, 'hybridbygg', NULL, NOW() - INTERVAL '40 days', NOW()),
(gen_random_uuid(), 'Offer', 'd0000013-0013-0013-0013-000000000013', 'Tilbud sendt Aibel', 'Offshore modul tilbud', NOW() - INTERVAL '35 days', 'Thomas Dahl', 'email', 'completed', NULL, NULL, NOW() - INTERVAL '35 days', NULL, 3, false, 'usr-thomas-dahl-0007', NULL, 'industri', NULL, NOW() - INTERVAL '35 days', NOW()),

-- Note activities
(gen_random_uuid(), 'Project', 'e0000007-0007-0007-0007-000000000007', 'Prosjekt satt pa vent', 'Finansiering ikke pa plass', NOW() - INTERVAL '30 days', 'Thomas Dahl', 'note', 'completed', NULL, NULL, NOW() - INTERVAL '30 days', NULL, 3, false, 'usr-thomas-dahl-0007', NULL, 'industri', NULL, NOW() - INTERVAL '30 days', NOW()),
(gen_random_uuid(), 'Project', 'e0000010-0010-0010-0010-000000000010', 'Prosjekt ferdigstilt', 'Skoletak overlevert til kunde', NOW() - INTERVAL '15 days', 'Ole Martinsen', 'note', 'completed', NULL, NULL, NOW() - INTERVAL '15 days', NULL, 2, false, 'usr-ole-martinsen-0009', NULL, 'tak', NULL, NOW() - INTERVAL '15 days', NOW()),
(gen_random_uuid(), 'Offer', 'd0000004-0004-0004-0004-000000000004', 'Tilbud vunnet', 'Gardermoen kontrakt signert', NOW() - INTERVAL '30 days', 'Lars Johansen', 'note', 'completed', NULL, NULL, NOW() - INTERVAL '30 days', NULL, 3, false, 'usr-lars-johansen-0003', NULL, 'stalbygg', NULL, NOW() - INTERVAL '30 days', NOW()),

-- Pending tasks
(gen_random_uuid(), 'Customer', 'a0000001-0001-0001-0001-000000000001', 'Folge opp Veidekke Q1', 'Avtale mote for Q1 planer', NOW() + INTERVAL '3 days', 'Lars Johansen', 'task', 'planned', NULL, NOW() + INTERVAL '5 days', NULL, NULL, 2, false, 'usr-lars-johansen-0003', 'usr-lars-johansen-0003', 'stalbygg', NULL, NOW() - INTERVAL '2 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000005-0005-0005-0005-000000000005', 'Ferdigstille tilbud Vestby', 'Komplett tilbud ma vaere klart', NOW() + INTERVAL '5 days', 'Lars Johansen', 'task', 'in_progress', NULL, NOW() + INTERVAL '7 days', NULL, NULL, 3, false, 'usr-lars-johansen-0003', 'usr-lars-johansen-0003', 'stalbygg', NULL, NOW() - INTERVAL '3 days', NOW()),
(gen_random_uuid(), 'Project', 'e0000001-0001-0001-0001-000000000001', 'Bestille stalbjelker fase 2', 'Innkjopsordre for neste fase', NOW() + INTERVAL '2 days', 'Kristine Berg', 'task', 'planned', NULL, NOW() + INTERVAL '4 days', NULL, NULL, 2, false, 'usr-kristine-berg-0004', 'usr-kristine-berg-0004', 'stalbygg', NULL, NOW() - INTERVAL '1 day', NOW()),
(gen_random_uuid(), 'Project', 'e0000004-0004-0004-0004-000000000004', 'Koordineringsmote underentreprenor', 'Mote med elektriker og rorlegger', NOW() + INTERVAL '7 days', 'Anders Nilsen', 'meeting', 'planned', NOW() + INTERVAL '7 days', NULL, NULL, 90, 2, false, 'usr-anders-nilsen-0005', 'usr-anders-nilsen-0005', 'hybridbygg', NULL, NOW(), NOW()),
(gen_random_uuid(), 'Deal', 'c0000008-0008-0008-0008-000000000008', 'Oppfolgingssamtale Selvaag', 'Sjekke status pa tilbudet', NOW() + INTERVAL '2 days', 'Ingrid Bakke', 'call', 'planned', NOW() + INTERVAL '2 days', NULL, NULL, 30, 2, false, 'usr-ingrid-bakke-0006', 'usr-ingrid-bakke-0006', 'hybridbygg', NULL, NOW(), NOW()),

-- Contact activities
(gen_random_uuid(), 'Contact', 'b0000001-0001-0001-0001-000000000001', 'Lunsj med Per', 'Uformell lunsj for relasjonsbygging', NOW() - INTERVAL '45 days', 'Lars Johansen', 'meeting', 'completed', NOW() - INTERVAL '45 days', NULL, NOW() - INTERVAL '45 days', 60, 1, false, 'usr-lars-johansen-0003', NULL, 'stalbygg', ARRAY['usr-lars-johansen-0003'], NOW() - INTERVAL '50 days', NOW()),
(gen_random_uuid(), 'Contact', 'b0000005-0005-0005-0005-000000000005', 'Teknisk samtale Skanska', 'Diskusjon om konstruksjonsmetoder', NOW() - INTERVAL '30 days', 'Marte Olsen', 'call', 'completed', NULL, NULL, NOW() - INTERVAL '30 days', 45, 1, false, 'usr-marte-olsen-0002', NULL, 'stalbygg', NULL, NOW() - INTERVAL '30 days', NOW()),
(gen_random_uuid(), 'Contact', 'b0000009-0009-0009-0009-000000000009', 'E-post oppfolging OBOS', 'Sendt info om kommende prosjekter', NOW() - INTERVAL '20 days', 'Anders Nilsen', 'email', 'completed', NULL, NULL, NOW() - INTERVAL '20 days', NULL, 1, false, 'usr-anders-nilsen-0005', NULL, 'hybridbygg', NULL, NOW() - INTERVAL '20 days', NOW()),

-- More diverse activities
(gen_random_uuid(), 'Customer', 'a0000010-0010-0010-0010-000000000010', 'Introduksjonsmote Trondheim', 'Forste mote med ny kontaktperson', NOW() - INTERVAL '50 days', 'Silje Haugen', 'meeting', 'completed', NOW() - INTERVAL '50 days', NULL, NOW() - INTERVAL '50 days', 60, 2, false, 'usr-silje-haugen-0010', NULL, 'tak', ARRAY['usr-silje-haugen-0010', 'usr-ole-martinsen-0009'], NOW() - INTERVAL '55 days', NOW()),
(gen_random_uuid(), 'Customer', 'a0000012-0012-0012-0012-000000000012', 'Rammeavtalemote Coop', 'Arlig gjennomgang av rammeavtale', NOW() - INTERVAL '40 days', 'Ole Martinsen', 'meeting', 'completed', NOW() - INTERVAL '40 days', NULL, NOW() - INTERVAL '40 days', 90, 2, false, 'usr-ole-martinsen-0009', NULL, 'tak', ARRAY['usr-ole-martinsen-0009', 'usr-silje-haugen-0010'], NOW() - INTERVAL '45 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000017-0017-0017-0017-000000000017', 'Kulturminneavklaring Bryggen', 'Mote med vernemyndigheter', NOW() - INTERVAL '12 days', 'Ole Martinsen', 'meeting', 'completed', NOW() - INTERVAL '12 days', NULL, NOW() - INTERVAL '12 days', 120, 2, false, 'usr-ole-martinsen-0009', NULL, 'tak', ARRAY['usr-ole-martinsen-0009'], NOW() - INTERVAL '14 days', NOW()),
(gen_random_uuid(), 'Project', 'e0000012-0012-0012-0012-000000000012', 'Ferdigstillelse Bodo', 'Prosjekt overlevert', NOW() - INTERVAL '10 days', 'Knut Svendsen', 'note', 'completed', NULL, NULL, NOW() - INTERVAL '10 days', NULL, 2, false, 'usr-knut-svendsen-0011', NULL, 'montasje', NULL, NOW() - INTERVAL '10 days', NOW()),
(gen_random_uuid(), 'Deal', 'c0000020-0020-0020-0020-000000000020', 'Forhandlingsrunde IKEA', 'Gjennomgang av vilkar', NOW() - INTERVAL '5 days', 'Knut Svendsen', 'meeting', 'completed', NOW() - INTERVAL '5 days', NULL, NOW() - INTERVAL '5 days', 60, 3, false, 'usr-knut-svendsen-0011', NULL, 'montasje', ARRAY['usr-knut-svendsen-0011', 'usr-maria-holm-0012'], NOW() - INTERVAL '6 days', NOW());

-- ============================================================================
-- PART 14: NOTIFICATIONS (20 notifications for users)
-- Note: notifications.user_id is UUID type
-- ============================================================================

-- First, we need to create a helper to convert user string IDs to UUIDs
-- Since notifications.user_id expects UUID, we'll use md5 to generate consistent UUIDs from user IDs

INSERT INTO notifications (id, user_id, type, title, message, read, entity_id, entity_type, created_at, updated_at) VALUES
-- Unread notifications (recent)
(gen_random_uuid(), md5('usr-lars-johansen-0003')::uuid, 'task_assigned', 'Ny oppgave tildelt', 'Du har fatt en ny oppgave: Folge opp Veidekke Q1 plan', false, 'a0000001-0001-0001-0001-000000000001', 'Customer', NOW() - INTERVAL '2 days', NOW()),
(gen_random_uuid(), md5('usr-lars-johansen-0003')::uuid, 'deal_stage_changed', 'Deal oppdatert', 'Logistikksenter Vestby er na i forhandling-fasen', false, 'c0000005-0005-0005-0005-000000000005', 'Deal', NOW() - INTERVAL '1 day', NOW()),
(gen_random_uuid(), md5('usr-kristine-berg-0004')::uuid, 'project_update', 'Prosjektoppdatering', 'Lagerbygg Gardermoen: 25% ferdigstilt', false, 'e0000001-0001-0001-0001-000000000001', 'Project', NOW() - INTERVAL '1 day', NOW()),
(gen_random_uuid(), md5('usr-anders-nilsen-0005')::uuid, 'activity_reminder', 'Motepaminnelse', 'Koordineringsmote underentreprenor om 2 dager', false, 'e0000004-0004-0004-0004-000000000004', 'Project', NOW() - INTERVAL '12 hours', NOW()),
(gen_random_uuid(), md5('usr-ingrid-bakke-0006')::uuid, 'task_assigned', 'Oppfolging planlagt', 'Husk oppfolgingssamtale med Selvaag', false, 'c0000008-0008-0008-0008-000000000008', 'Deal', NOW() - INTERVAL '6 hours', NOW()),
(gen_random_uuid(), md5('usr-thomas-dahl-0007')::uuid, 'deal_stage_changed', 'Deal fremgang', 'Offshore modul er na i tilbudsfasen', false, 'c0000013-0013-0013-0013-000000000013', 'Deal', NOW() - INTERVAL '3 days', NOW()),
(gen_random_uuid(), md5('usr-hanne-lie-0008')::uuid, 'project_update', 'Prosjekt nesten ferdig', 'Verksted Karmoy: 85% ferdigstilt', false, 'e0000006-0006-0006-0006-000000000006', 'Project', NOW() - INTERVAL '2 days', NOW()),
(gen_random_uuid(), md5('usr-ole-martinsen-0009')::uuid, 'budget_alert', 'Budsjettvarsel', 'Takprosjekt Coop narmer seg budsjettgrense', false, 'e0000009-0009-0009-0009-000000000009', 'Project', NOW() - INTERVAL '4 days', NOW()),
(gen_random_uuid(), md5('usr-silje-haugen-0010')::uuid, 'deal_stage_changed', 'Ny mulighet', 'Takutskifting Moholt er registrert som lead', false, 'c0000016-0016-0016-0016-000000000016', 'Deal', NOW() - INTERVAL '3 days', NOW()),
(gen_random_uuid(), md5('usr-knut-svendsen-0011')::uuid, 'deal_stage_changed', 'Deal i forhandling', 'IKEA innredning narmer seg signering', false, 'c0000020-0020-0020-0020-000000000020', 'Deal', NOW() - INTERVAL '1 day', NOW()),

-- Read notifications (historical)
(gen_random_uuid(), md5('usr-lars-johansen-0003')::uuid, 'offer_accepted', 'Tilbud akseptert!', 'Gratulerer! Lagerbygg Gardermoen er vunnet', true, 'd0000004-0004-0004-0004-000000000004', 'Offer', NOW() - INTERVAL '30 days', NOW() - INTERVAL '28 days'),
(gen_random_uuid(), md5('usr-marte-olsen-0002')::uuid, 'offer_rejected', 'Tilbud tapt', 'Flerbrukshall Drammen gikk til konkurrent', true, 'd0000006-0006-0006-0006-000000000006', 'Offer', NOW() - INTERVAL '40 days', NOW() - INTERVAL '38 days'),
(gen_random_uuid(), md5('usr-ingrid-bakke-0006')::uuid, 'offer_accepted', 'Tilbud akseptert!', 'Kontorbygg Majorstuen er vunnet', true, 'd0000009-0009-0009-0009-000000000009', 'Offer', NOW() - INTERVAL '45 days', NOW() - INTERVAL '43 days'),
(gen_random_uuid(), md5('usr-hanne-lie-0008')::uuid, 'offer_accepted', 'Tilbud akseptert!', 'Verksted Karmoy er vunnet', true, 'd0000012-0012-0012-0012-000000000012', 'Offer', NOW() - INTERVAL '60 days', NOW() - INTERVAL '58 days'),
(gen_random_uuid(), md5('usr-ole-martinsen-0009')::uuid, 'project_update', 'Prosjekt ferdig', 'Skoletak Stavanger er ferdigstilt og overlevert', true, 'e0000010-0010-0010-0010-000000000010', 'Project', NOW() - INTERVAL '15 days', NOW() - INTERVAL '13 days'),
(gen_random_uuid(), md5('usr-knut-svendsen-0011')::uuid, 'project_update', 'Prosjekt ferdig', 'Monteringsjobb Bodo er ferdigstilt', true, 'e0000012-0012-0012-0012-000000000012', 'Project', NOW() - INTERVAL '10 days', NOW() - INTERVAL '8 days'),
(gen_random_uuid(), md5('usr-kristine-berg-0004')::uuid, 'task_assigned', 'Oppgave fullfort', 'Du har fullfort: Oppstartsmote Gardermoen', true, 'e0000001-0001-0001-0001-000000000001', 'Project', NOW() - INTERVAL '25 days', NOW() - INTERVAL '24 days'),
(gen_random_uuid(), md5('usr-anders-nilsen-0005')::uuid, 'deal_stage_changed', 'Deal vunnet!', 'Universitetsbygg NTNU er na i forhandlingsfasen', true, 'c0000010-0010-0010-0010-000000000010', 'Deal', NOW() - INTERVAL '10 days', NOW() - INTERVAL '8 days'),
(gen_random_uuid(), md5('usr-erik-hansen-0001')::uuid, 'project_update', 'Systemstatus', 'Alle systemer opererer normalt', true, NULL, NULL, NOW() - INTERVAL '7 days', NOW() - INTERVAL '5 days'),
(gen_random_uuid(), md5('usr-maria-holm-0012')::uuid, 'project_update', 'Prosjektinformasjon', 'Du er lagt til i team for XXL Butikk Oslo', false, 'e0000015-0015-0015-0015-000000000015', 'Project', NOW() - INTERVAL '3 days', NOW());

-- ============================================================================
-- END OF SEED DATA
-- ============================================================================

-- Summary:
-- Users: 12
-- User Roles: 12
-- Customers: 18
-- Contacts: 36
-- Contact Relationships: 18
-- Deals: 25
-- Deal Stage History: 18
-- Offers: 18
-- Budget Dimensions (Offers): 22
-- Projects: 15
-- Budget Dimensions (Projects): 12
-- Activities: 40
-- Notifications: 20
--
-- Total: ~246 records
