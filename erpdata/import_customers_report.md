# Customer Import Report

Generated: 2025-12-10 20:26:39

## Summary

| Metric | Count |
|--------|-------|
| Total customers to import | 168 |
| From customers_20251210075553.xlsx | 99 |
| From Plan Straye Tak only | 60 |
| From offers (referenced only) | 14 |
| Used existing IDs | 38 |
| Created new IDs | 121 |
| Duplicates skipped | 16 |

## Data Sources

1. **customers_20251210075553.xlsx** - Full customer data with org numbers, addresses, contact info
2. **Plan Straye Tak new.xlsx** - Customer names only from offer tracking sheet
3. **import_offers.sql** - Customer IDs referenced in existing offers

## Issues Found

### Issue Summary

| Issue Type | Count | Severity |
|------------|-------|----------|
| Missing name | 1 | 🔴 High |
| Invalid org number | 0 | 🟡 Medium |
| Missing contact info | 32 | 🟡 Medium |
| New customer, no data | 49 | 🟡 Medium |
| Orphan from offers | 14 | 🟢 Low |

### Missing Names (Skipped)

- Row 30 in customers_export

### Missing Contact Information

These customers have no email or phone number:

- BOLIGSAMEIET RÅDHUSPLASS
- BRASETVIK BYGG AS
- Brick AS
- Christian Gran
- EIERSEKSJONSSAMEIET BERGRÅDVEIEN 5
- ENTER SOLUTION AS
- FJELLHAUG EIENDOM
- FLØYSAND TAK AS
- GREV WEDELS PLASS 9 AS
- Hent AS
- Holmskau Prosjekt AS
- Lillestrøm Tak og Membran AS
- Logistic Contractor Norge AS
- MITTEGETLOKALE PORSGRUNN AS
- NP BYGG AS
- Norbygg AS
- OSLOBYGG KF
- RANDEM & HÜBERT AS
- Rg Fjellsikring AS
- Rørlegger Sentralen AS
- SAMEIET KJØRBOKOLLEN 19-29
- SAMEIET TIDEMANDS GATE 28
- SOLCELLESPESIALISTEN AS
- Straye Gruppen AS
- Straye Hybridbygg AS
- Straye Industri AS
- Straye Montasje AS
- Straye Stålbygg AS
- Vest Entreprenør AS
- Vestliterrassen Boligsameie
- ØSTLANDSTAK AS
- Øyvind Bjørn Kristiansen

### New Customers with No Data

These customers exist only in Plan Straye Tak and need enrichment:

- A Bygg
- Arealbygg
- Betongbygg AS
- Betonmast Romerike
- Betonmast Trøndelag
- Bomekan
- Brick
- Byggekompaniet Østfold
- Byggekompaniet Østfold
- Byggkompaniet
- Enter Solutions AS
- Furuno AS
- Fusen
- Geir Nielsen (Holmskau)
- Gressvik Properties
- Grinda 9 Revidert
- Hallmaker
- Hallmaker AS/ Straye Stålbygg AS
- Hersleth Entreprenør
- Høstbakken 11 AS
- KM Bygg
- Matotalbygg AS
- Nordbygg AS
- PEAB
- PEAB AS
- Park & Anlegg
- Parketteksperten
- Peab/ Straye Stålbygg
- Sameie Hoffsveien 88
- Sameiet Kornmoenga
- Seby AS/ Veidekke AS
- Straye Hybrid AS
- Straye Industri
- Straye Industribygg AS
- Straye Stålbygg
- Straye Stålbygg AS / Hallmaker
- Straye Stålbygg AS/ Hallmaker
- StrayeIndustri AS
- TatalBygg Midt-norge AS
- Thermica AS/ Hallmaker
- Totalbetong
- Veidekke AS
- Veidekke Bygg- Vest
- Veidekke Entreprenør
- Veidekke Ålesund
- Vest Entreprenør
- Vestre Bærum Tennis
- Workman AS
- dpend/ Straye Stålbygg

### Orphan Customers from Offers

These customers are referenced in offers but not found in Excel files:

| Customer | UUID |
|----------|------|
| As Betongbygg | `157b35d0-5d9e-44a4-9e3f-7de033202cc9` |
| Betonmast Romerike As | `c00b8d38-c71a-4287-8182-007514a1e769` |
| Betonmast Trøndelag As | `b6039121-40fc-46f2-86a4-54cabe7caf43` |
| Bomekan As | `8f5b4c19-3531-49c7-8eb2-cd95bacc22a2` |
| Fusen As | `a27e4bf6-59c2-4a18-9fd3-f031fcdc738f` |
| Gressvik Properties As | `8481eca2-3af3-4780-9241-4bd38fbb3be6` |
| Høstbakken Eiendom As | `ec13ae1c-23a0-4d5e-993a-7d71215ed492` |
| Jesper Vogt Lorentzen | `32b751fb-bb21-448f-9a36-0ea2d6ad5490` |
| Parketteksperten As | `dd52c2be-fb94-4267-9abf-0f890e733845` |
| Peab Bygg As | `4f4d6062-7e83-4cf2-adb1-f0c6855203c7` |
| Sameiet Hoffsveien 88/90 | `fa874f0d-3239-4e2d-96d2-c14a06fdeedd` |
| Totalbetong Gruppen As | `923d08e0-2f7f-4774-ade3-bcccbbce9a4c` |
| Totalbygg Midt-Norge As | `d2d2784f-41fe-428b-a5fc-5dcac100ed1a` |
| Workman Norway As | `5e39e696-267c-4c87-9ae6-c29eb20cd065` |

## Recommendations

1. **Enrich Plan Straye Tak customers**: 60 customers from the offer tracking sheet have only names. Consider looking up their org numbers and contact info in Brønnøysund register.
2. **Update missing contact info**: 106 customers have no email or phone. This will limit communication capabilities.
3. **Verify orphan customers**: 14 customers from offers don't exist in Excel files - verify these are valid customers.

## Import Command

```bash
# Run from the API container or with database access
psql -U relation_user -d relation -f erpdata/import_customers.sql
```

The import uses `ON CONFLICT DO UPDATE` to safely update existing customers without overwriting non-empty fields.
