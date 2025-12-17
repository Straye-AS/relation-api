#!/usr/bin/env python3
"""
Import offers from ERP Excel export into the Relation API database.

This script:
1. Reads offer data from an Excel file (Plan Straye Tak format)
2. Matches customers by name (fuzzy match with normalization)
3. Skips offers where customer is not found (reports them)
4. Imports offers with status mapping

Usage:
    python import_offers.py [excel_file] [--dry-run] [--clear]

Environment variables:
    DATABASE_HOST     (default: localhost)
    DATABASE_PORT     (default: 5432)
    DATABASE_USER     (default: relation_user)
    DATABASE_PASSWORD (default: relation_password)
    DATABASE_NAME     (default: relation)
"""

import argparse
import os
import re
import sys
import uuid
from datetime import datetime, timezone

import pandas as pd
import psycopg2
from psycopg2.extras import execute_values


# Mapping of Excel status to our Phase and Status
# Note: "BUDSJETT" and "Tilbud" need sent_date to determine phase (handled in get_status_mapping)
# Phase meanings:
#   - in_progress: Being worked on internally (no sent_date)
#   - sent: Sent to customer (has sent_date)
#   - order: Customer accepted, work in progress
#   - completed: Order finished
#   - lost: Customer rejected
#   - expired: Offer expired (UTGÅR - skip these)
STATUS_MAPPING = {
    "Tilbud": ("_needs_sent_date", "active"),   # Depends on sent_date
    "Ferdig": ("completed", "active"),          # Completed order
    "Ordre": ("order", "active"),               # Active order
    "Tapt": ("lost", "lost"),                   # Lost
    "BUDSJETT": ("_needs_sent_date", "active"), # Depends on sent_date
    "UTGÅR": ("_skip", "expired"),              # Skip these entirely
    "UTLØPT": ("_skip", "expired"),             # Skip these entirely
}

# Responsible person initials to names
RESPONSIBLE_MAPPING = {
    "HSK": "Håkon Knutsen",
    "KL": "Kristoffer Larsen",
    "AB": "Anders Berg",
}

# Default company for this import (Straye Tak)
DEFAULT_COMPANY_ID = "tak"


def get_db_connection():
    """Create database connection from environment variables."""
    return psycopg2.connect(
        host=os.getenv("DATABASE_HOST", "localhost"),
        port=os.getenv("DATABASE_PORT", "5432"),
        user=os.getenv("DATABASE_USER", "relation_user"),
        password=os.getenv("DATABASE_PASSWORD", "relation_password"),
        dbname=os.getenv("DATABASE_NAME", "relation"),
    )


def clean_string(value) -> str:
    """Clean a string value, handling NaN and None."""
    if pd.isna(value) or value is None:
        return ""
    return str(value).strip()


def clean_float(value) -> float:
    """Clean a float value, returning 0 if not valid."""
    if pd.isna(value) or value is None:
        return 0.0
    try:
        return float(value)
    except (ValueError, TypeError):
        return 0.0


def clean_date(value) -> datetime | None:
    """Clean a date value, returning None if not valid."""
    if pd.isna(value) or value is None:
        return None
    try:
        if isinstance(value, datetime):
            return value.replace(tzinfo=timezone.utc)
        if isinstance(value, str):
            for fmt in ["%Y-%m-%d", "%d.%m.%Y", "%Y-%m-%d %H:%M:%S"]:
                try:
                    return datetime.strptime(value, fmt).replace(tzinfo=timezone.utc)
                except ValueError:
                    continue
        return None
    except Exception:
        return None


def get_status_mapping(excel_status: str, sent_date=None) -> tuple[str, str]:
    """Map Excel status to (phase, status) tuple.

    Args:
        excel_status: The status from Excel
        sent_date: The sent date (used to determine if offer was sent)

    Returns:
        Tuple of (phase, status). Phase can be:
        - "_skip": This offer should be skipped entirely
        - "_needs_sent_date": Handled here based on sent_date
        - Regular phase: in_progress, sent, order, completed, lost
    """
    excel_status = clean_string(excel_status)

    # Get base mapping
    phase, status = STATUS_MAPPING.get(excel_status, ("_needs_sent_date", "active"))

    # Handle status that depends on sent_date
    if phase == "_needs_sent_date":
        if sent_date is not None:
            phase = "sent"
        else:
            phase = "in_progress"

    return (phase, status)


