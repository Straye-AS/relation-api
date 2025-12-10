#!/usr/bin/env python3
"""
Import customers from ERP Excel export into the Relation API database.

This script:
1. Reads customer data from an Excel file
2. Clears existing customers (and related data via CASCADE)
3. Imports the new customers

Usage:
    python import_customers.py [excel_file] [--dry-run]

Environment variables:
    DATABASE_HOST     (default: localhost)
    DATABASE_PORT     (default: 5432)
    DATABASE_USER     (default: relation_user)
    DATABASE_PASSWORD (default: relation_password)
    DATABASE_NAME     (default: relation)
"""

import argparse
import os
import sys
import uuid
from datetime import datetime, timezone

import pandas as pd
import psycopg2
from psycopg2.extras import execute_values


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


def clean_org_number(value) -> str:
    """Clean org number, converting float to int string."""
    if pd.isna(value) or value is None:
        return ""
    # Convert float like 977195500.0 to "977195500"
    try:
        return str(int(float(value)))
    except (ValueError, TypeError):
        return ""


def clean_postal_code(value) -> str:
    """Clean postal code, ensuring 4 digits with leading zeros."""
    if pd.isna(value) or value is None:
        return ""
    try:
        return str(int(float(value))).zfill(4)
    except (ValueError, TypeError):
        return ""


def clean_phone(mobil, telefon) -> str:
    """Get phone number, preferring mobile over landline."""
    mobil = clean_string(mobil)
    telefon = clean_string(telefon)
    # Prefer mobile, fallback to landline
    return mobil if mobil else telefon


def clean_credit_limit(value) -> float | None:
    """Clean credit limit, returning None if not set."""
    if pd.isna(value) or value is None:
        return None
    try:
        return float(value)
    except (ValueError, TypeError):
        return None


def determine_status(inaktiv: bool) -> str:
    """Convert Inaktiv boolean to status string."""
    return "inactive" if inaktiv else "active"


def read_excel_customers(filepath: str) -> list[dict]:
    """Read and transform Excel data to customer records."""
    df = pd.read_excel(filepath)
    customers = []
    seen_org_numbers = set()  # Track org numbers to skip duplicates

    for _, row in df.iterrows():
        org_number = clean_org_number(row.get("Org.nr."))

        # Skip duplicate org numbers (keep first occurrence)
        if org_number and org_number in seen_org_numbers:
            name = clean_string(row.get("Kundenavn", ""))
            print(f"  Skipping duplicate org number {org_number}: {name}")
            continue
        if org_number:
            seen_org_numbers.add(org_number)

        customer = {
            "id": str(uuid.uuid4()),
            "name": clean_string(row.get("Kundenavn", "")),
            "org_number": clean_org_number(row.get("Org.nr.")),
            "email": clean_string(row.get("Epost", "")),
            "phone": clean_phone(row.get("Mobil"), row.get("Telefon")),
            "address": clean_string(row.get("Adresse", "")),
            "city": clean_string(row.get("Poststed", "")),
            "postal_code": clean_postal_code(row.get("Postnr.")),
            "country": clean_string(row.get("Land", "Norway")) or "Norway",
            "contact_person": clean_string(row.get("Hovedkontakt", "")),
            "contact_email": "",  # Not in Excel
            "contact_phone": "",  # Not in Excel
            "status": determine_status(row.get("Inaktiv", False)),
            "tier": "bronze",  # Default
            "industry": "",  # Not in Excel
            # New extended fields
            "notes": "",  # Kundenotat is float (NaN), so skip for now
            "customer_class": clean_string(row.get("Kundeklasse", "")),
            "credit_limit": clean_credit_limit(row.get("Kredittgrense")),
            "is_internal": bool(row.get("Internkunde", False)),
            "municipality": clean_string(row.get("Kommune", "")),
            "county": clean_string(row.get("Fylke", "")),
            "company_id": None,  # Can be set later
            "created_at": datetime.now(timezone.utc),
            "updated_at": datetime.now(timezone.utc),
        }

        # Skip rows without a name
        if not customer["name"]:
            print(f"  Skipping row with empty name")
            continue

        customers.append(customer)

    return customers


def clear_customers(conn, dry_run: bool = False):
    """Clear all existing customers (cascades to related tables)."""
    with conn.cursor() as cur:
        # Check what will be deleted
        cur.execute("SELECT COUNT(*) FROM customers")
        count = cur.fetchone()[0]
        print(f"  Found {count} existing customers to delete")

        if not dry_run:
            # Delete all customers - related data (contacts, projects, etc.)
            # will be deleted via CASCADE constraints
            cur.execute("DELETE FROM customers")
            print(f"  Deleted {count} customers")


def insert_customers(conn, customers: list[dict], dry_run: bool = False):
    """Insert customers into the database."""
    if not customers:
        print("  No customers to insert")
        return

    columns = [
        "id", "name", "org_number", "email", "phone", "address", "city",
        "postal_code", "country", "contact_person", "contact_email",
        "contact_phone", "status", "tier", "industry",
        "notes", "customer_class", "credit_limit", "is_internal",
        "municipality", "county", "company_id",
        "created_at", "updated_at"
    ]

    values = [
        (
            c["id"], c["name"], c["org_number"] or None, c["email"], c["phone"],
            c["address"], c["city"], c["postal_code"], c["country"],
            c["contact_person"], c["contact_email"], c["contact_phone"],
            c["status"], c["tier"], c["industry"] or None,
            c["notes"] or None, c["customer_class"] or None, c["credit_limit"],
            c["is_internal"], c["municipality"] or None, c["county"] or None,
            c["company_id"],
            c["created_at"], c["updated_at"]
        )
        for c in customers
    ]

    if dry_run:
        print(f"  Would insert {len(customers)} customers")
        for c in customers[:5]:
            org = c['org_number'] or 'no org nr'
            internal = " [INTERNAL]" if c['is_internal'] else ""
            print(f"    - {c['name']} ({org}) {c['municipality']}, {c['county']}{internal}")
        if len(customers) > 5:
            print(f"    ... and {len(customers) - 5} more")
        return

    with conn.cursor() as cur:
        insert_sql = f"""
            INSERT INTO customers ({', '.join(columns)})
            VALUES %s
        """
        execute_values(cur, insert_sql, values)
        print(f"  Inserted {len(customers)} customers")


def main():
    parser = argparse.ArgumentParser(
        description="Import customers from Excel to database"
    )
    parser.add_argument(
        "excel_file",
        nargs="?",
        default="../erpdata/customers_20251210075553.xlsx",
        help="Path to Excel file"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be done without making changes"
    )
    args = parser.parse_args()

    print(f"=== Customer Import Script ===")
    print(f"Excel file: {args.excel_file}")
    print(f"Dry run: {args.dry_run}")
    print()

    # Read Excel data
    print("1. Reading Excel file...")
    customers = read_excel_customers(args.excel_file)
    print(f"   Found {len(customers)} valid customers")
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
        # Clear existing customers
        print("3. Clearing existing customers...")
        clear_customers(conn, args.dry_run)
        print()

        # Insert new customers
        print("4. Inserting new customers...")
        insert_customers(conn, customers, args.dry_run)
        print()

        if not args.dry_run:
            conn.commit()
            print("5. Transaction committed successfully!")
        else:
            conn.rollback()
            print("5. Dry run - no changes made")

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
