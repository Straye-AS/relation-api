package domain

import (
	"time"

	"github.com/google/uuid"
)

// DTOs for API responses matching Norwegian spec

type CustomerDTO struct {
	ID            uuid.UUID        `json:"id"`
	Name          string           `json:"name"`
	OrgNumber     string           `json:"orgNumber"`
	Email         string           `json:"email"`
	Phone         string           `json:"phone"`
	Address       string           `json:"address,omitempty"`
	City          string           `json:"city,omitempty"`
	PostalCode    string           `json:"postalCode,omitempty"`
	Country       string           `json:"country"`
	ContactPerson string           `json:"contactPerson,omitempty"`
	ContactEmail  string           `json:"contactEmail,omitempty"`
	ContactPhone  string           `json:"contactPhone,omitempty"`
	Status        CustomerStatus   `json:"status"`
	Tier          CustomerTier     `json:"tier"`
	Industry      CustomerIndustry `json:"industry,omitempty"`
	Notes         string           `json:"notes,omitempty"`
	CustomerClass string           `json:"customerClass,omitempty"`
	CreditLimit   *float64         `json:"creditLimit,omitempty"`
	IsInternal    bool             `json:"isInternal"`
	Municipality  string           `json:"municipality,omitempty"`
	County        string           `json:"county,omitempty"`
	CreatedAt     string           `json:"createdAt"` // ISO 8601
	UpdatedAt     string           `json:"updatedAt"` // ISO 8601
	TotalValue    float64          `json:"totalValue,omitempty"`
	ActiveOffers  int              `json:"activeOffers,omitempty"`
}

// CustomerWithDetailsDTO includes customer data with related entities and statistics
type CustomerWithDetailsDTO struct {
	CustomerDTO
	Stats          *CustomerStatsDTO `json:"stats,omitempty"`
	Contacts       []ContactDTO      `json:"contacts,omitempty"`
	ActiveDeals    []DealDTO         `json:"activeDeals,omitempty"`
	ActiveProjects []ProjectDTO      `json:"activeProjects,omitempty"`
}

// CustomerStatsDTO holds aggregated statistics for a customer
type CustomerStatsDTO struct {
	TotalValue     float64 `json:"totalValue"`
	ActiveOffers   int     `json:"activeOffers"`
	ActiveDeals    int     `json:"activeDeals"`
	ActiveProjects int     `json:"activeProjects"`
	TotalContacts  int     `json:"totalContacts"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type ContactDTO struct {
	ID                     uuid.UUID                `json:"id"`
	FirstName              string                   `json:"firstName"`
	LastName               string                   `json:"lastName"`
	FullName               string                   `json:"fullName"`
	Email                  string                   `json:"email,omitempty"`
	Phone                  string                   `json:"phone,omitempty"`
	Mobile                 string                   `json:"mobile,omitempty"`
	Title                  string                   `json:"title,omitempty"`
	Department             string                   `json:"department,omitempty"`
	ContactType            ContactType              `json:"contactType"`
	PrimaryCustomerID      *uuid.UUID               `json:"primaryCustomerId,omitempty"`
	PrimaryCustomerName    string                   `json:"primaryCustomerName,omitempty"`
	Address                string                   `json:"address,omitempty"`
	City                   string                   `json:"city,omitempty"`
	PostalCode             string                   `json:"postalCode,omitempty"`
	Country                string                   `json:"country,omitempty"`
	LinkedInURL            string                   `json:"linkedInUrl,omitempty"`
	PreferredContactMethod string                   `json:"preferredContactMethod,omitempty"`
	Notes                  string                   `json:"notes,omitempty"`
	IsActive               bool                     `json:"isActive"`
	Relationships          []ContactRelationshipDTO `json:"relationships,omitempty"`
	CreatedAt              string                   `json:"createdAt"` // ISO 8601
	UpdatedAt              string                   `json:"updatedAt"` // ISO 8601
}

type ContactRelationshipDTO struct {
	ID         uuid.UUID         `json:"id"`
	ContactID  uuid.UUID         `json:"contactId"`
	EntityType ContactEntityType `json:"entityType"`
	EntityID   uuid.UUID         `json:"entityId"`
	Role       string            `json:"role,omitempty"`
	IsPrimary  bool              `json:"isPrimary"`
	CreatedAt  string            `json:"createdAt"`
}

type DealDTO struct {
	ID                 uuid.UUID           `json:"id"`
	Title              string              `json:"title"`
	Description        string              `json:"description,omitempty"`
	CustomerID         uuid.UUID           `json:"customerId"`
	CustomerName       string              `json:"customerName,omitempty"`
	CompanyID          CompanyID           `json:"companyId"`
	Stage              DealStage           `json:"stage"`
	Probability        int                 `json:"probability"`
	Value              float64             `json:"value"`
	WeightedValue      float64             `json:"weightedValue"`
	Currency           string              `json:"currency"`
	ExpectedCloseDate  *string             `json:"expectedCloseDate,omitempty"`
	ActualCloseDate    *string             `json:"actualCloseDate,omitempty"`
	OwnerID            string              `json:"ownerId"`
	OwnerName          string              `json:"ownerName,omitempty"`
	Source             string              `json:"source,omitempty"`
	Notes              string              `json:"notes,omitempty"`
	LostReason         string              `json:"lostReason,omitempty"`
	LossReasonCategory *LossReasonCategory `json:"lossReasonCategory,omitempty"`
	OfferID            *uuid.UUID          `json:"offerId,omitempty"`
	CreatedAt          string              `json:"createdAt"`
	UpdatedAt          string              `json:"updatedAt"`
}

type DealStageHistoryDTO struct {
	ID            uuid.UUID  `json:"id"`
	DealID        uuid.UUID  `json:"dealId"`
	FromStage     *DealStage `json:"fromStage,omitempty"`
	ToStage       DealStage  `json:"toStage"`
	ChangedByID   string     `json:"changedById"`
	ChangedByName string     `json:"changedByName,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	ChangedAt     string     `json:"changedAt"`
}

type OfferDTO struct {
	ID                    uuid.UUID      `json:"id"`
	Title                 string         `json:"title"`
	OfferNumber           string         `json:"offerNumber,omitempty"`       // Internal number, e.g., "TK-2025-001"
	ExternalReference     string         `json:"externalReference,omitempty"` // External/customer reference number
	CustomerID            uuid.UUID      `json:"customerId"`
	CustomerName          string         `json:"customerName,omitempty"`
	ProjectID             *uuid.UUID     `json:"projectId,omitempty"` // Link to project (nullable)
	ProjectName           string         `json:"projectName,omitempty"`
	CompanyID             CompanyID      `json:"companyId"`
	Phase                 OfferPhase     `json:"phase"`
	Probability           int            `json:"probability"`
	Value                 float64        `json:"value"`
	Status                OfferStatus    `json:"status"`
	CreatedAt             string         `json:"createdAt"` // ISO 8601
	UpdatedAt             string         `json:"updatedAt"` // ISO 8601
	ResponsibleUserID     string         `json:"responsibleUserId,omitempty"`
	ResponsibleUserName   string         `json:"responsibleUserName,omitempty"`
	Items                 []OfferItemDTO `json:"items"`
	Description           string         `json:"description,omitempty"`
	Notes                 string         `json:"notes,omitempty"`
	DueDate               *string        `json:"dueDate,omitempty"` // ISO 8601
	Cost                  float64        `json:"cost"`              // Internal cost
	Margin                float64        `json:"margin"`            // Calculated: Value - Cost
	MarginPercent         float64        `json:"marginPercent"`     // Dekningsgrad: (Value - Cost) / Value * 100
	Location              string         `json:"location,omitempty"`
	SentDate              *string        `json:"sentDate,omitempty"`       // ISO 8601
	ExpirationDate        *string        `json:"expirationDate,omitempty"` // ISO 8601 - When offer expires (default 60 days after sent)
	CustomerHasWonProject bool           `json:"customerHasWonProject"`    // Whether customer has won their project
}

type OfferItemDTO struct {
	ID          uuid.UUID `json:"id"`
	Discipline  string    `json:"discipline"`
	Cost        float64   `json:"cost"`
	Revenue     float64   `json:"revenue"`
	Margin      float64   `json:"margin"`
	Description string    `json:"description,omitempty"`
	Quantity    float64   `json:"quantity,omitempty"`
	Unit        string    `json:"unit,omitempty"`
}

// Budget Item DTOs