def get_responsible_name(initials: str) -> str:
    """Map initials to full name."""
    initials = clean_string(initials).upper()
    return RESPONSIBLE_MAPPING.get(initials, initials)


# Customer name aliases - maps variations to canonical names (normalized, lowercase, no AS suffix)
CUSTOMER_ALIASES = {
    # === Straye variations ===
    "straye hybrid": "straye hybridbygg",
    "straye hybrid as": "straye hybridbygg",
    "strayeindustri": "straye industri",
    "strayeindustri as": "straye industri",
    "straye industribygg": "straye industri",
    "straye industribygg as": "straye industri",

    # === Veidekke variations - all map to Veidekke Entreprenør ===
    "veidekke": "veidekke entreprenør",
    "veidekke as": "veidekke entreprenør",
    "veidekke bygg": "veidekke entreprenør",
    "veidekke bygg- vest": "veidekke entreprenør",
    "veidekke bygg vest": "veidekke entreprenør",
    "veidekke ålesund": "veidekke entreprenør",
    "seby as/ veidekke as": "veidekke entreprenør",

    # === PEAB variations ===
    "peab": "peab bygg",
    "peab as": "peab bygg",
    "peab/ straye stålbygg": "straye stålbygg",  # Internal project

    # === Section 2: Probable Matches ===
    "a bygg": "a bygg entreprenør",
    "betongbygg": "as betongbygg",
    "betongbygg as": "as betongbygg",
    "byggkompaniet": "byggkompaniet østfold",
    "enter solutions": "enter solution",
    "enter solutions as": "enter solution",
    "furuno": "furuno norge",
    "furuno as": "furuno norge",
    "fusen": "solenergi fusen",
    "totalbetong": "totalbetong gruppen",
    "vestre bærum tennis": "vestre bærum tennisklubb",
    "workman": "workman norway",
    "workman as": "workman norway",

    # === Joint/combined offers - assign to Hallmaker ===
    "hallmaker as/ straye stålbygg as": "hallmaker",
    "straye stålbygg as / hallmaker": "hallmaker",
    "straye stålbygg as/ hallmaker": "hallmaker",
    "thermica as/ hallmaker": "hallmaker",
    "dpend/ straye stålbygg": "straye stålbygg",

    # === Section 3: Needs Attention - Typos and variations ===
    "byggekompaniet østfold": "byggkompaniet østfold",  # Typo: extra 'e'
    "km bygg": "kopperud murtnes bygg",
    "km bygg??": "kopperud murtnes bygg",
    "arealbygg": "areal bygg",
    "geir nielsen (holmskau)": "geir nilsen",  # Spelling: Nielsen vs Nilsen
    "hansen & dahl": "hansen & dahl",  # Same but needs exact match
    "høstbakken 11": "høstbakken eiendom",
    "høstbakken 11 as": "høstbakken eiendom",
    "matotalbygg": "ma totalbygg",
    "matotalbygg as": "ma totalbygg",
    "nordbygg": "norbygg",
    "nordbygg as": "norbygg",
    "park & anlegg": "park og anlegg",
    "sameie hoffsveien 88": "sameiet hoffsveien 88/90",
    "sameiet kornmoenga": "kornmoenga 3 sameie",
    "tatalbygg midt-norge": "totalbygg midt-norge",
    "tatalbygg midt-norge as": "totalbygg midt-norge",
    "øm fjell": "ø.m. fjeld",
    "øm fjell as": "ø.m. fjeld",

    # === Unknown/placeholder mappings ===
    "grinda 9 revidert": "jesper vogt-lorentzen",  # Project name, not customer
    "flere- gikk til km bygg": "kopperud murtnes bygg",
    "flere": "kopperud murtnes bygg",
    "ukjent kunde": "kopperud murtnes bygg",
    "2581 hjelseth": "kopperud murtnes bygg",
}


def normalize_customer_name(name: str) -> str:
    """Normalize customer name for fuzzy matching."""
    if not name:
        return ""
    name = name.lower().strip()
    # Remove extra whitespace
    name = re.sub(r"\s+", " ", name)
    # Check for aliases BEFORE removing suffixes (for combined names)
    if name in CUSTOMER_ALIASES:
        name = CUSTOMER_ALIASES[name]
    # Remove common suffixes (AS, A/S, ANS, DA, SA)
    name = re.sub(r"\s+(as|a/s|ans|da|sa)$", "", name)
    # Check aliases again after suffix removal
    if name in CUSTOMER_ALIASES:
        name = CUSTOMER_ALIASES[name]
    return name


