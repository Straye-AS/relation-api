-- Users imported from MS365
-- Generated: 2025-12-17 10:00:12.408456
-- Total users: 219

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b18ba874-d3ac-41d9-8a08-79d3cbe167b1', 'b18ba874-d3ac-41d9-8a08-79d3cbe167b1', 'Adam Grygorcewicz', 'adam.grygorcewicz@straye.no', 'Adam', 'Grygorcewicz', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c8fff2dc-dcc9-478a-8071-86e5046d02d5', 'c8fff2dc-dcc9-478a-8071-86e5046d02d5', 'Adam Nahajowski', 'adam.nahajowski@straye.no', 'Adam', 'Nahajowski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('755f1a8d-f040-4d7e-b85e-4e56f5ace660', '755f1a8d-f040-4d7e-b85e-4e56f5ace660', 'Adam Wojdecki', 'adam.wojdecki@straye.no', 'Adam', 'Wojdecki', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a66f1112-a3e8-4b3e-af63-3adc9476654b', 'a66f1112-a3e8-4b3e-af63-3adc9476654b', 'Adrian Kurek', 'adrian.kurek@straye.no', 'Adrian', 'Kurek', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d84efdf7-a33d-453c-a1e8-f328b4ef70e4', 'd84efdf7-a33d-453c-a1e8-f328b4ef70e4', 'Ainars Zile', 'ainars.zile@straye.no', 'Ainars', 'Zile', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('bc92e60c-2ed8-438f-b45a-60811f640d00', 'bc92e60c-2ed8-438f-b45a-60811f640d00', 'Airydas Azanauskas', 'airydas.azanauskas@straye.no', 'Airydas', 'Azanauskas', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a7cb55f3-c6e5-4c91-a9f7-a5f783751894', 'a7cb55f3-c6e5-4c91-a9f7-a5f783751894', 'Aivars Lauss', 'aivars.lauss@straye.no', 'Aivars', 'Lauss', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ff0bfc57-32d2-4ca0-94fe-d399dcfe35d9', 'ff0bfc57-32d2-4ca0-94fe-d399dcfe35d9', 'Ajai Manoji', 'ajaymanoj@straye.no', 'Ajay', 'Manoji', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('dc83afb6-b02d-40e7-b7d8-5d5ecc9a372f', 'dc83afb6-b02d-40e7-b7d8-5d5ecc9a372f', 'Rexlin Ajitha', 'aji@straye.no', 'Rexlin', 'Ajitha', 'Straye India', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a8d94e1e-7712-4123-b6d4-89b2814af99c', 'a8d94e1e-7712-4123-b6d4-89b2814af99c', 'Ajith Babu', 'ajith@straye.no', 'Ajith', 'Babu', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('35ff9299-c9e2-4720-9072-b90aaa801e36', '35ff9299-c9e2-4720-9072-b90aaa801e36', 'Algirdas Azanauskas', 'algirdas.azanauskas@straye.no', 'Algirdas', 'Azanauskas', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('25ead86d-a57c-48bd-98e0-d22499745660', '25ead86d-a57c-48bd-98e0-d22499745660', 'Ali Aldohdar', 'ali@straye.no', 'Ali', 'Aldohdar', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('74d1714b-7c24-493c-8ad3-f1ad7d7345c8', '74d1714b-7c24-493c-8ad3-f1ad7d7345c8', 'Andreas Spetaas', 'andreas.s@straye.no', 'Andreas', 'Spetaas', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b4161477-6d90-40d6-8fe4-ae508f725aaf', 'b4161477-6d90-40d6-8fe4-ae508f725aaf', 'Andreas Bure', 'andreas@straye.no', 'Andreas', 'Bure', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('612d6017-c0fa-4038-a7cd-499f21264da8', '612d6017-c0fa-4038-a7cd-499f21264da8', 'Andrzej Kapowicz', 'andrzej.kapowicz@straye.no', 'Andrzej', 'Kapowicz', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('6a5cde36-483f-497a-9d11-207c199d429c', '6a5cde36-483f-497a-9d11-207c199d429c', 'Anish Babu', 'anishbabu@straye.no', 'Anish', 'Babu', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('1bc2781b-72c8-45c2-b003-b4e067ea6da4', '1bc2781b-72c8-45c2-b003-b4e067ea6da4', 'Anniken Lønnes (kontor)', 'anniken.l@straye.no', 'Anniken', 'Lønnes', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e07dbe8d-3a55-4b60-811b-b2963568f907', 'e07dbe8d-3a55-4b60-811b-b2963568f907', 'Anto Jerin', 'antojerin@straye.no', 'Anto', 'Jerin', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('bdc7812d-1611-44f8-b18c-4cf6f4bef7f5', 'bdc7812d-1611-44f8-b18c-4cf6f4bef7f5', 'Arne  Kollbye', 'arne@straye.no', 'Arne', 'Kollbye', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('1cd1eeae-2cbd-40de-9899-d68a2fe9ea65', '1cd1eeae-2cbd-40de-9899-d68a2fe9ea65', 'Artur Brzuzy', 'artur.brzuzy@straye.no', 'Artur', 'Brzuzy', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('8411e981-e120-4d2e-a09a-70b8e0c52ef2', '8411e981-e120-4d2e-a09a-70b8e0c52ef2', 'Arturas Grauzas', 'arturas@straye.no', 'Arturas', 'Grauzas', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('af8f45e2-572b-4d2b-b060-f6c0443b0618', 'af8f45e2-572b-4d2b-b060-f6c0443b0618', 'arturasgrauzas@gmail.com', 'arturasgrauzas@gmail.com', '', '', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a6d67a1a-43bb-4ff5-801f-34a22febb3e1', 'a6d67a1a-43bb-4ff5-801f-34a22febb3e1', 'Arturs Kapitanovs', 'arturs.kapitanovs@straye.no', 'Arturs', 'Kapitanovs', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('92d38106-5295-4229-804c-b11db8cc6a3a', '92d38106-5295-4229-804c-b11db8cc6a3a', 'Arun', 'arun@straye.no', 'Arun', '', 'Straye India', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('f8e90a46-ff58-4793-8ecb-e8a5f00e501e', 'f8e90a46-ff58-4793-8ecb-e8a5f00e501e', 'Asko', 'asko@straye.no', 'Asko', '', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e3b7565f-066c-4306-94d7-f10524a01cc6', 'e3b7565f-066c-4306-94d7-f10524a01cc6', 'Aslin', 'aslin@straye.no', 'Aslin', '', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e7e5836c-c385-44b9-9e08-669111f739d1', 'e7e5836c-c385-44b9-9e08-669111f739d1', 'Aswin', 'aswin@straye.no', 'Aswin', '', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('15befd24-c39e-494d-be5e-f7e443d686cb', '15befd24-c39e-494d-be5e-f7e443d686cb', 'Atle Håvard Gunnersen', 'atle.g@straye.no', 'Atle Håvard', 'Gunnersen', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b656634b-6d5e-4698-91e8-ef32079b02a5', 'b656634b-6d5e-4698-91e8-ef32079b02a5', 'Atle Berg', 'atle@straye.no', 'Atle', 'Berg', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('3bf87ee8-4f61-4f3f-87bd-667b0b645409', '3bf87ee8-4f61-4f3f-87bd-667b0b645409', 'Audrius Stukas', 'audrius@straye.no', 'Audrius', 'Stukas', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d603adbb-977c-421b-a8f3-3a4ad08bddb3', 'd603adbb-977c-421b-a8f3-3a4ad08bddb3', 'Azhakesh', 'azhakesh@straye.no', 'Azhakesh', '', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('91948f30-68a5-4a6f-8fad-26e14ae124e3', '91948f30-68a5-4a6f-8fad-26e14ae124e3', 'Bartlomiej Trawnik', 'bartlomiej.trawnik@straye.no', 'Bartlomiej', 'Trawnik', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c571c5f8-288a-4707-a720-505db0fbefb1', 'c571c5f8-288a-4707-a720-505db0fbefb1', 'Bjørn  Nordermoen', 'bjorn@straye.no', 'Bjørn', 'Nordermoen', 'DEAKTIVERT', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('882b92cc-999d-46d1-b033-8f676d300806', '882b92cc-999d-46d1-b033-8f676d300806', 'Camilla Bech', 'camilla@straye.no', 'Camilla', 'Bech', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('9b7bc95a-98b4-473d-b1e3-4f2daa5c2705', '9b7bc95a-98b4-473d-b1e3-4f2daa5c2705', 'Christer Svendsen', 'christer@straye.no', 'Christer', 'Svendsen', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b2afd6df-69bc-4bd2-a1b9-ff0220b17a10', 'b2afd6df-69bc-4bd2-a1b9-ff0220b17a10', 'Christian Quist Jensen', 'christian.quist@straye.no', 'Christian', 'Quist Jensen', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('63bc5e02-d469-4399-b47d-dc9b19169e1a', '63bc5e02-d469-4399-b47d-dc9b19169e1a', 'Christian Faye Lund', 'christian@straye.no', 'Christian', 'Faye Lund', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e15ea012-6016-4960-b31f-62f0d5252cff', 'e15ea012-6016-4960-b31f-62f0d5252cff', 'Dainius Girdauskas', 'dainius.girdauskas@straye.no', 'Dainius', 'Girdauskas', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b1f3c55a-53f8-4ca9-ad92-2c2686016efa', 'b1f3c55a-53f8-4ca9-ad92-2c2686016efa', 'Damian Bareja', 'damian.bareja@straye.no', 'Damian', 'Bareja', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('13ce535d-279c-4352-89d9-c5ffe6cb4907', '13ce535d-279c-4352-89d9-c5ffe6cb4907', 'Damian Nowaczyk', 'damian.nowaczyk@straye.no', 'Damian', 'Nowaczyk', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('f88c016b-8a45-438f-ab3f-392cef491834', 'f88c016b-8a45-438f-ab3f-392cef491834', 'Daniel Jamczuk', 'daniel.jamczuk@straye.no', 'Daniel', 'Jamczuk', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('1828447c-c1a8-43a2-94a3-b16c77348356', '1828447c-c1a8-43a2-94a3-b16c77348356', 'Daniel Karolczak', 'daniel.karolczak@straye.no', 'Daniel', 'Karolczak', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d98ba752-4c23-4a18-b737-5b12c983697e', 'd98ba752-4c23-4a18-b737-5b12c983697e', 'Daniel Skalle', 'daniel.s@straye.no', 'Daniel', 'Skalle', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('be7a622c-3bd0-4771-b9c2-3124c6d7a630', 'be7a622c-3bd0-4771-b9c2-3124c6d7a630', 'Daniel Faye Lund', 'daniel@straye.no', 'Daniel', 'Faye Lund', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c9eb40e8-f78b-4720-b899-caa51dfb72a5', 'c9eb40e8-f78b-4720-b899-caa51dfb72a5', 'Dariusz Lichosyt', 'dariusz.lichosyt@straye.no', 'Dariusz', 'Lichosyt', 'Straye Hybridbygg AS', false, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('3b8cd0e1-578a-419a-9f6f-faeb9558864d', '3b8cd0e1-578a-419a-9f6f-faeb9558864d', 'Dariusz Nowakowski', 'dariusz.nowakowski@straye.no', 'Dariusz', 'Nowakowski', 'Straye Tak AS', false, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('891377ce-a10e-463d-a40c-10e8babd8c85', '891377ce-a10e-463d-a40c-10e8babd8c85', 'David Terøy', 'david.teroy@straye.no', 'David', 'Terøy', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('89cb5639-8127-4272-a324-2afd7d136584', '89cb5639-8127-4272-a324-2afd7d136584', 'Dawid Zajac', 'dawid.zajac@straye.no', 'Dawid', 'Zajac', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('f7ecccc8-728a-41c7-a704-e6cce06e4b76', 'f7ecccc8-728a-41c7-a704-e6cce06e4b76', 'dawidduda.vip@gmail.com', 'dawidduda.vip@gmail.com', '', '', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('fa9d9f43-f885-4c2d-b976-30a72c9a3369', 'fa9d9f43-f885-4c2d-b976-30a72c9a3369', 'Dennis  Svendgård', 'dennis@straye.no', 'Dennis', 'Svendgård', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('1a74092f-cf63-4203-94f2-b593d39ba45f', '1a74092f-cf63-4203-94f2-b593d39ba45f', 'Digiflow Admin', 'digiflowadmin@straye.no', '', '', 'Systemtilgang', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('3b982281-00a7-4c40-9f14-8dfabe24d311', '3b982281-00a7-4c40-9f14-8dfabe24d311', 'Dominik Borowiec', 'dominik.borowiec@straye.no', 'Dominik', 'Borowiec', 'Straye Stålbygg AS', false, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('f78f0f7e-890b-413c-a3f6-a72a82049b97', 'f78f0f7e-890b-413c-a3f6-a72a82049b97', 'Dominik Chwastek', 'dominik.chwastek@straye.no', 'Dominik', 'Chwastek', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('80a24145-43bf-46ec-b5e7-ff80f8cdd73e', '80a24145-43bf-46ec-b5e7-ff80f8cdd73e', 'Dominika Samojedny', 'dominika@straye.no', 'Dominika', 'Samojedny', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a3882bbe-ad6b-4f07-ad27-e766b3713a80', 'a3882bbe-ad6b-4f07-ad27-e766b3713a80', 'Einars Kavoss', 'einars.kavoss@straye.no', 'Einars', 'Kavoss', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('2aa16c14-1f46-4bd6-92a3-93c9ce966020', '2aa16c14-1f46-4bd6-92a3-93c9ce966020', 'Einstein', 'einstein@straye.no', 'Einstein', '', 'Straye India', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c4439567-b123-42e9-9021-a604c50b9a22', 'c4439567-b123-42e9-9021-a604c50b9a22', 'Egil Jensen', 'ej@straye.no', 'Egil', 'Jensen', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('f1eb883b-7d43-498f-b486-080e624584d3', 'f1eb883b-7d43-498f-b486-080e624584d3', 'Elvis Ekmanis', 'elvis.ekmanis@straye.no', 'Elvis', 'Ekmanis', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('7b7d84c1-d48d-47ce-a972-592d63fd5f43', '7b7d84c1-d48d-47ce-a972-592d63fd5f43', 'Elviss  Racans', 'elviss.racans@straye.no', 'Elviss', 'Racans', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d723f81b-2432-43c8-bcdc-54f86f02dd7e', 'd723f81b-2432-43c8-bcdc-54f86f02dd7e', 'Eirik Ramdahl', 'erik.ramdahl@straye.no', 'Eirik', 'Ramdahl', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b8aa4f60-fefd-4487-b681-3e20f0d9e19a', 'b8aa4f60-fefd-4487-b681-3e20f0d9e19a', 'Espen Grindal', 'espen@straye.no', 'Espen', 'Grindal', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('fb644e87-ab0f-4a31-8df1-1b2050fcbd32', 'fb644e87-ab0f-4a31-8df1-1b2050fcbd32', 'Eva- Jeanette Fossen', 'eva@straye.no', 'Eva- Jeanette', 'Fossen', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('61deb2f2-1517-451a-a02c-4e3c3e02b6c6', '61deb2f2-1517-451a-a02c-4e3c3e02b6c6', 'Fabian Kaminski', 'fabian.kaminski@straye.no', 'Fabian', 'Kaminski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('eaa49133-76f4-43a6-8699-cb0508b188bb', 'eaa49133-76f4-43a6-8699-cb0508b188bb', 'Faktura', 'faktura@straye.no', 'Nina Faktura', '', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('34b705e8-e9ad-42b3-b451-48bff005c521', '34b705e8-e9ad-42b3-b451-48bff005c521', 'Fathima Aswin', 'fathima@straye.no', 'Fathima', 'Aswin', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('3e196271-a2b4-42c3-9a33-357fbe275ea0', '3e196271-a2b4-42c3-9a33-357fbe275ea0', 'Anto Felix', 'felix@straye.no', 'Anto', 'Felix', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d1bc27c8-d932-4137-a338-63046f8816c2', 'd1bc27c8-d932-4137-a338-63046f8816c2', 'Felles Kalender', 'felleskalender@straye.no', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('6eb81d10-92d4-4fc1-be11-5aa317d64372', '6eb81d10-92d4-4fc1-be11-5aa317d64372', 'Fredrik Eilertsen', 'fredrik.eilertsen@straye.no', 'Fredrik', 'Eilertsen', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('7f48ab20-fb6e-41aa-800b-b971330177c7', '7f48ab20-fb6e-41aa-800b-b971330177c7', 'Fredrik Smith', 'fredrik.smith@straye.no', 'Fredrik', 'Smith', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e76b7562-d17c-4028-bfba-6cbafc8f927f', 'e76b7562-d17c-4028-bfba-6cbafc8f927f', 'Frode Eskelund', 'frode@straye.no', 'Frode', 'Eskelund', 'Straye Industri AS', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('8538b03c-ddfe-4ea6-a7d7-b029f22871f3', '8538b03c-ddfe-4ea6-a7d7-b029f22871f3', 'Gediminas Zaliukas', 'gediminas.zaliukas@straye.no', 'Gediminas', 'Zaliukas', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('983dfb77-1384-420f-a0e2-20bc34a31e97', '983dfb77-1384-420f-a0e2-20bc34a31e97', 'Gjermund Lorensten Bakli', 'gjermund@straye.no', 'Gjermund', 'Lorensten Bakli', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ccbe1112-fb14-43e7-9f70-16c79288cd78', 'ccbe1112-fb14-43e7-9f70-16c79288cd78', 'Grzegorz Kwolek', 'grzegorz.kwolek@straye.no', 'Grzegorz', 'Kwolek', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d84793af-c4f4-432f-bb4d-9408f417ccfc', 'd84793af-c4f4-432f-bb4d-9408f417ccfc', 'Grzegorz Legutko', 'grzegorz.legutko@straye.no', 'Grzegorz', 'Legutko', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('912a0ac6-21c6-48f8-a97a-a3af16c7b132', '912a0ac6-21c6-48f8-a97a-a3af16c7b132', 'Grzegorz Łukaszewski', 'grzegorz.lukaszewski@straye.no', 'Grzegorz', 'Łukaszewski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b0f9c611-c814-448f-a5e6-b578c8f5c1fd', 'b0f9c611-c814-448f-a5e6-b578c8f5c1fd', 'Grzegorz Wawrzeniec', 'grzegorz.wawrzeniec@straye.no', 'Grzegorz', 'Wawrzeniec', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('df6045d6-ab75-4330-bc89-c31e8935cdab', 'df6045d6-ab75-4330-bc89-c31e8935cdab', 'Gytis Raugalas', 'gytis.raugalas@straye.no', 'Gytis', 'Raugalas', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('3b700556-cf3c-46fc-8d0e-251feac8ceff', '3b700556-cf3c-46fc-8d0e-251feac8ceff', 'Haroldas Butkus', 'haroldas.butkus@straye.no', 'Haroldas', 'Butkus', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ec62f84b-b33c-44d9-807b-c95132ec0033', 'ec62f84b-b33c-44d9-807b-c95132ec0033', 'Hauk Aleksander Olaussen', 'hauk@straye.no', 'Hauk Aleksander', 'Olaussen', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('db434bdc-2eed-4593-9406-5e1f6f8db533', 'db434bdc-2eed-4593-9406-5e1f6f8db533', 'Hege Widlund', 'hege@straye.no', 'Hege', 'Widlund', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b44b3686-1c7b-450c-a7ae-dc95190491ff', 'b44b3686-1c7b-450c-a7ae-dc95190491ff', 'Heidi Rustad', 'heidi@straye.no', 'Heidi', 'Rustad', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('9e3e4888-2e5b-4055-be08-ade3d83a7939', '9e3e4888-2e5b-4055-be08-ade3d83a7939', 'Henrik Carlsson', 'henrik.carlsson@straye.no', 'Henrik', 'Carlsson', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('aa8e2000-86a7-4fd3-8bba-09ee2de54b54', 'aa8e2000-86a7-4fd3-8bba-09ee2de54b54', 'Henrik Karlsen', 'henrik@straye.no', 'Henrik', 'Karlsen', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('2bdb7305-3bcf-4751-808f-0267bf80550b', '2bdb7305-3bcf-4751-808f-0267bf80550b', 'Irmantas Cepulis', 'irmantas.cepulis@straye.no', 'Irmantas', 'Cepulis', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('86dc802a-a11d-4281-b644-ad38a47ce7c4', '86dc802a-a11d-4281-b644-ad38a47ce7c4', 'Jacek Sztyler', 'jacek.sztyler@straye.no', 'Jacek', 'Sztyler', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('6ec2b6d7-5d80-4bc6-820f-090f0e4bdf97', '6ec2b6d7-5d80-4bc6-820f-090f0e4bdf97', 'Jagathis Rajamani', 'jagathis@straye.no', 'Jagathis', 'Rajamani', 'Straye India', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('9a6ec800-5db4-4340-95a6-ca575e72a4a0', '9a6ec800-5db4-4340-95a6-ca575e72a4a0', 'Jan Fredrik Smith', 'jan.smith@straye.no', 'Jan Fredrik', 'Smith', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('65d930e8-7052-44ad-abb5-fcd0d43864b5', '65d930e8-7052-44ad-abb5-fcd0d43864b5', 'Jarek Zientkiewicz', 'jarek.zientkiewicz@straye.no', 'Jarek', 'Zientkiewicz', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('450c5da2-98c0-4b34-b529-2bdd9c1ecca1', '450c5da2-98c0-4b34-b529-2bdd9c1ecca1', 'Jeanette Hjorton', 'jeanette@flexi.no', '', '', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('23c37685-8a57-4088-bb02-9c891a2b0fd7', '23c37685-8a57-4088-bb02-9c891a2b0fd7', 'Jenil Agnel Raj', 'jenil@straye.no', 'Jenil', 'Agnel Raj', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('9d536e7a-6171-4d44-8319-ac9ce46f437f', '9d536e7a-6171-4d44-8319-ac9ce46f437f', 'Jerin James', 'jerin@straye.no', 'Jerin', 'James', 'Straye India', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('10f6e56b-2277-45d2-8333-8a4e8009c572', '10f6e56b-2277-45d2-8333-8a4e8009c572', 'Joachim Richardsen', 'joachim.r@straye.no', 'Joachim', 'Richardsen', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('16ddab24-87df-45cb-a368-824fca35ba83', '16ddab24-87df-45cb-a368-824fca35ba83', 'Johnny Baardsen', 'johnny.baardsen@straye.no', 'Johnny', 'Baardsen', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('14f13f4e-9709-4db0-8043-3cc51db1a7c7', '14f13f4e-9709-4db0-8043-3cc51db1a7c7', 'Jon Erik Andresen', 'jon.a@straye.no', 'Jon Erik', 'Andresen', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('367914c2-457e-4112-b0d4-53a16a8da7cf', '367914c2-457e-4112-b0d4-53a16a8da7cf', 'Jon Kjellin', 'jon@straye.no', 'Jon', 'Kjellin', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('15f5806a-3244-4a9f-8b03-85822a2ce8cc', '15f5806a-3244-4a9f-8b03-85822a2ce8cc', 'jozefjaminski123@gmail.com', 'jozefjaminski123@gmail.com', '', '', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('9a61cff3-c368-40fc-af44-8401dc62453b', '9a61cff3-c368-40fc-af44-8401dc62453b', 'Julie Bøhaugen Evensen', 'julie@straye.no', 'Julie Bøhaugen', 'Evensen', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a6ef846f-4a81-43ef-a974-c70b954f5810', 'a6ef846f-4a81-43ef-a974-c70b954f5810', 'Milton Julin  Michael Raj', 'julin@straye.no', 'Milton Julin', 'Michael Raj', 'Straye India', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('f2611a77-7311-43f3-b02f-3602091c1dc0', 'f2611a77-7311-43f3-b02f-3602091c1dc0', 'Kacper Modrzejewski', 'kacper.modrzejewski@straye.no', 'Kacper', 'Modrzejewski', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b96af2c4-54d1-460f-a30c-379d59e008a8', 'b96af2c4-54d1-460f-a30c-379d59e008a8', 'Kamil Wiecek', 'kamil.wiecek@straye.no', 'Kamil', 'Wiecek', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('f01e5879-fa8f-4bb9-b8c1-34634d417730', 'f01e5879-fa8f-4bb9-b8c1-34634d417730', 'SES_Elektro-Solar', 'kampanje@straye.no', 'SES', '_Elektro-Solar', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('15020b50-e32f-41a9-9776-df54fe17734b', '15020b50-e32f-41a9-9776-df54fe17734b', 'Karoline Kroken', 'karoline@straye.no', 'Karoline', 'Kroken', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d25b7b4d-88ca-46f6-8dd2-8ab082f8016c', 'd25b7b4d-88ca-46f6-8dd2-8ab082f8016c', 'kazik693@poczta.fm', 'kazik693@poczta.fm', '', '', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a70e644e-4e5d-48b9-856e-fede30572359', 'a70e644e-4e5d-48b9-856e-fede30572359', 'Kenny Karlsson', 'kenny.karlsson@straye.no', 'Kenny', 'Karlsson', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('9ccbab5f-2a02-4498-a827-e139e93d5196', '9ccbab5f-2a02-4498-a827-e139e93d5196', 'Kim Ørjan Pettersen', 'kim@straye.no', 'Kim Ørjan', 'Pettersen', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('3f370c3b-8e44-4bbc-a4c7-61939d21e3a1', '3f370c3b-8e44-4bbc-a4c7-61939d21e3a1', 'Knut Jørgen Kollerød', 'knutkoll@straye.no', 'Knut Jørgen', 'Kollerød', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b66183ae-e1fc-47ac-9817-741893dec775', 'b66183ae-e1fc-47ac-9817-741893dec775', 'Kontrad  Kantorski', 'konrad.kantorski@straye.no', 'Kontrad', 'Kantorski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('242d5f66-b388-42ca-b8a1-5e5f716d6b4f', '242d5f66-b388-42ca-b8a1-5e5f716d6b4f', 'Konrad Legutko', 'konrad.legutko@straye.no', 'Konrad', 'Legutko', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('749b7a74-4729-436d-bbb9-4234c245643d', '749b7a74-4729-436d-bbb9-4234c245643d', 'Konrad Wichrowski', 'konrad.wichrowski@straye.no', 'Konrad', 'Wickrowski', 'Straye Tak AS', false, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('49c97465-6098-4f19-83ce-3421761acdc6', '49c97465-6098-4f19-83ce-3421761acdc6', 'kristoffer.k.doring', 'kristoffer.k.doring@hotmail.com', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('4ccb0cf7-6970-4980-82d8-7026dbe9aa95', '4ccb0cf7-6970-4980-82d8-7026dbe9aa95', 'Kristoffer Lund', 'kristoffer.lund@straye.no', 'Kristoffer', 'Lund', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('875a600f-a723-4ac6-8be8-9834150fdd02', '875a600f-a723-4ac6-8be8-9834150fdd02', 'krivs', 'krivs@hotmail.com', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ab82e103-b6a0-41fa-a5b7-fca2bd2a077b', 'ab82e103-b6a0-41fa-a5b7-fca2bd2a077b', 'Krzysztof Brzozowski', 'krzysztof.brzozowski@straye.no', 'Krzysztof', 'Brzozowski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('2f5e1999-30be-4923-b2e7-27b97a98e65f', '2f5e1999-30be-4923-b2e7-27b97a98e65f', 'Krzysztof Wojtas', 'krzysztof.wojtas@straye.no', 'Krzysztof', 'Wojtas', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a45c8e9d-d352-4ced-9674-bd2cf99ceb00', 'a45c8e9d-d352-4ced-9674-bd2cf99ceb00', 'Lånepc', 'laanepc@straye.no', 'Lånepc', '', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('de3be6bf-3e56-4abe-813a-07948fa1a9f0', 'de3be6bf-3e56-4abe-813a-07948fa1a9f0', 'Leif Sten Pettersen', 'leif@straye.no', 'Leif', 'Sten Pettersen', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('0af83981-42d2-40f8-8501-115614343afd', '0af83981-42d2-40f8-8501-115614343afd', 'Leszek Stadnik', 'leszek.stadnik@straye.no', 'Leszek', 'Stadnik', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a3c060d0-9145-40a7-ac9c-ded575cbe33e', 'a3c060d0-9145-40a7-ac9c-ded575cbe33e', 'Lasse Gylvik', 'lg@straye.no', 'Lasse', 'Gylvik', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('2568e714-f905-40f0-8c3b-38c5fefd64bb', '2568e714-f905-40f0-8c3b-38c5fefd64bb', 'Linus Kjell', 'linus@straye.no', 'Linus', 'Kjell', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b6c33491-aea1-42f8-809d-1e9f3fed6f52', 'b6c33491-aea1-42f8-809d-1e9f3fed6f52', 'Lukasz Damian Swieczak', 'lukasz.swieczak@straye.no', 'Lukasz', 'Swieczak', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('00bf235f-0d39-4f43-a115-86931e369b7d', '00bf235f-0d39-4f43-a115-86931e369b7d', 'Maciej Rawski', 'maciej.rawski@straye.no', 'Maciej', 'Rawski', 'Straye Stålbygg AS', false, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e99eecb2-2d1d-45e5-b8d9-8e90f530d751', 'e99eecb2-2d1d-45e5-b8d9-8e90f530d751', 'Maciej Cwyk', 'maciej@straye.no', 'Maciej', 'Cwyk', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('eab51761-32f8-4d95-80b4-2c7dbefb7d3f', 'eab51761-32f8-4d95-80b4-2c7dbefb7d3f', 'maciejtrzesniowski1982@gmail.com', 'maciejtrzesniowski1982@gmail.com', '', '', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('951192b1-a322-44c0-8879-0c1000f73456', '951192b1-a322-44c0-8879-0c1000f73456', 'Maksymilian Obodzinski', 'maksymilian@straye.no', 'Maksymilian', 'Obodzinski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('47306e34-1e54-4e5a-ba9b-d4cfd622630d', '47306e34-1e54-4e5a-ba9b-d4cfd622630d', 'Mantas Raulynaitis', 'mantas.raulynaitis@straye.no', 'Mantas', 'Raulynaitis', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('75aee5a0-5799-49a7-b2b6-dafa60609c60', '75aee5a0-5799-49a7-b2b6-dafa60609c60', 'Marciej Ordzianowski', 'marciej.ordzianowski@straye.no', 'Marciej', 'Ordzianowski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ffbb51c9-eed5-4e37-be04-6b3726ec328a', 'ffbb51c9-eed5-4e37-be04-6b3726ec328a', 'Marcin Kecik', 'marcin.kecik@straye.no', 'Marcin', 'Kecik', 'Straye Industri AS', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('6a8453ff-0980-49c5-937d-e5e66d1b3e66', '6a8453ff-0980-49c5-937d-e5e66d1b3e66', 'Marcin Lorek', 'marcin.lorek@straye.no', 'Marcin', 'Lorek', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('3573d58c-430e-43fd-8575-28f718f4a1e9', '3573d58c-430e-43fd-8575-28f718f4a1e9', 'Marcin Pawlowski', 'marcin.pawlowski@straye.no', 'Marcin', 'Pawlowski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('88eebb65-8a13-4b42-b65f-0bacfa951801', '88eebb65-8a13-4b42-b65f-0bacfa951801', 'Marcin Tomasiewicz', 'marcin.tomasiewicz@straye.no', 'Marcin', 'Tomasiewicz', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ce329154-d42f-4a3f-b528-722fcffb5988', 'ce329154-d42f-4a3f-b528-722fcffb5988', 'Marek Jankowski', 'marek.jankowski@straye.no', 'Marek', 'Jankowski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('645e44be-d066-4a7a-a503-7b2b03a41f1a', '645e44be-d066-4a7a-a503-7b2b03a41f1a', 'Marek Legutko', 'marek.legutko@straye.no', 'Marek', 'Legutko', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('549b8dd4-c72d-4639-bd6c-1872a9c62455', '549b8dd4-c72d-4639-bd6c-1872a9c62455', 'Marek Obodzinski', 'marek@straye.no', 'Marek', 'Obodzinski', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('0803dcf6-67b3-4536-aaa4-69249a1d6c1c', '0803dcf6-67b3-4536-aaa4-69249a1d6c1c', 'Maris Grinsteins', 'maris.grinsteins@straye.no', 'Maris', 'Grinsteins', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('529ea794-48ed-435f-b9af-0eba3b2d248a', '529ea794-48ed-435f-b9af-0eba3b2d248a', 'Marius Barzimiras', 'marius.barzimiras@straye.no', 'Marius', 'Barzimiras', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('fb5055b9-9084-4b75-a46a-ad12035099d0', 'fb5055b9-9084-4b75-a46a-ad12035099d0', 'Marius Filtvedt', 'marius.f@straye.no', 'Marius', 'Filtvedt', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('4bb300a9-9be6-494b-a416-c265f232182b', '4bb300a9-9be6-494b-a416-c265f232182b', 'Markus Johannesson', 'markus.johannesson@straye.no', 'Markus', 'Johannesson', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('5e530530-9cd5-4bc0-b3b0-e2526f1f69a4', '5e530530-9cd5-4bc0-b3b0-e2526f1f69a4', 'Martins Upmalis', 'martins.upmalis@straye.no', 'Martins', 'Upmalis', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('64ba9e6c-759b-4cd3-90f0-e55ab3456533', '64ba9e6c-759b-4cd3-90f0-e55ab3456533', 'Martins Zarins', 'martins.zarins@straye.no', 'Martins', 'Zarins', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('862c2e9e-9f06-492a-aa37-5f98bd2b3cc4', '862c2e9e-9f06-492a-aa37-5f98bd2b3cc4', 'Mateusz Jakubowski', 'mateusz.jakubowski@straye.no', 'Mateusz', 'Jakubowski', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('060a833e-0aa0-4535-87ef-4e3b10d3bb2e', '060a833e-0aa0-4535-87ef-4e3b10d3bb2e', 'Prince Mathew', 'mathew@straye.no', 'Prince', 'Mathew', 'Straye India', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('79df30d9-3037-4e62-920a-d5245a7e655f', '79df30d9-3037-4e62-920a-d5245a7e655f', 'Mershiya', 'mershiya@straye.no', 'Mershiya', '', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('9302ca9f-cca3-4138-a108-8de0cc65a6cc', '9302ca9f-cca3-4138-a108-8de0cc65a6cc', 'Michal Trawnik', 'michal.trawnik@straye.no', 'Michal', 'Trawnik', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('00bfc113-41dd-40af-919e-0242edf59ed7', '00bfc113-41dd-40af-919e-0242edf59ed7', 'Michal Pastor', 'michal@straye.no', 'Michal', 'Pastor', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('5562ded8-dc10-4a7e-8667-0c4306c4026f', '5562ded8-dc10-4a7e-8667-0c4306c4026f', 'Mikael Karlsson', 'mikael.karlsson@straye.no', 'Mikael', 'Karlsson', 'Straye Stålbygg', false, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('0075f355-1400-4aba-a312-1756c4ac8de1', '0075f355-1400-4aba-a312-1756c4ac8de1', 'Mindaugas Butkus', 'mindaugas.butkus@straye.no', 'Mindaugas', 'Butkus', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('7af22b97-d42c-429b-a48b-6e3818dcda2c', '7af22b97-d42c-429b-a48b-6e3818dcda2c', 'miroslaw.sulka88@gmail.com', 'miroslaw.sulka88@gmail.com', '', '', 'Straye Hybridbygg AS', true, ARRAY['user'], 'hybridbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d03c27a2-8ae9-4085-9746-f574ad40f891', 'd03c27a2-8ae9-4085-9746-f574ad40f891', 'mirzetsoft', 'mirzetsoft@gmail.com', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('4fb9f2c1-b63a-4ed4-b2bc-1dc849d4ad78', '4fb9f2c1-b63a-4ed4-b2bc-1dc849d4ad78', 'Møterom Lite1 HQ', 'moterom2@straye.no', '', '', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('21bbfa52-195c-49ce-b300-e0034c20a2a9', '21bbfa52-195c-49ce-b300-e0034c20a2a9', 'Møterom Lite2 HQ', 'moterom3@straye.no', '', '', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('98cbea6f-618d-4363-8782-6a27eecf2cd2', '98cbea6f-618d-4363-8782-6a27eecf2cd2', 'Møterom SON HQ', 'moteromson@straye.no', '', '', 'Gruppen', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d2465b6e-75ca-4428-87a0-7763fa3f8f0c', 'd2465b6e-75ca-4428-87a0-7763fa3f8f0c', 'Mirzet Softic', 'msof@equinor.com', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('4092b9ad-12ff-4126-91a6-a973c9478e8e', '4092b9ad-12ff-4126-91a6-a973c9478e8e', 'Niks Egle', 'niks.egle@straye.no', 'Niks', 'Egle', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('255cb01c-eb35-4c00-8a2f-0066c519d1dd', '255cb01c-eb35-4c00-8a2f-0066c519d1dd', 'Nina Bjørnerud Rustad', 'nina@straye.no', 'Nina', 'Bjørnerud Rustad', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c570b4e9-0f74-4877-913e-88fb8901e085', 'c570b4e9-0f74-4877-913e-88fb8901e085', 'Noreply', 'noreply@straye.no', 'Noreply', '', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('2afe3ef9-b0d4-47a8-8aae-a3a1f54f2b5d', '2afe3ef9-b0d4-47a8-8aae-a3a1f54f2b5d', 'Straye Tak- Driftsleder', 'om@straye.no', '', '', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('fad81ebe-adca-49bc-ae8e-15fdaaeb2332', 'fad81ebe-adca-49bc-ae8e-15fdaaeb2332', 'Oscar Semb Fredricsson', 'oscar@straye.no', 'Oscar', 'Semb Fredricsson', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('dac937fd-2c58-4e26-b43a-4880da15fa80', 'dac937fd-2c58-4e26-b43a-4880da15fa80', 'Pål-Martin Kastum', 'pal-martin@straye.no', 'Pål-Martin', 'Kastum', 'BEHOLDES SÅ SLETTES- Straye Trebygg AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('68c6cd9e-e733-4b6e-9fa5-a6b8ba4b6d13', '68c6cd9e-e733-4b6e-9fa5-a6b8ba4b6d13', 'Patryk Janik', 'patryk@straye.no', 'Patryk', 'Janik', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('4f812cf1-786f-43f8-b948-894a40ca30d7', '4f812cf1-786f-43f8-b948-894a40ca30d7', 'Pawel Krawczyk', 'pawel.krawczyk@straye.no', 'Pawel', 'Krawczyk', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c55c4a08-0637-46a9-9baa-19cd33b46b64', 'c55c4a08-0637-46a9-9baa-19cd33b46b64', 'Pawel Mazur', 'pawel.mazur@straye.no', 'Pawel', 'Mazur', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('2a1116d6-7bda-4577-8377-6c166084314a', '2a1116d6-7bda-4577-8377-6c166084314a', 'Pawel Nyga', 'pawel.nyga@straye.no', 'Pawel', 'Nyga', 'Straye Stålbygg', false, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('8f6253c7-d0a2-4063-bb43-c1d2a11f961c', '8f6253c7-d0a2-4063-bb43-c1d2a11f961c', 'Pawel Ulikowski', 'pawel.ulikowski@straye.no', 'Pawel', 'Ulikowski', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('1ec446b6-335a-41f9-900d-d38ac77ec39a', '1ec446b6-335a-41f9-900d-d38ac77ec39a', 'Petter Faye Lund', 'petter@straye.no', 'Petter', 'Faye Lund', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('70e84d29-8860-43ad-80b7-908c38a50508', '70e84d29-8860-43ad-80b7-908c38a50508', 'Piotr Mochnal', 'piotr.mochnal@straye.no', 'Piotr', 'Mochnal', 'Straye Industri AS', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('8be9cf56-7339-49b7-86ee-6ac487b9b496', '8be9cf56-7339-49b7-86ee-6ac487b9b496', 'Piotr Rajski', 'piotr.rajski@straye.no', 'Piotr', 'Rajski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('eb19962b-3cef-4171-9deb-47204f87d8f4', 'eb19962b-3cef-4171-9deb-47204f87d8f4', 'Piotr Stoltmann', 'piotr.stoltmann@straye.no', 'Piotr', 'Stoltmann', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('5813c4e3-60d9-454b-a100-f6efbf0db558', '5813c4e3-60d9-454b-a100-f6efbf0db558', 'Straye Gruppen AS', 'post@straye.no', 'Post', '', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('0066c67d-79d9-4a5c-ae06-0e4d1d65af82', '0066c67d-79d9-4a5c-ae06-0e4d1d65af82', 'Prosjekt', 'prosjekt@straye.no', 'Prosjekt', '', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a218844d-ea38-4fae-9289-84f0acabfb86', 'a218844d-ea38-4fae-9289-84f0acabfb86', 'Przemyslaw Szlag', 'przemyslaw.szlag@straye.no', 'Przemyslaw', 'Szlag', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c97b3088-79e2-4feb-b235-723edea7b787', 'c97b3088-79e2-4feb-b235-723edea7b787', 'Rafal Roman Jaworski', 'rafal.jaworski@straye.no', 'Rafal Roman', 'Jaworski', 'Straye Industri AS', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('bc7ffba7-9f56-4b8a-8871-086719c619c7', 'bc7ffba7-9f56-4b8a-8871-086719c619c7', 'Rafal Nieznanski', 'rafal.nieznanski@straye.no', 'Rafal', 'Nieznanski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('2c548f51-4ca7-48b7-8008-4e9fd665a007', '2c548f51-4ca7-48b7-8008-4e9fd665a007', 'Ragul .S', 'ragul@straye.no', 'Ragul', '', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('cfbe2f90-29f9-4278-97d9-15d1dfca17fb', 'cfbe2f90-29f9-4278-97d9-15d1dfca17fb', 'Raivis Buss', 'raivis.buss@straye.no', 'Raivis', 'Buss', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('241f770d-93c6-4ecf-9a9b-7ae4fcaa8e97', '241f770d-93c6-4ecf-9a9b-7ae4fcaa8e97', 'Rathiya', 'rathiya@straye.no', 'Rathiya', '', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ae76e1b4-bfd9-44d5-b871-f19f7d88275c', 'ae76e1b4-bfd9-44d5-b871-f19f7d88275c', 'Regimantas Butkus', 'regimantas.butkus@straye.no', 'Regimantas', 'Butkus', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('04f9a8e4-f2ae-4380-aa3f-5f27f0aa3ff5', '04f9a8e4-f2ae-4380-aa3f-5f27f0aa3ff5', 'Regnskap', 'regnskap@straye.no', '', '', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('1ba4af2a-5951-43ca-aa1f-2e1f72ed66ae', '1ba4af2a-5951-43ca-aa1f-2e1f72ed66ae', 'Reklamasjon', 'reklamasjon@straye.no', '', '', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e2fe69b7-92d4-409b-84ae-f8ad55f92857', 'e2fe69b7-92d4-409b-84ae-f8ad55f92857', 'Renars Sviska', 'renars.sviska@straye.no', 'Renars', 'Sviska', 'Straye Montasje AS', false, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('592efc94-fdeb-415b-b8a9-e5ba5b06d8b0', '592efc94-fdeb-415b-b8a9-e5ba5b06d8b0', 'Rexlin Subitha', 'rexlin@straye.no', 'Rexlin', 'Subitha', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('7812a55c-b187-4f59-ad73-1294fb895d7d', '7812a55c-b187-4f59-ad73-1294fb895d7d', 'Rihards Sviska', 'rihards.sviska@straye.no', 'Rihards', 'Sviska', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('4db52f6e-24ac-4169-967f-14dbbf7758bd', '4db52f6e-24ac-4169-967f-14dbbf7758bd', 'Rishikesh', 'rishikesh@straye.no', 'Rishikesh', '', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e0f289bd-0171-4b48-b24b-fdee294f6e37', 'e0f289bd-0171-4b48-b24b-fdee294f6e37', 'Robert Russvoll', 'robert.russvoll@straye.no', 'Robert', 'Russvoll', 'Straye Tak AS', true, ARRAY['user'], 'tak')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('5dc95d60-ed41-4d1c-b7d2-34ae9f296a63', '5dc95d60-ed41-4d1c-b7d2-34ae9f296a63', 'Robert Ziemianowicz', 'robert.ziemianowicz@straye.no', 'Robert', 'Ziemianowicz', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('0fd83545-9600-4528-a2a2-4e82a4c2ce42', '0fd83545-9600-4528-a2a2-4e82a4c2ce42', 'Straye Robot', 'robot@straye.no', 'Straye', 'Robot', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('e1581fd4-7a5d-4f78-9ee9-f9df2c7df49a', 'e1581fd4-7a5d-4f78-9ee9-f9df2c7df49a', 'Rolandas Raulynaitis', 'rolandas.raulynaitis@straye.no', 'Rolandas', 'Raulynaitis', 'Straye Montasje AS', true, ARRAY['user'], 'montasje')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c71ed1b5-f026-4ef6-9eba-d4b618b6f2f6', 'c71ed1b5-f026-4ef6-9eba-d4b618b6f2f6', 'Room adminacc', 'roomadmin@straye.no', 'Room', 'adminacc', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('04a28a40-fd9d-472b-b24e-b18913e98b54', '04a28a40-fd9d-472b-b24e-b18913e98b54', 'Ryszard Kurylowicz', 'ryszard.kurylowicz@straye.no', 'Ryszard', 'Kurylowicz', 'Straye Industri AS', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('7566ac58-fa06-4788-8201-6ab16b5cf2bf', '7566ac58-fa06-4788-8201-6ab16b5cf2bf', 'Salmir Berbic', 'salmir.berbic@straye.no', 'Salmir', 'Berbic', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('983787e7-d580-40ed-8979-57c09d8121e8', '983787e7-d580-40ed-8979-57c09d8121e8', 'Santhosh Kumar', 'santhosh@straye.no', 'Santhosh', 'Kumar', 'Straye India', false, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ed737f19-6f7f-4515-ac2a-6dc20474ddf2', 'ed737f19-6f7f-4515-ac2a-6dc20474ddf2', 'Scanner Straye', 'scanner@straye.no', 'Scanner', 'Straye', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('f59da0e9-b529-4066-abd4-3870b89c0a59', 'f59da0e9-b529-4066-abd4-3870b89c0a59', 'Sebastian Fras', 'sebastian.fras@straye.no', 'Sebastian', 'Fras', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('174959b6-77ee-4312-822e-eeb4f7671aff', '174959b6-77ee-4312-822e-eeb4f7671aff', 'sin.grotle', 'sin.grotle@gmail.com', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c5856247-8216-4b54-81a1-d63fa2d83518', 'c5856247-8216-4b54-81a1-d63fa2d83518', 'Siv Hege Jaren', 'siv@straye.no', 'Siv Hege', 'Jaren', 'Straye Industri AS', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('68ce6b90-66b4-4896-9779-0c8fe2a01c7c', '68ce6b90-66b4-4896-9779-0c8fe2a01c7c', 'SOsync', 'sosync@straye.no', 'SOsync', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('060f9f06-39f9-45e0-8259-74251aba7f77', '060f9f06-39f9-45e0-8259-74251aba7f77', 'Stefan Lindblad', 'stefan.lindblad@straye.no', 'Stefan', 'Lindblad', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('470954de-c78a-4c47-b2ec-4696c1b98c7a', '470954de-c78a-4c47-b2ec-4696c1b98c7a', 'Stefan Steisjö', 'stefan.s@straye.no', 'Stefan', 'Steisjö', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('63c826dc-4cd1-4b31-976d-bb7f9291cd7a', '63c826dc-4cd1-4b31-976d-bb7f9291cd7a', 'Stian Røvig Sletner', 'stian@ciservices.no', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b105a078-ba1c-466a-99a6-a474a0f2a1dc', 'b105a078-ba1c-466a-99a6-a474a0f2a1dc', 'Stian  Årvik', 'stian@sonark.no', 'Stian', 'Årvik', 'Son Arkitektkontor AS ', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c23707a0-4838-42aa-a4bf-14cfaa382509', 'c23707a0-4838-42aa-a4bf-14cfaa382509', 'Møterom Stort HQ', 'stortmoterom1@straye.no', '', '', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('c1958ce7-6a01-4b2f-9e29-ae8f265f6726', 'c1958ce7-6a01-4b2f-9e29-ae8f265f6726', 'STRAYE India Pvt Ltd', 'strayeindiapvtltd@straye.no', '', '', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('ae53b317-0644-4606-a9ba-e1e3339cfcf9', 'ae53b317-0644-4606-a9ba-e1e3339cfcf9', 'Sven Eirik Wilhelmsen', 'sven@straye.no', 'Sven Eirik', 'Wilhelmsen', 'Straye Industri AS', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('96215ead-1dc8-453f-8cc6-edc7cbc359d7', '96215ead-1dc8-453f-8cc6-edc7cbc359d7', 'swojtek249@gmail.com', 'swojtek249@gmail.com', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d0156f69-ede7-479e-a514-acd1e60f0d15', 'd0156f69-ede7-479e-a514-acd1e60f0d15', 'szymkid', 'szymkid@wp.pl', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('90b36bea-c0b7-4fd5-8b19-f00899686b02', '90b36bea-c0b7-4fd5-8b19-f00899686b02', 'Test Testesen', 'testslack@straye.no', 'Test', 'Testesen', 'Straye Gruppen AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('538cf627-34fd-4534-a867-26c86277cf96', '538cf627-34fd-4534-a867-26c86277cf96', 'Toab72', 'toab72@gmail.com', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('d26387ef-c183-4188-ac67-6b40d530fed2', 'd26387ef-c183-4188-ac67-6b40d530fed2', 'Tomasz Sobien', 'tomasz.sobien@straye.no', 'Tomasz', 'Sobien', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('eee59d97-c667-49ef-9029-790ab4f1ff49', 'eee59d97-c667-49ef-9029-790ab4f1ff49', 'Tommy Aas', 'tommy.aas@straye.no', 'Tommy', 'Aas', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('fe58ad71-6999-4582-916a-a573bed2a71c', 'fe58ad71-6999-4582-916a-a573bed2a71c', 'Tore Sten Olsen', 'tore.o@straye.no', 'Tore Sten', 'Olsen', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('22adcabe-5a6b-415d-8a32-7409a9087485', '22adcabe-5a6b-415d-8a32-7409a9087485', 'Trond Aas', 'trond@straye.no', 'Trond', 'Aas', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('47476a9d-1f94-4d7c-ad01-866ddb839c72', '47476a9d-1f94-4d7c-ad01-866ddb839c72', 'Straye Verktøy', 'verktoy@straye.no', '', '', 'IKKE AVKLART', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('42edafc8-062a-4b86-8e0d-a8efe15ea055', '42edafc8-062a-4b86-8e0d-a8efe15ea055', 'Vikar', 'vikar@straye.no', 'Vikar', '', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('eacd489c-d374-4739-8560-eabd359ce97a', 'eacd489c-d374-4739-8560-eabd359ce97a', 'Viktor Karlsson', 'viktor.karlsson@straye.no', 'Viktor', 'Karlsson', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('219e6b6f-6248-4c2d-8496-4a17190f1236', '219e6b6f-6248-4c2d-8496-4a17190f1236', 'Vilmantas Petrikas', 'vilmantas.petrikas@straye.no', 'Vilmantas', 'Petrikas', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('a4027941-32d5-48f2-9f2c-dfa139623137', 'a4027941-32d5-48f2-9f2c-dfa139623137', 'Jain Vimal Raj', 'vimal@straye.no', 'Jain Vimal', 'Raj', 'Straye India', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('7cb2bd69-72dc-485f-af79-8e8fdab93f43', '7cb2bd69-72dc-485f-af79-8e8fdab93f43', 'Wojciech Pisarek', 'wojciech.pisarek@straye.no', 'Wojciech', 'Pisarek', 'Straye Industri', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('29dca0ff-2fa4-4609-924b-7e20db760f7d', '29dca0ff-2fa4-4609-924b-7e20db760f7d', 'Wojciech Waz', 'wojciech.waz@straye.no', 'Wojciech', 'Waz', 'Straye Industri AS', true, ARRAY['user'], 'industri')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('b9d3c557-d029-4950-af0e-86785456862e', 'b9d3c557-d029-4950-af0e-86785456862e', 'Zarian A. Kristiansen', 'zarian@straye.no', 'Zarian A.', 'Kristiansen', 'Straye El & Solar AS', true, ARRAY['user'], NULL)
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('82045ada-9ed0-4a70-90a7-a438f8f26914', '82045ada-9ed0-4a70-90a7-a438f8f26914', 'Zbigniew Adamski', 'zbigniew.adamski@straye.no', 'Zbigniew', 'Adamski', 'Straye Stålbygg AS', true, ARRAY['user'], 'stalbygg')
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;

