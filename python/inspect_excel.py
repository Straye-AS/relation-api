#!/usr/bin/env python3
"""Inspect Excel file structure to understand columns and data."""

import pandas as pd
import sys

def inspect_excel(filepath: str):
    """Read and display Excel file structure."""
    df = pd.read_excel(filepath)

    print("=== COLUMNS ===")
    for i, col in enumerate(df.columns):
        print(f"{i}: {col}")

    print("\n=== FIRST 5 ROWS ===")
    print(df.head().to_string())

    print("\n=== DATA TYPES ===")
    print(df.dtypes)

    print(f"\n=== TOTAL ROWS: {len(df)} ===")

if __name__ == "__main__":
    filepath = sys.argv[1] if len(sys.argv) > 1 else "../erpdata/customers_20251210075553.xlsx"
    inspect_excel(filepath)
