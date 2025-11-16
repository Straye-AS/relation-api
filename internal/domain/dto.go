package domain

import (
	"time"

	"github.com/google/uuid"
)

// DTOs for API responses matching Norwegian spec

type CustomerDTO struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	OrgNumber     string    `json:"orgNumber"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address,omitempty"`
	City          string    `json:"city,omitempty"`
	PostalCode    string    `json:"postalCode,omitempty"`
	Country       string    `json:"country"`
	ContactPerson string    `json:"contactPerson,omitempty"`
	ContactEmail  string    `json:"contactEmail,omitempty"`
	ContactPhone  string    `json:"contactPhone,omitempty"`
	CreatedAt     string    `json:"createdAt"` // ISO 8601
	UpdatedAt     string    `json:"updatedAt"` // ISO 8601
	TotalValue    float64   `json:"totalValue,omitempty"`
	ActiveOffers  int       `json:"activeOffers,omitempty"`
}

type ContactDTO struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Role         string     `json:"role,omitempty"`
	CustomerID   *uuid.UUID `json:"customerId,omitempty"`
	CustomerName string     `json:"customerName,omitempty"`
	ProjectID    *uuid.UUID `json:"projectId,omitempty"`
	ProjectName  string     `json:"projectName,omitempty"`
	CreatedAt    string     `json:"createdAt"` // ISO 8601
	UpdatedAt    string     `json:"updatedAt"` // ISO 8601
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

type ProjectDTO struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	Summary         string        `json:"summary,omitempty"`
	Description     string        `json:"description,omitempty"`
	CustomerID      uuid.UUID     `json:"customerId"`
	CustomerName    string        `json:"customerName,omitempty"`
	CompanyID       CompanyID     `json:"companyId"`
	Status          ProjectStatus `json:"status"`
	StartDate       string        `json:"startDate"`         // ISO 8601
	EndDate         string        `json:"endDate,omitempty"` // ISO 8601
	Budget          float64       `json:"budget"`
	Spent           float64       `json:"spent"`
	ManagerID       string        `json:"managerId"`
	ManagerName     string        `json:"managerName,omitempty"`
	TeamMembers     []string      `json:"teamMembers,omitempty"`
	TeamsChannelID  string        `json:"teamsChannelId,omitempty"`
	TeamsChannelURL string        `json:"teamsChannelUrl,omitempty"`
	CreatedAt       string        `json:"createdAt"` // ISO 8601
	UpdatedAt       string        `json:"updatedAt"` // ISO 8601
	OfferID         *uuid.UUID    `json:"offerId,omitempty"`
}

type UserDTO struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	Department string   `json:"department,omitempty"`
	Avatar     string   `json:"avatar,omitempty"`
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
	Name          string `json:"name" validate:"required,max=200"`
	OrgNumber     string `json:"orgNumber" validate:"required,max=20"`
	Email         string `json:"email" validate:"required,email"`
	Phone         string `json:"phone" validate:"required,max=50"`
	Address       string `json:"address,omitempty" validate:"max=500"`
	City          string `json:"city,omitempty" validate:"max=100"`
	PostalCode    string `json:"postalCode,omitempty" validate:"max=20"`
	Country       string `json:"country" validate:"required,max=100"`
	ContactPerson string `json:"contactPerson,omitempty" validate:"max=200"`
	ContactEmail  string `json:"contactEmail,omitempty" validate:"omitempty,email"`
	ContactPhone  string `json:"contactPhone,omitempty" validate:"max=50"`
}

type UpdateCustomerRequest struct {
	Name          string `json:"name" validate:"required,max=200"`
	OrgNumber     string `json:"orgNumber" validate:"required,max=20"`
	Email         string `json:"email" validate:"required,email"`
	Phone         string `json:"phone" validate:"required,max=50"`
	Address       string `json:"address,omitempty" validate:"max=500"`
	City          string `json:"city,omitempty" validate:"max=100"`
	PostalCode    string `json:"postalCode,omitempty" validate:"max=20"`
	Country       string `json:"country" validate:"required,max=100"`
	ContactPerson string `json:"contactPerson,omitempty" validate:"max=200"`
	ContactEmail  string `json:"contactEmail,omitempty" validate:"omitempty,email"`
	ContactPhone  string `json:"contactPhone,omitempty" validate:"max=50"`
}