# Cache for customer lookups
_customer_cache: dict[str, tuple[str, str]] = {}  # normalized_name -> (id, original_name)
_customer_cache_loaded = False


def load_customer_cache(conn):
    """Load all customers into cache for fast lookups."""
    global _customer_cache, _customer_cache_loaded
    if _customer_cache_loaded:
        return

    with conn.cursor() as cur:
        cur.execute("SELECT id, name FROM customers")
        for row in cur.fetchall():
            customer_id, name = str(row[0]), row[1]
            normalized = normalize_customer_name(name)
            _customer_cache[normalized] = (customer_id, name)
    _customer_cache_loaded = True


def find_customer(conn, customer_name: str) -> tuple[str, str] | None:
    """Find customer ID by name (fuzzy match). Returns (id, matched_name) or None."""
    load_customer_cache(conn)

    normalized = normalize_customer_name(customer_name)
    if normalized in _customer_cache:
        return _customer_cache[normalized]

    return None


def read_excel_offers(filepath: str) -> tuple[list[dict], dict]:
    """Read and transform Excel data to offer records.

    Returns:
        Tuple of (offers_list, skip_stats) where skip_stats tracks skipped rows
    """
    df = pd.read_excel(filepath, header=7)
    offers = []
    skip_stats = {"empty_rows": 0, "utgaar": 0, "no_customer": 0}

    for _, row in df.iterrows():
        project_title = clean_string(row.get("Prosjekt", ""))
        if not project_title:
            skip_stats["empty_rows"] += 1
            continue

        # Split title into external_reference and name
        # Format is typically "<external_reference> <name>" e.g. "22000 Hjalmar Bjørges vei 105"
        external_reference = ""
        title = project_title
        parts = project_title.split(" ", 1)
        if len(parts) == 2 and parts[0].isdigit():
            external_reference = parts[0]
            title = parts[1]

        customer_name = clean_string(row.get("Kunde / Byggherre", ""))
        if not customer_name:
            customer_name = "Ukjent kunde"

        # Get sent_date first - needed for phase determination
        sent_date = clean_date(row.get("Sendt"))

        # Get status mapping (now depends on sent_date)
        excel_status = clean_string(row.get("Status", ""))
        phase, status = get_status_mapping(excel_status, sent_date)

        # Skip UTGÅR/UTLØPT entries
        if phase == "_skip":
            skip_stats["utgaar"] += 1
            continue

        value = clean_float(row.get("Tilbudspris", 0))
        margin_amount = clean_float(row.get("DB", 0))
        cost = value - margin_amount if value > 0 else 0

        due_date = clean_date(row.get("Vedståelses frist"))
        last_updated = clean_date(row.get("Sist oppdatert"))

        location = clean_string(row.get("Beliggenhet", ""))
        notes = clean_string(row.get("Beskrivelse / siste nytt", ""))
        responsible_initials = clean_string(row.get("Ansvarlig", ""))
        responsible_name = get_responsible_name(responsible_initials)

        # Set probability based on phase
        probability = 0
        if phase == "in_progress":
            probability = 20
        elif phase == "sent":
            probability = 50
        elif phase == "order":
            probability = 100  # Already accepted
        elif phase == "completed":
            probability = 100  # Finished

        offer = {
            "id": str(uuid.uuid4()),
            "title": title,
            "external_reference": external_reference,
            "customer_name": customer_name,
            "company_id": DEFAULT_COMPANY_ID,
            "phase": phase,
            "status": status,
            "probability": probability,
            "value": value,
            "cost": cost,
            "location": location,
            "notes": notes,
            "responsible_user_name": responsible_name,
            "sent_date": sent_date,
            "due_date": due_date,
            "created_at": sent_date or datetime.now(timezone.utc),
            "updated_at": last_updated or datetime.now(timezone.utc),
        }

        offers.append(offer)

    return offers, skip_stats


