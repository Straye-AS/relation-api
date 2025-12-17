#!/usr/bin/env python3
"""
Analyze customer name matches between Excel offers and database customers.

This script compares customer names from the Excel file with customers in the database
and generates a markdown report with three categories:
1. Exact matches (confident)
2. Probable matches (fuzzy match)
3. Needs attention (no good match found)
"""

import os
import re
import sys
from datetime import datetime

import pandas as pd
import psycopg2


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


def normalize_name(name: str) -> str:
    """Normalize customer name for comparison."""
    if not name:
        return ""
    name = name.lower().strip()
    # Remove extra whitespace
    name = re.sub(r"\s+", " ", name)
    # Remove common suffixes
    name = re.sub(r"\s+(as|a/s|ans|da|sa)$", "", name)
    return name


def similarity_score(name1: str, name2: str) -> float:
    """Calculate simple similarity score between two names."""
    n1 = normalize_name(name1)
    n2 = normalize_name(name2)

    if n1 == n2:
        return 1.0

    # Check if one contains the other
    if n1 in n2 or n2 in n1:
        return 0.9

    # Check word overlap
    words1 = set(n1.split())
    words2 = set(n2.split())

    if not words1 or not words2:
        return 0.0

    overlap = len(words1 & words2)
    total = len(words1 | words2)

    return overlap / total if total > 0 else 0.0


def get_excel_customers(filepath: str) -> dict:
    """Get unique customer names from Excel with offer count."""
    df = pd.read_excel(filepath, header=7)
    customer_counts = {}

    for _, row in df.iterrows():
        customer_name = clean_string(row.get("Kunde / Byggherre", ""))
        if customer_name:
            customer_counts[customer_name] = customer_counts.get(customer_name, 0) + 1

    return customer_counts


def get_db_customers(conn) -> list:
    """Get all customers from database."""
    with conn.cursor() as cur:
        cur.execute("SELECT id, name, org_number FROM customers ORDER BY name")
        return [(str(row[0]), row[1], row[2] or "") for row in cur.fetchall()]


def find_best_match(excel_name: str, db_customers: list) -> tuple:
    """Find best matching database customer for an Excel customer name."""
    best_match = None
    best_score = 0.0

    for customer_id, db_name, org_number in db_customers:
        score = similarity_score(excel_name, db_name)
        if score > best_score:
            best_score = score
            best_match = (customer_id, db_name, org_number)

    return best_match, best_score


def main():
    excel_file = "../erpdata/Plan Straye Tak last.xlsx"
    output_file = "../erpdata/customer_mapping_analysis.md"

    print("=== Customer Mapping Analysis ===")
    print()

    # Get Excel customers
    print("1. Reading Excel file...")
    excel_customers = get_excel_customers(excel_file)
    print(f"   Found {len(excel_customers)} unique customer names")
    print()

    # Get DB customers
    print("2. Connecting to database...")
    try:
        conn = get_db_connection()
        db_customers = get_db_customers(conn)
        print(f"   Found {len(db_customers)} customers in database")
    except Exception as e:
        print(f"   ERROR: Failed to connect: {e}")
        sys.exit(1)
    print()

    # Analyze matches
    print("3. Analyzing matches...")
    exact_matches = []      # Score >= 0.95
    probable_matches = []   # Score >= 0.6
    needs_attention = []    # Score < 0.6

    for excel_name, offer_count in sorted(excel_customers.items()):
        best_match, score = find_best_match(excel_name, db_customers)

        if best_match:
            db_id, db_name, org_number = best_match
            match_info = {
                "excel_name": excel_name,
                "db_name": db_name,
                "db_id": db_id,
                "org_number": org_number,
                "score": score,
                "offer_count": offer_count,
            }

            if score >= 0.95:
                exact_matches.append(match_info)
            elif score >= 0.6:
                probable_matches.append(match_info)
            else:
                needs_attention.append(match_info)
        else:
            needs_attention.append({
                "excel_name": excel_name,
                "db_name": "NO MATCH FOUND",
                "db_id": "",
                "org_number": "",
                "score": 0,
                "offer_count": offer_count,
            })

    print(f"   Exact matches: {len(exact_matches)}")
    print(f"   Probable matches: {len(probable_matches)}")
    print(f"   Needs attention: {len(needs_attention)}")
    print()

    # Generate markdown report
    print("4. Generating report...")

    with open(output_file, 'w') as f:
        f.write("# Customer Mapping Analysis\n\n")
        f.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
        f.write(f"- Excel customers: {len(excel_customers)}\n")
        f.write(f"- Database customers: {len(db_customers)}\n\n")
        f.write("---\n\n")

        # Table 1: Exact matches
        f.write("## 1. Exact Matches (Confident)\n\n")
        f.write("These customers match exactly or nearly exactly. No action needed.\n\n")
        f.write("| Excel Name | Database Name | Offers | Score |\n")
        f.write("|------------|---------------|--------|-------|\n")
        for m in sorted(exact_matches, key=lambda x: x['excel_name']):
            f.write(f"| {m['excel_name']} | {m['db_name']} | {m['offer_count']} | {m['score']:.0%} |\n")
        f.write(f"\n**Total: {len(exact_matches)} customers**\n\n")
        f.write("---\n\n")

        # Table 2: Probable matches
        f.write("## 2. Probable Matches (Please Verify)\n\n")
        f.write("These customers have a good match but may need verification.\n\n")
        f.write("| Excel Name | Database Name | Offers | Score | Action |\n")
        f.write("|------------|---------------|--------|-------|--------|\n")
        for m in sorted(probable_matches, key=lambda x: x['score'], reverse=True):
            f.write(f"| {m['excel_name']} | {m['db_name']} | {m['offer_count']} | {m['score']:.0%} | |\n")
        f.write(f"\n**Total: {len(probable_matches)} customers**\n\n")
        f.write("---\n\n")

        # Table 3: Needs attention
        f.write("## 3. Needs Attention (Manual Review Required)\n\n")
        f.write("These customers need manual mapping. Please specify the correct database customer.\n\n")
        f.write("| Excel Name | Best Match (if any) | Offers | Score | Correct Customer |\n")
        f.write("|------------|---------------------|--------|-------|------------------|\n")
        for m in sorted(needs_attention, key=lambda x: -x['offer_count']):
            f.write(f"| {m['excel_name']} | {m['db_name']} | {m['offer_count']} | {m['score']:.0%} | |\n")
        f.write(f"\n**Total: {len(needs_attention)} customers**\n\n")

        # Summary of needed aliases
        f.write("---\n\n")
        f.write("## Suggested Aliases to Add\n\n")
        f.write("Based on the analysis, add these aliases to `CUSTOMER_ALIASES` in `import_offers.py`:\n\n")
        f.write("```python\nCUSTOMER_ALIASES = {\n")
        for m in probable_matches:
            normalized_excel = normalize_name(m['excel_name'])
            normalized_db = normalize_name(m['db_name'])
            if normalized_excel != normalized_db:
                f.write(f'    "{normalized_excel}": "{normalized_db}",\n')
        f.write("}\n```\n")

    print(f"   Report saved to: {output_file}")
    print()
    print("=== Done! ===")

    conn.close()


if __name__ == "__main__":
    main()