type BudgetItemDTO struct {
	ID              uuid.UUID        `json:"id"`
	ParentType      BudgetParentType `json:"parentType"`
	ParentID        uuid.UUID        `json:"parentId"`
	Name            string           `json:"name"`
	ExpectedCost    float64          `json:"expectedCost"`
	ExpectedMargin  float64          `json:"expectedMargin"`
	ExpectedRevenue float64          `json:"expectedRevenue"`
	ExpectedProfit  float64          `json:"expectedProfit"`
	Quantity        *float64         `json:"quantity,omitempty"`
	PricePerItem    *float64         `json:"pricePerItem,omitempty"`
	Description     string           `json:"description,omitempty"`
	DisplayOrder    int              `json:"displayOrder"`
	CreatedAt       string           `json:"createdAt"`
	UpdatedAt       string           `json:"updatedAt"`
}

type BudgetSummaryDTO struct {
	ParentType    BudgetParentType `json:"parentType,omitempty"`
	ParentID      uuid.UUID        `json:"parentId,omitempty"`
	TotalCost     float64          `json:"totalCost"`
	TotalRevenue  float64          `json:"totalRevenue"`
	TotalProfit   float64          `json:"totalProfit"`
	MarginPercent float64          `json:"marginPercent"`
	ItemCount     int              `json:"itemCount"`
}

type ProjectDTO struct {
	ID                      uuid.UUID      `json:"id"`
	Name                    string         `json:"name"`
	ProjectNumber           string         `json:"projectNumber,omitempty"`
	Summary                 string         `json:"summary,omitempty"`
	Description             string         `json:"description,omitempty"`
	CustomerID              uuid.UUID      `json:"customerId"`
	CustomerName            string         `json:"customerName,omitempty"`
	CompanyID               CompanyID      `json:"companyId"`
	Status                  ProjectStatus  `json:"status"`
	Phase                   ProjectPhase   `json:"phase"`
	StartDate               string         `json:"startDate,omitempty"` // ISO 8601
	EndDate                 string         `json:"endDate,omitempty"`   // ISO 8601
	Value                   float64        `json:"value"`
	Cost                    float64        `json:"cost"`
	MarginPercent           float64        `json:"marginPercent"`
	Spent                   float64        `json:"spent"`
	ManagerID               *string        `json:"managerId,omitempty"`
	ManagerName             string         `json:"managerName,omitempty"`
	TeamMembers             []string       `json:"teamMembers,omitempty"`
	CreatedAt               string         `json:"createdAt"` // ISO 8601
	UpdatedAt               string         `json:"updatedAt"` // ISO 8601
	OfferID                 *uuid.UUID     `json:"offerId,omitempty"`
	DealID                  *uuid.UUID     `json:"dealId,omitempty"`
	HasDetailedBudget       bool           `json:"hasDetailedBudget"`
	Health                  *ProjectHealth `json:"health,omitempty"`
	CompletionPercent       *float64       `json:"completionPercent,omitempty"`
	EstimatedCompletionDate string         `json:"estimatedCompletionDate,omitempty"`
	// Phase-related fields for offer folder functionality
	WinningOfferID       *uuid.UUID `json:"winningOfferId,omitempty"`
	InheritedOfferNumber string     `json:"inheritedOfferNumber,omitempty"`
	CalculatedOfferValue float64    `json:"calculatedOfferValue"`
	WonAt                string     `json:"wonAt,omitempty"` // ISO 8601
	IsEconomicsEditable  bool       `json:"isEconomicsEditable"`
}

// Project Actual Cost DTOs

type ProjectActualCostDTO struct {
	ID               uuid.UUID  `json:"id"`
	ProjectID        uuid.UUID  `json:"projectId"`
	CostType         CostType   `json:"costType"`
	Description      string     `json:"description"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency"`
	CostDate         string     `json:"costDate"`
	PostingDate      string     `json:"postingDate,omitempty"`
	BudgetItemID     *uuid.UUID `json:"budgetItemId,omitempty"`
	ERPSource        ERPSource  `json:"erpSource"`
	ERPReference     string     `json:"erpReference,omitempty"`
	ERPTransactionID string     `json:"erpTransactionId,omitempty"`
	ERPSyncedAt      string     `json:"erpSyncedAt,omitempty"`
	IsApproved       bool       `json:"isApproved"`
	ApprovedByID     string     `json:"approvedById,omitempty"`
	ApprovedAt       string     `json:"approvedAt,omitempty"`
	Notes            string     `json:"notes,omitempty"`
	CreatedAt        string     `json:"createdAt"`
	UpdatedAt        string     `json:"updatedAt"`
}

type ProjectCostSummaryDTO struct {
	ProjectID        uuid.UUID `json:"projectId"`
	ProjectName      string    `json:"projectName"`
	Value            float64   `json:"value"`
	Cost             float64   `json:"cost"`
	MarginPercent    float64   `json:"marginPercent"`
	Spent            float64   `json:"spent"`
	ActualCosts      float64   `json:"actualCosts"`
	RemainingValue   float64   `json:"remainingValue"`
	ValueUsedPercent float64   `json:"valueUsedPercent"`
	CostEntryCount   int       `json:"costEntryCount"`
}

type UserDTO struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	Department string   `json:"department,omitempty"`
	Avatar     string   `json:"avatar,omitempty"`
}

// Auth DTOs

// AuthUserDTO represents the current authenticated user with full context
type AuthUserDTO struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Email          string      `json:"email"`
	Roles          []string    `json:"roles"`
	Company        *CompanyDTO `json:"company,omitempty"`
	Initials       string      `json:"initials"`
	IsSuperAdmin   bool        `json:"isSuperAdmin"`
	IsCompanyAdmin bool        `json:"isCompanyAdmin"`
}

// CompanyDTO represents a company in auth context (minimal version)
type CompanyDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CompanyDetailDTO represents a company with full details
type CompanyDetailDTO struct {
	ID                          string  `json:"id"`
	Name                        string  `json:"name"`
	ShortName                   string  `json:"shortName"`
	OrgNumber                   string  `json:"orgNumber,omitempty"`
	Color                       string  `json:"color"`
	Logo                        string  `json:"logo,omitempty"`
	IsActive                    bool    `json:"isActive"`
	DefaultOfferResponsibleID   *string `json:"defaultOfferResponsibleId,omitempty"`
	DefaultProjectResponsibleID *string `json:"defaultProjectResponsibleId,omitempty"`
	CreatedAt                   string  `json:"createdAt"`
	UpdatedAt                   string  `json:"updatedAt"`
}

// UpdateCompanyRequest contains the data for updating company settings
type UpdateCompanyRequest struct {
	DefaultOfferResponsibleID   *string `json:"defaultOfferResponsibleId,omitempty" validate:"omitempty,max=100"`
	DefaultProjectResponsibleID *string `json:"defaultProjectResponsibleId,omitempty" validate:"omitempty,max=100"`
}

// PermissionDTO represents a single permission
type PermissionDTO struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Allowed  bool   `json:"allowed"`
}

// PermissionsResponseDTO represents the full permissions response
type PermissionsResponseDTO struct {
	Permissions  []PermissionDTO `json:"permissions"`
	Roles        []string        `json:"roles"`
	IsSuperAdmin bool            `json:"isSuperAdmin"`
}

