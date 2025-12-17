#!/usr/bin/env python3
"""
Import users from Microsoft 365 (Azure AD) into the Relation API database.

Requirements:
    pip install msal requests psycopg2-binary

Azure AD Setup:
    1. Go to Azure Portal > App Registrations
    2. Create new registration or use existing app
    3. Add API permission: Microsoft Graph > Application > User.Read.All
    4. Grant admin consent
    5. Create a client secret under Certificates & Secrets

Environment variables:
    AZURE_TENANT_ID     - Your Azure AD tenant ID
    AZURE_CLIENT_ID     - App registration client ID
    AZURE_CLIENT_SECRET - App registration client secret
    DATABASE_HOST       - Database host (default: localhost)
    DATABASE_PORT       - Database port (default: 5432)
    DATABASE_USER       - Database user (default: relation_user)
    DATABASE_PASSWORD   - Database password (default: relation_password)
    DATABASE_NAME       - Database name (default: relation)

Usage:
    python import_users_from_ms365.py [--dry-run] [--output-sql FILE]
"""

import argparse
import os
import sys
from datetime import datetime
from pathlib import Path

import requests

# Load .env file from project root
env_file = Path(__file__).parent.parent / ".env"
if env_file.exists():
    with open(env_file) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith("#") and "=" in line:
                key, _, value = line.partition("=")
                os.environ.setdefault(key.strip(), value.strip())

try:
    from msal import ConfidentialClientApplication
except ImportError:
    print("ERROR: msal not installed. Run: pip install msal")
    sys.exit(1)

try:
    import psycopg2
except ImportError:
    print("ERROR: psycopg2 not installed. Run: pip install psycopg2-binary")
    sys.exit(1)


# Azure AD configuration
TENANT_ID = os.getenv("AZURE_TENANT_ID")
CLIENT_ID = os.getenv("AZURE_CLIENT_ID")
CLIENT_SECRET = os.getenv("AZURE_CLIENT_SECRET")

# MS Graph API
GRAPH_URL = "https://graph.microsoft.com/v1.0"
SCOPE = ["https://graph.microsoft.com/.default"]


def get_db_connection():
    """Create database connection from environment variables."""
    return psycopg2.connect(
        host=os.getenv("DATABASE_HOST", "localhost"),
        port=os.getenv("DATABASE_PORT", "5432"),
        user=os.getenv("DATABASE_USER", "relation_user"),
        password=os.getenv("DATABASE_PASSWORD", "relation_password"),
        dbname=os.getenv("DATABASE_NAME", "relation"),
    )


def get_ms_graph_token():
    """Get access token for MS Graph API using client credentials."""
    if not all([TENANT_ID, CLIENT_ID, CLIENT_SECRET]):
        print("ERROR: Missing Azure AD credentials.")
        print("Set environment variables: AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET")
        sys.exit(1)

    authority = f"https://login.microsoftonline.com/{TENANT_ID}"
    app = ConfidentialClientApplication(
        CLIENT_ID,
        authority=authority,
        client_credential=CLIENT_SECRET,
    )

    result = app.acquire_token_for_client(scopes=SCOPE)

    if "access_token" not in result:
        print(f"ERROR: Failed to get token: {result.get('error_description', 'Unknown error')}")
        sys.exit(1)

    return result["access_token"]


def fetch_users_from_ms365(token: str) -> list[dict]:
    """Fetch all users from MS365 using Graph API."""
    headers = {"Authorization": f"Bearer {token}"}
    users = []

    # Select specific fields to reduce payload
    url = f"{GRAPH_URL}/users?$select=id,displayName,mail,givenName,surname,department,jobTitle,accountEnabled"

    while url:
        response = requests.get(url, headers=headers)
        if response.status_code != 200:
            print(f"ERROR: Graph API request failed: {response.status_code}")
            print(response.text)
            sys.exit(1)

        data = response.json()
        users.extend(data.get("value", []))

        # Handle pagination
        url = data.get("@odata.nextLink")

    return users


def transform_user(ms_user: dict, require_department: bool = True) -> dict:
    """Transform MS365 user to our database format."""
    email = ms_user.get("mail") or ms_user.get("userPrincipalName", "")

    # Skip users without email (service accounts, etc.)
    if not email or "@" not in email:
        return None

    # Skip external users (guests)
    if "#EXT#" in email:
        return None

    department = ms_user.get("department") or ""

    # Skip users without department if required
    if require_department and not department:
        return None

    return {
        "id": ms_user["id"],  # Use Azure AD OID as ID
        "azure_ad_oid": ms_user["id"],
        "name": ms_user.get("displayName") or "",
        "email": email.lower(),
        "first_name": ms_user.get("givenName") or "",
        "last_name": ms_user.get("surname") or "",
        "department": department,
        "is_active": ms_user.get("accountEnabled", True),
        "roles": ["user"],  # Default role
    }


