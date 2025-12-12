# Straye Tak Offers Import Plan

Last updated: 2025-12-12

## Source Data

**File:** `Plan Straye Tak new.xlsx`
**Sheet:** `Utsendte tilbud` (Sent Offers)
**Header Row:** 8
**Data Rows:** 9-488

## Summary Statistics

| Category | Count |
|----------|-------|
| Total rows with customer data | 434 |
| Offers imported | 421 |
| Offers skipped (UTGÅR) | 13 |
| Unique customer names in Excel | 72 |

### Customer Match Summary

| Category | Customers | Offers | Status |
|----------|-----------|--------|--------|
| Confirmed Matches | 72 | 421 | ✅ Imported |

---

## Column Mapping: Excel -> Offer Model

| Excel Column | Excel Name | Offer Field | Notes |
|--------------|------------|-------------|-------|
| A (1) | Sendt | `sent_date` | Date when offer was sent |
| B (2) | Ansvarlig | `responsible_user_name` | Initials only (HSK, etc.) |
| C (3) | Kunde / Byggherre | `customer_id` + `customer_name` | Must match to DB customers |
| D (4) | Prosjekt | `title` | Project name/offer title |
| E (5) | Frist | `due_date` | Deadline |
| F (6) | Beliggenhet | `location` | Location |
| G (7) | m2 | `description` | Store in description |
| P (16) | Status | `phase` | See status mapping below |
| Q (17) | Tilbudspris | `value` | Offer value in NOK |
| U (21) | Vedstaelses frist | `expiration_date` | Formula: sent_date + 60 days |
| W (23) | Beskrivelse / siste nytt | `notes` | Description/latest update |

### Status Mapping

| Excel Status | Offer Phase | Condition |
|--------------|-------------|-----------|
| `BUDSJETT` | `in_progress` | No sent_date |
| `BUDSJETT` | `sent` | Has sent_date |
| `Tilbud` | `in_progress` | No sent_date |
| `Tilbud` | `sent` | Has sent_date |
| `Ordre` | `won` | - |
| `Ferdig` | `won` | - |
| `Tapt` | `lost` | - |
| `UTGAR` | **SKIP** | These are deleted/expired |

### Generated Fields

| Field | Value |
|-------|-------|
| `company_id` | `tak` (all offers are for Straye Tak) |
| `status` | `active` |
| `probability` | Based on phase: draft=10%, in_progress=30%, sent=50%, won=100%, lost=0% |
| `offer_number` | Generate new: `TK-{YEAR}-{SEQ}` |
| `currency` | `NOK` |

---

## Customer Matching Results

### CONFIRMED MATCHES - 72 Customers (421 Offers)

All mappings verified and ready for import. See `customer_mapping.json` for complete UUID mapping.

