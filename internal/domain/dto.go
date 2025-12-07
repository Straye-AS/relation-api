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
	ID                uuid.UUID  `json:"id"`
	Title             string     `json:"title"`
	Description       string     `json:"description,omitempty"`
	CustomerID        uuid.UUID  `json:"customerId"`
	CustomerName      string     `json:"customerName,omitempty"`
	CompanyID         CompanyID  `json:"companyId"`
	Stage             DealStage  `json:"stage"`
	Probability       int        `json:"probability"`
	Value             float64    `json:"value"`
	WeightedValue     float64    `json:"weightedValue"`
	Currency          string     `json:"currency"`
	ExpectedCloseDate *string    `json:"expectedCloseDate,omitempty"`
	ActualCloseDate   *string    `json:"actualCloseDate,omitempty"`
	OwnerID           string     `json:"ownerId"`
	OwnerName         string     `json:"ownerName,omitempty"`
	Source            string     `json:"source,omitempty"`
	Notes             string     `json:"notes,omitempty"`
	LostReason        string     `json:"lostReason,omitempty"`
	OfferID           *uuid.UUID `json:"offerId,omitempty"`
	CreatedAt         string     `json:"createdAt"`
	UpdatedAt         string     `json:"updatedAt"`
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
	CustomerID          uuid.UUID      `json:"customerId"`
	CustomerName        string         `json:"customerName,omitempty"`
	CompanyID           CompanyID      `json:"companyId"`
	Phase               OfferPhase     `json:"phase"`
	Probability         int            `json:"probability"`
	Value               float64        `json:"value"`
	Status              OfferStatus    `json:"status"`
	CreatedAt           string         `json:"createdAt"` // ISO 8601
	UpdatedAt           string         `json:"updatedAt"` // ISO 8601
	ResponsibleUserID   string         `json:"responsibleUserId"`
	ResponsibleUserName string         `json:"responsibleUserName,omitempty"`
	Items               []OfferItemDTO `json:"items"`
	Description         string         `json:"description,omitempty"`
	Notes               string         `json:"notes,omitempty"`
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

// CompanyDTO represents a company in auth context
type CompanyDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
	Offers        []OfferDTO `json:"offers"`
}

type DashboardMetrics struct {
	TotalOffers           int                 `json:"totalOffers"`
	ActiveOffers          int                 `json:"activeOffers"`
	WonOffers             int                 `json:"wonOffers"`
	LostOffers            int                 `json:"lostOffers"`
	TotalValue            float64             `json:"totalValue"`
	WeightedValue         float64             `json:"weightedValue"`
	AverageProbability    float64             `json:"averageProbability"`
	OffersByPhase         map[OfferPhase]int  `json:"offersByPhase"`
	Pipeline              []PipelinePhaseData `json:"pipeline"`
	OfferReserve          float64             `json:"offerReserve"`
	WinRate               float64             `json:"winRate"`
	RevenueForecast30Days float64             `json:"revenueForecast30Days"`
	RevenueForecast90Days float64             `json:"revenueForecast90Days"`
	TopDisciplines        []DisciplineStats   `json:"topDisciplines"`
	ActiveProjects        []ProjectDTO        `json:"activeProjects"`
	TopCustomers          []CustomerDTO       `json:"topCustomers"`
	TeamPerformance       []TeamMemberStats   `json:"teamPerformance"`
	RecentOffers          []OfferDTO          `json:"recentOffers"`
	RecentProjects        []ProjectDTO        `json:"recentProjects"`
	RecentActivities      []NotificationDTO   `json:"recentActivities"`
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
	Email         string           `json:"email" validate:"required,email"`
	Phone         string           `json:"phone" validate:"required,max=50"`
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
	Email         string           `json:"email" validate:"required,email"`
	Phone         string           `json:"phone" validate:"required,max=50"`
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

type CreateProjectRequest struct {
	Name                    string         `json:"name" validate:"required,max=200"`
	ProjectNumber           string         `json:"projectNumber,omitempty" validate:"max=50"`
	Summary                 string         `json:"summary,omitempty"`
	Description             string         `json:"description,omitempty"`
	CustomerID              uuid.UUID      `json:"customerId" validate:"required"`
	CompanyID               CompanyID      `json:"companyId" validate:"required"`
	Status                  ProjectStatus  `json:"status" validate:"required"`
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
	CompanyID         CompanyID                `json:"companyId" validate:"required"`
	Phase             OfferPhase               `json:"phase" validate:"required"`
	Probability       int                      `json:"probability" validate:"min=0,max=100"`
	Status            OfferStatus              `json:"status" validate:"required"`
	ResponsibleUserID string                   `json:"responsibleUserId" validate:"required"`
	Items             []CreateOfferItemRequest `json:"items" validate:"required,min=1"`
	Description       string                   `json:"description,omitempty"`
	Notes             string                   `json:"notes,omitempty"`
}

type UpdateOfferRequest struct {
	Title             string      `json:"title" validate:"required,max=200"`
	Phase             OfferPhase  `json:"phase" validate:"required"`
	Probability       int         `json:"probability" validate:"min=0,max=100"`
	Status            OfferStatus `json:"status" validate:"required"`
	ResponsibleUserID string      `json:"responsibleUserId" validate:"required"`
	Description       string      `json:"description,omitempty"`
	Notes             string      `json:"notes,omitempty"`
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
	ID              uuid.UUID          `json:"id"`
	TargetType      ActivityTargetType `json:"targetType"`
	TargetID        uuid.UUID          `json:"targetId"`
	Title           string             `json:"title"`
	Body            string             `json:"body,omitempty"`
	OccurredAt      string             `json:"occurredAt"`
	CreatorName     string             `json:"creatorName,omitempty"`
	CreatedAt       string             `json:"createdAt"`
	ActivityType    ActivityType       `json:"activityType"`
	Status          ActivityStatus     `json:"status"`
	ScheduledAt     string             `json:"scheduledAt,omitempty"`
	DueDate         string             `json:"dueDate,omitempty"`
	CompletedAt     string             `json:"completedAt,omitempty"`
	DurationMinutes *int               `json:"durationMinutes,omitempty"`
	Priority        int                `json:"priority"`
	IsPrivate       bool               `json:"isPrivate"`
	CreatorID       string             `json:"creatorId,omitempty"`
	AssignedToID    string             `json:"assignedToId,omitempty"`
	CompanyID       *CompanyID         `json:"companyId,omitempty"`
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

// CustomerWithStatsResponse extends CustomerDTO with statistics
type CustomerWithStatsResponse struct {
	CustomerDTO
	Stats CustomerStatsResponse `json:"stats"`
}
