#!/usr/bin/env python3
"""Check how well offer customers match database customers."""

import pandas as pd
import psycopg2
import re

# Import customer aliases from import_offers
from import_offers import CUSTOMER_ALIASES


def normalize_name(name):
    """Normalize customer name for matching (uses aliases from import_offers)."""
    if not name:
        return ""
    name = name.lower().strip()
    # Remove extra whitespace
    name = re.sub(r"\s+", " ", name)
    # Check aliases BEFORE removing suffixes
    if name in CUSTOMER_ALIASES:
        name = CUSTOMER_ALIASES[name]
    # Remove common suffixes
    name = re.sub(r"\s+(as|a/s|ans|da|sa)$", "", name)
    # Check aliases AFTER removing suffixes
    if name in CUSTOMER_ALIASES:
        name = CUSTOMER_ALIASES[name]
    return name


def main():
    # Read offer customers from Excel
    df = pd.read_excel("../erpdata/Plan Straye Tak last.xlsx", header=7)
    offer_customers = df["Kunde / Byggherre"].dropna().str.strip().unique()

    # Get existing customers from database
    conn = psycopg2.connect(
        host="localhost",
        port="5432",
        user="relation_user",
        password="relation_password",
        dbname="relation",
    )
    cur = conn.cursor()
    cur.execute("SELECT name FROM customers")
    db_customers = [row[0] for row in cur.fetchall()]
    conn.close()

    # Create normalized lookup
    db_normalized = {normalize_name(c): c for c in db_customers}

    # Check matches
    matched = []
    unmatched = []
    for cust in offer_customers:
        norm = normalize_name(cust)
        if norm in db_normalized:
            matched.append((cust, db_normalized[norm]))
        else:
            unmatched.append(cust)

    print(f"=== MATCHED: {len(matched)} ===")
    variations = [(orig, db) for orig, db in matched if orig != db]
    if variations:
        print("  Variations matched:")
        for orig, db in sorted(variations)[:15]:
            print(f'    "{orig}" -> "{db}"')

    print()
    print(f"=== STILL UNMATCHED: {len(unmatched)} ===")
    for m in sorted(unmatched):
        print(f"  - {m}")


if __name__ == "__main__":
    main()