type NotificationDTO struct {
	ID         uuid.UUID  `json:"id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Message    string     `json:"message"`
	Read       bool       `json:"read"`
	CreatedAt  string     `json:"createdAt"` // ISO 8601
	EntityID   *uuid.UUID `json:"entityId,omitempty"`
	EntityType string     `json:"entityType,omitempty"`
}

// CreateNotificationRequest contains the data needed to create a notification
type CreateNotificationRequest struct {
	UserID     uuid.UUID        `json:"userId" validate:"required"`
	Type       NotificationType `json:"type" validate:"required"`
	Title      string           `json:"title" validate:"required,max=200"`
	Message    string           `json:"message" validate:"required,max=500"`
	EntityID   *uuid.UUID       `json:"entityId,omitempty"`
	EntityType string           `json:"entityType,omitempty" validate:"max=50"`
}

// UnreadCountDTO represents the count of unread notifications
type UnreadCountDTO struct {
	Count int `json:"count"`
}

// Dashboard DTOs

// TimeRange represents the time range for dashboard metrics
type TimeRange string

const (
	// TimeRangeRolling12Months is the default 12-month rolling window
	TimeRangeRolling12Months TimeRange = "rolling12months"
	// TimeRangeAllTime calculates metrics without any date filter
	TimeRangeAllTime TimeRange = "allTime"
)

// IsValid checks if the TimeRange value is valid
func (t TimeRange) IsValid() bool {
	switch t {
	case TimeRangeRolling12Months, TimeRangeAllTime:
		return true
	default:
		return false
	}
}

type DisciplineStats struct {
	Name         string  `json:"name"`
	TotalValue   float64 `json:"totalValue"`
	OfferCount   int     `json:"offerCount"`
	ProjectCount int     `json:"projectCount"`
	AvgMargin    float64 `json:"avgMargin"`
}

type TeamMemberStats struct {
	UserID     string  `json:"userId"`
	Name       string  `json:"name"`
	Avatar     string  `json:"avatar,omitempty"`
	OfferCount int     `json:"offerCount"`
	WonCount   int     `json:"wonCount"`
	TotalValue float64 `json:"totalValue"`
	WonValue   float64 `json:"wonValue"`
	WinRate    float64 `json:"winRate"`
}

type PipelinePhaseData struct {
	Phase         OfferPhase `json:"phase"`
	Count         int        `json:"count"`         // Total offer count in this phase
	ProjectCount  int        `json:"projectCount"`  // Unique projects in this phase (excludes orphan offers)
	TotalValue    float64    `json:"totalValue"`    // Sum of best offer value per project (avoids double-counting)
	WeightedValue float64    `json:"weightedValue"` // Weighted by probability
	Offers        []OfferDTO `json:"offers,omitempty"`
}

// WinRateMetrics holds win/loss statistics for transparency
type WinRateMetrics struct {
	WonCount        int     `json:"wonCount"`
	LostCount       int     `json:"lostCount"`
	WonValue        float64 `json:"wonValue"`
	LostValue       float64 `json:"lostValue"`
	WinRate         float64 `json:"winRate"`         // won_count / (won_count + lost_count), 0-1 scale
	EconomicWinRate float64 `json:"economicWinRate"` // won_value / (won_value + lost_value), 0-1 scale
}

// TopCustomerDTO represents a customer with offer statistics for the dashboard
type TopCustomerDTO struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	OrgNumber     string    `json:"orgNumber,omitempty"`
	OfferCount    int       `json:"offerCount"`
	EconomicValue float64   `json:"economicValue"` // Total value of their offers
}

// DashboardMetrics contains all metrics for the dashboard
// By default, metrics use a rolling 12-month window from the current date.
// When timeRange is "allTime", metrics are calculated without any date filter.
// Drafts and expired offers are excluded from all calculations.
//
// IMPORTANT: Pipeline metrics use aggregation to avoid double-counting.
// When a project has multiple offers, only the highest value offer per phase is counted.
// Orphan offers (without project) are included at full value.
type DashboardMetrics struct {
	// TimeRange indicates the time range used for the metrics
	TimeRange TimeRange `json:"timeRange"` // "rolling12months" (default) or "allTime"

	// Offer Metrics (excluding drafts and expired)
	TotalOfferCount      int     `json:"totalOfferCount"`      // Count of offers excluding drafts and expired
	TotalProjectCount    int     `json:"totalProjectCount"`    // Count of unique projects with offers (excludes orphan offers)
	OfferReserve         float64 `json:"offerReserve"`         // Total value of active offers - best per project (avoids double-counting)
	WeightedOfferReserve float64 `json:"weightedOfferReserve"` // Sum of (value * probability/100) for active offers
	AverageProbability   float64 `json:"averageProbability"`   // Average probability of active offers

	// Pipeline Data (phases: in_progress, sent, won, lost - excludes draft and expired)
	// Uses aggregation: for projects with multiple offers, only the highest value per phase is counted
	Pipeline []PipelinePhaseData `json:"pipeline"`

	// Win Rate Metrics
	WinRateMetrics WinRateMetrics `json:"winRateMetrics"`

	// Order Reserve (from active projects)
	OrderReserve float64 `json:"orderReserve"` // Sum of (budget - spent) on active projects

	// Financial Summary
	TotalInvoiced float64 `json:"totalInvoiced"` // Sum of "spent" on all projects in time range
	TotalValue    float64 `json:"totalValue"`    // orderReserve + totalInvoiced

	// Recent Lists (limit 5 each)
	RecentOffers     []OfferDTO    `json:"recentOffers"`     // Last created offers (excluding drafts)
	RecentProjects   []ProjectDTO  `json:"recentProjects"`   // Last created projects
	RecentActivities []ActivityDTO `json:"recentActivities"` // Last activities

	// Top Customers (limit 5)
	TopCustomers []TopCustomerDTO `json:"topCustomers"` // Ranked by offer count
}

// Search DTOs

type SearchResults struct {
	Customers []CustomerDTO `json:"customers"`
	Projects  []ProjectDTO  `json:"projects"`
	Offers    []OfferDTO    `json:"offers"`
	Contacts  []ContactDTO  `json:"contacts"`
	Total     int           `json:"total"`
}

// Pagination response wrapper
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// API Response wrapper
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
}

// Request DTOs

type CreateCustomerRequest struct {
	Name          string           `json:"name" validate:"required,max=200"`
	OrgNumber     string           `json:"orgNumber,omitempty" validate:"max=20"`
	Email         string           `json:"email,omitempty" validate:"omitempty,email"`
	Phone         string           `json:"phone,omitempty" validate:"max=50"`
	Address       string           `json:"address,omitempty" validate:"max=500"`
	City          string           `json:"city,omitempty" validate:"max=100"`
	PostalCode    string           `json:"postalCode,omitempty" validate:"max=20"`
	Country       string           `json:"country" validate:"required,max=100"`
	ContactPerson string           `json:"contactPerson,omitempty" validate:"max=200"`
	ContactEmail  string           `json:"contactEmail,omitempty" validate:"omitempty,email"`
	ContactPhone  string           `json:"contactPhone,omitempty" validate:"max=50"`
	Status        CustomerStatus   `json:"status,omitempty"`
	Tier          CustomerTier     `json:"tier,omitempty"`
	Industry      CustomerIndustry `json:"industry,omitempty"`
	Notes         string           `json:"notes,omitempty"`
	CustomerClass string           `json:"customerClass,omitempty" validate:"max=50"`
	CreditLimit   *float64         `json:"creditLimit,omitempty"`
	IsInternal    bool             `json:"isInternal,omitempty"`
	Municipality  string           `json:"municipality,omitempty" validate:"max=100"`
	County        string           `json:"county,omitempty" validate:"max=100"`
}

type UpdateCustomerRequest struct {
	Name          string           `json:"name" validate:"required,max=200"`
	OrgNumber     string           `json:"orgNumber,omitempty" validate:"max=20"`
	Email         string           `json:"email,omitempty" validate:"omitempty,email"`
	Phone         string           `json:"phone,omitempty" validate:"max=50"`
	Address       string           `json:"address,omitempty" validate:"max=500"`
	City          string           `json:"city,omitempty" validate:"max=100"`
	PostalCode    string           `json:"postalCode,omitempty" validate:"max=20"`
	Country       string           `json:"country" validate:"required,max=100"`
	ContactPerson string           `json:"contactPerson,omitempty" validate:"max=200"`
	ContactEmail  string           `json:"contactEmail,omitempty" validate:"omitempty,email"`
	ContactPhone  string           `json:"contactPhone,omitempty" validate:"max=50"`
	Status        CustomerStatus   `json:"status,omitempty"`
	Tier          CustomerTier     `json:"tier,omitempty"`
	Industry      CustomerIndustry `json:"industry,omitempty"`
	Notes         string           `json:"notes,omitempty"`
	CustomerClass string           `json:"customerClass,omitempty" validate:"max=50"`
	CreditLimit   *float64         `json:"creditLimit,omitempty"`
	IsInternal    bool             `json:"isInternal,omitempty"`
	Municipality  string           `json:"municipality,omitempty" validate:"max=100"`
	County        string           `json:"county,omitempty" validate:"max=100"`
}

type CreateContactRequest struct {
	FirstName              string      `json:"firstName" validate:"required,max=100"`
	LastName               string      `json:"lastName" validate:"required,max=100"`
	Email                  string      `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone                  string      `json:"phone,omitempty" validate:"max=50"`
	Mobile                 string      `json:"mobile,omitempty" validate:"max=50"`
	Title                  string      `json:"title,omitempty" validate:"max=100"`
	Department             string      `json:"department,omitempty" validate:"max=100"`
	ContactType            ContactType `json:"contactType,omitempty"`
	PrimaryCustomerID      *uuid.UUID  `json:"primaryCustomerId,omitempty"`
	Address                string      `json:"address,omitempty" validate:"max=500"`
	City                   string      `json:"city,omitempty" validate:"max=100"`
	PostalCode             string      `json:"postalCode,omitempty" validate:"max=20"`
	Country                string      `json:"country,omitempty" validate:"max=100"`
	LinkedInURL            string      `json:"linkedInUrl,omitempty" validate:"max=500"`
	PreferredContactMethod string      `json:"preferredContactMethod,omitempty" validate:"max=50"`
	Notes                  string      `json:"notes,omitempty"`
}