def assign_offer_numbers(offers: list[dict], company_prefix: str = "TK") -> list[dict]:
    """Assign sequential offer numbers based on sent_date.

    Format: {PREFIX}-{YEAR}-{SEQ:03d}
    Example: TK-2023-001, TK-2023-002, TK-2024-001

    Offers are sorted by sent_date (nulls last, then by external_reference).
    Numbers are assigned per year.
    """
    # Sort offers: by sent_date (nulls last), then by external_reference
    def sort_key(o):
        sent = o.get("sent_date")
        ext_ref = o.get("external_reference", "")
        # Use a far future date for nulls so they sort last
        if sent is None:
            return (datetime(9999, 12, 31, tzinfo=timezone.utc), ext_ref)
        return (sent, ext_ref)

    sorted_offers = sorted(offers, key=sort_key)

    # Track sequence numbers per year
    year_sequences: dict[int, int] = {}

    for offer in sorted_offers:
        # Determine year from sent_date or created_at
        sent_date = offer.get("sent_date")
        created_at = offer.get("created_at")

        if sent_date:
            year = sent_date.year
        elif created_at:
            year = created_at.year
        else:
            year = datetime.now().year

        # Get next sequence for this year
        seq = year_sequences.get(year, 0) + 1
        year_sequences[year] = seq

        # Generate offer number
        offer["offer_number"] = f"{company_prefix}-{year}-{seq:03d}"

    return sorted_offers


def clear_offers(conn, company_id: str, dry_run: bool = False):
    """Clear all existing offers for a company."""
    with conn.cursor() as cur:
        cur.execute("SELECT COUNT(*) FROM offers WHERE company_id = %s", (company_id,))
        count = cur.fetchone()[0]
        print(f"  Found {count} existing offers for {company_id} to delete")

        if not dry_run:
            cur.execute("DELETE FROM offers WHERE company_id = %s", (company_id,))
            print(f"  Deleted {count} offers")


def insert_offers(conn, offers: list[dict], dry_run: bool = False) -> list[dict]:
    """Insert offers into the database. Returns list of skipped offers."""
    if not offers:
        print("  No offers to insert")
        return []

    # Load customer cache
    load_customer_cache(conn)

    # Separate offers into those we can import and those we can't
    importable = []
    skipped = []

    for offer in offers:
        customer_match = find_customer(conn, offer["customer_name"])
        if customer_match:
            customer_id, matched_name = customer_match
            offer["customer_id"] = customer_id
            offer["matched_customer_name"] = matched_name
            importable.append(offer)
        else:
            skipped.append(offer)

    # Show stats
    stats = {"in_progress": 0, "sent": 0, "order": 0, "completed": 0, "lost": 0}
    total_value = 0
    for offer in importable:
        stats[offer["phase"]] = stats.get(offer["phase"], 0) + 1
        total_value += offer["value"]

    print(f"  Can import: {len(importable)} offers")
    print(f"  Skipped (no customer match): {len(skipped)} offers")
    print()
    print(f"  Importable breakdown:")
    print(f"    - In Progress: {stats.get('in_progress', 0)}")
    print(f"    - Sent: {stats.get('sent', 0)}")
    print(f"    - Order: {stats.get('order', 0)}")
    print(f"    - Completed: {stats.get('completed', 0)}")
    print(f"    - Lost: {stats.get('lost', 0)}")
    print(f"    - Total value: {total_value:,.0f} NOK")

    if dry_run:
        print("\n  Sample importable offers:")
        for o in importable[:5]:
            print(f"    - {o['title'][:40]} | {o['customer_name'][:25]} | {o['phase']} | {o['value']:,.0f} NOK")
        if len(importable) > 5:
            print(f"    ... and {len(importable) - 5} more")
        return skipped

    if not importable:
        return skipped

    # Insert offers
    columns = [
        "id", "title", "customer_id", "customer_name", "company_id",
        "phase", "status", "probability", "value", "cost", "location",
        "notes", "responsible_user_name", "sent_date", "due_date",
        "created_at", "updated_at"
    ]

    values_to_insert = [
        (
            offer["id"],
            offer["title"],
            offer["customer_id"],
            offer["matched_customer_name"],  # Use the matched name from DB
            offer["company_id"],
            offer["phase"],
            offer["status"],
            offer["probability"],
            offer["value"],
            offer["cost"],
            offer["location"],
            offer["notes"],
            offer["responsible_user_name"],
            offer["sent_date"],
            offer["due_date"],
            offer["created_at"],
            offer["updated_at"],
        )
        for offer in importable
    ]

    with conn.cursor() as cur:
        insert_sql = f"""
            INSERT INTO offers ({', '.join(columns)})
            VALUES %s
        """
        execute_values(cur, insert_sql, values_to_insert)
        print(f"\n  Inserted {len(importable)} offers")

    return skipped


