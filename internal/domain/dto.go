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
	ID                  uuid.UUID      `json:"id"`
	Title               string         `json:"title"`
	OfferNumber         string         `json:"offerNumber,omitempty"` // Unique per company, e.g., "STB-2024-001"
	CustomerID          uuid.UUID      `json:"customerId"`
	CustomerName        string         `json:"customerName,omitempty"`
	ProjectID           *uuid.UUID     `json:"projectId,omitempty"` // Link to project (nullable)
	CompanyID           CompanyID      `json:"companyId"`
	Phase               OfferPhase     `json:"phase"`
	Probability         int            `json:"probability"`
	Value               float64        `json:"value"`
	Status              OfferStatus    `json:"status"`
	CreatedAt           string         `json:"createdAt"` // ISO 8601
	UpdatedAt           string         `json:"updatedAt"` // ISO 8601
	ResponsibleUserID   string         `json:"responsibleUserId,omitempty"`
	ResponsibleUserName string         `json:"responsibleUserName,omitempty"`
	Items               []OfferItemDTO `json:"items"`
	Description         string         `json:"description,omitempty"`
	Notes               string         `json:"notes,omitempty"`
	DueDate             *string        `json:"dueDate,omitempty"` // ISO 8601
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

// Budget Dimension DTOs

type BudgetDimensionCategoryDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	DisplayOrder int    `json:"displayOrder"`
	IsActive     bool   `json:"isActive"`
}

type BudgetDimensionDTO struct {
	ID                  uuid.UUID                   `json:"id"`
	ParentType          BudgetParentType            `json:"parentType"`
	ParentID            uuid.UUID                   `json:"parentId"`
	CategoryID          *string                     `json:"categoryId,omitempty"`
	Category            *BudgetDimensionCategoryDTO `json:"category,omitempty"`
	CustomName          string                      `json:"customName,omitempty"`
	Name                string                      `json:"name"` // Resolved name (category or custom)
	Cost                float64                     `json:"cost"`
	Revenue             float64                     `json:"revenue"`
	TargetMarginPercent *float64                    `json:"targetMarginPercent,omitempty"`
	MarginOverride      bool                        `json:"marginOverride"`
	MarginPercent       float64                     `json:"marginPercent"`
	Description         string                      `json:"description,omitempty"`
	Quantity            *float64                    `json:"quantity,omitempty"`
	Unit                string                      `json:"unit,omitempty"`
	DisplayOrder        int                         `json:"displayOrder"`
	CreatedAt           string                      `json:"createdAt"`
	UpdatedAt           string                      `json:"updatedAt"`
}

type BudgetSummaryDTO struct {
	ParentType           BudgetParentType `json:"parentType"`
	ParentID             uuid.UUID        `json:"parentId"`
	DimensionCount       int              `json:"dimensionCount"`
	TotalCost            float64          `json:"totalCost"`
	TotalRevenue         float64          `json:"totalRevenue"`
	OverallMarginPercent float64          `json:"overallMarginPercent"`
	TotalProfit          float64          `json:"totalProfit"`
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
	StartDate               string         `json:"startDate"`         // ISO 8601
	EndDate                 string         `json:"endDate,omitempty"` // ISO 8601
	Budget                  float64        `json:"budget"`
	Spent                   float64        `json:"spent"`
	ManagerID               string         `json:"managerId"`
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
}

// Project Actual Cost DTOs