type UpdateContactRequest struct {
	FirstName              string      `json:"firstName" validate:"required,max=100"`
	LastName               string      `json:"lastName" validate:"required,max=100"`
	Email                  string      `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone                  string      `json:"phone,omitempty" validate:"max=50"`
	Mobile                 string      `json:"mobile,omitempty" validate:"max=50"`
	Title                  string      `json:"title,omitempty" validate:"max=100"`
	Department             string      `json:"department,omitempty" validate:"max=100"`
	ContactType            ContactType `json:"contactType,omitempty"`
	PrimaryCustomerID      *uuid.UUID  `json:"primaryCustomerId,omitempty"`
	Address                string      `json:"address,omitempty" validate:"max=500"`
	City                   string      `json:"city,omitempty" validate:"max=100"`
	PostalCode             string      `json:"postalCode,omitempty" validate:"max=20"`
	Country                string      `json:"country,omitempty" validate:"max=100"`
	LinkedInURL            string      `json:"linkedInUrl,omitempty" validate:"max=500"`
	PreferredContactMethod string      `json:"preferredContactMethod,omitempty" validate:"max=50"`
	Notes                  string      `json:"notes,omitempty"`
	IsActive               *bool       `json:"isActive,omitempty"`
}

// Contact relationship request DTOs
type AddContactRelationshipRequest struct {
	EntityType ContactEntityType `json:"entityType" validate:"required"`
	EntityID   uuid.UUID         `json:"entityId" validate:"required"`
	Role       string            `json:"role,omitempty" validate:"max=100"`
	IsPrimary  bool              `json:"isPrimary,omitempty"`
}

// Deal request DTOs
type CreateDealRequest struct {
	Title             string     `json:"title" validate:"required,max=200"`
	Description       string     `json:"description,omitempty"`
	CustomerID        uuid.UUID  `json:"customerId" validate:"required"`
	CompanyID         CompanyID  `json:"companyId" validate:"required"`
	Stage             DealStage  `json:"stage,omitempty"`
	Probability       int        `json:"probability,omitempty" validate:"min=0,max=100"`
	Value             float64    `json:"value,omitempty" validate:"gte=0"`
	Currency          string     `json:"currency,omitempty" validate:"max=3"`
	ExpectedCloseDate *time.Time `json:"expectedCloseDate,omitempty"`
	OwnerID           string     `json:"ownerId" validate:"required,max=100"`
	Source            string     `json:"source,omitempty" validate:"max=100"`
	Notes             string     `json:"notes,omitempty"`
	OfferID           *uuid.UUID `json:"offerId,omitempty"`
}

type UpdateDealRequest struct {
	Title             string     `json:"title" validate:"required,max=200"`
	Description       string     `json:"description,omitempty"`
	Stage             DealStage  `json:"stage,omitempty"`
	Probability       int        `json:"probability,omitempty" validate:"min=0,max=100"`
	Value             float64    `json:"value,omitempty" validate:"gte=0"`
	Currency          string     `json:"currency,omitempty" validate:"max=3"`
	ExpectedCloseDate *time.Time `json:"expectedCloseDate,omitempty"`
	ActualCloseDate   *time.Time `json:"actualCloseDate,omitempty"`
	OwnerID           string     `json:"ownerId,omitempty" validate:"max=100"`
	Source            string     `json:"source,omitempty" validate:"max=100"`
	Notes             string     `json:"notes,omitempty"`
	LostReason        string     `json:"lostReason,omitempty" validate:"max=500"`
}

type UpdateDealStageRequest struct {
	Stage DealStage `json:"stage" validate:"required"`
	Notes string    `json:"notes,omitempty"`
}

// LoseDealRequest contains the categorized reason and detailed notes for losing a deal
type LoseDealRequest struct {
	Reason LossReasonCategory `json:"reason" validate:"required,oneof=price timing competitor requirements other" example:"competitor"`
	Notes  string             `json:"notes" validate:"required,min=10,max=500" example:"Lost to competitor XYZ who offered lower price"`
}

type CreateProjectRequest struct {
	Name                    string         `json:"name" validate:"required,max=200"`
	ProjectNumber           string         `json:"projectNumber,omitempty" validate:"max=50"`
	Summary                 string         `json:"summary,omitempty"`
	Description             string         `json:"description,omitempty"`
	CustomerID              uuid.UUID      `json:"customerId" validate:"required"`
	CompanyID               CompanyID      `json:"companyId" validate:"required"`
	Status                  ProjectStatus  `json:"status" validate:"required,oneof=planning active on_hold completed cancelled"`
	Phase                   ProjectPhase   `json:"phase,omitempty" validate:"omitempty,oneof=tilbud active completed cancelled"`
	StartDate               *time.Time     `json:"startDate,omitempty"`
	EndDate                 *time.Time     `json:"endDate,omitempty"`
	Value                   float64        `json:"value" validate:"gte=0"`
	Cost                    float64        `json:"cost" validate:"gte=0"`
	Spent                   float64        `json:"spent" validate:"gte=0"`
	ManagerID               *string        `json:"managerId,omitempty"`
	TeamMembers             []string       `json:"teamMembers,omitempty"`
	OfferID                 *uuid.UUID     `json:"offerId,omitempty"`
	DealID                  *uuid.UUID     `json:"dealId,omitempty"`
	HasDetailedBudget       bool           `json:"hasDetailedBudget,omitempty"`
	Health                  *ProjectHealth `json:"health,omitempty"`
	CompletionPercent       *float64       `json:"completionPercent,omitempty" validate:"omitempty,gte=0,lte=100"`
	EstimatedCompletionDate *time.Time     `json:"estimatedCompletionDate,omitempty"`
}

type UpdateProjectRequest struct {
	Name                    string         `json:"name" validate:"required,max=200"`
	ProjectNumber           string         `json:"projectNumber,omitempty" validate:"max=50"`
	Summary                 string         `json:"summary,omitempty"`
	Description             string         `json:"description,omitempty"`
	CompanyID               CompanyID      `json:"companyId" validate:"required"`
	Status                  ProjectStatus  `json:"status" validate:"required"`
	StartDate               *time.Time     `json:"startDate,omitempty"`
	EndDate                 *time.Time     `json:"endDate,omitempty"`
	Value                   float64        `json:"value" validate:"gte=0"`
	Cost                    float64        `json:"cost" validate:"gte=0"`
	Spent                   float64        `json:"spent" validate:"gte=0"`
	ManagerID               *string        `json:"managerId,omitempty"`
	TeamMembers             []string       `json:"teamMembers,omitempty"`
	DealID                  *uuid.UUID     `json:"dealId,omitempty"`
	HasDetailedBudget       *bool          `json:"hasDetailedBudget,omitempty"`
	Health                  *ProjectHealth `json:"health,omitempty"`
	CompletionPercent       *float64       `json:"completionPercent,omitempty" validate:"omitempty,gte=0,lte=100"`
	EstimatedCompletionDate *time.Time     `json:"estimatedCompletionDate,omitempty"`
}

type CreateOfferRequest struct {
	Title             string                   `json:"title" validate:"required,max=200"`
	CustomerID        *uuid.UUID               `json:"customerId,omitempty"` // Optional if projectId is provided (inherits from project)
	CompanyID         CompanyID                `json:"companyId,omitempty"`
	Phase             OfferPhase               `json:"phase,omitempty"`
	Probability       *int                     `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`
	Status            OfferStatus              `json:"status,omitempty"`
	ResponsibleUserID string                   `json:"responsibleUserId,omitempty"`
	ProjectID         *uuid.UUID               `json:"projectId,omitempty"` // Link to existing project (auto-created if not provided and phase != draft)
	Items             []CreateOfferItemRequest `json:"items,omitempty"`
	Description       string                   `json:"description,omitempty"`
	Notes             string                   `json:"notes,omitempty"`
	DueDate           *time.Time               `json:"dueDate,omitempty"`
	Cost              float64                  `json:"cost,omitempty" validate:"gte=0"`
	Location          string                   `json:"location,omitempty" validate:"max=200"`
	SentDate          *time.Time               `json:"sentDate,omitempty"`
	ExpirationDate    *time.Time               `json:"expirationDate,omitempty"` // Optional, defaults to 60 days after sent date
}

type UpdateOfferRequest struct {
	Title             string      `json:"title" validate:"required,max=200"`
	Phase             OfferPhase  `json:"phase" validate:"required"`
	Probability       int         `json:"probability" validate:"min=0,max=100"`
	Status            OfferStatus `json:"status" validate:"required"`
	ResponsibleUserID string      `json:"responsibleUserId" validate:"required"`
	Description       string      `json:"description,omitempty"`
	Notes             string      `json:"notes,omitempty"`
	DueDate           *time.Time  `json:"dueDate,omitempty"`
	Cost              float64     `json:"cost,omitempty" validate:"gte=0"`
	Location          string      `json:"location,omitempty" validate:"max=200"`
	SentDate          *time.Time  `json:"sentDate,omitempty"`
	ExpirationDate    *time.Time  `json:"expirationDate,omitempty"` // Optional, defaults to 60 days after sent date
}

type CreateOfferItemRequest struct {
	Discipline  string  `json:"discipline" validate:"required,max=200"`
	Cost        float64 `json:"cost" validate:"required,gte=0"`
	Revenue     float64 `json:"revenue" validate:"required,gte=0"`
	Description string  `json:"description,omitempty"`
	Quantity    float64 `json:"quantity,omitempty" validate:"gte=0"`
	Unit        string  `json:"unit,omitempty" validate:"max=50"`
}

type UpdateOfferItemRequest struct {
	Discipline  string  `json:"discipline" validate:"required,max=200"`
	Cost        float64 `json:"cost" validate:"required,gte=0"`
	Revenue     float64 `json:"revenue" validate:"required,gte=0"`
	Description string  `json:"description,omitempty"`
	Quantity    float64 `json:"quantity,omitempty" validate:"gte=0"`
	Unit        string  `json:"unit,omitempty" validate:"max=50"`
}

// Budget Item Request DTOs

type CreateBudgetItemRequest struct {
	ParentType     BudgetParentType `json:"parentType" validate:"required"`
	ParentID       uuid.UUID        `json:"parentId" validate:"required"`
	Name           string           `json:"name" validate:"required,max=200"`
	ExpectedCost   float64          `json:"expectedCost" validate:"gte=0"`
	ExpectedMargin float64          `json:"expectedMargin" validate:"gte=0,lte=100"`
	Quantity       *float64         `json:"quantity,omitempty" validate:"omitempty,gte=0"`
	PricePerItem   *float64         `json:"pricePerItem,omitempty" validate:"omitempty,gte=0"`
	Description    string           `json:"description,omitempty"`
	DisplayOrder   int              `json:"displayOrder,omitempty" validate:"gte=0"`
}

type UpdateBudgetItemRequest struct {
	Name           string   `json:"name" validate:"required,max=200"`
	ExpectedCost   float64  `json:"expectedCost" validate:"gte=0"`
	ExpectedMargin float64  `json:"expectedMargin" validate:"gte=0,lte=100"`
	Quantity       *float64 `json:"quantity,omitempty" validate:"omitempty,gte=0"`
	PricePerItem   *float64 `json:"pricePerItem,omitempty" validate:"omitempty,gte=0"`
	Description    string   `json:"description,omitempty"`
	DisplayOrder   int      `json:"displayOrder,omitempty" validate:"gte=0"`
}

// ReorderBudgetItemsRequest contains the ordered list of budget item IDs
type ReorderBudgetItemsRequest struct {
	OrderedIDs []uuid.UUID `json:"orderedIds" validate:"required,min=1"`
}

// AddOfferBudgetItemRequest is the simplified request for adding budget items to an offer
// ParentType and ParentID are inferred from the URL
type AddOfferBudgetItemRequest struct {
	Name           string   `json:"name" validate:"required,max=200"`
	ExpectedCost   float64  `json:"expectedCost" validate:"gte=0"`
	ExpectedMargin float64  `json:"expectedMargin" validate:"gte=0,lte=100"`
	Quantity       *float64 `json:"quantity,omitempty" validate:"omitempty,gte=0"`
	PricePerItem   *float64 `json:"pricePerItem,omitempty" validate:"omitempty,gte=0"`
	Description    string   `json:"description,omitempty"`
	DisplayOrder   int      `json:"displayOrder,omitempty" validate:"gte=0"`
}

// Project Actual Cost Request DTOs

type CreateProjectActualCostRequest struct {
	ProjectID        uuid.UUID  `json:"projectId" validate:"required"`
	CostType         CostType   `json:"costType" validate:"required"`
	Description      string     `json:"description" validate:"required,max=500"`
	Amount           float64    `json:"amount" validate:"required"`
	Currency         string     `json:"currency,omitempty" validate:"max=3"`
	CostDate         time.Time  `json:"costDate" validate:"required"`
	PostingDate      *time.Time `json:"postingDate,omitempty"`
	BudgetItemID     *uuid.UUID `json:"budgetItemId,omitempty"`
	ERPSource        ERPSource  `json:"erpSource,omitempty"`
	ERPReference     string     `json:"erpReference,omitempty" validate:"max=100"`
	ERPTransactionID string     `json:"erpTransactionId,omitempty" validate:"max=100"`
	Notes            string     `json:"notes,omitempty"`
}

type UpdateProjectActualCostRequest struct {
	CostType         CostType   `json:"costType" validate:"required"`
	Description      string     `json:"description" validate:"required,max=500"`
	Amount           float64    `json:"amount" validate:"required"`
	Currency         string     `json:"currency,omitempty" validate:"max=3"`
	CostDate         time.Time  `json:"costDate" validate:"required"`
	PostingDate      *time.Time `json:"postingDate,omitempty"`
	BudgetItemID     *uuid.UUID `json:"budgetItemId,omitempty"`
	ERPSource        ERPSource  `json:"erpSource,omitempty"`
	ERPReference     string     `json:"erpReference,omitempty" validate:"max=100"`
	ERPTransactionID string     `json:"erpTransactionId,omitempty" validate:"max=100"`
	Notes            string     `json:"notes,omitempty"`
}

type ApproveProjectActualCostRequest struct {
	IsApproved bool `json:"isApproved" validate:"required"`
}

// Additional DTOs for compatibility

type OfferWithItemsDTO struct {
	ID                  uuid.UUID      `json:"id"`
	Title               string         `json:"title"`
	CustomerID          uuid.UUID      `json:"customerId"`
	CustomerName        string         `json:"customerName,omitempty"`
	CompanyID           CompanyID      `json:"companyId"`
	Phase               OfferPhase     `json:"phase"`
	Probability         int            `json:"probability"`
	Value               float64        `json:"value"`
	Status              OfferStatus    `json:"status"`
	CreatedAt           string         `json:"createdAt"`
	UpdatedAt           string         `json:"updatedAt"`
	ResponsibleUserID   string         `json:"responsibleUserId"`
	ResponsibleUserName string         `json:"responsibleUserName,omitempty"`
	Items               []OfferItemDTO `json:"items"`
	Description         string         `json:"description,omitempty"`
	Notes               string         `json:"notes,omitempty"`
}

type AdvanceOfferRequest struct {
	Phase     OfferPhase `json:"phase" validate:"required"`
	ProjectID *uuid.UUID `json:"projectId,omitempty"` // Link to existing project (auto-created if not provided and advancing to in_progress)
}

// Offer Lifecycle DTOs

// CloneOfferRequest contains options for cloning an offer
type CloneOfferRequest struct {
	NewTitle      string `json:"newTitle,omitempty" validate:"max=200"`
	IncludeBudget *bool  `json:"includeBudget,omitempty"` // Default true - clone budget items (nil treated as true)
	IncludeFiles  bool   `json:"includeFiles"`            // Default false - files are not cloned by default
}

// AcceptOfferRequest contains options when accepting an offer
type AcceptOfferRequest struct {
	CreateProject bool   `json:"createProject"` // If true, create a project from this offer
	ProjectName   string `json:"projectName,omitempty" validate:"max=200"`
	ManagerID     string `json:"managerId,omitempty" validate:"max=100"`
}

// AcceptOfferResponse contains the result of accepting an offer
type AcceptOfferResponse struct {
	Offer   *OfferDTO   `json:"offer"`
	Project *ProjectDTO `json:"project,omitempty"` // Only present if createProject was true
}

// RejectOfferRequest contains the reason for rejecting an offer
type RejectOfferRequest struct {
	Reason string `json:"reason,omitempty" validate:"max=500"`
}

// OfferWithProjectResponse contains an offer and optionally an auto-created project
// Used when creating/advancing offers that trigger project auto-creation
type OfferWithProjectResponse struct {
	Offer          *OfferDTO   `json:"offer"`
	Project        *ProjectDTO `json:"project,omitempty"`        // Present if a project was auto-created
	ProjectCreated bool        `json:"projectCreated,omitempty"` // True if a new project was created
}

// WinOfferRequest contains options when winning an offer within a project context
// This is used when the offer belongs to a project (offer folder model)
type WinOfferRequest struct {
	// Notes is an optional note about why this offer was selected
	Notes string `json:"notes,omitempty" validate:"max=500"`
}

// WinOfferResponse contains the result of winning an offer
type WinOfferResponse struct {
	Offer         *OfferDTO   `json:"offer"`
	Project       *ProjectDTO `json:"project,omitempty"`
	ExpiredOffers []OfferDTO  `json:"expiredOffers,omitempty"` // Sibling offers that were expired
	ExpiredCount  int         `json:"expiredCount"`            // Count of sibling offers that were expired
}

// ProjectOffersDTO contains a project with its associated offers
type ProjectOffersDTO struct {
	Project *ProjectDTO `json:"project"`
	Offers  []OfferDTO  `json:"offers"`
}

// OfferDetailDTO includes offer with budget items and summary
type OfferDetailDTO struct {
	OfferDTO
	BudgetItems   []BudgetItemDTO   `json:"budgetItems,omitempty"`
	BudgetSummary *BudgetSummaryDTO `json:"budgetSummary,omitempty"`
	FilesCount    int               `json:"filesCount"`
}

type ProjectBudgetDTO struct {
	Value         float64 `json:"value"`
	Cost          float64 `json:"cost"`
	MarginPercent float64 `json:"marginPercent"`
	Spent         float64 `json:"spent"`
	Remaining     float64 `json:"remaining"`
	PercentUsed   float64 `json:"percentUsed"`
}

type ActivityDTO struct {
	ID               uuid.UUID          `json:"id"`
	TargetType       ActivityTargetType `json:"targetType"`
	TargetID         uuid.UUID          `json:"targetId"`
	Title            string             `json:"title"`
	Body             string             `json:"body,omitempty"`
	OccurredAt       string             `json:"occurredAt"`
	CreatorName      string             `json:"creatorName,omitempty"`
	CreatedAt        string             `json:"createdAt"`
	ActivityType     ActivityType       `json:"activityType"`
	Status           ActivityStatus     `json:"status"`
	ScheduledAt      string             `json:"scheduledAt,omitempty"`
	DueDate          string             `json:"dueDate,omitempty"`
	CompletedAt      string             `json:"completedAt,omitempty"`
	DurationMinutes  *int               `json:"durationMinutes,omitempty"`
	Priority         int                `json:"priority"`
	IsPrivate        bool               `json:"isPrivate"`
	CreatorID        string             `json:"creatorId,omitempty"`
	AssignedToID     string             `json:"assignedToId,omitempty"`
	CompanyID        *CompanyID         `json:"companyId,omitempty"`
	Attendees        []string           `json:"attendees,omitempty"`
	ParentActivityID *uuid.UUID         `json:"parentActivityId,omitempty"`
}

// UserRole DTOs

type UserRoleDTO struct {
	ID        uuid.UUID    `json:"id"`
	UserID    string       `json:"userId"`
	Role      UserRoleType `json:"role"`
	CompanyID *CompanyID   `json:"companyId,omitempty"`
	GrantedBy string       `json:"grantedBy,omitempty"`
	GrantedAt string       `json:"grantedAt"`
	ExpiresAt string       `json:"expiresAt,omitempty"`
	IsActive  bool         `json:"isActive"`
}

type UserPermissionDTO struct {
	ID         uuid.UUID      `json:"id"`
	UserID     string         `json:"userId"`
	Permission PermissionType `json:"permission"`
	CompanyID  *CompanyID     `json:"companyId,omitempty"`
	IsGranted  bool           `json:"isGranted"`
	GrantedBy  string         `json:"grantedBy,omitempty"`
	GrantedAt  string         `json:"grantedAt"`
	ExpiresAt  string         `json:"expiresAt,omitempty"`
	Reason     string         `json:"reason,omitempty"`
}

type AuditLogDTO struct {
	ID          uuid.UUID   `json:"id"`
	UserID      string      `json:"userId,omitempty"`
	UserEmail   string      `json:"userEmail,omitempty"`
	UserName    string      `json:"userName,omitempty"`
	Action      AuditAction `json:"action"`
	EntityType  string      `json:"entityType"`
	EntityID    *uuid.UUID  `json:"entityId,omitempty"`
	EntityName  string      `json:"entityName,omitempty"`
	CompanyID   *CompanyID  `json:"companyId,omitempty"`
	OldValues   interface{} `json:"oldValues,omitempty"`
	NewValues   interface{} `json:"newValues,omitempty"`
	Changes     interface{} `json:"changes,omitempty"`
	IPAddress   string      `json:"ipAddress,omitempty"`
	UserAgent   string      `json:"userAgent,omitempty"`
	RequestID   string      `json:"requestId,omitempty"`
	PerformedAt string      `json:"performedAt"`
}

type FileDTO struct {
	ID          uuid.UUID  `json:"id"`
	Filename    string     `json:"filename"`
	ContentType string     `json:"contentType"`
	Size        int64      `json:"size"`
	OfferID     *uuid.UUID `json:"offerId,omitempty"`
	CreatedAt   string     `json:"createdAt"`
}

type SearchResult struct {
	ID       uuid.UUID `json:"id"`
	Type     string    `json:"type"`
	Title    string    `json:"title"`
	Subtitle string    `json:"subtitle,omitempty"`
	Metadata string    `json:"metadata,omitempty"`
}

// Customer Stats DTOs

// CustomerStatsResponse contains aggregated statistics for a customer
type CustomerStatsResponse struct {
	ActiveDealsCount    int64   `json:"activeDealsCount"`
	TotalDealsCount     int64   `json:"totalDealsCount"`
	TotalDealValue      float64 `json:"totalDealValue"`
	WonDealsValue       float64 `json:"wonDealsValue"`
	ActiveProjectsCount int64   `json:"activeProjectsCount"`
	TotalProjectsCount  int64   `json:"totalProjectsCount"`
}

// Activity Filters for advanced querying

// ActivityFilters provides comprehensive filtering options for activity queries
type ActivityFilters struct {
	ActivityType  *ActivityType       `json:"activityType,omitempty"`
	Status        *ActivityStatus     `json:"status,omitempty"`
	TargetType    *ActivityTargetType `json:"targetType,omitempty"`
	TargetID      *uuid.UUID          `json:"targetId,omitempty"`
	AssignedToID  *string             `json:"assignedToId,omitempty"`
	CreatorID     *string             `json:"creatorId,omitempty"`
	DueDateFrom   *time.Time          `json:"dueDateFrom,omitempty"`
	DueDateTo     *time.Time          `json:"dueDateTo,omitempty"`
	ScheduledFrom *time.Time          `json:"scheduledFrom,omitempty"`
	ScheduledTo   *time.Time          `json:"scheduledTo,omitempty"`
	IsPrivate     *bool               `json:"isPrivate,omitempty"`
	Priority      *int                `json:"priority,omitempty"`
}

// ActivityStatusCounts holds counts grouped by activity status for dashboard statistics
type ActivityStatusCounts struct {
	Planned    int `json:"planned"`
	InProgress int `json:"inProgress"`
	Completed  int `json:"completed"`
	Cancelled  int `json:"cancelled"`
}

// CustomerWithStatsResponse extends CustomerDTO with statistics
type CustomerWithStatsResponse struct {
	CustomerDTO
	Stats CustomerStatsResponse `json:"stats"`
}

// Activity Request DTOs

// CreateActivityRequest contains the data needed to create a new activity
type CreateActivityRequest struct {
	TargetType      ActivityTargetType `json:"targetType" validate:"required"`
	TargetID        uuid.UUID          `json:"targetId" validate:"required"`
	Title           string             `json:"title" validate:"required,max=200"`
	Body            string             `json:"body,omitempty" validate:"max=2000"`
	ActivityType    ActivityType       `json:"activityType" validate:"required"`
	Status          ActivityStatus     `json:"status,omitempty"`
	ScheduledAt     *time.Time         `json:"scheduledAt,omitempty"`
	DueDate         *time.Time         `json:"dueDate,omitempty"`
	DurationMinutes *int               `json:"durationMinutes,omitempty" validate:"omitempty,min=1"`
	Priority        int                `json:"priority,omitempty" validate:"min=0,max=5"`
	IsPrivate       bool               `json:"isPrivate,omitempty"`
	AssignedToID    string             `json:"assignedToId,omitempty" validate:"max=100"`
	CompanyID       *CompanyID         `json:"companyId,omitempty"`
	Attendees       []string           `json:"attendees,omitempty"`
}

// UpdateActivityRequest contains the data for updating an existing activity
type UpdateActivityRequest struct {
	Title           string         `json:"title" validate:"required,max=200"`
	Body            string         `json:"body,omitempty" validate:"max=2000"`
	Status          ActivityStatus `json:"status,omitempty"`
	ScheduledAt     *time.Time     `json:"scheduledAt,omitempty"`
	DueDate         *time.Time     `json:"dueDate,omitempty"`
	DurationMinutes *int           `json:"durationMinutes,omitempty" validate:"omitempty,min=1"`
	Priority        int            `json:"priority,omitempty" validate:"min=0,max=5"`
	IsPrivate       bool           `json:"isPrivate,omitempty"`
	AssignedToID    string         `json:"assignedToId,omitempty" validate:"max=100"`
	Attendees       []string       `json:"attendees,omitempty"`
}

// CompleteActivityRequest contains optional outcome when completing an activity
type CompleteActivityRequest struct {
	Outcome string `json:"outcome,omitempty" validate:"max=500"`
}

// AddAttendeeRequest contains the user ID to add as an attendee
type AddAttendeeRequest struct {
	UserID string `json:"userId" validate:"required,max=100"`
}

// CreateFollowUpRequest contains the data for creating a follow-up task from a completed activity
type CreateFollowUpRequest struct {
	Title        string     `json:"title" validate:"required,max=200"`
	Description  string     `json:"description,omitempty" validate:"max=2000"`
	DueDate      *time.Time `json:"dueDate,omitempty"`
	AssignedToID *string    `json:"assignedToId,omitempty" validate:"omitempty,max=100"`
}

// Project Request DTOs

// UpdateProjectStatusRequest contains data for updating project status and health
type UpdateProjectStatusRequest struct {
	Status            ProjectStatus  `json:"status" validate:"required"`
	Health            *ProjectHealth `json:"health,omitempty"`
	CompletionPercent *float64       `json:"completionPercent,omitempty" validate:"omitempty,gte=0,lte=100"`
}

// InheritBudgetRequest for POST /projects/{id}/inherit-budget
type InheritBudgetRequest struct {
	OfferID uuid.UUID `json:"offerId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
}

