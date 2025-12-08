package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Base model with common fields
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// CustomerStatus represents the status of a customer
type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusLead     CustomerStatus = "lead"
	CustomerStatusChurned  CustomerStatus = "churned"
)

// CustomerTier represents the tier/importance level of a customer
type CustomerTier string

const (
	CustomerTierBronze   CustomerTier = "bronze"
	CustomerTierSilver   CustomerTier = "silver"
	CustomerTierGold     CustomerTier = "gold"
	CustomerTierPlatinum CustomerTier = "platinum"
)

// CustomerIndustry represents the industry sector of a customer
type CustomerIndustry string

const (
	CustomerIndustryConstruction  CustomerIndustry = "construction"
	CustomerIndustryManufacturing CustomerIndustry = "manufacturing"
	CustomerIndustryRetail        CustomerIndustry = "retail"
	CustomerIndustryLogistics     CustomerIndustry = "logistics"
	CustomerIndustryAgriculture   CustomerIndustry = "agriculture"
	CustomerIndustryEnergy        CustomerIndustry = "energy"
	CustomerIndustryPublicSector  CustomerIndustry = "public_sector"
	CustomerIndustryRealEstate    CustomerIndustry = "real_estate"
	CustomerIndustryOther         CustomerIndustry = "other"
)

// CompanyID represents Straye group companies
type CompanyID string

const (
	CompanyAll        CompanyID = "all"
	CompanyGruppen    CompanyID = "gruppen"
	CompanyStalbygg   CompanyID = "stalbygg"
	CompanyHybridbygg CompanyID = "hybridbygg"
	CompanyIndustri   CompanyID = "industri"
	CompanyTak        CompanyID = "tak"
	CompanyMontasje   CompanyID = "montasje"
)

// Company represents a Straye group company (stored in database)
type Company struct {
	ID        CompanyID `gorm:"type:varchar(50);primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(200);not null" json:"name"`
	ShortName string    `gorm:"type:varchar(50);not null;column:short_name" json:"shortName"`
	OrgNumber string    `gorm:"type:varchar(20);column:org_number" json:"orgNumber,omitempty"`
	Color     string    `gorm:"type:varchar(20);not null;default:'#000000'" json:"color"`
	Logo      string    `gorm:"type:varchar(500)" json:"logo,omitempty"`
	IsActive  bool      `gorm:"not null;default:true;column:is_active" json:"isActive"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// Customer represents an organization in the CRM
type Customer struct {
	BaseModel
	Name          string           `gorm:"type:varchar(200);not null;index"`
	OrgNumber     string           `gorm:"type:varchar(20);unique;index"`
	Email         string           `gorm:"type:varchar(255);not null"`
	Phone         string           `gorm:"type:varchar(50);not null"`
	Address       string           `gorm:"type:varchar(500)"`
	City          string           `gorm:"type:varchar(100)"`
	PostalCode    string           `gorm:"type:varchar(20)"`
	Country       string           `gorm:"type:varchar(100);not null;default:'Norway'"`
	ContactPerson string           `gorm:"type:varchar(200)"`
	ContactEmail  string           `gorm:"type:varchar(255)"`
	ContactPhone  string           `gorm:"type:varchar(50)"`
	Status        CustomerStatus   `gorm:"type:varchar(50);not null;default:'active';index"`
	Tier          CustomerTier     `gorm:"type:varchar(50);not null;default:'bronze';index"`
	Industry      CustomerIndustry `gorm:"type:varchar(50);index"`
	CompanyID     *CompanyID       `gorm:"type:varchar(50);column:company_id;index"`
	Company       *Company         `gorm:"foreignKey:CompanyID"`
	Contacts      []Contact        `gorm:"foreignKey:PrimaryCustomerID;constraint:OnDelete:CASCADE"`
	Projects      []Project        `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"`
	Offers        []Offer          `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"`
}

// ContactType represents the classification of a contact
type ContactType string

const (
	ContactTypePrimary   ContactType = "primary"
	ContactTypeSecondary ContactType = "secondary"
	ContactTypeBilling   ContactType = "billing"
	ContactTypeTechnical ContactType = "technical"
	ContactTypeExecutive ContactType = "executive"
	ContactTypeOther     ContactType = "other"
)