def save_skipped_offers(skipped: list[dict], output_file: str):
    """Save skipped offers to a CSV file for manual review."""
    if not skipped:
        return

    df = pd.DataFrame([
        {
            "customer_name": o["customer_name"],
            "title": o["title"],
            "value": o["value"],
            "phase": o["phase"],
            "location": o["location"],
        }
        for o in skipped
    ])

    df.to_csv(output_file, index=False)
    print(f"  Saved {len(skipped)} skipped offers to: {output_file}")


def escape_sql_string(value: str) -> str:
    """Escape a string for SQL insertion."""
    if value is None:
        return ""
    return str(value).replace("'", "''")


def format_sql_date(dt: datetime | None) -> str:
    """Format a datetime for SQL, or NULL if None."""
    if dt is None:
        return "NULL"
    return f"'{dt.strftime('%Y-%m-%d %H:%M:%S')}'"


def resolve_customer_name_for_sql(excel_name: str, db_customers: dict[str, str]) -> str:
    """Resolve Excel customer name to actual database customer name.

    Args:
        excel_name: Customer name from Excel
        db_customers: Dict mapping normalized names to actual DB names

    Returns:
        The actual database customer name, or the normalized Excel name as fallback
    """
    normalized = normalize_customer_name(excel_name)
    if normalized in db_customers:
        return db_customers[normalized]
    return excel_name


def generate_sql_file(offers: list[dict], output_file: str, skip_stats: dict):
    """Generate a SQL file with INSERT statements for all offers."""

    # Connect to database to get actual customer names
    print("  Loading customer names from database...")
    try:
        conn = get_db_connection()
        with conn.cursor() as cur:
            cur.execute("SELECT name FROM customers")
            db_customers = {}
            for row in cur.fetchall():
                name = row[0]
                normalized = normalize_customer_name(name)
                db_customers[normalized] = name
        conn.close()
        print(f"  Loaded {len(db_customers)} customers")
    except Exception as e:
        print(f"  WARNING: Could not load customers from DB: {e}")
        print(f"  SQL will use Excel customer names (may not match)")
        db_customers = {}

    # Stats for the header
    stats = {"in_progress": 0, "sent": 0, "order": 0, "completed": 0, "lost": 0}
    unmatched = []
    for offer in offers:
        stats[offer["phase"]] = stats.get(offer["phase"], 0) + 1
        # Check if customer will match
        normalized = normalize_customer_name(offer['customer_name'])
        if normalized not in db_customers:
            unmatched.append(offer['customer_name'])

    if unmatched:
        print(f"  WARNING: {len(set(unmatched))} unique customers won't match:")
        for name in sorted(set(unmatched))[:10]:
            print(f"    - {name}")
        if len(set(unmatched)) > 10:
            print(f"    ... and {len(set(unmatched)) - 10} more")

    with open(output_file, 'w') as f:
        # Write header
        f.write(f"-- Import offers from Plan Straye Tak last.xlsx\n")
        f.write(f"-- Generated: {datetime.now()}\n")
        f.write(f"-- To import: {len(offers)}\n")
        f.write(f"-- Skipped UTGÅR/UTLØPT: {skip_stats.get('utgaar', 0)}\n")
        f.write(f"-- Skipped empty rows: {skip_stats.get('empty_rows', 0)}\n")
        f.write(f"-- Phase breakdown: in_progress={stats['in_progress']}, sent={stats['sent']}, order={stats['order']}, completed={stats['completed']}, lost={stats['lost']}\n")
        f.write(f"-- Unmatched customers: {len(set(unmatched))}\n\n")

        # Write INSERT statements
        for offer in offers:
            # Resolve to actual DB customer name
            db_customer_name = resolve_customer_name_for_sql(offer['customer_name'], db_customers)

            customer_name_escaped = escape_sql_string(db_customer_name)
            title_escaped = escape_sql_string(offer['title'])
            notes_escaped = escape_sql_string(offer.get('notes', ''))
            location_escaped = escape_sql_string(offer.get('location', ''))
            external_ref_escaped = escape_sql_string(offer.get('external_reference', ''))
            offer_number_escaped = escape_sql_string(offer.get('offer_number', ''))

            sent_date_sql = format_sql_date(offer.get('sent_date'))

            sql = f"""INSERT INTO offers (id, title, customer_id, customer_name, company_id, phase, probability, value, status, description, notes, created_at, updated_at, external_reference, offer_number, cost, location, sent_date)
VALUES ('{offer['id']}', '{title_escaped}', (SELECT id FROM customers WHERE name = '{customer_name_escaped}'), '{customer_name_escaped}', '{offer['company_id']}', '{offer['phase']}', {offer['probability']}, {offer['value']}, '{offer['status']}', '', '{notes_escaped}', NOW(), NOW(), '{external_ref_escaped}', '{offer_number_escaped}', {offer['cost']}, '{location_escaped}', {sent_date_sql}) ON CONFLICT (id) DO NOTHING;\n"""
            f.write(sql)

    print(f"  Generated SQL file: {output_file}")
    print(f"  Total INSERT statements: {len(offers)}")


