# Quick Start - Straye Relation API (Norwegian Spec)

## ✅ What's Ready

The API has been **partially migrated** to the Norwegian Straye Relation Portal specification:

### Completed:
- ✅ Domain models with new Norwegian fields
- ✅ DTOs matching TypeScript interfaces  
- ✅ Database migrations for new schema
- ✅ Company service (static Straye companies)
- ✅ Updated mappers with calculations
- ✅ Configuration system
- ✅ Authentication (JWT + API Key)
- ✅ Docker setup

### Needs Completion:
- ⚠️ Repositories (field mapping updates)
- ⚠️ Services (business logic updates)
- ⚠️ HTTP Handlers (endpoint updates)
- ⚠️ Router (new routes)
- ⚠️ Main app (wiring)

## 🚀 Next Steps

### Option 1: Let Me Finish (Recommended)

**Say:** "Continue and complete the migration"

I'll update all remaining files (~2-3 hours of work) to get the API fully functional.

### Option 2: Manual Completion

Follow `MIGRATION_TO_NORWEGIAN_SPEC.md` for step-by-step instructions.

### Option 3: Incremental

Start with what's working and gradually add features.

## 📝 Key Changes Made

### Customer
```go
// OLD
type Customer struct {
    Name     string
    Industry string
    Website  string
}

// NEW - Norwegian spec
type Customer struct {
    Name          string
    OrgNumber     string  // Norwegian org number
    Email         string  // Now required
    Phone         string  // Now required
    Address       string
    City          string
    PostalCode    string
    Country       string
    ContactPerson string
    ContactEmail  string
    ContactPhone  string
}
```

### Offer
```go
// OLD
type Offer struct {
    Title       string
    TotalAmount float64
    Phase       OfferPhase // "Draft", "Sent", etc.
}

// NEW - Norwegian spec
type Offer struct {
    Title               string
    CompanyID           CompanyID    // stalbygg, hybridbygg, etc.
    Phase               OfferPhase   // "draft", "in_progress", "sent", "won", "lost", "expired"
    Probability         int          // 0-100
    Value               float64      // Auto-calculated from items
    Status              OfferStatus  // "active", "inactive", "archived"
    ResponsibleUserID   string
    ResponsibleUserName string
}
```

### OfferItem
```go
// OLD
type OfferItem struct {
    Name      string
    Quantity  float64
    UnitPrice float64
}

// NEW - Norwegian spec  
type OfferItem struct {
    Discipline  string  // "Yttervegg", "Tak", etc.
    Cost        float64
    Revenue     float64
    Margin      float64 // Calculated: ((revenue - cost) / revenue) * 100
    Quantity    float64
    Unit        string  // "m²", "stk", etc.
}
```

### Project
```go
// NEW fields
type Project struct {
    CompanyID       CompanyID
    TeamMembers     []string  // Array of user IDs
    TeamsChannelID  string
    TeamsChannelURL string
    ManagerID       string
    ManagerName     string
    OfferID         *uuid.UUID // Link back to original offer
}
```

## 🏢 Straye Companies (Static Data)

```json
[
  {"id": "all", "name": "Straye Gruppen", "color": "#1e40af"},
  {"id": "stalbygg", "name": "Straye Stålbygg", "color": "#dc2626"},
  {"id": "hybridbygg", "name": "Straye Hybridbygg", "color": "#16a34a"},
  {"id": "industri", "name": "Straye Industri", "color": "#9333ea"},
  {"id": "tak", "name": "Straye Tak", "color": "#ea580c"},
  {"id": "montasje", "name": "Straye Montasje", "color": "#0891b2"}
]
```

## 📊 New Calculations

### Margin
```go
margin = ((revenue - cost) / revenue) * 100
```

### Offer Value (auto-calculated)
```go
value = SUM(items[].revenue)
```

### Weighted Pipeline Value
```go
weightedValue = value * (probability / 100)
```

### Win Rate
```go
winRate = (wonOffers / (wonOffers + lostOffers)) * 100
```

## 🌐 API Endpoints (When Complete)

### New
- `GET /companies` - List Straye companies
- `GET /search?q=<query>` - Global search
- `GET /search/recent` - Recent items
- `GET /contacts` - Top-level (was nested under customers)

### Updated
- `GET /offers?companyId=<id>&probability=<min>-<max>` - New filters
- `GET /projects?companyId=<id>` - Company filter
- `GET /dashboard?companyId=<id>` - Enhanced metrics
- `POST /offers` - Value auto-calculated from items

### Removed
- `/customers/{id}/contacts` → Now `/contacts?customerId=<id>`

## 🗄️ Database

```bash
# Fresh start (required - new schema)
docker-compose down -v
docker-compose up -d postgres
sleep 5

# Run migrations
go run ./cmd/migrate up

# Check status
go run ./cmd/migrate status
```

## 🧪 Test When Ready

```bash
# Start API
go run ./cmd/api

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/companies
curl -H "x-api-key: dev-admin-key-change-in-production" \
  http://localhost:8080/customers
```

## 📚 Documentation Files

- `README_CURRENT_STATE.md` - Current status
- `MIGRATION_TO_NORWEGIAN_SPEC.md` - Full migration guide
- `CONTRIBUTING.md` - Development guidelines

## ⚡ Speed Up Completion

If you want this done quickly, tell me to continue and I'll:

1. Update all repositories (30 min)
2. Update all services (45 min)
3. Update all handlers (30 min)
4. Update router & main (15 min)
5. Test & verify (30 min)

**Total: ~2.5 hours of focused work**

---

**Current Status:** 🟡 Foundation complete, needs wiring

**Next Action:** Say "continue" or follow manual steps in MIGRATION_TO_NORWEGIAN_SPEC.md