// Contact represents an individual person
type Contact struct {
	BaseModel
	FirstName              string                `gorm:"type:varchar(100);not null;column:first_name"`
	LastName               string                `gorm:"type:varchar(100);not null;column:last_name"`
	Email                  string                `gorm:"type:varchar(255);uniqueIndex"`
	Phone                  string                `gorm:"type:varchar(50)"`
	Mobile                 string                `gorm:"type:varchar(50)"`
	Title                  string                `gorm:"type:varchar(100)"`
	Department             string                `gorm:"type:varchar(100)"`
	ContactType            ContactType           `gorm:"type:varchar(50);not null;default:'primary';column:contact_type;index"`
	PrimaryCustomerID      *uuid.UUID            `gorm:"type:uuid;column:primary_customer_id"`
	PrimaryCustomer        *Customer             `gorm:"foreignKey:PrimaryCustomerID"`
	Address                string                `gorm:"type:varchar(500)"`
	City                   string                `gorm:"type:varchar(100)"`
	PostalCode             string                `gorm:"type:varchar(20)"`
	Country                string                `gorm:"type:varchar(100);default:'Norway'"`
	LinkedInURL            string                `gorm:"type:varchar(500);column:linkedin_url"`
	PreferredContactMethod string                `gorm:"type:varchar(50);default:'email';column:preferred_contact_method"`
	Notes                  string                `gorm:"type:text"`
	IsActive               bool                  `gorm:"not null;default:true;column:is_active"`
	Relationships          []ContactRelationship `gorm:"foreignKey:ContactID"`
}

// FullName returns the contact's full name
func (c *Contact) FullName() string {
	return c.FirstName + " " + c.LastName
}

// ContactEntityType represents the type of entity a contact can be related to
type ContactEntityType string

const (
	ContactEntityCustomer ContactEntityType = "customer"
	ContactEntityProject  ContactEntityType = "project"
	ContactEntityDeal     ContactEntityType = "deal"
)

// ContactRelationship represents a polymorphic relationship between a contact and an entity
type ContactRelationship struct {
	ID         uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactID  uuid.UUID         `gorm:"type:uuid;not null;index;column:contact_id"`
	Contact    *Contact          `gorm:"foreignKey:ContactID"`
	EntityType ContactEntityType `gorm:"type:varchar(50);not null;column:entity_type"`
	EntityID   uuid.UUID         `gorm:"type:uuid;not null;column:entity_id"`
	Role       string            `gorm:"type:varchar(100)"`
	IsPrimary  bool              `gorm:"not null;default:false;column:is_primary"`
	CreatedAt  time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// DealStage represents the stage of a deal in the sales pipeline
type DealStage string

const (
	DealStageLead        DealStage = "lead"
	DealStageQualified   DealStage = "qualified"
	DealStageProposal    DealStage = "proposal"
	DealStageNegotiation DealStage = "negotiation"
	DealStageWon         DealStage = "won"
	DealStageLost        DealStage = "lost"
)

// LossReasonCategory represents the categorized reason for losing a deal
type LossReasonCategory string

const (
	LossReasonPrice        LossReasonCategory = "price"
	LossReasonTiming       LossReasonCategory = "timing"
	LossReasonCompetitor   LossReasonCategory = "competitor"
	LossReasonRequirements LossReasonCategory = "requirements"
	LossReasonOther        LossReasonCategory = "other"
)

// IsValid checks if the LossReasonCategory is a valid enum value
func (lrc LossReasonCategory) IsValid() bool {
	switch lrc {
	case LossReasonPrice, LossReasonTiming, LossReasonCompetitor, LossReasonRequirements, LossReasonOther:
		return true
	}
	return false
}

// Deal represents a sales opportunity in the pipeline
type Deal struct {
	BaseModel
	Title              string              `gorm:"type:varchar(200);not null"`
	Description        string              `gorm:"type:text"`
	CustomerID         uuid.UUID           `gorm:"type:uuid;not null;index;column:customer_id"`
	Customer           *Customer           `gorm:"foreignKey:CustomerID"`
	CompanyID          CompanyID           `gorm:"type:varchar(50);not null;index;column:company_id"`
	Company            *Company            `gorm:"foreignKey:CompanyID"`
	CustomerName       string              `gorm:"type:varchar(200);column:customer_name"`
	Stage              DealStage           `gorm:"type:varchar(50);not null;default:'lead'"`
	Probability        int                 `gorm:"type:int;not null;default:0"`
	Value              float64             `gorm:"type:decimal(15,2);not null;default:0"`
	WeightedValue      float64             `gorm:"type:decimal(15,2);column:weighted_value;->"` // Read-only, computed by DB
	Currency           string              `gorm:"type:varchar(3);not null;default:'NOK'"`
	ExpectedCloseDate  *time.Time          `gorm:"type:date;column:expected_close_date"`
	ActualCloseDate    *time.Time          `gorm:"type:date;column:actual_close_date"`
	OwnerID            string              `gorm:"type:varchar(100);not null;column:owner_id"`
	OwnerName          string              `gorm:"type:varchar(200);column:owner_name"`
	Source             string              `gorm:"type:varchar(100)"`
	Notes              string              `gorm:"type:text"`
	LostReason         string              `gorm:"type:varchar(500);column:lost_reason"`
	LossReasonCategory *LossReasonCategory `gorm:"type:varchar(50);column:loss_reason_category"`
	OfferID            *uuid.UUID          `gorm:"type:uuid;index;column:offer_id"`
	Offer              *Offer              `gorm:"foreignKey:OfferID"`
}

// DealStageHistory tracks stage changes for audit purposes
type DealStageHistory struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DealID        uuid.UUID  `gorm:"type:uuid;not null;index;column:deal_id"`
	Deal          *Deal      `gorm:"foreignKey:DealID"`
	FromStage     *DealStage `gorm:"type:varchar(50);column:from_stage"`
	ToStage       DealStage  `gorm:"type:varchar(50);not null;column:to_stage"`
	ChangedByID   string     `gorm:"type:varchar(100);not null;column:changed_by_id"`
	ChangedByName string     `gorm:"type:varchar(200);column:changed_by_name"`
	Notes         string     `gorm:"type:text"`
	ChangedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;column:changed_at"`
}

