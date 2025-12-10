-- Import customers from customers_20251210075553.xlsx and Plan Straye Tak new.xlsx
-- Generated: 2025-12-10 20:25:34.069200
-- Total customers: 168
-- From export file: 99
-- From Plan Straye Tak only: 60
-- From offers (referenced but not in Excel files): 14
-- Used existing IDs (from offers): 38
-- Created new IDs: 121
-- Skipped duplicates: 16
-- Skipped no name: 1
--
-- PROBLEMS IDENTIFIED:
-- Invalid org numbers: 0
-- Missing contact info: 106
-- Internal customers: 0
-- Inactive customers: 0

INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('df9f59fb-af94-4825-b7ba-4c0808dd41fc', 'A Bygg', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('5b3dbcde-d66c-44fa-addc-76861cac9a05', 'Alvimveien 61 AS', '918129707', '', '+4791344353', 'c/o Toftenes Eiendom ASBruksveien 33', '1390', 'VOLLEN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'ASKER', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('c0701700-1315-4e57-b2eb-032d2449c035', 'Apilar Logistics AS', '990044112', '', '+4767583080', 'Johan Follestads vei 7', '3474', 'ÅROS', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'ASKER', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d9337320-529d-4fa9-934a-646c0516e21f', 'Areal Bygg AS', '985731926', '', '+4797109535', 'Stamveien 7', '1481', 'HAGAN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'NITTEDAL', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('2bcee1e6-3003-4131-89ff-5e0366fd057e', 'Arealbygg', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('157b35d0-5d9e-44a4-9e3f-7de033202cc9', 'As Betongbygg', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d6c6a19f-9e3e-465f-8ab0-754ca69cb9c9', 'Asko Bygg Vestby AS', '884133572', '', '+472425', 'Postboks 164', '1541', 'VESTBY', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'VESTBY', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('de0481c4-716c-4130-b119-11e4bbea5883', 'BILLINGSTADLIA BOLIGSAMEIE II', '989734601', 'pel@mam.no', '+47 95777974', 'c/o Enqvist Boligforvaltning AS, Konghellegata 3', '0569', 'OSLO', 'Norway', 'BILLINGSTADLIA BOLIGSAMEIE II', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('692c8315-fd31-4ba5-92c3-9f968485cb7c', 'BOLIGSAMEIET RÅDHUSPLASS', '990503818', '', '', 'v/Regnskapssentralen AS, Kongens gate 3', '1530', 'MOSS', 'Norway', 'BOLIGSAMEIET RÅDHUSPLASS', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'MOSS', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('6696f84c-0050-4bff-a24b-93c705ca7058', 'BRASETVIK BYGG AS', '991465693', '', '', 'Gartnerveien 20', '3478', 'NÆRSNES', 'Norway', 'BRASETVIK BYGG AS', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'ASKER', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('2e0e548d-7ec9-4165-a3b6-264abcb5cb2b', 'Backegården DA', '983868088', '', '+4724028000', 'c/o Malling & ForvaltningPostboks 1883 Vika', '0124', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('4a764310-a5be-41ac-ac39-a05472c77095', 'Betongbygg AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b692a516-25a3-4555-a050-af5d063eb3f4', 'Betonmast Romerike', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('c00b8d38-c71a-4287-8182-007514a1e769', 'Betonmast Romerike As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b3abdce7-80f8-4851-ae04-d9de3477fcce', 'Betonmast Trøndelag', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b6039121-40fc-46f2-86a4-54cabe7caf43', 'Betonmast Trøndelag As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('abccbf58-b56f-4c0d-b9ea-3eeadb8e18df', 'Bjerke Panorama Sameie', '998410177', '', '+4722983800', 'Postboks 8944 Youngstorget', '0028', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('732c6ca2-6d90-4cb2-946a-62700217822e', 'Bomekan', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('8f5b4c19-3531-49c7-8eb2-cd95bacc22a2', 'Bomekan As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('8de71774-7a5f-4160-b91e-e5173f2b40ff', 'Brick', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('24bb4656-b775-4d4b-9e5b-1e0cf6054965', 'Brick AS', '983100767', '', '', 'Grålumveien 125', '1712', 'GRÅLUM', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'SARPSBORG', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('0ddffc86-4b83-44c0-8b5c-30aa7d62f19d', 'Byggekompaniet Østfold', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('f0726280-6b66-4552-a555-60dd7580112d', 'Byggkompaniet', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('fcc6104b-a9e2-4182-a54b-c5fd0ce49457', 'Byggkompaniet Østfold AS', '970902643', '', '+4769353388', 'Rosenlund 55A', '1617', 'FREDRIKSTAD', 'Norway', 'Jon Bjørgul', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('7c9bc1fc-b8e6-4724-a906-15c3af7cbf87', 'Christian Gran', '', '', '', '', '', '', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('70804119-c863-4310-bed8-0b1b38e13f26', 'Dan blikk AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b1e07c6a-ae59-49dc-9836-8e82a7db9042', 'Dybvig AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('64308ded-dc31-4910-a97c-5823c381764f', 'EIERSEKSJONSSAMEIET BERGRÅDVEIEN 5', '988003700', '', '', 'c/o Obos Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'EIERSEKSJONSSAMEIET BERGRÅDVEIEN 5', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b8fe5e4b-89e8-4202-b36d-b291d1103291', 'ENTER SOLUTION AS', '932757303', '', '', 'Rønningveien 14', '1664', 'ROLVSØY', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('877deaf5-447f-4730-a782-f7f238c4ca1b', 'Eli Østberg', '', '', '+4795997576', 'Grinda 9D', '0861', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('f1dab971-cdef-4ecc-8c28-0635f8bfff42', 'Enter Solutions AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('9d3255de-1721-483e-91df-60510bc83d32', 'FJELLHAUG EIENDOM', '930625132', '', '', 'Sinsenveien 25', '0572', 'OSLO', 'Norway', 'FJELLHAUG EIENDOM', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('cb90aa10-7627-4f5b-97e2-a2956a197144', 'FLØYSAND TAK AS', '892289522', '', '', 'Industrivegen 63', '5210', 'OS', 'Norway', 'Mathias Meek', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'BJØRNAFJORDEN', 'VESTLAND', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('99292045-d41f-4b9f-b943-e54c283ffa11', 'FREDENSBORG SAMEIE 1', '987283718', 'fredensborg1@styrerommet.no', '+4799164418', 'v/OBOS Eiendomsforvaltning AS', '0179', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('a6606813-3a96-49a5-8407-623a0ad33ffa', 'Fossum Terrasse Boligsameie', '984953267', '', '+4790088515', 'c/o OBOS Eiendomsforvaltning ASPostboks 6666 St Olavs plass', '0129', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b62d9b80-f3f9-48a3-abff-87e0703c4e53', 'Furuno AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('072a5b29-9e17-4c72-9643-136afc395ebc', 'Furuno Norge AS', '927200724', '', '+4770102950', 'Postboks 1511', '6025', 'ÅLESUND', 'Norway', 'Finn Helge Stene', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'ÅLESUND', 'MØRE OG ROMSDAL', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('499ad391-98a2-41d4-a0e6-7dc5a18c66cb', 'Fusen', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('a27e4bf6-59c2-4a18-9fd3-f031fcdc738f', 'Fusen As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('cd5ab57e-945b-4f2c-8263-a766a7ae1a66', 'GREV WEDELS PLASS 9 AS', '993511080', '', '', 'Professor Kohts vei 9', '1366', 'LYSAKER', 'Norway', 'GREV WEDELS PLASS 9 AS', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'BÆRUM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('0b8c6db4-7c53-4848-bc5e-21b4c3667e09', 'Ga Meknett AS', '970888160', '', '+4722646550', 'Masteveien 6', '1481', 'HAGAN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'NITTEDAL', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b296011f-a6e2-4977-a274-944e7a1dba77', 'Geir Nielsen (Holmskau)', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('c0f9735e-6b66-4022-80bb-a78a6868e9ac', 'Goenveien 2 Rygge Boligsameie', '823276192', '', '+4795266200', 'Varnaveien 34', '1523', 'MOSS', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'MOSS', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('de5de100-f667-4613-8c29-bd87b3243009', 'Gressvik Properties', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('8481eca2-3af3-4780-9241-4bd38fbb3be6', 'Gressvik Properties As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('60013ebb-ce78-4fca-8365-b1457344755b', 'Gresvik If', '977195500', 'kontoret@gresvikif.no', '', 'Granliveien 23', '1621', 'GRESSVIK', 'Norway', 'Terje Johansen', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('e769c56b-bd61-40c5-83db-811cc1081dd0', 'Grinda 9 Revidert', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('247ffc15-aada-4b7c-8c5a-feef271ae334', 'HAUGER PARK BOLIGSAMEIE', '990474885', 'thorleifka@gmail.com', '', 'Kinoveien 3 A', '1337', 'SANDVIKA', 'Norway', 'HAUGER PARK BOLIGSAMEIE', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'BÆRUM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d9241b6c-1864-42bb-92bf-8729ad4ed13f', 'HEIMANSÅSEN BORETTSLAG', '997003306', 'tor.inge.skoglund@ibis.no', '+4795178465', 'OBOS', '0179', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('4b9492f4-cad3-47c6-87e0-fb57a02fc29e', 'HELLERUDPARKEN BOLIGSAMEIE', '988552216', 'hellerudparken@styrerommet.no', '+47 93204447', 'v/OBOS Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'HELLERUDPARKEN BOLIGSAMEIE', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('eefd93c2-f90f-4f30-9e1f-690649f490dc', 'Hallgruppen AS', '915846432', 'post@hallgruppen.no', '+4721561465', 'Karoline Eggens vei 3', '2016', 'FROGNER', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LILLESTRØM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('50710bb2-d9ee-4629-8b40-29b520197752', 'Hallmaker', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('1c4a030a-77f0-464d-b9f8-e875123b4660', 'Hallmaker AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('e8219f8f-67b2-41a2-9e47-665dd4ad4043', 'Hallmaker AS/ Straye Stålbygg AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b1f799aa-0880-47f8-a4b3-502447778d21', 'Hans Magnus Lie', '', '', '+4797697880', 'Goenveien 4', '1580', 'RYGGE', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'MOSS', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('67e59e8e-f041-40c9-8d7a-925318256a58', 'Hauk Aleksander Olaussen', '', 'Hauk@straye.no', '+47 95000207', 'Langstien 17A', '1715', 'YVEN', 'Norway', 'Hauk Aleksander Olaussen', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'SARPSBORG', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d9602b69-8527-41a5-916e-bc0ac9004b7c', 'Hent AS', '990749655', '', '', 'Vestre Rosten 69', '7072', 'HEIMDAL', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'TRONDHEIM', 'TRØNDELAG', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('13ce7f4e-8f6e-4db8-993d-cca50f6490b3', 'Hersleth Entreprenør', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('6844deb3-0d8f-422b-9d59-542c5054deff', 'Hersleth Entreprenør AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('fa36ad33-1b00-49bd-88df-088aaa83582f', 'Holmskau Prosjekt AS', '991273166', '', '', 'Postboks 206', '1662', 'ROLVSØY', 'Norway', 'Geir Nielsen', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('56dc9583-8ea3-49ec-92ee-f344e5ec9ed1', 'Hyllebærstien Borettslag', '988549584', '', '+4791123911', 'Postboks 313', '1401', 'SKI', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'NORDRE FOLLO', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('4cabbf7a-e0e7-41fa-8114-24bc76c74967', 'Høstbakken 11 AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('ec13ae1c-23a0-4d5e-993a-7d71215ed492', 'Høstbakken Eiendom As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('542f17be-5d1e-4308-8208-d746d643ef26', 'ILDJERNÅSEN SAMEIE', '924004207', 'ildjernasen@styrerommet.no', '', 'v/OBOS Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'ILDJERNÅSEN SAMEIE', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('8a54bfe0-008b-418e-8eee-4185ae0e9c62', 'Ind. Veien 27 D AS', '997615441', '', '+4740407393', 'Industriveien 19', '1481', 'HAGAN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'NITTEDAL', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('489d3ba0-c2e8-49f2-b878-b23adc6c3ee7', 'Jan Bremer Øvrebø', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('0950e6c1-586e-4deb-8fd6-2e63ca8cc5aa', 'Jan Fredrik Smith', '', 'jan.smith@straye.no', '+4791160120', 'Buskogen 72', '1675', 'KRÅKERØY', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b96b28f7-72bc-4be4-a661-2e84af899507', 'Jan Olav Martinsen', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('ae756bbb-26c7-42bb-8625-cf67f5c5666c', 'Jan Svendsen', '', '', '+4790041004', '', '2016', 'FROGNER', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LILLESTRØM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('a803f498-e07b-4c54-8764-6dc6410a7806', 'Jan-Erik Tørmoen', '', 'janerik@tormoen.no', '+4790840847', 'Skjettenveien 114A', '2013', 'SKJETTEN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LILLESTRØM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('0abd2e69-df6b-4589-b6d1-9d34902ffb3e', 'Janne Ekeberg', '', 'janne.ekeberg@gmail.com', '+4790950313', 'Hurrødåsen 1', '1621', 'GRESSVIK', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d0bb4a39-40d9-44f7-9b7d-7fe448885d9d', 'Jensen Bygg & Eiendom AS', '', 'tmjensen@jenseneiendom.no', '+4792048948', '', '', '', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('32b751fb-bb21-448f-9a36-0ea2d6ad5490', 'Jesper Vogt Lorentzen', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('1890279a-7544-40d3-88af-bf7b970b0266', 'Jesper Vogt-lorentzen', '', '', '+4748012336', 'Grinda 9B', '0861', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('1f4706f1-7810-42a7-904b-6921c37338a6', 'Jowa Bygg og Eiendom AS', '916045158', '', '+4794886596', 'Formann Hansens vei 1', '1621', 'GRESSVIK', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('6002f01b-3789-4dac-9d8e-becea97b1e36', 'KM Bygg', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('9f12cf30-cd1d-4eba-9238-650c171dbe9d', 'KOPPERUD MURTNES BYGG AS', '963652313', 'firmapost@km.no', '', 'Grenseveien 11', '1890', 'RAKKESTAD', 'Norway', 'Knut Damm Lyngstad', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'RAKKESTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('2f6086e5-6817-4cae-9742-26c3b16f562b', 'KORNMOENGA 3 SAMEIE', '998799449', 'kornmoenga.sameie@gmail.com', '+47 91328074', 'v/OBOS Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'KORNMOENGA 3 SAMEIE', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('490cea49-3e33-44f2-9b35-7a36f7aa5b26', 'Konsmo fabrikker AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('70dc4fd4-7e5d-4bd9-9871-9ccf724a4f5a', 'Kyrre Johansen', '', '', '+47 99536510', 'Husarveien 35', '1396', 'BILLINGSTAD', 'Norway', 'Kyrre Johansen', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'ASKER', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('0eeb296f-df83-47d1-af01-98caf9ed6d04', 'Lillestrøm Tak og Membran AS', '998586178', '', '', 'Lønsvollveien 80', '1480', 'SLATTUM', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'NITTEDAL', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('acf4c3c1-2225-4723-9d33-71f46f2dbe98', 'Logistic Contractor Norge AS', '915448879', '', '', 'Wergelandsveien 3', '0167', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('9a33ed56-0a59-45c7-8174-d3266863bb7d', 'Loyds Eiendom AS', '994241869', '', '+4790579717', 'Bredmyra 3', '1739', 'BORGENHAUGEN', 'Norway', 'Lasse Hansen', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'SARPSBORG', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('bba63293-8351-4516-b016-753dccb6b8fd', 'MA Totalbygg as', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('4947c3ca-3415-4629-8d3d-3eb07de68ed9', 'MITTEGETLOKALE PORSGRUNN AS', '921810563', '', '', 'Ole Deviks vei 4', '0666', 'OSLO', 'Norway', 'MITTEGETLOKALE PORSGRUNN AS', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('9bf0f685-8c98-4b47-9bdc-f99cfeeae293', 'Matotalbygg AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('3a950106-d4cd-45b1-be1b-417fac45c419', 'Mk Eiendom AS', '996668355', '', '+4790988557', 'Bergsbygdavegen 188', '3949', 'PORSGRUNN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'PORSGRUNN', 'TELEMARK', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('46d814a4-e938-4f3e-a9d0-a557b969f4ad', 'Moelven Byggmodul AS', '941809219', '', '+4762347000', 'Industrivegen 12', '2390', 'MOELV', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'RINGSAKER', 'INNLANDET', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('c8cf9ef3-d692-4671-bc48-8721e12e1dad', 'Morten Andre Kristiansen', '', 'mawaak@gamail.com', '+47 97753733', 'Grinda 9C', '0861', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('32adb13d-4c54-4c7e-9f03-cd44c964bd02', 'NP BYGG AS', '942273711', '', '', 'Tvetenveien 11', '0661', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('ffddf73e-66df-486d-a60c-30f4a2713568', 'NSU NORDIC SERVICE UNION AS', '918305521', '', '+47 94212600', 'Elvesvingen 10', '2003', 'LILLESTRØM', 'Norway', 'NSU NORDIC SERVICE UNION AS', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LILLESTRØM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('4e6ddf96-bca9-4184-9c7e-3836151d8af9', 'Newsec AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('5495cc85-b7a7-4378-a19a-730d9ebca4a2', 'Norbygg AS', '923728902', '', '', 'Bjørnstadmyra 12', '1712', 'GRÅLUM', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'SARPSBORG', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('0649ee1f-0820-41d2-be80-d7e8c13ef134', 'Nordbygg AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('9ba05276-5a2c-48a7-a7e9-c198bccd9059', 'OSLOBYGG KF', '924599545', '', '', 'Grenseveien 82', '0663', 'OSLO', 'Norway', 'OSLOBYGG KF', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('da293e67-4228-428f-9990-d344d57db5c5', 'Ocab AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('787eea47-5fa9-476a-b002-34a2063dbe13', 'PARELIUSVEIEN 2 SAMEIE', '996734781', 'barbros1@getmail.com', '+47 92454400', 'c/o Obos Eiendomsforvaltning as, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'PARELIUSVEIEN 2 SAMEIE', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('11f35db9-ba42-4e5c-ad21-b8ee268bdc04', 'PEAB', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('422284a7-bab9-4f7b-a221-bcdcc9d83add', 'PEAB AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('f24bb15e-1d3b-462d-bf43-aae329e008ee', 'Park & Anlegg', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('49b09bc7-f3f0-4ce1-8ed7-5d284ee7fd33', 'Parketteksperten', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('dd52c2be-fb94-4267-9abf-0f890e733845', 'Parketteksperten As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('4f4d6062-7e83-4cf2-adb1-f0c6855203c7', 'Peab Bygg As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('4e170e58-5c02-4581-8fca-8b7258bd9eea', 'Peab/ Straye Stålbygg', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d2c9c4ae-3df1-4eed-99a8-4c9ca1124b92', 'Per Bremer Øvrebø', '', 'per@heireklame.no', '+47 90037320', '', '1177', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('68888ff1-3b97-4db3-8262-9268d66dab66', 'Per Thormod Skogstad', '', '', '+4790691185', 'Hvalsodden 29', '1394', 'NESBRU', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'ASKER', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('051899a5-2a0b-4139-9b01-5bc716d22574', 'RANDEM & HÜBERT AS', '989653245', '', '', 'Stanseveien 11', '0975', 'OSLO', 'Norway', 'RANDEM & HÜBERT AS', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('52d83e4d-2147-4fb3-ba3b-a0d496590e1f', 'Rg Fjellsikring AS', '924823755', '', '', 'Hølen Verft 15', '', '', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('52d729e1-d174-4879-93a9-ffc8f693a93c', 'Rigmor Frøystad', '', 'rigmfr@online.no', '+47 92252905', 'Løkenåsringen 20', '1473', 'LØRENSKOG', 'Norway', 'Rigmor Frøystad', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LØRENSKOG', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('1ee74ce5-7a00-4aa3-88c0-93349f0b669e', 'Rygge Senior Bo', '989480146', '', '+4797697880', 'c/o Hans Magnus LieGoenveien 4', '1580', 'RYGGE', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'MOSS', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('9625ee7d-f3bb-438f-ae56-1b50ec701ce8', 'Rygge Seniortun Boligsameie', '896957732', '', '+4795266200', 'Varnaveien 34', '1523', 'MOSS', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'MOSS', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('cc25d61e-fe62-4e30-a1d0-cb22618949ea', 'Rørlegger Sentralen AS', '998942683', '', '', 'Lundliveien 11A', '0584', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('da445b6b-7d46-4e36-bdaf-9982e850caa8', 'SAMEIET BEKKESTUA SYD 2', '917708444', 'terjehauff@gmail.com', '+47 90722955', 'c/o Enqvist Boligforvaltning AS, Konghellegata 3', '0569', 'OSLO', 'Norway', 'SAMEIET BEKKESTUA SYD 2', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('066834c1-0bf5-4aec-a021-fc01d7c79d63', 'SAMEIET KJØRBOKOLLEN 19-29', '990179999', '', '', 'v/ Obos Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'SAMEIET KJØRBOKOLLEN 19-29', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('0eb2b8b5-8930-4d90-bca8-6d8827a26647', 'SAMEIET SKOVVEIEN 35', '971280816', 'runeedvin@outlook.com', '+47 91559718', 'v/Obos Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'SAMEIET SKOVVEIEN 35', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('2876f5e9-9d25-4c40-bdce-dfddf4174a1b', 'SAMEIET SØNDRE SKRENTEN 3', '975466795', 'sondreskrenten3@styrerommet.no', '+47 95764730', 'v/ OBOS Eiendomsforvaltning AS, Haugenveien 13B', '1423', 'SKI', 'Norway', 'SAMEIET SØNDRE SKRENTEN 3', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'NORDRE FOLLO', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('35b1d2a1-87f4-4c0b-8906-6cc5e0b78618', 'SAMEIET TIDEMANDS GATE 28', '982706270', '', '', 'c/o ECIT Norian AS, Rosenkrantz'' gate 16', '0160', 'OSLO', 'Norway', 'SAMEIET TIDEMANDS GATE 28', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('9970e837-4831-44b8-ae7a-8905aa833c0b', 'SAMEIET ØSTERÅSBOLIGER I', '971258713', 'pal@fritzon.as', '+4791633130', 'Nedre Storgate 15', '3015', 'DRAMMEN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'DRAMMEN', 'BUSKERUD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('2d5f10d1-4a18-4c7a-a9aa-5db5b9c629a7', 'SOLCELLESPESIALISTEN AS', '930520837', '', '', 'Dikeveien 52', '1661', 'ROLVSØY', 'Norway', 'SOLCELLESPESIALISTEN AS', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b550b16d-da8b-4c91-a90a-4ae042419680', 'STRØMMEN TERRASSE SAMEIE STRØMSVEIEN 93 95 97', '986038167', 'ai@washify.no', '+4790884000', 'Strømsveien 97B', '2010', 'STRØMMEN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LILLESTRØM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('c5214d31-aa5b-4fdd-9847-80a0bd4d620b', 'Sameie Hoffsveien 88', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('0d9cc34d-8a47-458e-aed4-f034c7b4ca75', 'Sameiet Hafrsfjordgate 3', '971271418', '', '+4795079533', 'c/o Nor Forvaltning ASAlnaparkveien 11', '1081', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('fa874f0d-3239-4e2d-96d2-c14a06fdeedd', 'Sameiet Hoffsveien 88/90', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('8af30d75-2cec-4788-8fa5-43b8e42cfef6', 'Sameiet Kornmoenga', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('31d99e7f-3a86-44ea-832c-2bfa61d8fc34', 'Seby AS/ Veidekke AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('54f7730f-7c6f-48da-bbc3-e10100c93e75', 'Seltor AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('c2bd8849-5f41-4e6c-8779-1c9838c312df', 'St. Marie Gate 95 AS', '930117021', 'daniel@straye.no', '', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('89ac134d-9e2d-4b7f-8dcf-3b4d853d107c', 'Straye Gruppen AS', '922249733', '', '', 'Postboks 808', '1670', 'KRÅKERØY', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('326c5bd0-9318-4998-8973-16947187f115', 'Straye Hybrid AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('532e5993-12e7-4a5c-9a55-9487c7552b9a', 'Straye Hybridbygg AS', '932538105', '', '', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', 'Christer Svendsen', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('e77100d8-6610-482f-81de-ab574c5bcb95', 'Straye Industri', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('1c5feb6b-cf61-4dff-86e7-be6d93dea528', 'Straye Industri AS', '931004603', '', '', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d618bf4c-a480-422c-b636-94dfce84e905', 'Straye Industribygg AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('20e942a5-22cb-4bd5-98d1-d582a2425d81', 'Straye Montasje AS', '927378957', '', '', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('74108e6c-cf66-4bdf-a1f1-5429476198e7', 'Straye Stålbygg', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('a5963405-df43-4cf1-98bf-47b36e858569', 'Straye Stålbygg AS', '991664459', '', '', 'Postboks 808', '1670', 'KRÅKERØY', 'Norway', 'Fredrik Eilertsen', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('11f7d96d-8ad3-44ec-86aa-0e77e211f678', 'Straye Stålbygg AS / Hallmaker', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d8caf565-861d-4fbc-8202-74eb4a65fb08', 'Straye Stålbygg AS/ Hallmaker', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('f72f542f-dacd-46b2-851c-196dfd8cf000', 'Straye Tak AS', '929418514', 'henrik@straye.no', '+4747685198', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', 'Henrik Karlsen', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'FREDRIKSTAD', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d377136a-1750-455c-bc1d-eddafea2fdda', 'Straye Trebygg AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('27f3311f-8a84-437f-8232-46d9a4bd6003', 'StrayeIndustri AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('7e4346dc-dc7e-46da-a9bd-a8dfbece042a', 'Sunday Power AS', '922629323', 'kjetil@sundaypower.no', '+4793236123', 'Akersgata 32', '0180', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('4737c611-9755-497c-a74b-1fd1b6a0e0f9', 'T.r Eiendom AS', '924632763', 'hosam@betongspesialisten.com', '+4748512996', 'Torvbanen 9', '1640', 'RÅDE', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'RÅDE', 'ØSTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('ab5b907c-573d-47c7-a8f2-5ff40996a332', 'TatalBygg Midt-norge AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('a02f8ac2-c235-48b5-b662-f8b75cf2fdda', 'Thermica AS', '997933273', 'post@thermica.no', '+4794879592', 'Ringeriksveien 20', '3414', 'LIERSTRANDA', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LIER', 'BUSKERUD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b91d3780-f5af-473a-9a71-d836d761d4cf', 'Thermica AS/ Hallmaker', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('853baedc-7401-4611-9bc2-70e815aa793b', 'Totalbetong', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('923d08e0-2f7f-4774-ade3-bcccbbce9a4c', 'Totalbetong Gruppen As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d2d2784f-41fe-428b-a5fc-5dcac100ed1a', 'Totalbygg Midt-Norge As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('01746014-0418-4dd9-bba6-32c88edebaeb', 'Unil AS', '885316522', '', '+4724113555', 'Karenslyst allé 12', '0278', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('8de95d7d-3e44-4958-99ca-a272d2ccb066', 'Valdresgata Borettslag', '946827908', 'styret@valdresgataborettslag.no', '+47 90504804', 'c/o OBOS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'Valdresgata Borettslag', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('43ca9d23-e63c-40cf-bff5-c29bb180878d', 'Veidekke AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('939ef7ee-7853-4083-875f-2cfebe2793c5', 'Veidekke Bygg- Vest', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('fb395917-f94e-4eab-9654-65d615720781', 'Veidekke Entreprenør', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('b3019c17-69b2-4eb3-b450-1adf92b09805', 'Veidekke Entreprenør AS', '984024290', '', '+4721055000', 'Postboks 506 Skøyen', '0214', 'OSLO', 'Norway', 'Sigbjørn Dahl Helland', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('39485272-b21d-44ef-9c60-a9b443f05955', 'Veidekke Logistikkbygg AS', '971203587', '', '+4733291900', 'Faret 20', '3271', 'LARVIK', 'Norway', 'Raymond Stulen Løberg', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LARVIK', 'VESTFOLD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('f302fa95-549b-446a-a2e1-efa1698763ff', 'Veidekke Ålesund', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('a5fa3a64-ab70-4dfc-b00e-967cce0f71e2', 'Vest Entreprenør', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('3c6e0fe2-4f31-4345-ae0b-db2f31765260', 'Vest Entreprenør AS', '924581611', '', '', 'Gamle Forusveien 10A', '4031', 'STAVANGER', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'STAVANGER', 'ROGALAND', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('dc9a5cfd-2f19-4c87-a813-2f34e8da42f7', 'Vestliterrassen Boligsameie', '884066662', '', '', 'v/OBOS Eiendomsforvaltning ASPostboks 6666, St. Olavs plass', '0129', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('1d715685-c18c-49b0-b0d4-8469875c57d2', 'Vestre Bærum Tennis', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('a107c296-7b9c-40b3-a9a3-f72326ca2f2e', 'Vestre Bærum Tennisklubb', '982088631', 'markus@vbtk.no', '+4791003200', 'Paal Bergs vei 125', '1348', 'RYKKINN', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'BÆRUM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('fa25e1cd-cf5c-452a-a14b-12ebd620840c', 'Workman AS', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('5e39e696-267c-4c87-9ae6-c29eb20cd065', 'Workman Norway As', '', '', '', '', '', '', 'Norway', '', 'Imported from offers - needs enrichment', 'active', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('d1ec04ea-d796-4957-a736-918a2b5313f1', 'dpend/ Straye Stålbygg', '', '', '', '', '', '', 'Norway', '', 'Imported from Plan Straye Tak - needs enrichment', 'lead', 'bronze', 'construction', '', NULL, false, '', '', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('adf5c2d1-c8d2-45f6-bda2-29f35c770ea5', 'ØSTLANDSTAK AS', '926962906', '', '', 'Markveien 35', '3060', 'SVELVIK', 'Norway', 'ØSTLANDSTAK AS', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'DRAMMEN', 'BUSKERUD', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('fceb3d12-a037-486b-bd82-f67ebcb0842b', 'ØYERNBLIKK 2 BOLIGSAMEIE', '916744536', 'oyernblikk2@gmail.com', '+47 48127317', 'c/o BORI BBL, Tærudgata 16', '2004', 'LILLESTRØM', 'Norway', 'ØYERNBLIKK 2 BOLIGSAMEIE', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'LILLESTRØM', 'AKERSHUS', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, notes, status, tier, industry, customer_class, credit_limit, is_internal, municipality, county, company_id, created_at, updated_at)
VALUES ('63b04c74-911b-42d9-8200-262d27c6ed90', 'Øyvind Bjørn Kristiansen', '', '', '', 'Grinda 9', '0861', 'OSLO', 'Norway', '', '', 'active', 'bronze', 'construction', 'Standard', NULL, false, 'OSLO', 'OSLO', 'tak', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  org_number = CASE WHEN customers.org_number = '' OR customers.org_number IS NULL THEN EXCLUDED.org_number ELSE customers.org_number END,
  email = CASE WHEN customers.email = '' OR customers.email IS NULL THEN EXCLUDED.email ELSE customers.email END,
  phone = CASE WHEN customers.phone = '' OR customers.phone IS NULL THEN EXCLUDED.phone ELSE customers.phone END,
  address = CASE WHEN customers.address = '' OR customers.address IS NULL THEN EXCLUDED.address ELSE customers.address END,
  postal_code = CASE WHEN customers.postal_code = '' OR customers.postal_code IS NULL THEN EXCLUDED.postal_code ELSE customers.postal_code END,
  city = CASE WHEN customers.city = '' OR customers.city IS NULL THEN EXCLUDED.city ELSE customers.city END,
  contact_person = CASE WHEN customers.contact_person = '' OR customers.contact_person IS NULL THEN EXCLUDED.contact_person ELSE customers.contact_person END,
  municipality = CASE WHEN customers.municipality = '' OR customers.municipality IS NULL THEN EXCLUDED.municipality ELSE customers.municipality END,
  county = CASE WHEN customers.county = '' OR customers.county IS NULL THEN EXCLUDED.county ELSE customers.county END,
  updated_at = NOW();