type ProjectActualCostDTO struct {
	ID                uuid.UUID  `json:"id"`
	ProjectID         uuid.UUID  `json:"projectId"`
	CostType          CostType   `json:"costType"`
	Description       string     `json:"description"`
	Amount            float64    `json:"amount"`
	Currency          string     `json:"currency"`
	CostDate          string     `json:"costDate"`
	PostingDate       string     `json:"postingDate,omitempty"`
	BudgetDimensionID *uuid.UUID `json:"budgetDimensionId,omitempty"`
	ERPSource         ERPSource  `json:"erpSource"`
	ERPReference      string     `json:"erpReference,omitempty"`
	ERPTransactionID  string     `json:"erpTransactionId,omitempty"`
	ERPSyncedAt       string     `json:"erpSyncedAt,omitempty"`
	IsApproved        bool       `json:"isApproved"`
	ApprovedByID      string     `json:"approvedById,omitempty"`
	ApprovedAt        string     `json:"approvedAt,omitempty"`
	Notes             string     `json:"notes,omitempty"`
	CreatedAt         string     `json:"createdAt"`
	UpdatedAt         string     `json:"updatedAt"`
}

type ProjectCostSummaryDTO struct {
	ProjectID         uuid.UUID `json:"projectId"`
	ProjectName       string    `json:"projectName"`
	Budget            float64   `json:"budget"`
	Spent             float64   `json:"spent"`
	ActualCosts       float64   `json:"actualCosts"`
	RemainingBudget   float64   `json:"remainingBudget"`
	BudgetUsedPercent float64   `json:"budgetUsedPercent"`
	CostEntryCount    int       `json:"costEntryCount"`
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
	Count         int        `json:"count"`
	TotalValue    float64    `json:"totalValue"`
	WeightedValue float64    `json:"weightedValue"`
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
// All metrics use a rolling 12-month window from the current date
// Drafts and expired offers are excluded from all calculations
type DashboardMetrics struct {
	// Offer Metrics (12-month window, excluding drafts and expired)
	TotalOfferCount      int     `json:"totalOfferCount"`      // Count of offers excluding drafts and expired
	OfferReserve         float64 `json:"offerReserve"`         // Total value of active offers (in_progress, sent)
	WeightedOfferReserve float64 `json:"weightedOfferReserve"` // Sum of (value * probability/100) for active offers
	AverageProbability   float64 `json:"averageProbability"`   // Average probability of active offers

	// Pipeline Data (phases: in_progress, sent, won, lost - excludes draft and expired)
	Pipeline []PipelinePhaseData `json:"pipeline"`

	// Win Rate Metrics (12-month window)
	WinRateMetrics WinRateMetrics `json:"winRateMetrics"`

	// Order Reserve (from active projects)
	OrderReserve float64 `json:"orderReserve"` // Sum of (budget - spent) on active projects

	// Financial Summary
	TotalInvoiced float64 `json:"totalInvoiced"` // Sum of "spent" on all projects in 12-month window
	TotalValue    float64 `json:"totalValue"`    // orderReserve + totalInvoiced

	// Recent Lists (12-month window, limit 10 each)
	RecentOffers     []OfferDTO    `json:"recentOffers"`     // Last created offers (excluding drafts)
	RecentProjects   []ProjectDTO  `json:"recentProjects"`   // Last created projects
	RecentActivities []ActivityDTO `json:"recentActivities"` // Last activities

	// Top Customers (12-month window, limit 10)
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
	OrgNumber     string           `json:"orgNumber" validate:"required,max=20"`
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
}

type UpdateCustomerRequest struct {
	Name          string           `json:"name" validate:"required,max=200"`
	OrgNumber     string           `json:"orgNumber" validate:"required,max=20"`
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
	StartDate               time.Time      `json:"startDate" validate:"required"`
	EndDate                 *time.Time     `json:"endDate,omitempty"`
	Budget                  float64        `json:"budget" validate:"gte=0"`
	Spent                   float64        `json:"spent" validate:"gte=0"`
	ManagerID               string         `json:"managerId" validate:"required"`
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
	StartDate               time.Time      `json:"startDate" validate:"required"`
	EndDate                 *time.Time     `json:"endDate,omitempty"`
	Budget                  float64        `json:"budget" validate:"gte=0"`
	Spent                   float64        `json:"spent" validate:"gte=0"`
	ManagerID               string         `json:"managerId" validate:"required"`
	TeamMembers             []string       `json:"teamMembers,omitempty"`
	DealID                  *uuid.UUID     `json:"dealId,omitempty"`
	HasDetailedBudget       *bool          `json:"hasDetailedBudget,omitempty"`
	Health                  *ProjectHealth `json:"health,omitempty"`
	CompletionPercent       *float64       `json:"completionPercent,omitempty" validate:"omitempty,gte=0,lte=100"`
	EstimatedCompletionDate *time.Time     `json:"estimatedCompletionDate,omitempty"`
}

type CreateOfferRequest struct {
	Title             string                   `json:"title" validate:"required,max=200"`
	CustomerID        uuid.UUID                `json:"customerId" validate:"required"`
	CompanyID         CompanyID                `json:"companyId,omitempty"`
	Phase             OfferPhase               `json:"phase,omitempty"`
	Probability       *int                     `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`
	Status            OfferStatus              `json:"status,omitempty"`
	ResponsibleUserID string                   `json:"responsibleUserId,omitempty"`
	Items             []CreateOfferItemRequest `json:"items,omitempty"`
	Description       string                   `json:"description,omitempty"`
	Notes             string                   `json:"notes,omitempty"`
	DueDate           *time.Time               `json:"dueDate,omitempty"`
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

// Budget Dimension Request DTOs

type CreateBudgetDimensionRequest struct {
	ParentType          BudgetParentType `json:"parentType" validate:"required"`
	ParentID            uuid.UUID        `json:"parentId" validate:"required"`
	CategoryID          *string          `json:"categoryId,omitempty" validate:"omitempty,max=50"`
	CustomName          string           `json:"customName,omitempty" validate:"max=200"`
	Cost                float64          `json:"cost" validate:"gte=0"`
	Revenue             float64          `json:"revenue,omitempty" validate:"gte=0"`
	TargetMarginPercent *float64         `json:"targetMarginPercent,omitempty" validate:"omitempty,gte=0,lt=100"`
	MarginOverride      bool             `json:"marginOverride,omitempty"`
	Description         string           `json:"description,omitempty"`
	Quantity            *float64         `json:"quantity,omitempty" validate:"omitempty,gte=0"`
	Unit                string           `json:"unit,omitempty" validate:"max=50"`
	DisplayOrder        int              `json:"displayOrder,omitempty" validate:"gte=0"`
}

type UpdateBudgetDimensionRequest struct {
	CategoryID          *string  `json:"categoryId,omitempty" validate:"omitempty,max=50"`
	CustomName          string   `json:"customName,omitempty" validate:"max=200"`
	Cost                float64  `json:"cost" validate:"gte=0"`
	Revenue             float64  `json:"revenue,omitempty" validate:"gte=0"`
	TargetMarginPercent *float64 `json:"targetMarginPercent,omitempty" validate:"omitempty,gte=0,lt=100"`
	MarginOverride      bool     `json:"marginOverride,omitempty"`
	Description         string   `json:"description,omitempty"`
	Quantity            *float64 `json:"quantity,omitempty" validate:"omitempty,gte=0"`
	Unit                string   `json:"unit,omitempty" validate:"max=50"`
	DisplayOrder        int      `json:"displayOrder,omitempty" validate:"gte=0"`
}

// ReorderDimensionsRequest contains the ordered list of dimension IDs
type ReorderDimensionsRequest struct {
	OrderedIDs []uuid.UUID `json:"orderedIds" validate:"required,min=1"`
}

// AddOfferBudgetDimensionRequest is the simplified request for adding dimensions to an offer
// ParentType and ParentID are inferred from the URL
type AddOfferBudgetDimensionRequest struct {
	CategoryID          *string  `json:"categoryId,omitempty" validate:"omitempty,max=50"`
	CustomName          string   `json:"customName,omitempty" validate:"max=200"`
	Cost                float64  `json:"cost" validate:"gt=0"`
	Revenue             float64  `json:"revenue,omitempty" validate:"gte=0"`
	TargetMarginPercent *float64 `json:"targetMarginPercent,omitempty" validate:"omitempty,gte=0,lt=100"`
	MarginOverride      bool     `json:"marginOverride,omitempty"`
	Description         string   `json:"description,omitempty"`
	Quantity            *float64 `json:"quantity,omitempty" validate:"omitempty,gte=0"`
	Unit                string   `json:"unit,omitempty" validate:"max=50"`
	DisplayOrder        int      `json:"displayOrder,omitempty" validate:"gte=0"`
}

// Project Actual Cost Request DTOs

type CreateProjectActualCostRequest struct {
	ProjectID         uuid.UUID  `json:"projectId" validate:"required"`
	CostType          CostType   `json:"costType" validate:"required"`
	Description       string     `json:"description" validate:"required,max=500"`
	Amount            float64    `json:"amount" validate:"required"`
	Currency          string     `json:"currency,omitempty" validate:"max=3"`
	CostDate          time.Time  `json:"costDate" validate:"required"`
	PostingDate       *time.Time `json:"postingDate,omitempty"`
	BudgetDimensionID *uuid.UUID `json:"budgetDimensionId,omitempty"`
	ERPSource         ERPSource  `json:"erpSource,omitempty"`
	ERPReference      string     `json:"erpReference,omitempty" validate:"max=100"`
	ERPTransactionID  string     `json:"erpTransactionId,omitempty" validate:"max=100"`
	Notes             string     `json:"notes,omitempty"`
}

type UpdateProjectActualCostRequest struct {
	CostType          CostType   `json:"costType" validate:"required"`
	Description       string     `json:"description" validate:"required,max=500"`
	Amount            float64    `json:"amount" validate:"required"`
	Currency          string     `json:"currency,omitempty" validate:"max=3"`
	CostDate          time.Time  `json:"costDate" validate:"required"`
	PostingDate       *time.Time `json:"postingDate,omitempty"`
	BudgetDimensionID *uuid.UUID `json:"budgetDimensionId,omitempty"`
	ERPSource         ERPSource  `json:"erpSource,omitempty"`
	ERPReference      string     `json:"erpReference,omitempty" validate:"max=100"`
	ERPTransactionID  string     `json:"erpTransactionId,omitempty" validate:"max=100"`
	Notes             string     `json:"notes,omitempty"`
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
	Phase OfferPhase `json:"phase" validate:"required"`
}

// Offer Lifecycle DTOs

// CloneOfferRequest contains options for cloning an offer
type CloneOfferRequest struct {
	NewTitle          string `json:"newTitle,omitempty" validate:"max=200"`
	IncludeDimensions *bool  `json:"includeDimensions,omitempty"` // Default true - clone budget dimensions (nil treated as true)
	IncludeFiles      bool   `json:"includeFiles"`                // Default false - files are not cloned by default
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

// OfferDetailDTO includes offer with budget dimensions and summary
type OfferDetailDTO struct {
	OfferDTO
	BudgetDimensions []BudgetDimensionDTO `json:"budgetDimensions,omitempty"`
	BudgetSummary    *BudgetSummaryDTO    `json:"budgetSummary,omitempty"`
	FilesCount       int                  `json:"filesCount"`
}

type ProjectBudgetDTO struct {
	Budget      float64 `json:"budget"`
	Spent       float64 `json:"spent"`
	Remaining   float64 `json:"remaining"`
	PercentUsed float64 `json:"percentUsed"`
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
	Project         *ProjectDTO `json:"project"`
	DimensionsCount int         `json:"dimensionsCount"`
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

// UpdateOfferDescriptionRequest for updating offer description
type UpdateOfferDescriptionRequest struct {
	Description string `json:"description" validate:"max=10000"`
}

// UpdateOfferProjectRequest for linking offer to a project
type UpdateOfferProjectRequest struct {
	ProjectID uuid.UUID `json:"projectId" validate:"required"`
}