// TableName overrides the default table name to match the migration
func (DealStageHistory) TableName() string {
	return "deal_stage_history"
}

// ProjectStatus represents the status of a project
type ProjectStatus string

const (
	ProjectStatusPlanning  ProjectStatus = "planning"
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusCancelled ProjectStatus = "cancelled"
)

// ProjectHealth represents the health status of a project
type ProjectHealth string

const (
	ProjectHealthOnTrack    ProjectHealth = "on_track"
	ProjectHealthAtRisk     ProjectHealth = "at_risk"
	ProjectHealthDelayed    ProjectHealth = "delayed"
	ProjectHealthOverBudget ProjectHealth = "over_budget"
)

// Project represents work being performed for a customer
type Project struct {
	BaseModel
	Name                    string         `gorm:"type:varchar(200);not null;index"`
	ProjectNumber           string         `gorm:"type:varchar(50);unique;index;column:project_number"` // External reference number for ERP/accounting systems
	Summary                 string         `gorm:"type:varchar(500)"`
	Description             string         `gorm:"type:text"`
	CustomerID              uuid.UUID      `gorm:"type:uuid;not null;index"`
	Customer                *Customer      `gorm:"foreignKey:CustomerID"`
	CustomerName            string         `gorm:"type:varchar(200)"`
	CompanyID               CompanyID      `gorm:"type:varchar(50);not null;index"`
	Status                  ProjectStatus  `gorm:"type:varchar(50);not null;index"`
	StartDate               time.Time      `gorm:"type:date;not null"`
	EndDate                 *time.Time     `gorm:"type:date"`
	Budget                  float64        `gorm:"type:decimal(15,2);not null;default:0"`
	Spent                   float64        `gorm:"type:decimal(15,2);not null;default:0"`
	ManagerID               string         `gorm:"type:varchar(100);not null"`
	ManagerName             string         `gorm:"type:varchar(200)"`
	TeamMembers             pq.StringArray `gorm:"type:text[]"`
	OfferID                 *uuid.UUID     `gorm:"type:uuid;index"`
	Offer                   *Offer         `gorm:"foreignKey:OfferID"`
	DealID                  *uuid.UUID     `gorm:"type:uuid;index;column:deal_id"`
	Deal                    *Deal          `gorm:"foreignKey:DealID"`
	HasDetailedBudget       bool           `gorm:"not null;default:false;column:has_detailed_budget"`
	Health                  *ProjectHealth `gorm:"type:project_health;default:'on_track'"`
	CompletionPercent       *float64       `gorm:"type:decimal(5,2);default:0;column:completion_percent"`
	EstimatedCompletionDate *time.Time     `gorm:"type:date;column:estimated_completion_date"`
}

