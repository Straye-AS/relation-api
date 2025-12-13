#!/usr/bin/env python3
"""
Add missing customers found from Brønnøysundregistrene.
"""

import os
import uuid
from datetime import datetime, timezone

import psycopg2

# Customers found from Brønnøysund searches
CUSTOMERS_TO_ADD = [
    # (name, org_number, city, county)
    ("Hallmaker AS", "937920040", "Lysaker", "AKERSHUS"),
    ("PEAB Bygg AS", "943672520", "Lysaker", "AKERSHUS"),
    ("Veidekke Entreprenør AS", "984024290", "Oslo", "OSLO"),
    ("Betonmast Romerike AS", "993925314", "Lillestrøm", "AKERSHUS"),
    ("Straye Hybridbygg AS", "922249733", "Fredrikstad", "ØSTFOLD"),
    ("Straye Trebygg AS", "822249752", "Fredrikstad", "ØSTFOLD"),
    ("Park & Anlegg AS", "982382955", "Grålum", "ØSTFOLD"),
    ("Betonmast Trøndelag AS", "859739822", "Trondheim", "TRØNDELAG"),
    ("Seltor AS", "915617344", "Porsgrunn", "TELEMARK"),
    ("Hersleth Entreprenør AS", "964602360", "Hølen", "AKERSHUS"),
    ("Gressvik Properties AS", "925208469", "Fredrikstad", "ØSTFOLD"),
    ("Workman Norway AS", "927371138", "Slemmestad", "AKERSHUS"),
    ("Thermica AS", "997933273", "Lierstranda", "BUSKERUD"),
    # Additional companies we need (without org numbers for now)
    ("Enter Solution AS", "", "Sarpsborg", "ØSTFOLD"),
    ("Nordbygg AS", "", "Sofiemyr", "AKERSHUS"),
    ("A Bygg AS", "", "", ""),
    ("Arealbygg AS", "", "", ""),
    ("Betongbygg AS", "", "", ""),
    ("Bomekan AS", "", "", ""),
    ("Byggekompaniet Østfold AS", "", "", "ØSTFOLD"),
    ("Byggkompaniet AS", "", "", ""),
    ("Dan Blikk AS", "", "", ""),
    ("Dybvig AS", "", "", ""),
    ("Furuno Norge AS", "", "", ""),
    ("Fusen AS", "", "", ""),
    ("Høstbakken 11 AS", "", "Halden", "ØSTFOLD"),
    ("KM Bygg AS", "", "", ""),
    ("Konsmo Fabrikker AS", "", "", ""),
    ("MA Totalbygg AS", "", "", ""),
    ("Newsec AS", "", "", ""),
    ("Ocab AS", "", "", ""),
    ("Parketteksperten AS", "", "", ""),
    ("Sameie Hoffsveien 88", "", "Oslo", "OSLO"),
    ("Sameiet Kornmoenga", "", "", ""),
    ("Totalbetong AS", "", "", ""),
    ("Vestre Bærum Tennisklubb", "", "Bærum", "AKERSHUS"),
]


def get_db_connection():
    return psycopg2.connect(
        host=os.getenv("DATABASE_HOST", "localhost"),
        port=os.getenv("DATABASE_PORT", "5432"),
        user=os.getenv("DATABASE_USER", "relation_user"),
        password=os.getenv("DATABASE_PASSWORD", "relation_password"),
        dbname=os.getenv("DATABASE_NAME", "relation"),
    )


def normalize_name(name):
    """Normalize for matching."""
    import re
    if not name:
        return ""
    name = name.lower().strip()
    name = re.sub(r"\s+(as|a/s|ans|da|sa)$", "", name)
    name = re.sub(r"\s+", " ", name)
    return name


def main():
    conn = get_db_connection()
    conn.autocommit = False

    # Load existing customers
    cur = conn.cursor()
    cur.execute("SELECT LOWER(name), org_number FROM customers")
    existing = {}
    for row in cur.fetchall():
        existing[normalize_name(row[0])] = row[1]

    added = 0
    skipped = 0
    now = datetime.now(timezone.utc)

    for name, org_number, city, county in CUSTOMERS_TO_ADD:
        normalized = normalize_name(name)

        # Check if exists
        if normalized in existing:
            print(f"  SKIP (exists): {name}")
            skipped += 1
            continue

        # Check org number exists
        if org_number and org_number in [v for v in existing.values() if v]:
            print(f"  SKIP (org# exists): {name} ({org_number})")
            skipped += 1
            continue

        # Insert
        customer_id = str(uuid.uuid4())
        cur.execute(
            """INSERT INTO customers
               (id, name, org_number, email, phone, city, county, country, status, tier, created_at, updated_at)
               VALUES (%s, %s, %s, '', '', %s, %s, 'Norway', 'active', 'bronze', %s, %s)""",
            (customer_id, name, org_number or None, city or None, county or None, now, now)
        )
        print(f"  ADD: {name} ({org_number or 'no org#'})")
        existing[normalized] = org_number
        added += 1

    conn.commit()
    conn.close()

    print()
    print(f"=== Done: Added {added}, Skipped {skipped} ===")


if __name__ == "__main__":
    main()