def main():
    parser = argparse.ArgumentParser(
        description="Import offers from Excel to database"
    )
    parser.add_argument(
        "excel_file",
        nargs="?",
        default="../erpdata/Plan Straye Tak last.xlsx",
        help="Path to Excel file"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be done without making changes"
    )
    parser.add_argument(
        "--clear",
        action="store_true",
        help="Clear existing offers for this company before import"
    )
    parser.add_argument(
        "--output-sql",
        type=str,
        metavar="FILE",
        help="Generate SQL file instead of inserting to database"
    )
    args = parser.parse_args()

    print(f"=== Offer Import Script ===")
    print(f"Excel file: {args.excel_file}")
    print(f"Company: {DEFAULT_COMPANY_ID}")
    if args.output_sql:
        print(f"Output SQL: {args.output_sql}")
    else:
        print(f"Dry run: {args.dry_run}")
        print(f"Clear existing: {args.clear}")
    print()

    # Read Excel data
    print("1. Reading Excel file...")
    offers, skip_stats = read_excel_offers(args.excel_file)
    print(f"   Found {len(offers)} valid offers in file")
    print(f"   Skipped: {skip_stats['empty_rows']} empty rows, {skip_stats['utgaar']} UTGÅR/UTLØPT")
    print()

    # Assign sequential offer numbers
    print("2. Assigning offer numbers...")
    offers = assign_offer_numbers(offers, company_prefix="TK")
    # Count offers per year
    year_counts: dict[int, int] = {}
    for o in offers:
        year = int(o["offer_number"].split("-")[1])
        year_counts[year] = year_counts.get(year, 0) + 1
    for year in sorted(year_counts.keys()):
        print(f"   {year}: {year_counts[year]} offers")
    print()

    # If --output-sql is specified, generate SQL file and exit
    if args.output_sql:
        print("3. Generating SQL file...")
        generate_sql_file(offers, args.output_sql, skip_stats)
        print()
        print("=== Done! ===")
        return

    # Connect to database
    print("2. Connecting to database...")
    try:
        conn = get_db_connection()
        conn.autocommit = False
        print("   Connected successfully")
    except Exception as e:
        print(f"   ERROR: Failed to connect: {e}")
        sys.exit(1)
    print()

    try:
        # Clear existing offers if requested
        if args.clear:
            print("3. Clearing existing offers...")
            clear_offers(conn, DEFAULT_COMPANY_ID, args.dry_run)
            print()
            step = 4
        else:
            step = 3

        # Insert new offers
        print(f"{step}. Processing offers...")
        skipped = insert_offers(conn, offers, args.dry_run)
        print()

        # Save skipped offers for manual review
        if skipped:
            print(f"{step + 1}. Saving skipped offers for review...")
            save_skipped_offers(skipped, "skipped_offers.csv")
            print()
            final_step = step + 2
        else:
            final_step = step + 1

        if not args.dry_run:
            conn.commit()
            print(f"{final_step}. Transaction committed successfully!")
        else:
            conn.rollback()
            print(f"{final_step}. Dry run - no changes made")

    except Exception as e:
        conn.rollback()
        print(f"ERROR: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
    finally:
        conn.close()

    print()
    print("=== Done! ===")


if __name__ == "__main__":
    main()