// InheritBudgetResponse contains the result of budget inheritance from an offer
type InheritBudgetResponse struct {
	Project    *ProjectDTO `json:"project"`
	ItemsCount int         `json:"itemsCount"`
}

// ResyncFromOfferResponse contains the result of syncing project economics from its best offer
type ResyncFromOfferResponse struct {
	Project     *ProjectDTO `json:"project"`
	OfferID     uuid.UUID   `json:"offerId"`
	OfferTitle  string      `json:"offerTitle"`
	OfferPhase  string      `json:"offerPhase"`
	SyncedValue float64     `json:"syncedValue"`
	SyncedCost  float64     `json:"syncedCost"`
}

// ProjectWithDetailsDTO includes project data with related entities and budget summary
type ProjectWithDetailsDTO struct {
	ProjectDTO
	BudgetSummary    *BudgetSummaryDTO `json:"budgetSummary,omitempty"`
	RecentActivities []ActivityDTO     `json:"recentActivities,omitempty"`
	Offer            *OfferDTO         `json:"offer,omitempty"`
	Deal             *DealDTO          `json:"deal,omitempty"`
}

// CreateOfferFromDealRequest contains options for creating an offer from a deal
type CreateOfferFromDealRequest struct {
	TemplateOfferID *uuid.UUID `json:"templateOfferId,omitempty"` // Optional: copy budget dimensions from this offer
	Title           string     `json:"title,omitempty" validate:"max=200"`
}