| Excel Customer Name | Database Customer Name |
|---------------------|------------------------|
| A Bygg | A BYGG ENTREPRENØR AS |
| Arealbygg | Areal Bygg AS |
| Betongbygg AS | AS Betongbygg |
| Betonmast Romerike | BETONMAST ROMERIKE AS |
| Betonmast Trøndelag | BETONMAST TRØNDELAG AS |
| Bomekan | BOMEKAN AS |
| Brick | Brick AS |
| Byggekompaniet Østfold | Byggkompaniet Østfold AS |
| Byggkompaniet | Byggkompaniet Østfold AS |
| Christian Gran | Christian Gran |
| Dan blikk AS | DAN BLIKK AS |
| Dybvig AS | DYBVIG AS |
| Enter Solutions AS | ENTER SOLUTION AS |
| Fjellhaug Eiendom | FJELLHAUG EIENDOM |
| Furuno AS | Furuno Norge AS |
| Fusen | SOLENERGI FUSEN AS |
| Geir Nielsen (Holmskau) | Holmskau Prosjekt AS |
| Grinda 9 Revidert | Jesper Vogt-Lorentzen |
| Gressvik Properties | GRESSVIK PROPERTIES AS |
| Hallgruppen AS | Hallgruppen AS |
| Hallmaker / Hallmaker AS | HALLMAKER AS |
| Hallmaker AS/ Straye Stålbygg AS | HALLMAKER AS (joint venture) |
| Hersleth Entreprenør (AS) | HERSLETH ENTREPRENØR AS |
| Høstbakken 11 AS | HØSTBAKKEN EIENDOM AS |
| Jan Bremer Øvrebø | Jan Bremer Øvrebø |
| Jan Olav Martinsen | Jan Olav Martinsen |
| Jan Svendsen | Jan Svendsen |
| KM Bygg | KOPPERUD MURTNES BYGG AS |
| Konsmo fabrikker AS | KONSMO FABRIKKER AS |
| MA Totalbygg as / Matotalbygg AS | MA Totalbygg AS |
| MK Eiendom AS | Mk Eiendom AS |
| Newsec AS | NEWSEC AS |
| Nordbygg AS | Norbygg AS |
| Ocab AS | OCAB AS |
| PEAB / PEAB AS | PEAB BYGG AS |
| Park & Anlegg | Park og Anlegg A/S |
| Parketteksperten | PARKETTEKSPERTEN AS |
| Peab/ Straye Stålbygg | PEAB BYGG AS (joint venture) |
| Sameie Hoffsveien 88 | SAMEIET HOFFSVEIEN 88/90 |
| Sameiet Kornmoenga | KORNMOENGA 3 SAMEIE |
| Seby AS/ Veidekke AS | Veidekke Entreprenør AS (joint venture) |
| Seltor AS | SELTOR AS |
| Straye Hybrid AS / Hybridbygg AS | Straye Hybridbygg AS |
| Straye Industri (AS) / Industribygg AS | Straye Industri AS |
| Straye Stålbygg (AS) | Straye Stålbygg AS |
| Straye Stålbygg AS / Hallmaker | Straye Stålbygg AS (joint venture) |
| Straye Tak AS | Straye Tak AS |
| Straye Trebygg AS | STRAYE TREBYGG AS |
| T.r Eiendom AS | T.r Eiendom AS |
| TatalBygg Midt-norge AS | TOTALBYGG MIDT-NORGE AS |
| Thermica AS | Thermica AS |
| Thermica AS/ Hallmaker | Thermica AS (joint venture) |
| Totalbetong | TOTALBETONG GRUPPEN AS |
| Veidekke AS | Veidekke Entreprenør AS |
| Veidekke Bygg- Vest | Veidekke Entreprenør AS |
| Veidekke Entreprenør | Veidekke Entreprenør AS |
| Veidekke Ålesund | Veidekke Entreprenør AS |
| Veidekke Logistikkbygg AS | Veidekke Logistikkbygg AS |
| Vest Entreprenør | Vest Entreprenør AS |
| Vestre Bærum Tennis(klubb) | Vestre Bærum Tennisklubb |
| Workman AS | WORKMAN NORWAY AS |
| dpend/ Straye Stålbygg | Straye Stålbygg AS (joint venture) |

---

## Status Distribution (Excluding UTGAR)

| Status | Count | Target Phase |
|--------|-------|--------------|
| Tilbud | 259 | in_progress/sent |
| Ferdig | 61 | won |
| Tapt | 55 | lost |
| Ordre | 31 | won |
| BUDSJETT | 15 | in_progress/sent |

---

## Import Implementation Plan

### Phase 1: Preparation ✅ Complete
1. [x] Review and approve all customer mappings
2. [x] Decide on joint venture handling strategy (mapped to primary customer)
3. [x] Resolve "Grinda 9 Revidert" -> mapped to Jesper Vogt-Lorentzen
4. [x] Responsible user names stored (HSK, etc.) - IDs left NULL for historical data

### Phase 2: Generate Import SQL ✅ Complete
1. [x] Parse Excel data using customer_mapping.json
2. [x] Apply status mapping logic
3. [x] Generate offer_number sequence (TK-YEAR-SEQ)
4. [x] Calculate expiration_date from sent_date + 60 days
5. [x] Generate INSERT statements (434 offers in 5 batches)

### Phase 3: Import
1. [ ] Run import SQL against database
2. [ ] Verify offer counts match expectations
3. [ ] Spot-check random offers for data accuracy

---

## Notes

- All offers will have `company_id = 'tak'`
- Currency is NOK for all offers
- The "Ansvarlig" column contains user initials (HSK) that need mapping to user IDs
- m2 (square meters) will be stored in the description field
- Joint ventures are mapped to the first/primary customer in the pair

---

## Files

| File | Purpose |
|------|---------|
| `offers.md` | This implementation plan |
| `customer_mapping.json` | Machine-readable customer name to UUID mapping (72 confirmed) |
| `import_offers.sql` | SQL import script with 421 offers (9298 lines, 5 batches) |

## Usage

To import offers after a database wipe:
```bash
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_offers.sql
```

## Offer Numbers Generated

| Year | Count | Range |
|------|-------|-------|
| 2022 | 13 | TK-2022-0001 to TK-2022-0013 |
| 2023 | 85 | TK-2023-0001 to TK-2023-0085 |
| 2024 | 143 | TK-2024-0001 to TK-2024-0143 |
| 2025 | 180 | TK-2025-0001 to TK-2025-0180 |