// ERPSource represents the source ERP system for cost data
type ERPSource string

const (
	ERPSourceManual      ERPSource = "manual"
	ERPSourceTripletex   ERPSource = "tripletex"
	ERPSourceVisma       ERPSource = "visma"
	ERPSourcePowerOffice ERPSource = "poweroffice"
	ERPSourceOther       ERPSource = "other"
)

// CostType represents the type of cost entry
type CostType string

const (
	CostTypeLabor         CostType = "labor"
	CostTypeMaterials     CostType = "materials"
	CostTypeEquipment     CostType = "equipment"
	CostTypeSubcontractor CostType = "subcontractor"
	CostTypeTravel        CostType = "travel"
	CostTypeOverhead      CostType = "overhead"
	CostTypeOther         CostType = "other"
)

// ProjectActualCost represents an actual cost entry for a project (from ERP or manual)
type ProjectActualCost struct {
	ID                uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID         uuid.UUID        `gorm:"type:uuid;not null;index;column:project_id"`
	Project           *Project         `gorm:"foreignKey:ProjectID"`
	CostType          CostType         `gorm:"type:cost_type;not null;column:cost_type"`
	Description       string           `gorm:"type:varchar(500);not null"`
	Amount            float64          `gorm:"type:decimal(15,2);not null"`
	Currency          string           `gorm:"type:varchar(3);not null;default:'NOK'"`
	CostDate          time.Time        `gorm:"type:date;not null;column:cost_date"`
	PostingDate       *time.Time       `gorm:"type:date;column:posting_date"`
	BudgetDimensionID *uuid.UUID       `gorm:"type:uuid;column:budget_dimension_id"`
	BudgetDimension   *BudgetDimension `gorm:"foreignKey:BudgetDimensionID"`
	ERPSource         ERPSource        `gorm:"type:erp_source;not null;default:'manual';column:erp_source"`
	ERPReference      string           `gorm:"type:varchar(100);column:erp_reference"`
	ERPTransactionID  string           `gorm:"type:varchar(100);column:erp_transaction_id"`
	ERPSyncedAt       *time.Time       `gorm:"column:erp_synced_at"`
	IsApproved        bool             `gorm:"not null;default:false;column:is_approved"`
	ApprovedByID      string           `gorm:"type:varchar(100);column:approved_by_id"`
	ApprovedAt        *time.Time       `gorm:"column:approved_at"`
	Notes             string           `gorm:"type:text"`
	CreatedAt         time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// OfferPhase represents the phase of an offer in the sales pipeline
type OfferPhase string

const (
	OfferPhaseDraft      OfferPhase = "draft"
	OfferPhaseInProgress OfferPhase = "in_progress"
	OfferPhaseSent       OfferPhase = "sent"
	OfferPhaseWon        OfferPhase = "won"
	OfferPhaseLost       OfferPhase = "lost"
	OfferPhaseExpired    OfferPhase = "expired"
)

// OfferStatus represents the status of an offer
type OfferStatus string

const (
	OfferStatusActive   OfferStatus = "active"
	OfferStatusInactive OfferStatus = "inactive"
	OfferStatusArchived OfferStatus = "archived"
)

// Offer represents a sales proposal
type Offer struct {
	BaseModel
	Title               string      `gorm:"type:varchar(200);not null;index"`
	CustomerID          uuid.UUID   `gorm:"type:uuid;not null;index"`
	Customer            *Customer   `gorm:"foreignKey:CustomerID"`
	CustomerName        string      `gorm:"type:varchar(200)"`
	CompanyID           CompanyID   `gorm:"type:varchar(50);not null;index"`
	Phase               OfferPhase  `gorm:"type:varchar(50);not null;index"`
	Probability         int         `gorm:"type:int;not null;default:0"`
	Value               float64     `gorm:"type:decimal(15,2);not null;default:0"`
	Status              OfferStatus `gorm:"type:varchar(50);not null;index"`
	ResponsibleUserID   string      `gorm:"type:varchar(100);not null;index"`
	ResponsibleUserName string      `gorm:"type:varchar(200)"`
	Description         string      `gorm:"type:text"`
	Notes               string      `gorm:"type:text"`
	Items               []OfferItem `gorm:"foreignKey:OfferID;constraint:OnDelete:CASCADE"`
	Files               []File      `gorm:"foreignKey:OfferID"`
}

// OfferItem represents a line item in an offer (legacy - being replaced by BudgetDimension)
type OfferItem struct {
	BaseModel
	OfferID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Offer       *Offer    `gorm:"foreignKey:OfferID"`
	Discipline  string    `gorm:"type:varchar(200);not null"`
	Cost        float64   `gorm:"type:decimal(15,2);not null"`
	Revenue     float64   `gorm:"type:decimal(15,2);not null"`
	Margin      float64   `gorm:"type:decimal(5,2);not null"`
	Description string    `gorm:"type:text"`
	Quantity    float64   `gorm:"type:decimal(10,2)"`
	Unit        string    `gorm:"type:varchar(50)"`
}

// BudgetDimensionCategory represents a predefined budget line type
// Categories can be company-specific (CompanyID set) or global (CompanyID null)
type BudgetDimensionCategory struct {
	ID           string     `gorm:"type:varchar(50);primaryKey" json:"id"`
	CompanyID    *CompanyID `gorm:"type:varchar(50);column:company_id;index" json:"companyId,omitempty"` // nil = global category available to all companies
	Name         string     `gorm:"type:varchar(200);not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description,omitempty"`
	DisplayOrder int        `gorm:"not null;default:0;column:display_order" json:"displayOrder"`
	IsActive     bool       `gorm:"not null;default:true;column:is_active" json:"isActive"`
	CreatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns the table name for BudgetDimensionCategory
func (BudgetDimensionCategory) TableName() string {
	return "budget_dimension_categories"
}

// BudgetCategoryID represents known budget dimension category IDs
// These are reference constants - actual seed data is loaded separately
type BudgetCategoryID string

const (
	BudgetCategorySteelStructure  BudgetCategoryID = "steel_structure"
	BudgetCategoryHybridStructure BudgetCategoryID = "hybrid_structure"
	BudgetCategoryRoofing         BudgetCategoryID = "roofing"
	BudgetCategoryCladding        BudgetCategoryID = "cladding"
	BudgetCategoryFoundation      BudgetCategoryID = "foundation"
	BudgetCategoryAssembly        BudgetCategoryID = "assembly"
	BudgetCategoryTransport       BudgetCategoryID = "transport"
	BudgetCategoryEngineering     BudgetCategoryID = "engineering"
	BudgetCategoryProjectMgmt     BudgetCategoryID = "project_management"
	BudgetCategoryCraneRigging    BudgetCategoryID = "crane_rigging"
	BudgetCategoryMiscellaneous   BudgetCategoryID = "miscellaneous"
	BudgetCategoryContingency     BudgetCategoryID = "contingency"
)

// BudgetParentType represents the type of entity a budget dimension belongs to
type BudgetParentType string

const (
	BudgetParentOffer   BudgetParentType = "offer"
	BudgetParentProject BudgetParentType = "project"
)

// BudgetDimension represents a budget line item that can belong to an offer or project
type BudgetDimension struct {
	ID                  uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ParentType          BudgetParentType         `gorm:"type:budget_parent_type;not null;column:parent_type" json:"parentType"`
	ParentID            uuid.UUID                `gorm:"type:uuid;not null;column:parent_id" json:"parentId"`
	CategoryID          *string                  `gorm:"type:varchar(50);column:category_id" json:"categoryId,omitempty"`
	Category            *BudgetDimensionCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CustomName          string                   `gorm:"type:varchar(200);column:custom_name" json:"customName,omitempty"`
	Cost                float64                  `gorm:"type:decimal(15,2);not null;default:0" json:"cost"`
	Revenue             float64                  `gorm:"type:decimal(15,2);not null;default:0" json:"revenue"`
	TargetMarginPercent *float64                 `gorm:"type:decimal(5,2);column:target_margin_percent" json:"targetMarginPercent,omitempty"`
	MarginOverride      bool                     `gorm:"not null;default:false;column:margin_override" json:"marginOverride"`
	MarginPercent       float64                  `gorm:"type:decimal(5,2);column:margin_percent;->" json:"marginPercent"` // Read-only, computed by DB
	Description         string                   `gorm:"type:text" json:"description,omitempty"`
	Quantity            *float64                 `gorm:"type:decimal(10,2)" json:"quantity,omitempty"`
	Unit                string                   `gorm:"type:varchar(50)" json:"unit,omitempty"`
	DisplayOrder        int                      `gorm:"not null;default:0;column:display_order" json:"displayOrder"`
	CreatedAt           time.Time                `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt           time.Time                `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// GetName returns the display name (category name or custom name)
func (bd *BudgetDimension) GetName() string {
	if bd.Category != nil {
		return bd.Category.Name
	}
	return bd.CustomName
}

// BudgetSummary holds aggregated budget totals for a parent entity (offer or project)
type BudgetSummary struct {
	TotalCost      float64 `json:"totalCost"`
	TotalRevenue   float64 `json:"totalRevenue"`
	TotalMargin    float64 `json:"totalMargin"`   // Revenue - Cost
	MarginPercent  float64 `json:"marginPercent"` // ((Revenue - Cost) / Revenue) * 100, 0 if revenue=0
	DimensionCount int     `json:"dimensionCount"`
}

// ActivityTargetType represents the type of entity an activity is associated with
type ActivityTargetType string

const (
	ActivityTargetCustomer     ActivityTargetType = "Customer"
	ActivityTargetContact      ActivityTargetType = "Contact"
	ActivityTargetProject      ActivityTargetType = "Project"
	ActivityTargetOffer        ActivityTargetType = "Offer"
	ActivityTargetDeal         ActivityTargetType = "Deal"
	ActivityTargetFile         ActivityTargetType = "File"
	ActivityTargetNotification ActivityTargetType = "Notification"
)

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeMeeting ActivityType = "meeting"
	ActivityTypeCall    ActivityType = "call"
	ActivityTypeEmail   ActivityType = "email"
	ActivityTypeTask    ActivityType = "task"
	ActivityTypeNote    ActivityType = "note"
	ActivityTypeSystem  ActivityType = "system"
)

// ActivityStatus represents the status of an activity
type ActivityStatus string

const (
	ActivityStatusPlanned    ActivityStatus = "planned"
	ActivityStatusInProgress ActivityStatus = "in_progress"
	ActivityStatusCompleted  ActivityStatus = "completed"
	ActivityStatusCancelled  ActivityStatus = "cancelled"
)

// IsValid checks if the ActivityType is a valid enum value
func (at ActivityType) IsValid() bool {
	switch at {
	case ActivityTypeMeeting, ActivityTypeCall, ActivityTypeEmail, ActivityTypeTask, ActivityTypeNote, ActivityTypeSystem:
		return true
	}
	return false
}

// IsValid checks if the ActivityStatus is a valid enum value
func (as ActivityStatus) IsValid() bool {
	switch as {
	case ActivityStatusPlanned, ActivityStatusInProgress, ActivityStatusCompleted, ActivityStatusCancelled:
		return true
	}
	return false
}

// Activity represents an event log entry for any entity
type Activity struct {
	BaseModel
	TargetType      ActivityTargetType `gorm:"type:varchar(50);not null;index;column:target_type"`
	TargetID        uuid.UUID          `gorm:"type:uuid;not null;index;column:target_id"`
	Title           string             `gorm:"type:varchar(200);not null"`
	Body            string             `gorm:"type:varchar(2000)"`
	OccurredAt      time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP;index;column:occurred_at"`
	CreatorName     string             `gorm:"type:varchar(200);column:creator_name"`
	ActivityType    ActivityType       `gorm:"type:activity_type;not null;default:'note';column:activity_type"`
	Status          ActivityStatus     `gorm:"type:activity_status;not null;default:'completed'"`
	ScheduledAt     *time.Time         `gorm:"column:scheduled_at"`
	DueDate         *time.Time         `gorm:"type:date;column:due_date"`
	CompletedAt     *time.Time         `gorm:"column:completed_at"`
	DurationMinutes *int               `gorm:"column:duration_minutes"`
	Priority        int                `gorm:"default:0"`
	IsPrivate       bool               `gorm:"not null;default:false;column:is_private"`
	CreatorID       string             `gorm:"type:varchar(100);column:creator_id"`
	AssignedToID    string             `gorm:"type:varchar(100);column:assigned_to_id"`
	CompanyID       *CompanyID         `gorm:"type:varchar(50);column:company_id"`
	Company         *Company           `gorm:"foreignKey:CompanyID"`
}

// UserRoleType represents a role a user can have
type UserRoleType string

const (
	RoleSuperAdmin     UserRoleType = "super_admin"
	RoleCompanyAdmin   UserRoleType = "company_admin"
	RoleManager        UserRoleType = "manager"
	RoleMarket         UserRoleType = "market"
	RoleProjectManager UserRoleType = "project_manager"
	RoleProjectLeader  UserRoleType = "project_leader"
	RoleViewer         UserRoleType = "viewer"
	RoleAPIService     UserRoleType = "api_service"
)

// UserRole represents a role assignment for a user
type UserRole struct {
	ID        uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    string       `gorm:"type:varchar(100);not null;index;column:user_id"`
	User      *User        `gorm:"foreignKey:UserID"`
	Role      UserRoleType `gorm:"type:user_role;not null"`
	CompanyID *CompanyID   `gorm:"type:varchar(50);column:company_id"`
	Company   *Company     `gorm:"foreignKey:CompanyID"`
	GrantedBy string       `gorm:"type:varchar(100);column:granted_by"`
	GrantedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP;column:granted_at"`
	ExpiresAt *time.Time   `gorm:"column:expires_at"`
	IsActive  bool         `gorm:"not null;default:true;column:is_active"`
	CreatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// PermissionType represents a specific permission
type PermissionType string

const (
	// Customer permissions
	PermissionCustomersRead   PermissionType = "customers:read"
	PermissionCustomersWrite  PermissionType = "customers:write"
	PermissionCustomersDelete PermissionType = "customers:delete"

	// Contact permissions
	PermissionContactsRead   PermissionType = "contacts:read"
	PermissionContactsWrite  PermissionType = "contacts:write"
	PermissionContactsDelete PermissionType = "contacts:delete"

	// Deal permissions
	PermissionDealsRead   PermissionType = "deals:read"
	PermissionDealsWrite  PermissionType = "deals:write"
	PermissionDealsDelete PermissionType = "deals:delete"

	// Offer permissions
	PermissionOffersRead    PermissionType = "offers:read"
	PermissionOffersWrite   PermissionType = "offers:write"
	PermissionOffersDelete  PermissionType = "offers:delete"
	PermissionOffersApprove PermissionType = "offers:approve"

	// Project permissions
	PermissionProjectsRead   PermissionType = "projects:read"
	PermissionProjectsWrite  PermissionType = "projects:write"
	PermissionProjectsDelete PermissionType = "projects:delete"

	// Budget permissions
	PermissionBudgetsRead  PermissionType = "budgets:read"
	PermissionBudgetsWrite PermissionType = "budgets:write"

	// Activity permissions
	PermissionActivitiesRead   PermissionType = "activities:read"
	PermissionActivitiesWrite  PermissionType = "activities:write"
	PermissionActivitiesDelete PermissionType = "activities:delete"

	// User management permissions
	PermissionUsersRead        PermissionType = "users:read"
	PermissionUsersWrite       PermissionType = "users:write"
	PermissionUsersManageRoles PermissionType = "users:manage_roles"

	// Company management permissions
	PermissionCompaniesRead  PermissionType = "companies:read"
	PermissionCompaniesWrite PermissionType = "companies:write"

	// Reports and analytics
	PermissionReportsView   PermissionType = "reports:view"
	PermissionReportsExport PermissionType = "reports:export"

	// System administration
	PermissionSystemAdmin     PermissionType = "system:admin"
	PermissionSystemAuditLogs PermissionType = "system:audit_logs"
)

// UserPermission represents a permission override for a user
type UserPermission struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     string         `gorm:"type:varchar(100);not null;index;column:user_id"`
	User       *User          `gorm:"foreignKey:UserID"`
	Permission PermissionType `gorm:"type:permission_type;not null"`
	CompanyID  *CompanyID     `gorm:"type:varchar(50);column:company_id"`
	Company    *Company       `gorm:"foreignKey:CompanyID"`
	IsGranted  bool           `gorm:"not null;default:true;column:is_granted"`
	GrantedBy  string         `gorm:"type:varchar(100);column:granted_by"`
	GrantedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;column:granted_at"`
	ExpiresAt  *time.Time     `gorm:"column:expires_at"`
	Reason     string         `gorm:"type:text"`
	CreatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// AuditAction represents the type of audit action
type AuditAction string

const (
	AuditActionCreate           AuditAction = "create"
	AuditActionUpdate           AuditAction = "update"
	AuditActionDelete           AuditAction = "delete"
	AuditActionLogin            AuditAction = "login"
	AuditActionLogout           AuditAction = "logout"
	AuditActionPermissionGrant  AuditAction = "permission_grant"
	AuditActionPermissionRevoke AuditAction = "permission_revoke"
	AuditActionRoleAssign       AuditAction = "role_assign"
	AuditActionRoleRemove       AuditAction = "role_remove"
	AuditActionExport           AuditAction = "export"
	AuditActionImport           AuditAction = "import"
	AuditActionAPICall          AuditAction = "api_call"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID          uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      string      `gorm:"type:varchar(100);column:user_id"`
	UserEmail   string      `gorm:"type:varchar(255);column:user_email"`
	UserName    string      `gorm:"type:varchar(200);column:user_name"`
	Action      AuditAction `gorm:"type:audit_action;not null"`
	EntityType  string      `gorm:"type:varchar(50);not null;column:entity_type"`
	EntityID    *uuid.UUID  `gorm:"type:uuid;column:entity_id"`
	EntityName  string      `gorm:"type:varchar(200);column:entity_name"`
	CompanyID   *CompanyID  `gorm:"type:varchar(50);column:company_id"`
	OldValues   string      `gorm:"type:jsonb;column:old_values"`
	NewValues   string      `gorm:"type:jsonb;column:new_values"`
	Changes     string      `gorm:"type:jsonb"`
	IPAddress   string      `gorm:"type:inet;column:ip_address"`
	UserAgent   string      `gorm:"type:text;column:user_agent"`
	RequestID   string      `gorm:"type:varchar(100);column:request_id"`
	Metadata    string      `gorm:"type:jsonb"`
	PerformedAt time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP;column:performed_at"`
	CreatedAt   time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// File represents an uploaded file
type File struct {
	BaseModel
	Filename    string     `gorm:"type:varchar(255);not null"`
	ContentType string     `gorm:"type:varchar(100);not null"`
	Size        int64      `gorm:"not null"`
	StoragePath string     `gorm:"type:varchar(500);not null;unique"`
	OfferID     *uuid.UUID `gorm:"type:uuid;index"`
	Offer       *Offer     `gorm:"foreignKey:OfferID"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeTaskAssigned     NotificationType = "task_assigned"
	NotificationTypeBudgetAlert      NotificationType = "budget_alert"
	NotificationTypeDealStageChanged NotificationType = "deal_stage_changed"
	NotificationTypeOfferAccepted    NotificationType = "offer_accepted"
	NotificationTypeOfferRejected    NotificationType = "offer_rejected"
	NotificationTypeActivityReminder NotificationType = "activity_reminder"
	NotificationTypeProjectUpdate    NotificationType = "project_update"
)

// Notification represents a user notification
type Notification struct {
	BaseModel
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Type       string    `gorm:"type:varchar(50);not null"`
	Title      string    `gorm:"type:varchar(200);not null"`
	Message    string    `gorm:"type:varchar(500);not null"`
	Read       bool      `gorm:"column:read;not null;default:false;index"`
	ReadAt     *time.Time
	EntityID   *uuid.UUID `gorm:"type:uuid"`
	EntityType string     `gorm:"type:varchar(50)"`
}

// User represents a user in the system
type User struct {
	ID          string         `gorm:"type:varchar(100);primaryKey" json:"id"`
	AzureADOID  string         `gorm:"type:varchar(100);unique;column:azure_ad_oid" json:"azureAdOid,omitempty"`
	Email       string         `gorm:"type:varchar(255);not null;unique" json:"email"`
	FirstName   string         `gorm:"type:varchar(100);column:first_name" json:"firstName,omitempty"`
	LastName    string         `gorm:"type:varchar(100);column:last_name" json:"lastName,omitempty"`
	DisplayName string         `gorm:"type:varchar(200);not null;column:name" json:"displayName"`
	Roles       pq.StringArray `gorm:"type:text[];not null" json:"roles"`
	Department  string         `gorm:"type:varchar(100)" json:"department,omitempty"`
	Avatar      string         `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	CompanyID   *CompanyID     `gorm:"type:varchar(50);column:company_id" json:"companyId,omitempty"`
	Company     *Company       `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	IsActive    bool           `gorm:"not null;default:true;column:is_active" json:"isActive"`
	LastLoginAt *time.Time     `gorm:"column:last_login_at" json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// FullName returns the user's full name, or display name if first/last not set
func (u *User) FullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.DisplayName
}