def determine_company_id(user: dict) -> str | None:
    """Determine company_id based on email domain or department."""
    email = user.get("email", "").lower()
    dept = (user.get("department") or "").lower()

    # Map departments to companies
    dept_mapping = {
        "tak": "tak",
        "stålbygg": "stalbygg",
        "hybridbygg": "hybridbygg",
        "industri": "industri",
        "montasje": "montasje",
        "straye gruppen": "gruppen",
        "gruppen": "gruppen",
    }

    for key, company in dept_mapping.items():
        if key in dept:
            return company

    # Default to gruppen for unmatched departments (they still work at Straye)
    return "gruppen"


def insert_users(conn, users: list[dict], dry_run: bool = False):
    """Insert or update users in database."""
    if not users:
        print("  No users to insert")
        return

    inserted = 0
    updated = 0

    with conn.cursor() as cur:
        for user in users:
            company_id = determine_company_id(user)

            if dry_run:
                print(f"  Would upsert: {user['name']} <{user['email']}> (company: {company_id})")
                continue

            cur.execute("""
                INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (azure_ad_oid) DO UPDATE SET
                    name = EXCLUDED.name,
                    email = EXCLUDED.email,
                    first_name = EXCLUDED.first_name,
                    last_name = EXCLUDED.last_name,
                    department = EXCLUDED.department,
                    is_active = EXCLUDED.is_active,
                    company_id = COALESCE(EXCLUDED.company_id, users.company_id),
                    updated_at = CURRENT_TIMESTAMP
                RETURNING (xmax = 0) as inserted
            """, (
                user["id"],
                user["azure_ad_oid"],
                user["name"],
                user["email"],
                user["first_name"],
                user["last_name"],
                user["department"],
                user["is_active"],
                user["roles"],
                company_id,
            ))

            result = cur.fetchone()
            if result and result[0]:
                inserted += 1
            else:
                updated += 1

    if not dry_run:
        print(f"  Inserted: {inserted} new users")
        print(f"  Updated: {updated} existing users")


def generate_sql_file(users: list[dict], output_file: str):
    """Generate SQL file for manual import."""
    with open(output_file, 'w') as f:
        f.write(f"-- Users imported from MS365\n")
        f.write(f"-- Generated: {datetime.now()}\n")
        f.write(f"-- Total users: {len(users)}\n\n")

        for user in users:
            company_id = determine_company_id(user)
            company_sql = f"'{company_id}'" if company_id else "NULL"
            roles_sql = "ARRAY[" + ", ".join(f"'{r}'" for r in user["roles"]) + "]"

            f.write(f"""INSERT INTO users (id, azure_ad_oid, name, email, first_name, last_name, department, is_active, roles, company_id)
VALUES ('{user["id"]}', '{user["azure_ad_oid"]}', '{user["name"].replace("'", "''")}', '{user["email"]}', '{user.get("first_name", "").replace("'", "''")}', '{user.get("last_name", "").replace("'", "''")}', '{user.get("department", "").replace("'", "''")}', {str(user["is_active"]).lower()}, {roles_sql}, {company_sql})
ON CONFLICT (azure_ad_oid) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, department = EXCLUDED.department, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP;\n\n""")

    print(f"  Generated SQL file: {output_file}")
    print(f"  Total users: {len(users)}")


def main():
    parser = argparse.ArgumentParser(description="Import users from MS365")
    parser.add_argument("--dry-run", action="store_true", help="Show what would be done")
    parser.add_argument("--output-sql", type=str, help="Generate SQL file instead of inserting")
    args = parser.parse_args()

    print("=== MS365 User Import ===")
    print()

    # Get MS Graph token
    print("1. Authenticating with MS Graph API...")
    token = get_ms_graph_token()
    print("   Authenticated successfully")
    print()

    # Fetch users
    print("2. Fetching users from MS365...")
    ms_users = fetch_users_from_ms365(token)
    print(f"   Found {len(ms_users)} users in Azure AD")
    print()

    # Transform users
    print("3. Transforming users...")
    users = []
    skipped = 0
    for ms_user in ms_users:
        user = transform_user(ms_user)
        if user:
            users.append(user)
        else:
            skipped += 1
    print(f"   Valid users: {len(users)}")
    print(f"   Skipped (no email/external): {skipped}")
    print()

    # Output
    if args.output_sql:
        print("4. Generating SQL file...")
        generate_sql_file(users, args.output_sql)
    else:
        print("4. Importing to database...")
        conn = get_db_connection()
        conn.autocommit = False
        try:
            insert_users(conn, users, args.dry_run)
            if not args.dry_run:
                conn.commit()
                print("   Committed successfully")
            else:
                conn.rollback()
                print("   Dry run - no changes made")
        except Exception as e:
            conn.rollback()
            print(f"ERROR: {e}")
            sys.exit(1)
        finally:
            conn.close()

    print()
    print("=== Done! ===")


if __name__ == "__main__":
    main()