// Pipeline Analytics DTOs

// StageSummaryDTO represents aggregated data for a single stage
type StageSummaryDTO struct {
	Stage          string  `json:"stage"`
	DealCount      int64   `json:"dealCount"`
	TotalValue     float64 `json:"totalValue"`
	WeightedValue  float64 `json:"weightedValue"`
	AvgProbability float64 `json:"avgProbability"`
	AvgDealValue   float64 `json:"avgDealValue"`
	OverdueCount   int64   `json:"overdueCount"`
}

// RevenueForecastDTO represents forecast for a time period
type RevenueForecastDTO struct {
	Period        string  `json:"period"` // "30d" or "90d"
	DealCount     int64   `json:"dealCount"`
	TotalValue    float64 `json:"totalValue"`
	WeightedValue float64 `json:"weightedValue"`
}

// ConversionRateDTO represents stage-to-stage conversion
type ConversionRateDTO struct {
	FromStage      string  `json:"fromStage"`
	ToStage        string  `json:"toStage"`
	ConversionRate float64 `json:"conversionRate"`
	DealsConverted int64   `json:"dealsConverted"`
	TotalDeals     int64   `json:"totalDeals"`
}

// WinRateAnalysisDTO represents win/loss analysis
type WinRateAnalysisDTO struct {
	TotalClosed      int64   `json:"totalClosed"`
	TotalWon         int64   `json:"totalWon"`
	TotalLost        int64   `json:"totalLost"`
	WinRate          float64 `json:"winRate"`
	WonValue         float64 `json:"wonValue"`
	LostValue        float64 `json:"lostValue"`
	AvgWonDealValue  float64 `json:"avgWonDealValue"`
	AvgLostDealValue float64 `json:"avgLostDealValue"`
	AvgDaysToClose   float64 `json:"avgDaysToClose"`
}

