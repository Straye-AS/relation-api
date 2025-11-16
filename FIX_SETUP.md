# Fix Setup Issues

## 1. Fix go.sum (Required)

The `go.sum` file was malformed. Run this command in your terminal:

```bash
cd /Users/straye-as/Documents/straye/relation-api
go mod tidy
```

This will download all dependencies and regenerate a proper `go.sum` file.

## 2. Using migrate-create Command

After fixing `go.sum`, you can create migrations using:

### Method 1: Using Make (Recommended)
```bash
make migrate-create name=add_new_field
```

### Method 2: Direct Go Command
```bash
go run ./cmd/migrate create add_new_field
```

### Example: Create a migration to add a field
```bash
make migrate-create name=add_customer_vat_number
```

This will create two files in the `migrations/` directory:
- `YYYYMMDDHHMMSS_add_customer_vat_number.sql` (with up and down sections)

## 3. Database Name Alignment

**Note:** Your `docker-compose.yml` uses database name `relation` but `config.json` might use `relation_db`. 

Update `config.json` to match:

```json
{
  "database": {
    "name": "relation"
  }
}
```

## 4. Quick Test After Fixing

```bash
# 1. Verify go.sum is fixed
go mod verify

# 2. Try creating a test migration
make migrate-create name=test

# 3. Start database
docker-compose up -d postgres

# 4. Run existing migrations
make migrate-up

# 5. Check migration status
make migrate-status
```

## 5. If You Get Certificate Errors

If you see certificate errors (x509), try:

```bash
# Option A: Update certificates
go clean -modcache
go mod download

# Option B: Bypass proxy (not recommended for production)
export GOPROXY=direct
go mod tidy
```

## Common Issues & Solutions

### Issue: "malformed go.sum"
**Solution:** Delete `go.sum` and run `go mod tidy`

### Issue: "wrong number of fields"
**Solution:** The go.sum file had placeholder text. Run `go mod tidy` to regenerate.

### Issue: "make migrate-create doesn't work"
**Solution:** Must provide `name` parameter:
```bash
make migrate-create name=your_migration_name
```

### Issue: Database connection failed
**Solution:** Make sure database name matches in both `config.json` and `docker-compose.yml`

## PostgreSQL 16 ✅

Your docker-compose.yml is already using PostgreSQL 16:
```yaml
postgres:
  image: postgres:16-alpine
```

All set! 🎉

