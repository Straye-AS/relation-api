-- Straye Tak Customer Import (Full Database Dump)
-- Generated: 2025-12-12
-- Total customers: 121
-- Source of truth for customer data restoration

-- =====================================================
-- CUSTOMERS WITH ORG NUMBER (UPSERT)
-- =====================================================

TRUNCATE TABLE customers;

INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, status, tier, notes, is_internal, municipality, county)
VALUES
('38762824-0b7c-4ab7-a292-da5a88d53119', 'A BYGG ENTREPRENØR AS', '989575716', '', '', 'Ulvenveien 82E', '0581', 'OSLO', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('dc126d03-b3b5-4cae-a8e3-59e912f138d7', 'AS Betongbygg', '929257650', '', '', 'Ringtunveien 8', '1712', 'GRÅLUM', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('052083ee-69fb-4e0e-b01c-4b303d32ae92', 'Alvimveien 61 AS', '918129707', '', '+4791344353', 'c/o Toftenes Eiendom AS, Bruksveien 33', '1390', 'VOLLEN', 'Norway', '', 'active', 'bronze', '', false, 'ASKER', 'AKERSHUS'),
('6a1fbfca-1964-449c-9f7b-82c05462bb6c', 'Apilar Logistics AS', '990044112', '', '+4767583080', 'Johan Follestads vei 7', '3474', 'ÅROS', 'Norway', '', 'active', 'bronze', '', false, 'ASKER', 'AKERSHUS'),
('432877c3-e257-4a79-ad69-7ad653d9a7b5', 'Areal Bygg AS', '985731926', '', '+4797109535', 'Stamveien 7', '1481', 'HAGAN', 'Norway', '', 'active', 'bronze', '', false, 'NITTEDAL', 'AKERSHUS'),
('592a1903-54a7-4a85-a59c-9ad26681ca72', 'Asko Bygg Vestby AS', '884133572', '', '+472425', 'Postboks 164', '1541', 'VESTBY', 'Norway', '', 'active', 'bronze', '', false, 'VESTBY', 'AKERSHUS'),
('3e71b0aa-c2dc-46d4-954d-8cf173d55679', 'BETONMAST ROMERIKE AS', '993925314', '', '', 'Storgata 10', '2000', 'LILLESTRØM', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('c7de4beb-c46b-4a3e-9af1-429dc9ad1761', 'BETONMAST TRØNDELAG AS', '859739822', '', '', 'Falkenborgvegen 36B', '7044', 'TRONDHEIM', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('930d7e01-13ca-43ad-8942-c5189359af11', 'BILLINGSTADLIA BOLIGSAMEIE II', '989734601', 'pel@mam.no', '+47 95777974', 'c/o Enqvist Boligforvaltning AS, Konghellegata 3', '0569', 'OSLO', 'Norway', 'BILLINGSTADLIA BOLIGSAMEIE II', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('19775c3d-0f89-4b91-9233-652fe9fd72be', 'BOLIGSAMEIET RÅDHUSPLASS', '990503818', '', '', 'v/Regnskapssentralen AS, Kongens gate 3', '1530', 'MOSS', 'Norway', 'BOLIGSAMEIET RÅDHUSPLASS', 'active', 'bronze', '', false, 'MOSS', 'ØSTFOLD'),
('a8743c2d-a5e7-470a-b4fc-7bb6640baa5b', 'BOMEKAN AS', '985338442', '', '', 'Industriveien 5', '3090', 'HOF', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('a5ce0c9c-e3d5-4e07-a6a3-aff8ac0d40dc', 'BRASETVIK BYGG AS', '991465693', '', '', 'Gartnerveien 20', '3478', 'NÆRSNES', 'Norway', 'BRASETVIK BYGG AS', 'active', 'bronze', '', false, 'ASKER', 'AKERSHUS'),
('593b4627-09c8-49eb-aa05-325f55c5569f', 'Backegården DA', '983868088', '', '+4724028000', 'c/o Malling & Forvaltning, Postboks 1883 Vika', '0124', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('b469448f-6c38-4a70-8a36-4d2082b13194', 'Bjerke Panorama Sameie', '998410177', '', '+4722983800', 'Postboks 8944 Youngstorget', '0028', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('f1e045ef-65df-45d0-9dd3-83137337404a', 'Brick AS', '983100767', '', '', 'Grålumveien 125', '1712', 'GRÅLUM', 'Norway', '', 'active', 'bronze', '', false, 'SARPSBORG', 'ØSTFOLD'),
('2ed1d049-cef5-42ff-b5f7-faa56d9bbdba', 'Byggkompaniet Østfold AS', '970902643', '', '+4769353388', 'Rosenlund 55A', '1617', 'FREDRIKSTAD', 'Norway', 'Jon Bjørgul', 'active', 'bronze', '', false, 'FREDRIKSTAD', 'ØSTFOLD'),
('8621edcf-1548-40c4-a79a-ccf337c0c262', 'DAN BLIKK AS', '968850601', '', '', 'Pottemakerveien 2', '0954', 'OSLO', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('91fca0ac-31db-4b6e-a1ef-ecc6653e1fc0', 'DYBVIG AS', '976700163', '', '', 'Enebakkveien 304', '1188', 'OSLO', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('8f6067bf-5c5c-4645-946a-a48f1855d755', 'EIERSEKSJONSSAMEIET BERGRÅDVEIEN 5', '988003700', '', '', 'c/o Obos Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'EIERSEKSJONSSAMEIET BERGRÅDVEIEN 5', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('11093e33-e63e-4a63-a336-4f05e1dda1a6', 'ENTER SOLUTION AS', '932757303', '', '', 'Rønningveien 14', '1664', 'ROLVSØY', 'Norway', '', 'active', 'bronze', '', false, 'FREDRIKSTAD', 'ØSTFOLD'),
('fca9a3a1-7098-4816-99c4-ed243171e12a', 'FJELLHAUG EIENDOM', '930625132', '', '', 'Sinsenveien 25', '0572', 'OSLO', 'Norway', 'FJELLHAUG EIENDOM', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('f80eb4d7-8ca9-421c-a37a-177af90093f6', 'FLØYSAND TAK AS', '892289522', '', '', 'Industrivegen 63', '5210', 'OS', 'Norway', 'Mathias Meek', 'active', 'bronze', '', false, 'BJØRNAFJORDEN', 'VESTLAND'),
('f434b835-9db3-4eb7-8a65-ba153ee7cf4a', 'FREDENSBORG SAMEIE 1', '987283718', 'fredensborg1@styrerommet.no', '+4799164418', 'v/OBOS Eiendomsforvaltning AS', '0179', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('ad6dd2d3-66fc-41a5-a21f-c41e1a9a64e8', 'Fossum Terrasse Boligsameie', '984953267', '', '+4790088515', 'c/o OBOS Eiendomsforvaltning AS, Postboks 6666 St Olavs plass', '0129', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('19e0dc83-8d34-4ea3-9967-eef03428f2fc', 'Furuno Norge AS', '927200724', '', '+4770102950', 'Postboks 1511', '6025', 'ÅLESUND', 'Norway', 'Finn Helge Stene', 'active', 'bronze', '', false, 'ÅLESUND', 'MØRE OG ROMSDAL'),
('14ae0185-7a9d-43b3-8eb4-436a2f7056c7', 'GRESSVIK PROPERTIES AS', '925208469', '', '', 'Hjalmar Bjørges vei 105', '1604', 'FREDRIKSTAD', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('6867f86f-20be-466b-9d74-fdda7b9286a1', 'GREV WEDELS PLASS 9 AS', '993511080', '', '', 'Professor Kohts vei 9', '1366', 'LYSAKER', 'Norway', 'GREV WEDELS PLASS 9 AS', 'active', 'bronze', '', false, 'BÆRUM', 'AKERSHUS'),
('b81322d0-fc0f-4ceb-be30-0cd937de7051', 'Ga Meknett AS', '970888160', '', '+4722646550', 'Masteveien 6', '1481', 'HAGAN', 'Norway', '', 'active', 'bronze', '', false, 'NITTEDAL', 'AKERSHUS'),
('caf6894a-b589-4a52-b82e-a48dc843a736', 'Goenveien 2 Rygge Boligsameie', '823276192', '', '+4795266200', 'Varnaveien 34', '1523', 'MOSS', 'Norway', '', 'active', 'bronze', '', false, 'MOSS', 'ØSTFOLD'),
('622be78e-df88-415c-b503-b30a2fdf7aa9', 'Gresvik If', '977195500', 'kontoret@gresvikif.no', '', 'Granliveien 23', '1621', 'GRESSVIK', 'Norway', 'Terje Johansen', 'active', 'bronze', '', false, 'FREDRIKSTAD', 'ØSTFOLD')
ON CONFLICT (org_number) DO UPDATE SET
    name = EXCLUDED.name,
    email = COALESCE(NULLIF(EXCLUDED.email, ''), customers.email),
    phone = COALESCE(NULLIF(EXCLUDED.phone, ''), customers.phone),
    address = COALESCE(NULLIF(EXCLUDED.address, ''), customers.address),
    postal_code = COALESCE(NULLIF(EXCLUDED.postal_code, ''), customers.postal_code),
    city = COALESCE(NULLIF(EXCLUDED.city, ''), customers.city),
    country = COALESCE(NULLIF(EXCLUDED.country, ''), customers.country),
    contact_person = COALESCE(NULLIF(EXCLUDED.contact_person, ''), customers.contact_person),
    municipality = COALESCE(NULLIF(EXCLUDED.municipality, ''), customers.municipality),
    county = COALESCE(NULLIF(EXCLUDED.county, ''), customers.county),
    is_internal = EXCLUDED.is_internal;

-- Batch 2
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, status, tier, notes, is_internal, municipality, county)
VALUES
('e9e55461-4c71-473a-8a42-e03a73d6d229', 'HALLMAKER AS', '937920040', '', '', 'Fornebuveien 5', '1366', 'LYSAKER', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('c78f479f-1644-46ad-a817-994ee0255ba6', 'HAUGER PARK BOLIGSAMEIE', '990474885', 'thorleifka@gmail.com', '', 'Kinoveien 3 A', '1337', 'SANDVIKA', 'Norway', 'HAUGER PARK BOLIGSAMEIE', 'active', 'bronze', '', false, 'BÆRUM', 'AKERSHUS'),
('951888fc-feca-4014-916d-bcf35af4c285', 'HEIMANSÅSEN BORETTSLAG', '997003306', 'tor.inge.skoglund@ibis.no', '+4795178465', 'OBOS', '0179', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('49701643-1b09-4f8c-b036-3fd1a9dd5e71', 'HELLERUDPARKEN BOLIGSAMEIE', '988552216', 'hellerudparken@styrerommet.no', '+47 93204447', 'v/OBOS Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'HELLERUDPARKEN BOLIGSAMEIE', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('b98e98dc-e830-48bf-a1f6-bf728741d1e1', 'HERSLETH ENTREPRENØR AS', '964602360', '', '', 'Hobølveien 4', '1550', 'HØLEN', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('d7f04f69-582e-480f-a4e0-5520167c6b76', 'Hallgruppen AS', '915846432', 'post@hallgruppen.no', '+4721561465', 'Karoline Eggens vei 3', '2016', 'FROGNER', 'Norway', '', 'active', 'bronze', '', false, 'LILLESTRØM', 'AKERSHUS'),
('0a5157ff-15a1-459e-a7bf-c30777e55e2d', 'Hent AS', '990749655', '', '', 'Vestre Rosten 69', '7072', 'HEIMDAL', 'Norway', '', 'active', 'bronze', '', false, 'TRONDHEIM', 'TRØNDELAG'),
('a9d26045-affa-4b67-823f-4f3d8aa01d93', 'Holmskau Prosjekt AS', '991273166', '', '', 'Postboks 206', '1662', 'ROLVSØY', 'Norway', 'Geir Nielsen', 'active', 'bronze', '', false, 'FREDRIKSTAD', 'ØSTFOLD'),
('223f42ea-bac4-4539-9e65-6cd88b4fe00d', 'Hyllebærstien Borettslag', '988549584', '', '+4791123911', 'Postboks 313', '1401', 'SKI', 'Norway', '', 'active', 'bronze', '', false, 'NORDRE FOLLO', 'AKERSHUS'),
('d0c42899-f81f-4cb7-82de-87f675c998c8', 'HØSTBAKKEN EIENDOM AS', '991045252', '', '', 'Høstbakken 11', '1793', 'TISTEDAL', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('20b3c74d-ec2d-4ab5-a65c-c0e4170abe0c', 'ILDJERNÅSEN SAMEIE', '924004207', 'ildjernasen@styrerommet.no', '', 'v/OBOS Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'ILDJERNÅSEN SAMEIE', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('8ec85043-fbc0-4381-8482-982d2cc8d036', 'Ind. Veien 27 D AS', '997615441', '', '+4740407393', 'Industriveien 19', '1481', 'HAGAN', 'Norway', '', 'active', 'bronze', '', false, 'NITTEDAL', 'AKERSHUS'),
('2e648a50-1416-4d6b-872c-a8b1d977fb9a', 'JENSEN BYGG & EIENDOM AS', '980779033', '', '', 'Agentgaten 5', '1607', 'FREDRIKSTAD', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('5e56939b-c7ca-4d49-9d8d-066286a4ec25', 'Jowa Bygg og Eiendom AS', '916045158', '', '+4794886596', 'Formann Hansens vei 1', '1621', 'GRESSVIK', 'Norway', '', 'active', 'bronze', '', false, 'FREDRIKSTAD', 'ØSTFOLD'),
('e3a1ac0d-368a-4efd-9c00-778bbe8da0ed', 'KONSMO FABRIKKER AS', '950167866', '', '', 'Breilimoen 15', '4525', 'KONSMO', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('9617b98c-d8a3-4458-a637-3bd9b2e4c8bf', 'KOPPERUD MURTNES BYGG AS', '963652313', 'firmapost@km.no', '', 'Grenseveien 11', '1890', 'RAKKESTAD', 'Norway', 'Knut Damm Lyngstad', 'active', 'bronze', '', false, 'RAKKESTAD', 'ØSTFOLD'),
('78cccdaa-8fb5-40ac-b00a-ef65a18f54c4', 'KORNMOENGA 3 SAMEIE', '998799449', 'kornmoenga.sameie@gmail.com', '+47 91328074', 'v/OBOS Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'KORNMOENGA 3 SAMEIE', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('e5fc085a-7e2b-4f58-abe1-64388e0b3965', 'Lillestrøm Tak og Membran AS', '998586178', '', '', 'Lønsvollveien 80', '1480', 'SLATTUM', 'Norway', '', 'active', 'bronze', '', false, 'NITTEDAL', 'AKERSHUS'),
('b0f19287-c466-4c77-ad30-1b6c5703e970', 'Logistic Contractor Norge AS', '915448879', '', '', 'Wergelandsveien 3', '0167', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('96442513-dd2a-4c5c-af66-1a292a2933d3', 'Loyds Eiendom AS', '994241869', '', '+4790579717', 'Bredmyra 3', '1739', 'BORGENHAUGEN', 'Norway', 'Lasse Hansen', 'active', 'bronze', '', false, 'SARPSBORG', 'ØSTFOLD'),
('0348d815-e8f5-4617-9cf5-142847128c6d', 'MA Totalbygg AS', '999194664', '', '', '', '0272', 'Oslo', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('09d0cc60-9685-4d4d-8962-58b5cc1184c0', 'MITTEGETLOKALE PORSGRUNN AS', '921810563', '', '', 'Ole Deviks vei 4', '0666', 'OSLO', 'Norway', 'MITTEGETLOKALE PORSGRUNN AS', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('32bd85bf-c272-47f9-9520-654232763907', 'Mk Eiendom AS', '996668355', '', '+4790988557', 'Bergsbygdavegen 188', '3949', 'PORSGRUNN', 'Norway', '', 'active', 'bronze', '', false, 'PORSGRUNN', 'TELEMARK'),
('72857098-3b72-4cda-89ee-e49ab5a12372', 'Moelven Byggmodul AS', '941809219', '', '+4762347000', 'Industrivegen 12', '2390', 'MOELV', 'Norway', '', 'active', 'bronze', '', false, 'RINGSAKER', 'INNLANDET'),
('19f7f4e4-39c0-4af3-a30e-5ef432725ee7', 'NEWSEC AS', '986033033', '', '', 'Haakon VIIs gate 2', '0161', 'OSLO', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('a3192144-8e88-4dc9-afdb-6acc4df15ab4', 'NP BYGG AS', '942273711', '', '', 'Tvetenveien 11', '0661', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('4eae41cf-934f-402a-a769-ffe34af41336', 'NSU NORDIC SERVICE UNION AS', '918305521', '', '+47 94212600', 'Elvesvingen 10', '2003', 'LILLESTRØM', 'Norway', 'NSU NORDIC SERVICE UNION AS', 'active', 'bronze', '', false, 'LILLESTRØM', 'AKERSHUS'),
('53daeb79-8cec-4053-bbda-72e2336f8e9d', 'Norbygg AS', '923728902', '', '', 'Bjørnstadmyra 12', '1712', 'GRÅLUM', 'Norway', '', 'active', 'bronze', '', false, 'SARPSBORG', 'ØSTFOLD'),
('78d10ac9-d9c5-4d45-add4-3a7be502508b', 'OCAB AS', '916326211', '', '', 'Jongsåsveien 3', '1338', 'SANDVIKA', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('b9e1fb3d-b98e-461e-8650-bd3e45a63b53', 'OSLOBYGG KF', '924599545', '', '', 'Grenseveien 82', '0663', 'OSLO', 'Norway', 'OSLOBYGG KF', 'active', 'bronze', '', false, 'OSLO', 'OSLO')
ON CONFLICT (org_number) DO UPDATE SET
    name = EXCLUDED.name,
    email = COALESCE(NULLIF(EXCLUDED.email, ''), customers.email),
    phone = COALESCE(NULLIF(EXCLUDED.phone, ''), customers.phone),
    address = COALESCE(NULLIF(EXCLUDED.address, ''), customers.address),
    postal_code = COALESCE(NULLIF(EXCLUDED.postal_code, ''), customers.postal_code),
    city = COALESCE(NULLIF(EXCLUDED.city, ''), customers.city),
    country = COALESCE(NULLIF(EXCLUDED.country, ''), customers.country),
    contact_person = COALESCE(NULLIF(EXCLUDED.contact_person, ''), customers.contact_person),
    municipality = COALESCE(NULLIF(EXCLUDED.municipality, ''), customers.municipality),
    county = COALESCE(NULLIF(EXCLUDED.county, ''), customers.county),
    is_internal = EXCLUDED.is_internal;

-- Batch 3
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, status, tier, notes, is_internal, municipality, county)
VALUES
('a9a739eb-ce6c-4087-a0f6-85aa9af30b12', 'PARELIUSVEIEN 2 SAMEIE', '996734781', 'barbros1@getmail.com', '+47 92454400', 'c/o Obos Eiendomsforvaltning as, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'PARELIUSVEIEN 2 SAMEIE', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('5fe75c17-ba11-4750-81ec-076f38144d3d', 'PARKETTEKSPERTEN AS', '818037562', '', '', 'Strykerveien 22B', '1658', 'TORP', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('1b2a860f-e551-430a-ab31-ddf4c16c82be', 'PEAB BYGG AS', '943672520', '', '', 'Hjalmar Johansens gate 25', '9007', 'TROMSØ', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('1f48be94-23de-4a25-89a4-b63d24035160', 'Park og Anlegg A/S', '947585533', '', '', 'Sandflatvegen 32', '7036', 'TRONDHEIM', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('ad765551-e313-47a8-82e2-5949636cf92d', 'RANDEM & HÜBERT AS', '989653245', '', '', 'Stanseveien 11', '0975', 'OSLO', 'Norway', 'RANDEM & HÜBERT AS', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('3e8af49f-b8d1-442b-914f-7f1bd1dd36f6', 'Rg Fjellsikring AS', '924823755', '', '', 'Hølen Verft 15', '', '', 'Norway', '', 'active', 'bronze', '', false, '', ''),
('461cabf0-5a7f-4061-a042-9396572920dc', 'Rygge Senior Bo', '989480146', '', '+4797697880', 'c/o Hans Magnus Lie, Goenveien 4', '1580', 'RYGGE', 'Norway', '', 'active', 'bronze', '', false, 'MOSS', 'ØSTFOLD'),
('e0c5ee48-433c-40ec-a052-841ffe20dd84', 'Rygge Seniortun Boligsameie', '896957732', '', '+4795266200', 'Varnaveien 34', '1523', 'MOSS', 'Norway', '', 'active', 'bronze', '', false, 'MOSS', 'ØSTFOLD'),
('da328f78-8386-48e8-b4b9-de2371ceb07e', 'Rørlegger Sentralen AS', '998942683', '', '', 'Lundliveien 11A', '0584', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('a221c305-4287-4dc7-a655-1dacffd6ed0c', 'SAMEIET BEKKESTUA SYD 2', '917708444', 'terjehauff@gmail.com', '+47 90722955', 'c/o Enqvist Boligforvaltning AS, Konghellegata 3', '0569', 'OSLO', 'Norway', 'SAMEIET BEKKESTUA SYD 2', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('d7febd27-d0ba-4a8b-ab83-2dffe4dd505f', 'SAMEIET HOFFSVEIEN 88/90', '981498631', '', '', 'Hoffsveien 88/90', '0377', 'OSLO', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('8273a737-ffab-48fe-9bd3-0ad9ea8b489c', 'SAMEIET KJØRBOKOLLEN 19-29', '990179999', '', '', 'v/ Obos Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'SAMEIET KJØRBOKOLLEN 19-29', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('aa47374c-72b2-429c-82b6-a7b7d4ba9bf6', 'SAMEIET SKOVVEIEN 35', '971280816', 'runeedvin@outlook.com', '+47 91559718', 'v/Obos Eiendomsforvaltning AS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'SAMEIET SKOVVEIEN 35', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('7af3abb7-4885-426f-a325-7790cf7385bd', 'SAMEIET SØNDRE SKRENTEN 3', '975466795', 'sondreskrenten3@styrerommet.no', '+47 95764730', 'v/ OBOS Eiendomsforvaltning AS, Haugenveien 13B', '1423', 'SKI', 'Norway', 'SAMEIET SØNDRE SKRENTEN 3', 'active', 'bronze', '', false, 'NORDRE FOLLO', 'AKERSHUS'),
('c7cbec09-1c5b-4642-b27e-a8f240ce8c2e', 'SAMEIET TIDEMANDS GATE 28', '982706270', '', '', 'c/o ECIT Norian AS, Rosenkrantz'' gate 16', '0160', 'OSLO', 'Norway', 'SAMEIET TIDEMANDS GATE 28', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('3e10404f-5bc5-448a-9613-53d0d40ac839', 'SAMEIET ØSTERÅSBOLIGER I', '971258713', 'pal@fritzon.as', '+4791633130', 'Nedre Storgate 15', '3015', 'DRAMMEN', 'Norway', '', 'active', 'bronze', '', false, 'DRAMMEN', 'BUSKERUD'),
('2c0095d2-5678-40e6-b394-443e573cf0f6', 'SELTOR AS', '915617344', '', '', 'Dokkvegen 8', '3920', 'PORSGRUNN', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('dcc6db9f-dc5a-45cf-bea6-5f885c44cd38', 'SOLCELLESPESIALISTEN AS', '930520837', '', '', 'Dikeveien 52', '1661', 'ROLVSØY', 'Norway', 'SOLCELLESPESIALISTEN AS', 'active', 'bronze', '', false, 'FREDRIKSTAD', 'ØSTFOLD'),
('d2e0b843-a8fd-4436-be46-d658648cbc7e', 'SOLENERGI FUSEN AS', '999009123', '', '', 'Østensjøveien 15D', '0661', 'OSLO', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('e1611dbd-40b8-4728-8c84-071f0aee0fee', 'STRAYE TREBYGG AS', '822249752', '', '', 'Kråkerøyveien 2A', '1671', 'KRÅKERØY', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('9ae80957-0cf2-4e1e-afa6-611e88942d4c', 'STRØMMEN TERRASSE SAMEIE STRØMSVEIEN 93 95 97', '986038167', 'ai@washify.no', '+4790884000', 'Strømsveien 97B', '2010', 'STRØMMEN', 'Norway', '', 'active', 'bronze', '', false, 'LILLESTRØM', 'AKERSHUS'),
('054dc550-13bc-4dcf-b4be-c41952a0feb8', 'Sameiet Hafrsfjordgate 3', '971271418', '', '+4795079533', 'c/o Nor Forvaltning AS, Alnaparkveien 11', '1081', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('87abe3cd-dd60-4411-802a-6a20d1c0f4b4', 'St. Marie Gate 95 AS', '930117021', 'daniel@straye.no', '', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', '', 'active', 'bronze', '', false, 'FREDRIKSTAD', 'ØSTFOLD'),
('43ceae70-9288-4a68-b2a2-587c122947ef', 'Sunday Power AS', '922629323', 'kjetil@sundaypower.no', '+4793236123', 'Akersgata 32', '0180', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('6f56b8c7-d2b2-4af4-80f3-667f8c3b1a25', 'T.r Eiendom AS', '924632763', 'hosam@betongspesialisten.com', '+4748512996', 'Torvbanen 9', '1640', 'RÅDE', 'Norway', '', 'active', 'bronze', '', false, 'RÅDE', 'ØSTFOLD'),
('f9b60f04-0f55-42ae-8894-51c448d35a2a', 'TOTALBETONG GRUPPEN AS', '921153023', '', '', 'Martin Vagles veg 7', '4344', 'BRYNE', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('6782ac44-0095-4fc0-8ba1-2519b94e3fcf', 'TOTALBYGG MIDT-NORGE AS', '999239390', '', '', 'Lissbjørknesvegen 105', '7970', 'KOLVEREID', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('be02dd33-e4e2-4be8-8420-3310c4bd75b7', 'Thermica AS', '997933273', 'post@thermica.no', '+4794879592', 'Ringeriksveien 20', '3414', 'LIERSTRANDA', 'Norway', '', 'active', 'bronze', '', false, 'LIER', 'BUSKERUD'),
('09155d00-0e04-49a1-aa08-e57917bee76d', 'Unil AS', '885316522', '', '+4724113555', 'Karenslyst allé 12', '0278', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('c575c9be-991a-4737-9355-75fc9927d385', 'Valdresgata Borettslag', '946827908', 'styret@valdresgataborettslag.no', '+47 90504804', 'c/o OBOS, Hammersborg torg 1', '0179', 'OSLO', 'Norway', 'Valdresgata Borettslag', 'active', 'bronze', '', false, 'OSLO', 'OSLO')
ON CONFLICT (org_number) DO UPDATE SET
    name = EXCLUDED.name,
    email = COALESCE(NULLIF(EXCLUDED.email, ''), customers.email),
    phone = COALESCE(NULLIF(EXCLUDED.phone, ''), customers.phone),
    address = COALESCE(NULLIF(EXCLUDED.address, ''), customers.address),
    postal_code = COALESCE(NULLIF(EXCLUDED.postal_code, ''), customers.postal_code),
    city = COALESCE(NULLIF(EXCLUDED.city, ''), customers.city),
    country = COALESCE(NULLIF(EXCLUDED.country, ''), customers.country),
    contact_person = COALESCE(NULLIF(EXCLUDED.contact_person, ''), customers.contact_person),
    municipality = COALESCE(NULLIF(EXCLUDED.municipality, ''), customers.municipality),
    county = COALESCE(NULLIF(EXCLUDED.county, ''), customers.county),
    is_internal = EXCLUDED.is_internal;

-- Batch 4: Remaining companies + Internal Straye companies
INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, contact_person, status, tier, notes, is_internal, municipality, county)
VALUES
('45799f54-0340-4d3b-8e4b-74b9657e15d0', 'Veidekke Entreprenør AS', '984024290', '', '+4721055000', 'Postboks 506 Skøyen', '0214', 'OSLO', 'Norway', 'Sigbjørn Dahl Helland', 'active', 'bronze', '', false, '', ''),
('0fa0880e-303d-4b25-8577-5ad0c5421690', 'Veidekke Logistikkbygg AS', '971203587', '', '+4733291900', 'Faret 20', '3271', 'LARVIK', 'Norway', 'Raymond Stulen Løberg', 'active', 'bronze', '', false, 'LARVIK', 'VESTFOLD'),
('43678b5f-eeb2-440b-beed-b4b8e095d436', 'Vest Entreprenør AS', '924581611', '', '', 'Gamle Forusveien 10A', '4031', 'STAVANGER', 'Norway', '', 'active', 'bronze', '', false, 'STAVANGER', 'ROGALAND'),
('98b56491-e351-4110-b0d2-4152dfe524cd', 'Vestliterrassen Boligsameie', '884066662', '', '', 'v/OBOS Eiendomsforvaltning AS, Postboks 6666, St. Olavs plass', '0129', 'OSLO', 'Norway', '', 'active', 'bronze', '', false, 'OSLO', 'OSLO'),
('e75ae140-0b21-479c-994a-1793eb430377', 'Vestre Bærum Tennisklubb', '982088631', 'markus@vbtk.no', '+4791003200', 'Paal Bergs vei 125', '1348', 'RYKKINN', 'Norway', '', 'active', 'bronze', '', false, 'BÆRUM', 'AKERSHUS'),
('bfee114e-577c-4972-ae13-318e5043ab5f', 'WORKMAN NORWAY AS', '927371138', '', '', 'Bjerkås Næringspark', '3470', 'SLEMMESTAD', 'Norge', '', 'active', 'bronze', '', false, '', ''),
('1a582c30-eb6e-4470-a7d7-5880a005d98a', 'ØSTLANDSTAK AS', '926962906', '', '', 'Markveien 35', '3060', 'SVELVIK', 'Norway', 'ØSTLANDSTAK AS', 'active', 'bronze', '', false, 'DRAMMEN', 'BUSKERUD'),
('441a2089-aa2d-4f8e-9b9f-0977a573b1c8', 'ØYERNBLIKK 2 BOLIGSAMEIE', '916744536', 'oyernblikk2@gmail.com', '+47 48127317', 'c/o BORI BBL, Tærudgata 16', '2004', 'LILLESTRØM', 'Norway', 'ØYERNBLIKK 2 BOLIGSAMEIE', 'active', 'bronze', '', false, 'LILLESTRØM', 'AKERSHUS'),
-- Internal Straye companies
('835dfa92-97c1-44a2-9dc4-6841eeecee96', 'Straye Gruppen AS', '922249733', '', '', 'Postboks 808', '1670', 'KRÅKERØY', 'Norway', '', 'active', 'bronze', '', true, 'FREDRIKSTAD', 'ØSTFOLD'),
('9ed5ed8d-fafb-4037-a14f-afe9bee8b7db', 'Straye Hybridbygg AS', '932538105', '', '', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', 'Christer Svendsen', 'active', 'bronze', '', true, 'FREDRIKSTAD', 'ØSTFOLD'),
('30bfcd26-82ec-4a07-9e89-e4d7527da01e', 'Straye Industri AS', '931004603', '', '', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', '', 'active', 'bronze', '', true, 'FREDRIKSTAD', 'ØSTFOLD'),
('4501b39e-f22b-4c9d-802e-b90c10b0dbea', 'Straye Montasje AS', '927378957', '', '', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', '', 'active', 'bronze', '', true, 'FREDRIKSTAD', 'ØSTFOLD'),
('e7659d18-bf4f-4071-abe4-13c637edeae3', 'Straye Stålbygg AS', '991664459', '', '', 'Postboks 808', '1670', 'KRÅKERØY', 'Norway', 'Fredrik Eilertsen', 'active', 'bronze', '', true, 'FREDRIKSTAD', 'ØSTFOLD'),
('2f84849e-11a2-42bc-b7c6-917b547b2932', 'Straye Tak AS', '929418514', 'henrik@straye.no', '+4747685198', 'Kråkerøyveien 4', '1671', 'KRÅKERØY', 'Norway', 'Henrik Karlsen', 'active', 'bronze', '', true, 'FREDRIKSTAD', 'ØSTFOLD')
ON CONFLICT (org_number) DO UPDATE SET
    name = EXCLUDED.name,
    email = COALESCE(NULLIF(EXCLUDED.email, ''), customers.email),
    phone = COALESCE(NULLIF(EXCLUDED.phone, ''), customers.phone),
    address = COALESCE(NULLIF(EXCLUDED.address, ''), customers.address),
    postal_code = COALESCE(NULLIF(EXCLUDED.postal_code, ''), customers.postal_code),
    city = COALESCE(NULLIF(EXCLUDED.city, ''), customers.city),
    country = COALESCE(NULLIF(EXCLUDED.country, ''), customers.country),
    contact_person = COALESCE(NULLIF(EXCLUDED.contact_person, ''), customers.contact_person),
    municipality = COALESCE(NULLIF(EXCLUDED.municipality, ''), customers.municipality),
    county = COALESCE(NULLIF(EXCLUDED.county, ''), customers.county),
    is_internal = EXCLUDED.is_internal;

-- =====================================================
-- PRIVATE PERSONS (no org number - INSERT with ON CONFLICT id)
-- =====================================================

INSERT INTO customers (id, name, org_number, email, phone, address, postal_code, city, country, status, tier, notes, is_internal, municipality, county)
VALUES
('17051d9c-cdd8-49dc-82cb-0b678bc6af06', 'Christian Gran', NULL, '', '', '', '', '', 'Norway', 'active', 'bronze', 'Private person', false, '', ''),
('c801ccf6-4b6d-45a5-b662-b93af503324c', 'Eli Østberg', NULL, '', '+4795997576', 'Grinda 9D', '0861', 'OSLO', 'Norway', 'active', 'bronze', 'Private person', false, 'OSLO', 'OSLO'),
('7898be31-ad09-4cc7-847e-400fd8d73e33', 'Hans Magnus Lie', NULL, '', '+4797697880', 'Goenveien 4', '1580', 'RYGGE', 'Norway', 'active', 'bronze', 'Private person', false, 'MOSS', 'ØSTFOLD'),
('993d3b6d-31ae-4f36-a1f0-bab5be353f42', 'Hauk Aleksander Olaussen', NULL, 'Hauk@straye.no', '+47 95000207', 'Langstien 17A', '1715', 'YVEN', 'Norway', 'active', 'bronze', 'Private person - Internal employee', false, 'SARPSBORG', 'ØSTFOLD'),
('840bbcd2-549a-4a7c-a38e-a2c190ea20ab', 'Jan Bremer Øvrebø', NULL, '', '', '', '', '', 'Norway', 'lead', 'bronze', 'Private person', false, '', ''),
('a6a50826-5a2b-44c8-9b45-53e84074c55c', 'Jan Fredrik Smith', NULL, 'jan.smith@straye.no', '+4791160120', 'Buskogen 72', '1675', 'KRÅKERØY', 'Norway', 'active', 'bronze', 'Private person - Internal employee', false, 'FREDRIKSTAD', 'ØSTFOLD'),
('09a9e65c-61cb-47ae-b301-901a7be5d976', 'Jan Olav Martinsen', NULL, '', '', '', '', '', 'Norway', 'lead', 'bronze', 'Private person', false, '', ''),
('10df8669-9e05-4ec1-8685-3598dd59bfb8', 'Jan Svendsen', NULL, '', '+4790041004', '', '2016', 'FROGNER', 'Norway', 'active', 'bronze', 'Private person', false, 'LILLESTRØM', 'AKERSHUS'),
('94e9af46-58f8-4516-9699-09251175209f', 'Jan-Erik Tørmoen', NULL, 'janerik@tormoen.no', '+4790840847', 'Skjettenveien 114A', '2013', 'SKJETTEN', 'Norway', 'active', 'bronze', 'Private person', false, 'LILLESTRØM', 'AKERSHUS'),
('0bd8ccac-623e-4661-af46-9631148d9b8a', 'Janne Ekeberg', NULL, 'janne.ekeberg@gmail.com', '+4790950313', 'Hurrødåsen 1', '1621', 'GRESSVIK', 'Norway', 'active', 'bronze', 'Private person', false, 'FREDRIKSTAD', 'ØSTFOLD'),
('ce2c1349-951b-43ad-a541-3ed44e10dcb4', 'Jesper Vogt-Lorentzen', NULL, '', '+4748012336', 'Grinda 9B', '0861', 'OSLO', 'Norway', 'active', 'bronze', 'Private person', false, 'OSLO', 'OSLO'),
('9845ba7b-7f3a-407d-8808-42be10832ee4', 'Kyrre Johansen', NULL, '', '+47 99536510', 'Husarveien 35', '1396', 'BILLINGSTAD', 'Norway', 'active', 'bronze', 'Private person', false, 'ASKER', 'AKERSHUS'),
('e03690b0-75cf-4e1b-a794-829640be91f6', 'Morten Andre Kristiansen', NULL, 'mawaak@gamail.com', '+47 97753733', 'Grinda 9C', '0861', 'OSLO', 'Norway', 'active', 'bronze', 'Private person', false, 'OSLO', 'OSLO'),
('dc0e9517-8e49-478b-8c56-22ef37fcc184', 'Per Bremer Øvrebø', NULL, 'per@heireklame.no', '+47 90037320', '', '1177', 'OSLO', 'Norway', 'active', 'bronze', 'Private person', false, 'OSLO', 'OSLO'),
('0bba8a0f-4d10-4321-8631-282f4bc2a20a', 'Per Thormod Skogstad', NULL, '', '+4790691185', 'Hvalsodden 29', '1394', 'NESBRU', 'Norway', 'active', 'bronze', 'Private person', false, 'ASKER', 'AKERSHUS'),
('5f664bab-907f-493f-b71f-0254b91b25f0', 'Rigmor Frøystad', NULL, 'rigmfr@online.no', '+47 92252905', 'Løkenåsringen 20', '1473', 'LØRENSKOG', 'Norway', 'active', 'bronze', 'Private person', false, 'LØRENSKOG', 'AKERSHUS'),
('c86641e9-1fff-4489-9073-1d5401dc15d2', 'Øyvind Bjørn Kristiansen', NULL, '', '', 'Grinda 9', '0861', 'OSLO', 'Norway', 'active', 'bronze', 'Private person', false, 'OSLO', 'OSLO')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    email = COALESCE(NULLIF(EXCLUDED.email, ''), customers.email),
    phone = COALESCE(NULLIF(EXCLUDED.phone, ''), customers.phone),
    address = COALESCE(NULLIF(EXCLUDED.address, ''), customers.address),
    postal_code = COALESCE(NULLIF(EXCLUDED.postal_code, ''), customers.postal_code),
    city = COALESCE(NULLIF(EXCLUDED.city, ''), customers.city),
    country = COALESCE(NULLIF(EXCLUDED.country, ''), customers.country),
    notes = COALESCE(NULLIF(EXCLUDED.notes, ''), customers.notes),
    municipality = COALESCE(NULLIF(EXCLUDED.municipality, ''), customers.municipality),
    county = COALESCE(NULLIF(EXCLUDED.county, ''), customers.county);