// PipelineAnalyticsDTO is the comprehensive analytics response
type PipelineAnalyticsDTO struct {
	Summary         []StageSummaryDTO   `json:"summary"`
	Forecast30Days  RevenueForecastDTO  `json:"forecast30Days"`
	Forecast90Days  RevenueForecastDTO  `json:"forecast90Days"`
	ConversionRates []ConversionRateDTO `json:"conversionRates"`
	WinRateAnalysis WinRateAnalysisDTO  `json:"winRateAnalysis"`
	GeneratedAt     string              `json:"generatedAt"`
}

// PipelineAnalyticsFilters contains optional filters for analytics queries
type PipelineAnalyticsFilters struct {
	CompanyID *CompanyID `json:"companyId,omitempty"`
	OwnerID   *string    `json:"ownerId,omitempty"`
	DateFrom  *time.Time `json:"dateFrom,omitempty"`
	DateTo    *time.Time `json:"dateTo,omitempty"`
}

// CreateOfferFromDealResponse contains the result of creating an offer from a deal
type CreateOfferFromDealResponse struct {
	Offer *OfferDTO `json:"offer"`
	Deal  *DealDTO  `json:"deal"`
}

// ============================================================================
// Inquiry (Draft Offer) DTOs
// ============================================================================

// InquiryDTO represents an inquiry (offer in draft phase) - alias for clarity
type InquiryDTO = OfferDTO

