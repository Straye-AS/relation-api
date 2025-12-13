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
STATUS_MAPPING = {
    "Tilbud": ("sent", "active"),       # Offer sent, awaiting response
    "Ferdig": ("won", "won"),            # Won and completed
    "Ordre": ("won", "won"),             # Order received = won
    "Tapt": ("lost", "lost"),            # Lost
    "BUDSJETT": ("draft", "active"),     # Still budgeting/drafting
    "UTGÅR": ("expired", "expired"),     # Expired
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


def get_status_mapping(excel_status: str) -> tuple[str, str]:
    """Map Excel status to (phase, status) tuple."""
    excel_status = clean_string(excel_status)
    return STATUS_MAPPING.get(excel_status, ("draft", "active"))


def get_responsible_name(initials: str) -> str:
    """Map initials to full name."""
    initials = clean_string(initials).upper()
    return RESPONSIBLE_MAPPING.get(initials, initials)


# Customer name aliases - maps variations to canonical names
CUSTOMER_ALIASES = {
    # Straye variations
    "straye hybrid": "straye hybridbygg",
    "straye hybrid as": "straye hybridbygg",
    "strayeindustri": "straye industri",
    "strayeindustri as": "straye industri",
    "straye industribygg": "straye industri",
    "straye industribygg as": "straye industri",
    # Veidekke variations - all map to Veidekke Entreprenør
    "veidekke": "veidekke entreprenør",
    "veidekke as": "veidekke entreprenør",
    "veidekke bygg": "veidekke entreprenør",
    "veidekke bygg- vest": "veidekke entreprenør",
    "veidekke bygg vest": "veidekke entreprenør",
    "veidekke ålesund": "veidekke entreprenør",
    # PEAB variations
    "peab": "peab bygg",
    "peab as": "peab bygg",
    # Typos and variations
    "enter solutions": "enter solution",
    "enter solutions as": "enter solution",
    "workman": "workman norway",
    "workman as": "workman norway",
    "furuno": "furuno norge",
    "furuno as": "furuno norge",
    "vestre bærum tennis": "vestre bærum tennisklubb",
    "km bygg??": "km bygg",
    "matotalbygg": "ma totalbygg",
    "matotalbygg as": "ma totalbygg",
    # Joint/combined offers - assign to primary customer
    "seby as/ veidekke as": "veidekke entreprenør",
    "straye stålbygg as / hallmaker": "straye stålbygg",
    "hallmaker as/ straye stålbygg as": "straye stålbygg",
    "straye stålbygg as/ hallmaker": "straye stålbygg",
    "thermica as/ hallmaker": "hallmaker",
    "peab/ straye stålbygg": "straye stålbygg",
    "dpend/ straye stålbygg": "straye stålbygg",
    "flere- gikk til km bygg": "km bygg",
    # More fixes
    "geir nielsen (holmskau)": "holmskau prosjekt",
    "tatalbygg midt-norge": "totalbygg midt-norge",
    "tatalbygg midt-norge as": "totalbygg midt-norge",
    # Reassigned offers
    "flere": "km bygg",
    "ukjent kunde": "km bygg",
    "grinda 9 revidert": "jesper vogt lorentzen",
    "2581 hjelseth": "km bygg",
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


def read_excel_offers(filepath: str) -> list[dict]:
    """Read and transform Excel data to offer records."""
    df = pd.read_excel(filepath, header=7)
    offers = []

    for _, row in df.iterrows():
        project_title = clean_string(row.get("Prosjekt", ""))
        if not project_title:
            continue

        customer_name = clean_string(row.get("Kunde / Byggherre", ""))
        if not customer_name:
            customer_name = "Ukjent kunde"

        excel_status = clean_string(row.get("Status", ""))
        phase, status = get_status_mapping(excel_status)

        value = clean_float(row.get("Tilbudspris", 0))
        margin_amount = clean_float(row.get("DB", 0))
        cost = value - margin_amount if value > 0 else 0

        sent_date = clean_date(row.get("Sendt"))
        due_date = clean_date(row.get("Vedståelses frist"))
        last_updated = clean_date(row.get("Sist oppdatert"))

        location = clean_string(row.get("Beliggenhet", ""))
        notes = clean_string(row.get("Beskrivelse / siste nytt", ""))
        responsible_initials = clean_string(row.get("Ansvarlig", ""))
        responsible_name = get_responsible_name(responsible_initials)

        probability = 0
        if phase == "draft":
            probability = 10
        elif phase == "sent":
            probability = 30
        elif phase == "won":
            probability = 100

        offer = {
            "id": str(uuid.uuid4()),
            "title": project_title,
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

    return offers


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
    stats = {"draft": 0, "sent": 0, "won": 0, "lost": 0, "expired": 0}
    total_value = 0
    for offer in importable:
        stats[offer["phase"]] = stats.get(offer["phase"], 0) + 1
        total_value += offer["value"]

    print(f"  Can import: {len(importable)} offers")
    print(f"  Skipped (no customer match): {len(skipped)} offers")
    print()
    print(f"  Importable breakdown:")
    print(f"    - Draft: {stats.get('draft', 0)}")
    print(f"    - Sent: {stats.get('sent', 0)}")
    print(f"    - Won: {stats.get('won', 0)}")
    print(f"    - Lost: {stats.get('lost', 0)}")
    print(f"    - Expired: {stats.get('expired', 0)}")
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


def main():
    parser = argparse.ArgumentParser(
        description="Import offers from Excel to database"
    )
    parser.add_argument(
        "excel_file",
        nargs="?",
        default="../erpdata/Plan Straye Tak new.xlsx",
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
    args = parser.parse_args()

    print(f"=== Offer Import Script ===")
    print(f"Excel file: {args.excel_file}")
    print(f"Company: {DEFAULT_COMPANY_ID}")
    print(f"Dry run: {args.dry_run}")
    print(f"Clear existing: {args.clear}")
    print()

    # Read Excel data
    print("1. Reading Excel file...")
    offers = read_excel_offers(args.excel_file)
    print(f"   Found {len(offers)} valid offers in file")
    print()

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