type CreateContactRequest struct {
	Name       string     `json:"name" validate:"required,max=200"`
	Email      string     `json:"email" validate:"required,email"`
	Phone      string     `json:"phone" validate:"required,max=50"`
	Role       string     `json:"role,omitempty" validate:"max=120"`
	CustomerID *uuid.UUID `json:"customerId,omitempty"`
	ProjectID  *uuid.UUID `json:"projectId,omitempty"`
}

type UpdateContactRequest struct {
	Name       string     `json:"name" validate:"required,max=200"`
	Email      string     `json:"email" validate:"required,email"`
	Phone      string     `json:"phone" validate:"required,max=50"`
	Role       string     `json:"role,omitempty" validate:"max=120"`
	CustomerID *uuid.UUID `json:"customerId,omitempty"`
	ProjectID  *uuid.UUID `json:"projectId,omitempty"`
}

type CreateProjectRequest struct {
	Name            string        `json:"name" validate:"required,max=200"`
	Summary         string        `json:"summary,omitempty"`
	Description     string        `json:"description,omitempty"`
	CustomerID      uuid.UUID     `json:"customerId" validate:"required"`
	CompanyID       CompanyID     `json:"companyId" validate:"required"`
	Status          ProjectStatus `json:"status" validate:"required"`
	StartDate       time.Time     `json:"startDate" validate:"required"`
	EndDate         *time.Time    `json:"endDate,omitempty"`
	Budget          float64       `json:"budget" validate:"gte=0"`
	Spent           float64       `json:"spent" validate:"gte=0"`
	ManagerID       string        `json:"managerId" validate:"required"`
	TeamMembers     []string      `json:"teamMembers,omitempty"`
	TeamsChannelID  string        `json:"teamsChannelId,omitempty"`
	TeamsChannelURL string        `json:"teamsChannelUrl,omitempty"`
	OfferID         *uuid.UUID    `json:"offerId,omitempty"`
}

type UpdateProjectRequest struct {
	Name            string        `json:"name" validate:"required,max=200"`
	Summary         string        `json:"summary,omitempty"`
	Description     string        `json:"description,omitempty"`
	CompanyID       CompanyID     `json:"companyId" validate:"required"`
	Status          ProjectStatus `json:"status" validate:"required"`
	StartDate       time.Time     `json:"startDate" validate:"required"`
	EndDate         *time.Time    `json:"endDate,omitempty"`
	Budget          float64       `json:"budget" validate:"gte=0"`
	Spent           float64       `json:"spent" validate:"gte=0"`
	ManagerID       string        `json:"managerId" validate:"required"`
	TeamMembers     []string      `json:"teamMembers,omitempty"`
	TeamsChannelID  string        `json:"teamsChannelId,omitempty"`
	TeamsChannelURL string        `json:"teamsChannelUrl,omitempty"`
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

type ProjectBudgetDTO struct {
	Budget      float64 `json:"budget"`
	Spent       float64 `json:"spent"`
	Remaining   float64 `json:"remaining"`
	PercentUsed float64 `json:"percentUsed"`
}

type ActivityDTO struct {
	ID          uuid.UUID    `json:"id"`
	TargetType  ActivityType `json:"targetType"`
	TargetID    uuid.UUID    `json:"targetId"`
	Title       string       `json:"title"`
	Body        string       `json:"body,omitempty"`
	OccurredAt  string       `json:"occurredAt"`
	CreatorName string       `json:"creatorName,omitempty"`
	CreatedAt   string       `json:"createdAt"`
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