// CreateInquiryRequest contains the data needed to create a new inquiry (draft offer)
// Minimal fields required - responsibleUserId and companyId are optional
type CreateInquiryRequest struct {
	Title       string     `json:"title" validate:"required,max=200"`
	CustomerID  uuid.UUID  `json:"customerId" validate:"required"`
	Description string     `json:"description,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
}

// ConvertInquiryRequest contains options for converting an inquiry to an offer
type ConvertInquiryRequest struct {
	ResponsibleUserID *string    `json:"responsibleUserId,omitempty" validate:"omitempty,max=100"`
	CompanyID         *CompanyID `json:"companyId,omitempty"`
}

// ConvertInquiryResponse contains the result of converting an inquiry to an offer
type ConvertInquiryResponse struct {
	Offer       *OfferDTO `json:"offer"`
	OfferNumber string    `json:"offerNumber"`
}

// ============================================================================
// Offer Property Update Request DTOs
// ============================================================================

// UpdateOfferProbabilityRequest for updating offer probability
type UpdateOfferProbabilityRequest struct {
	Probability int `json:"probability" validate:"required,min=0,max=100"`
}

// UpdateOfferTitleRequest for updating offer title
type UpdateOfferTitleRequest struct {
	Title string `json:"title" validate:"required,max=200"`
}

// UpdateOfferResponsibleRequest for updating offer responsible user
type UpdateOfferResponsibleRequest struct {
	ResponsibleUserID string `json:"responsibleUserId" validate:"required,max=100"`
}

// UpdateOfferCustomerRequest for updating offer customer
type UpdateOfferCustomerRequest struct {
	CustomerID uuid.UUID `json:"customerId" validate:"required"`
}

// UpdateOfferValueRequest for updating offer value
type UpdateOfferValueRequest struct {
	Value float64 `json:"value" validate:"required,gte=0"`
}

// UpdateOfferDueDateRequest for updating offer due date
type UpdateOfferDueDateRequest struct {
	DueDate *time.Time `json:"dueDate"` // nullable to allow clearing
}

// UpdateOfferExpirationDateRequest for updating/extending offer expiration date
type UpdateOfferExpirationDateRequest struct {
	ExpirationDate *time.Time `json:"expirationDate"` // nullable to allow clearing (though not recommended)
}

// UpdateOfferDescriptionRequest for updating offer description
type UpdateOfferDescriptionRequest struct {
	Description string `json:"description" validate:"max=10000"`
}

// UpdateOfferProjectRequest for linking offer to a project
type UpdateOfferProjectRequest struct {
	ProjectID uuid.UUID `json:"projectId" validate:"required"`
}

// UpdateOfferCustomerHasWonProjectRequest for toggling customer has won project flag
type UpdateOfferCustomerHasWonProjectRequest struct {
	CustomerHasWonProject bool `json:"customerHasWonProject"`
}

// UpdateOfferNumberRequest for updating the internal offer number
type UpdateOfferNumberRequest struct {
	OfferNumber string `json:"offerNumber" validate:"required,max=50"`
}

// UpdateOfferExternalReferenceRequest for updating the external reference
type UpdateOfferExternalReferenceRequest struct {
	ExternalReference string `json:"externalReference" validate:"max=100"`
}

// NextOfferNumberResponse contains the next offer number preview
type NextOfferNumberResponse struct {
	NextOfferNumber string    `json:"nextOfferNumber"` // The next offer number that would be assigned, e.g., "TK-2025-001"
	CompanyID       CompanyID `json:"companyId"`       // The company ID this number is for
	Year            int       `json:"year"`            // The year component of the number
}

// ============================================================================
// Customer Property Update Request DTOs
// ============================================================================

// UpdateCustomerStatusRequest for updating customer status
type UpdateCustomerStatusRequest struct {
	Status CustomerStatus `json:"status" validate:"required,oneof=active inactive lead churned"`
}

// UpdateCustomerTierRequest for updating customer tier
type UpdateCustomerTierRequest struct {
	Tier CustomerTier `json:"tier" validate:"required,oneof=bronze silver gold platinum"`
}

// UpdateCustomerIndustryRequest for updating customer industry
type UpdateCustomerIndustryRequest struct {
	Industry CustomerIndustry `json:"industry" validate:"omitempty,oneof=construction manufacturing retail logistics agriculture energy public_sector real_estate other"`
}

// UpdateCustomerNotesRequest for updating customer notes
type UpdateCustomerNotesRequest struct {
	Notes string `json:"notes"`
}

// UpdateCustomerCompanyRequest for assigning customer to a company
type UpdateCustomerCompanyRequest struct {
	CompanyID *CompanyID `json:"companyId"` // nullable to allow unassignment
}

// UpdateCustomerClassRequest for updating customer class
type UpdateCustomerClassRequest struct {
	CustomerClass string `json:"customerClass" validate:"max=50"`
}

// UpdateCustomerCreditLimitRequest for updating customer credit limit
type UpdateCustomerCreditLimitRequest struct {
	CreditLimit *float64 `json:"creditLimit"` // nullable to allow clearing
}

// UpdateCustomerIsInternalRequest for updating customer internal flag
type UpdateCustomerIsInternalRequest struct {
	IsInternal bool `json:"isInternal"`
}

// UpdateCustomerAddressRequest for updating customer address fields
type UpdateCustomerAddressRequest struct {
	Address    string `json:"address" validate:"max=500"`
	City       string `json:"city" validate:"max=100"`
	PostalCode string `json:"postalCode" validate:"max=20"`
	Country    string `json:"country" validate:"max=100"`
}

// UpdateCustomerPostalCodeRequest for updating customer postal code only
type UpdateCustomerPostalCodeRequest struct {
	PostalCode string `json:"postalCode" validate:"max=20"`
}

// UpdateCustomerCityRequest for updating customer city only
type UpdateCustomerCityRequest struct {
	City string `json:"city" validate:"max=100"`
}

// UpdateCustomerContactInfoRequest for updating customer contact information
type UpdateCustomerContactInfoRequest struct {
	ContactPerson string `json:"contactPerson" validate:"max=200"`
	ContactEmail  string `json:"contactEmail" validate:"omitempty,email"`
	ContactPhone  string `json:"contactPhone" validate:"max=50"`
}

// ============================================================================
// Project Property Update Request DTOs
// ============================================================================

// UpdateProjectNameRequest for updating project name
type UpdateProjectNameRequest struct {
	Name string `json:"name" validate:"required,min=1,max=200"`
}

// UpdateProjectDescriptionRequest for updating project description and summary
type UpdateProjectDescriptionRequest struct {
	Summary     string `json:"summary" validate:"max=500"`
	Description string `json:"description"`
}

// UpdateProjectPhaseRequest for updating project phase
type UpdateProjectPhaseRequest struct {
	Phase ProjectPhase `json:"phase" validate:"required,oneof=tilbud working active completed cancelled"`
}

// UpdateProjectManagerRequest for updating project manager
type UpdateProjectManagerRequest struct {
	ManagerID string `json:"managerId" validate:"max=100"`
}

// UpdateProjectDatesRequest for updating project start and end dates
type UpdateProjectDatesRequest struct {
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

// UpdateProjectBudgetRequest for updating project budget (only in active phase)
type UpdateProjectBudgetRequest struct {
	Budget float64 `json:"budget" validate:"min=0"`
}

// UpdateProjectSpentRequest for updating project spent amount (only in active phase)
type UpdateProjectSpentRequest struct {
	Spent float64 `json:"spent" validate:"min=0"`
}

// UpdateProjectTeamMembersRequest for updating project team members
type UpdateProjectTeamMembersRequest struct {
	TeamMembers []string `json:"teamMembers"`
}

// UpdateProjectHealthRequest for updating project health
type UpdateProjectHealthRequest struct {
	Health ProjectHealth `json:"health" validate:"required,oneof=on_track at_risk over_budget"`
}

// UpdateProjectCompletionPercentRequest for updating project completion percentage
type UpdateProjectCompletionPercentRequest struct {
	CompletionPercent float64 `json:"completionPercent" validate:"min=0,max=100"`
}

// UpdateProjectEstimatedCompletionDateRequest for updating estimated completion date
type UpdateProjectEstimatedCompletionDateRequest struct {
	EstimatedCompletionDate *time.Time `json:"estimatedCompletionDate"`
}

// UpdateProjectNumberRequest for updating project number
type UpdateProjectNumberRequest struct {
	ProjectNumber string `json:"projectNumber" validate:"max=50"`
}

// UpdateProjectCompanyRequest for updating project company assignment
type UpdateProjectCompanyRequest struct {
	CompanyID CompanyID `json:"companyId" validate:"omitempty,oneof=gruppen stalbygg hybridbygg industri tak montasje"`
}

// ReopenProjectRequest for reopening a completed or cancelled project
type ReopenProjectRequest struct {
	TargetPhase ProjectPhase `json:"targetPhase" validate:"required,oneof=tilbud working"`
	Notes       string       `json:"notes" validate:"max=1000"`
}

// ReopenProjectResponse is the response when reopening a project
type ReopenProjectResponse struct {
	Project           *ProjectDTO `json:"project"`
	PreviousPhase     string      `json:"previousPhase"`
	RevertedOffer     *OfferDTO   `json:"revertedOffer,omitempty"` // Offer that was reverted to sent (if any)
	ClearedOfferID    bool        `json:"clearedOfferId"`          // Whether WinningOfferID was cleared
	ClearedOfferValue bool        `json:"clearedOfferValue"`       // Whether economic values were cleared
}
