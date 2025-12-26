package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
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

// IsValidCompanyID checks if a string is a valid company ID
func IsValidCompanyID(id string) bool {
	switch CompanyID(id) {
	case CompanyGruppen, CompanyStalbygg, CompanyHybridbygg, CompanyIndustri, CompanyTak, CompanyMontasje:
		return true
	default:
		return false
	}
}

// Company represents a Straye group company (stored in database)
type Company struct {
	ID                          CompanyID `gorm:"type:varchar(50);primaryKey" json:"id"`
	Name                        string    `gorm:"type:varchar(200);not null" json:"name"`
	ShortName                   string    `gorm:"type:varchar(50);not null;column:short_name" json:"shortName"`
	OrgNumber                   string    `gorm:"type:varchar(20);column:org_number" json:"orgNumber,omitempty"`
	Color                       string    `gorm:"type:varchar(20);not null;default:'#000000'" json:"color"`
	Logo                        string    `gorm:"type:varchar(500)" json:"logo,omitempty"`
	IsActive                    bool      `gorm:"not null;default:true;column:is_active" json:"isActive"`
	DefaultOfferResponsibleID   *string   `gorm:"type:varchar(100);column:default_offer_responsible_id" json:"defaultOfferResponsibleId,omitempty"`
	DefaultProjectResponsibleID *string   `gorm:"type:varchar(100);column:default_project_responsible_id" json:"defaultProjectResponsibleId,omitempty"`
	CreatedAt                   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt                   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// Customer represents an organization in the CRM
// Supports soft delete - deleted customers are hidden but their data is preserved
// for historical reference in related projects and offers.
type Customer struct {
	BaseModel
	DeletedAt     gorm.DeletedAt   `gorm:"index"` // Soft delete support
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
	Notes         string           `gorm:"type:text"`
	CustomerClass string           `gorm:"type:varchar(50);column:customer_class"`
	CreditLimit   *float64         `gorm:"type:decimal(15,2);column:credit_limit"`
	IsInternal    bool             `gorm:"type:boolean;not null;default:false;column:is_internal"`
	Municipality  string           `gorm:"type:varchar(100)"`
	County        string           `gorm:"type:varchar(100)"`
	Website       string           `gorm:"type:varchar(500)"`
	CompanyID     *CompanyID       `gorm:"type:varchar(50);column:company_id;index"`
	Company       *Company         `gorm:"foreignKey:CompanyID"`
	// User tracking fields
	CreatedByID   string `gorm:"type:varchar(100);column:created_by_id;index"`
	CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
	UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
	UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
	// Relations - no cascade delete, projects/offers are preserved when customer is soft deleted
	Contacts []Contact `gorm:"foreignKey:PrimaryCustomerID"`
	Projects []Project `gorm:"foreignKey:CustomerID"`
	Offers   []Offer   `gorm:"foreignKey:CustomerID"`
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
	FirstName              string      `gorm:"type:varchar(100);not null;column:first_name"`
	LastName               string      `gorm:"type:varchar(100);not null;column:last_name"`
	Email                  string      `gorm:"type:varchar(255);uniqueIndex"`
	Phone                  string      `gorm:"type:varchar(50)"`
	Mobile                 string      `gorm:"type:varchar(50)"`
	Title                  string      `gorm:"type:varchar(100)"`
	Department             string      `gorm:"type:varchar(100)"`
	ContactType            ContactType `gorm:"type:varchar(50);not null;default:'primary';column:contact_type;index"`
	PrimaryCustomerID      *uuid.UUID  `gorm:"type:uuid;column:primary_customer_id"`
	PrimaryCustomer        *Customer   `gorm:"foreignKey:PrimaryCustomerID"`
	Address                string      `gorm:"type:varchar(500)"`
	City                   string      `gorm:"type:varchar(100)"`
	PostalCode             string      `gorm:"type:varchar(20)"`
	Country                string      `gorm:"type:varchar(100);default:'Norway'"`
	LinkedInURL            string      `gorm:"type:varchar(500);column:linkedin_url"`
	PreferredContactMethod string      `gorm:"type:varchar(50);default:'email';column:preferred_contact_method"`
	Notes                  string      `gorm:"type:text"`
	IsActive               bool        `gorm:"not null;default:true;column:is_active"`
	// User tracking fields
	CreatedByID   string `gorm:"type:varchar(100);column:created_by_id;index"`
	CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
	UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
	UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
	// Relations
	Relationships []ContactRelationship `gorm:"foreignKey:ContactID"`
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

// ProjectPhase represents the lifecycle phase of a project
// Projects are now lightweight containers for offers. Economic tracking moved to Offer.
// - tilbud: Offer/bidding phase (default). Project collecting offers.
// - working: Active project with work in progress
// - on_hold: Project temporarily paused
// - completed: Project is finished
// - cancelled: Project was cancelled
type ProjectPhase string

const (
	ProjectPhaseTilbud    ProjectPhase = "tilbud"
	ProjectPhaseWorking   ProjectPhase = "working"
	ProjectPhaseOnHold    ProjectPhase = "on_hold"
	ProjectPhaseCompleted ProjectPhase = "completed"
	ProjectPhaseCancelled ProjectPhase = "cancelled"
)

// IsValid checks if the ProjectPhase is a valid enum value
func (p ProjectPhase) IsValid() bool {
	switch p {
	case ProjectPhaseTilbud, ProjectPhaseWorking, ProjectPhaseOnHold, ProjectPhaseCompleted, ProjectPhaseCancelled:
		return true
	}
	return false
}

// IsEditablePhase returns true if the project can be modified in this phase
func (p ProjectPhase) IsEditablePhase() bool {
	return p == ProjectPhaseTilbud || p == ProjectPhaseWorking || p == ProjectPhaseOnHold
}

// IsActivePhase returns true if the phase represents an active project
func (p ProjectPhase) IsActivePhase() bool {
	return p == ProjectPhaseWorking
}

// IsClosedPhase returns true if the phase represents a closed/terminal state
func (p ProjectPhase) IsClosedPhase() bool {
	return p == ProjectPhaseCompleted || p == ProjectPhaseCancelled
}

// CanTransitionTo checks if a phase transition is valid
// Valid transitions:
// - tilbud -> working (start work), on_hold (pause), cancelled
// - working -> on_hold, completed, cancelled, tilbud (revert)
// - on_hold -> working (resume), cancelled, completed
// - completed -> working (can reopen)
// - cancelled -> (terminal state, no transitions)
func (p ProjectPhase) CanTransitionTo(target ProjectPhase) bool {
	if p == target {
		return true // Same phase is always valid
	}

	switch p {
	case ProjectPhaseTilbud:
		return target == ProjectPhaseWorking ||
			target == ProjectPhaseOnHold ||
			target == ProjectPhaseCancelled
	case ProjectPhaseWorking:
		return target == ProjectPhaseOnHold ||
			target == ProjectPhaseCompleted ||
			target == ProjectPhaseCancelled ||
			target == ProjectPhaseTilbud
	case ProjectPhaseOnHold:
		return target == ProjectPhaseWorking ||
			target == ProjectPhaseCancelled ||
			target == ProjectPhaseCompleted
	case ProjectPhaseCompleted:
		return target == ProjectPhaseWorking
	case ProjectPhaseCancelled:
		return false // Cancelled is a terminal state - no transitions allowed
	}
	return false
}

// Project represents a container for related offers. Projects are cross-company.
// Economic tracking (value, cost, spent, invoiced) has moved to the Offer model.
type Project struct {
	BaseModel
	Name              string       `gorm:"type:varchar(200);not null;index"`
	ProjectNumber     string       `gorm:"type:varchar(50);unique;index;column:project_number"` // External reference number for ERP/accounting systems
	Summary           string       `gorm:"type:varchar(500)"`
	Description       string       `gorm:"type:text"`
	CustomerID        *uuid.UUID   `gorm:"type:uuid;index"` // Optional - projects can be cross-company without specific customer
	Customer          *Customer    `gorm:"foreignKey:CustomerID"`
	CustomerName      string       `gorm:"type:varchar(200)"`
	Phase             ProjectPhase `gorm:"type:project_phase;not null;default:'tilbud';index"`
	StartDate         time.Time    `gorm:"type:date"`
	EndDate           *time.Time   `gorm:"type:date"`
	Location          string       `gorm:"type:varchar(200)"`
	DealID            *uuid.UUID   `gorm:"type:uuid;index;column:deal_id"`
	Deal              *Deal        `gorm:"foreignKey:DealID"`
	ExternalReference string       `gorm:"type:varchar(100);column:external_reference"`
	// User tracking fields
	CreatedByID   string `gorm:"type:varchar(100);column:created_by_id;index"`
	CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
	UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
	UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
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
	ID               uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID        uuid.UUID   `gorm:"type:uuid;not null;index;column:project_id"`
	Project          *Project    `gorm:"foreignKey:ProjectID"`
	CostType         CostType    `gorm:"type:cost_type;not null;column:cost_type"`
	Description      string      `gorm:"type:varchar(500);not null"`
	Amount           float64     `gorm:"type:decimal(15,2);not null"`
	Currency         string      `gorm:"type:varchar(3);not null;default:'NOK'"`
	CostDate         time.Time   `gorm:"type:date;not null;column:cost_date"`
	PostingDate      *time.Time  `gorm:"type:date;column:posting_date"`
	BudgetItemID     *uuid.UUID  `gorm:"type:uuid;column:budget_item_id"`
	BudgetItem       *BudgetItem `gorm:"foreignKey:BudgetItemID"`
	ERPSource        ERPSource   `gorm:"type:erp_source;not null;default:'manual';column:erp_source"`
	ERPReference     string      `gorm:"type:varchar(100);column:erp_reference"`
	ERPTransactionID string      `gorm:"type:varchar(100);column:erp_transaction_id"`
	ERPSyncedAt      *time.Time  `gorm:"column:erp_synced_at"`
	IsApproved       bool        `gorm:"not null;default:false;column:is_approved"`
	ApprovedByID     string      `gorm:"type:varchar(100);column:approved_by_id"`
	ApprovedAt       *time.Time  `gorm:"column:approved_at"`
	Notes            string      `gorm:"type:text"`
	CreatedAt        time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// OfferPhase represents the phase of an offer in the sales pipeline
// Pipeline: draft -> in_progress -> sent -> order -> completed (or lost/expired)
type OfferPhase string

const (
	OfferPhaseDraft      OfferPhase = "draft"
	OfferPhaseInProgress OfferPhase = "in_progress"
	OfferPhaseSent       OfferPhase = "sent"
	OfferPhaseOrder      OfferPhase = "order"     // Customer accepted, work in progress
	OfferPhaseCompleted  OfferPhase = "completed" // Work finished
	OfferPhaseLost       OfferPhase = "lost"
	OfferPhaseExpired    OfferPhase = "expired"
)

// IsValid checks if the OfferPhase is a valid enum value
func (p OfferPhase) IsValid() bool {
	switch p {
	case OfferPhaseDraft, OfferPhaseInProgress, OfferPhaseSent, OfferPhaseOrder, OfferPhaseCompleted, OfferPhaseLost, OfferPhaseExpired:
		return true
	}
	return false
}

// IsActivePhase returns true if the offer is in an active working phase
func (p OfferPhase) IsActivePhase() bool {
	return p == OfferPhaseOrder
}

// IsClosedPhase returns true if the offer is in a terminal state
func (p OfferPhase) IsClosedPhase() bool {
	return p == OfferPhaseCompleted || p == OfferPhaseLost || p == OfferPhaseExpired
}

// IsSalesPhase returns true if the offer is in the sales pipeline (not yet order)
func (p OfferPhase) IsSalesPhase() bool {
	return p == OfferPhaseDraft || p == OfferPhaseInProgress || p == OfferPhaseSent
}

// CanTransitionTo checks if an offer phase transition is valid
// Valid transitions:
// - draft -> in_progress, lost
// - in_progress -> sent, draft, lost
// - sent -> order, lost, expired, in_progress
// - order -> completed, lost
// - completed -> order (can reopen)
// - lost -> draft (can restart)
// - expired -> draft (can restart)
func (p OfferPhase) CanTransitionTo(target OfferPhase) bool {
	if p == target {
		return true // Same phase is always valid
	}

	switch p {
	case OfferPhaseDraft:
		return target == OfferPhaseInProgress || target == OfferPhaseLost
	case OfferPhaseInProgress:
		return target == OfferPhaseSent || target == OfferPhaseDraft || target == OfferPhaseLost
	case OfferPhaseSent:
		return target == OfferPhaseOrder || target == OfferPhaseLost || target == OfferPhaseExpired || target == OfferPhaseInProgress
	case OfferPhaseOrder:
		return target == OfferPhaseCompleted || target == OfferPhaseLost
	case OfferPhaseCompleted:
		return target == OfferPhaseOrder // Can reopen completed offer
	case OfferPhaseLost:
		return target == OfferPhaseDraft // Can restart lost offer
	case OfferPhaseExpired:
		return target == OfferPhaseDraft // Can restart expired offer
	}
	return false
}

// OfferHealth represents the health status of an offer in execution (order phase)
type OfferHealth string

const (
	OfferHealthOnTrack    OfferHealth = "on_track"
	OfferHealthAtRisk     OfferHealth = "at_risk"
	OfferHealthDelayed    OfferHealth = "delayed"
	OfferHealthOverBudget OfferHealth = "over_budget"
)

// IsValid checks if the OfferHealth is a valid enum value
func (h OfferHealth) IsValid() bool {
	switch h {
	case OfferHealthOnTrack, OfferHealthAtRisk, OfferHealthDelayed, OfferHealthOverBudget:
		return true
	}
	return false
}

// OfferStatus represents the status of an offer
type OfferStatus string

const (
	OfferStatusActive   OfferStatus = "active"
	OfferStatusInactive OfferStatus = "inactive"
	OfferStatusArchived OfferStatus = "archived"
)

// OfferWarning represents a warning code for offer data discrepancies
// Warning codes are string-based for extensibility
type OfferWarning string

const (
	// OfferWarningValueNotEqualsDWTotalIncome indicates that the offer's Value
	// does not match the DWTotalIncome from the data warehouse.
	// This warning is only applicable when the offer is in the "order" phase.
	OfferWarningValueNotEqualsDWTotalIncome OfferWarning = "value.not.equals.dwTotalIncome"
)

// Offer represents a sales proposal and, when in order phase, the execution of work
type Offer struct {
	BaseModel
	Title                 string      `gorm:"type:varchar(200);not null;index"`
	OfferNumber           string      `gorm:"type:varchar(50);column:offer_number;index"`  // Internal number, e.g., "TK-2025-001"
	ExternalReference     string      `gorm:"type:varchar(100);column:external_reference"` // External/customer reference number
	CustomerID            *uuid.UUID  `gorm:"type:uuid;index"`                             // Optional - offer can exist without customer when linked to project
	Customer              *Customer   `gorm:"foreignKey:CustomerID"`
	CustomerName          string      `gorm:"type:varchar(200)"`
	ProjectID             *uuid.UUID  `gorm:"type:uuid;index;column:project_id"` // Nullable - offer can exist without project
	Project               *Project    `gorm:"foreignKey:ProjectID"`
	ProjectName           string      `gorm:"type:varchar(200);column:project_name"`
	CompanyID             CompanyID   `gorm:"type:varchar(50);not null;index"`
	Phase                 OfferPhase  `gorm:"type:varchar(50);not null;index"`
	Probability           int         `gorm:"type:int;not null;default:0"`
	Value                 float64     `gorm:"type:decimal(15,2);not null;default:0"`
	Status                OfferStatus `gorm:"type:varchar(50);not null;index"`
	ResponsibleUserID     string      `gorm:"type:varchar(100);index"` // Optional for inquiries (draft phase)
	ResponsibleUserName   string      `gorm:"type:varchar(200)"`
	Description           string      `gorm:"type:text"`
	Notes                 string      `gorm:"type:text"`
	DueDate               *time.Time  `gorm:"type:timestamp;index"`
	Cost                  float64     `gorm:"type:decimal(15,2);default:0"`                               // Internal cost
	MarginPercent         float64     `gorm:"type:decimal(8,4);not null;default:0;column:margin_percent"` // Dekningsgrad: (value - cost) / value * 100, auto-calculated
	Location              string      `gorm:"type:varchar(200)"`
	SentDate              *time.Time  `gorm:"type:timestamp;index;column:sent_date"`
	ExpirationDate        *time.Time  `gorm:"type:timestamp;index;column:expiration_date"` // When the offer expires (default: 60 days after sent_date)
	CustomerHasWonProject bool        `gorm:"not null;default:false;column:customer_has_won_project"`
	// Order phase execution fields (used when phase = "order" or "completed")
	ManagerID               *string        `gorm:"type:varchar(100);column:manager_id"`
	ManagerName             string         `gorm:"type:varchar(200);column:manager_name"`
	TeamMembers             pq.StringArray `gorm:"type:text[];column:team_members"`
	Spent                   float64        `gorm:"type:decimal(15,2);not null;default:0"`                 // Actual costs incurred
	Invoiced                float64        `gorm:"type:decimal(15,2);not null;default:0"`                 // Amount invoiced to customer
	OrderReserve            float64        `gorm:"type:decimal(15,2);column:order_reserve;->"`            // Generated column: value - invoiced (read-only)
	Health                  *OfferHealth   `gorm:"type:varchar(20);default:'on_track'"`                   // Health status during execution
	CompletionPercent       *float64       `gorm:"type:decimal(5,2);default:0;column:completion_percent"` // 0-100 progress indicator
	StartDate               *time.Time     `gorm:"type:date;column:start_date"`                           // When work started
	EndDate                 *time.Time     `gorm:"type:date;column:end_date"`                             // Planned end date
	EstimatedCompletionDate *time.Time     `gorm:"type:date;column:estimated_completion_date"`            // Current estimate for completion
	// User tracking fields
	CreatedByID   string `gorm:"type:varchar(100);column:created_by_id;index"`
	CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
	UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
	UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
	// Data Warehouse synced fields - populated by periodic sync from external ERP system
	DWTotalIncome   float64    `gorm:"column:dw_total_income;default:0"`   // Income from DW (accounts 3000-3999)
	DWMaterialCosts float64    `gorm:"column:dw_material_costs;default:0"` // Material costs (accounts 4000-4999)
	DWEmployeeCosts float64    `gorm:"column:dw_employee_costs;default:0"` // Employee costs (accounts 5000-5999)
	DWOtherCosts       float64    `gorm:"column:dw_other_costs;default:0"`       // Other costs (accounts >= 6000)
	DWNetResult        float64    `gorm:"column:dw_net_result;default:0"`        // Net result (income - costs)
	DWTotalFixedPrice  float64    `gorm:"column:dw_total_fixed_price;default:0"` // Sum of FixedPriceAmount from synced assignments
	DWLastSyncedAt     *time.Time `gorm:"column:dw_last_synced_at"`              // Last successful sync timestamp
	// Relations
	Items []OfferItem `gorm:"foreignKey:OfferID;constraint:OnDelete:CASCADE"`
	Files []File      `gorm:"foreignKey:OfferID"`
}

// CalculateMarginPercent calculates the dekningsgrad based on value and cost.
// Formula: (value - cost) / value * 100
// Edge cases:
//   - cost=0 and value>0: returns 100%
//   - value=0: returns 0%
//   - both 0: returns 0%
func (o *Offer) CalculateMarginPercent() float64 {
	if o.Value > 0 {
		return ((o.Value - o.Cost) / o.Value) * 100
	}
	return 0
}

// CalculateMarginPercentFromValues is a helper function to calculate margin percent
// from value and cost without needing an Offer instance.
func CalculateMarginPercentFromValues(value, cost float64) float64 {
	if value > 0 {
		return ((value - cost) / value) * 100
	}
	return 0
}

// NumberSequence tracks the last used sequence number per company per year
// This sequence is SHARED between offers and projects to ensure unique numbers
// across both entity types within a company/year combination.
// Format: {PREFIX}-{YEAR}-{SEQUENCE} e.g., "ST-2025-001"
type NumberSequence struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID    CompanyID `gorm:"type:varchar(50);not null;column:company_id"`
	Year         int       `gorm:"not null"`
	LastSequence int       `gorm:"not null;default:0;column:last_sequence"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for NumberSequence
func (NumberSequence) TableName() string {
	return "number_sequences"
}

// OfferNumberSequence is an alias for backwards compatibility
// Deprecated: Use NumberSequence instead
type OfferNumberSequence = NumberSequence

// GetCompanyPrefix returns the offer number prefix for a company
// Format: 2-letter slug used in offer numbers e.g., ST-2025-001
func GetCompanyPrefix(companyID CompanyID) string {
	switch companyID {
	case CompanyStalbygg:
		return "ST"
	case CompanyHybridbygg:
		return "HB"
	case CompanyIndustri:
		return "IN"
	case CompanyTak:
		return "TK"
	case CompanyMontasje:
		return "MO"
	case CompanyGruppen:
		return "GR"
	default:
		return "GR"
	}
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

// BudgetParentType represents the type of entity a budget item belongs to
type BudgetParentType string

const (
	BudgetParentOffer   BudgetParentType = "offer"
	BudgetParentProject BudgetParentType = "project"
)

// BudgetItem represents a flexible budget line item that can belong to an offer or project
// Users define their own budget items with name, cost, margin, and optional quantity/price fields
type BudgetItem struct {
	ID              uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ParentType      BudgetParentType `gorm:"type:budget_parent_type;not null;column:parent_type" json:"parentType"`
	ParentID        uuid.UUID        `gorm:"type:uuid;not null;column:parent_id" json:"parentId"`
	Name            string           `gorm:"type:varchar(200);not null" json:"name"`
	ExpectedCost    float64          `gorm:"type:decimal(15,2);not null;default:0;column:expected_cost" json:"expectedCost"`
	ExpectedMargin  float64          `gorm:"type:decimal(5,2);not null;default:0;column:expected_margin" json:"expectedMargin"` // Percentage 0-100
	ExpectedRevenue float64          `gorm:"type:decimal(15,2);column:expected_revenue;->" json:"expectedRevenue"`              // Computed: cost / (1 - margin/100)
	ExpectedProfit  float64          `gorm:"type:decimal(15,2);column:expected_profit;->" json:"expectedProfit"`                // Computed: revenue - cost
	Quantity        *float64         `gorm:"type:decimal(10,2)" json:"quantity,omitempty"`
	PricePerItem    *float64         `gorm:"type:decimal(15,2);column:price_per_item" json:"pricePerItem,omitempty"`
	Description     string           `gorm:"type:text" json:"description,omitempty"`
	DisplayOrder    int              `gorm:"not null;default:0;column:display_order" json:"displayOrder"`
	CreatedAt       time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt       time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns the table name for BudgetItem
func (BudgetItem) TableName() string {
	return "budget_items"
}

// BudgetSummary holds aggregated budget totals for a parent entity (offer or project)
type BudgetSummary struct {
	TotalCost     float64 `json:"totalCost"`
	TotalRevenue  float64 `json:"totalRevenue"`
	TotalProfit   float64 `json:"totalProfit"`
	MarginPercent float64 `json:"marginPercent"` // (Profit / Revenue) * 100, 0 if revenue=0
	ItemCount     int     `json:"itemCount"`
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
	ActivityTargetSupplier     ActivityTargetType = "Supplier"
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
	TargetType       ActivityTargetType `gorm:"type:varchar(50);not null;index;column:target_type"`
	TargetID         uuid.UUID          `gorm:"type:uuid;not null;index;column:target_id"`
	TargetName       string             `gorm:"type:varchar(255);column:target_name"`
	Title            string             `gorm:"type:varchar(200);not null"`
	Body             string             `gorm:"type:varchar(2000)"`
	OccurredAt       time.Time          `gorm:"not null;default:CURRENT_TIMESTAMP;index;column:occurred_at"`
	CreatorName      string             `gorm:"type:varchar(200);column:creator_name"`
	ActivityType     ActivityType       `gorm:"type:activity_type;not null;default:'note';column:activity_type"`
	Status           ActivityStatus     `gorm:"type:activity_status;not null;default:'completed'"`
	ScheduledAt      *time.Time         `gorm:"column:scheduled_at"`
	DueDate          *time.Time         `gorm:"type:date;column:due_date"`
	CompletedAt      *time.Time         `gorm:"column:completed_at"`
	DurationMinutes  *int               `gorm:"column:duration_minutes"`
	Priority         int                `gorm:"default:0"`
	IsPrivate        bool               `gorm:"not null;default:false;column:is_private"`
	CreatorID        string             `gorm:"type:varchar(100);column:creator_id"`
	AssignedToID     string             `gorm:"type:varchar(100);column:assigned_to_id"`
	CompanyID        *CompanyID         `gorm:"type:varchar(50);column:company_id"`
	Company          *Company           `gorm:"foreignKey:CompanyID"`
	Attendees        pq.StringArray     `gorm:"type:text[];column:attendees"`
	ParentActivityID *uuid.UUID         `gorm:"type:uuid;column:parent_activity_id"`
	ParentActivity   *Activity          `gorm:"foreignKey:ParentActivityID"`
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
	IPAddress   string      `gorm:"type:varchar(100);column:ip_address"`
	UserAgent   string      `gorm:"type:text;column:user_agent"`
	RequestID   string      `gorm:"type:varchar(100);column:request_id"`
	Metadata    string      `gorm:"type:jsonb"`
	PerformedAt time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP;column:performed_at"`
	CreatedAt   time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// File represents an uploaded file attached to an entity.
// Exactly one of OfferID, CustomerID, ProjectID, or SupplierID should be set.
// CompanyID is required and indicates which company the file belongs to.
type File struct {
	BaseModel
	Filename        string         `gorm:"type:varchar(255);not null"`
	ContentType     string         `gorm:"type:varchar(100);not null"`
	Size            int64          `gorm:"not null"`
	StoragePath     string         `gorm:"type:varchar(500);not null;unique"`
	CompanyID       CompanyID      `gorm:"type:varchar(50);not null;index:idx_files_company_id;column:company_id"`
	Company         *Company       `gorm:"foreignKey:CompanyID"`
	OfferID         *uuid.UUID     `gorm:"type:uuid;index:idx_files_offer_id"`
	Offer           *Offer         `gorm:"foreignKey:OfferID"`
	CustomerID      *uuid.UUID     `gorm:"type:uuid;index:idx_files_customer_id"`
	Customer        *Customer      `gorm:"foreignKey:CustomerID"`
	ProjectID       *uuid.UUID     `gorm:"type:uuid;index:idx_files_project_id"`
	Project         *Project       `gorm:"foreignKey:ProjectID"`
	SupplierID      *uuid.UUID     `gorm:"type:uuid;index:idx_files_supplier_id"`
	Supplier        *Supplier      `gorm:"foreignKey:SupplierID"`
	OfferSupplierID *uuid.UUID     `gorm:"type:uuid;index:idx_files_offer_supplier_id;column:offer_supplier_id"`
	OfferSupplier   *OfferSupplier `gorm:"foreignKey:OfferSupplierID"`
}

// GetEntityType returns the type of entity this file is attached to
func (f *File) GetEntityType() string {
	switch {
	case f.OfferSupplierID != nil:
		return "offer_supplier"
	case f.CustomerID != nil:
		return "customer"
	case f.ProjectID != nil:
		return "project"
	case f.OfferID != nil:
		return "offer"
	case f.SupplierID != nil:
		return "supplier"
	default:
		return ""
	}
}

// GetEntityID returns the ID of the entity this file is attached to
func (f *File) GetEntityID() *uuid.UUID {
	switch {
	case f.OfferSupplierID != nil:
		return f.OfferSupplierID
	case f.CustomerID != nil:
		return f.CustomerID
	case f.ProjectID != nil:
		return f.ProjectID
	case f.OfferID != nil:
		return f.OfferID
	case f.SupplierID != nil:
		return f.SupplierID
	default:
		return nil
	}
}

// HasExactlyOneEntityID returns true if exactly one entity ID is set
func (f *File) HasExactlyOneEntityID() bool {
	count := 0
	if f.OfferID != nil {
		count++
	}
	if f.CustomerID != nil {
		count++
	}
	if f.ProjectID != nil {
		count++
	}
	if f.SupplierID != nil {
		count++
	}
	if f.OfferSupplierID != nil {
		count++
	}
	return count == 1
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
	ID            string         `gorm:"type:varchar(100);primaryKey" json:"id"`
	AzureADOID    string         `gorm:"type:varchar(100);unique;column:azure_ad_oid" json:"azureAdOid,omitempty"`
	Email         string         `gorm:"type:varchar(255);not null;unique" json:"email"`
	FirstName     string         `gorm:"type:varchar(100);column:first_name" json:"firstName,omitempty"`
	LastName      string         `gorm:"type:varchar(100);column:last_name" json:"lastName,omitempty"`
	DisplayName   string         `gorm:"type:varchar(200);not null;column:name" json:"displayName"`
	Roles         pq.StringArray `gorm:"type:text[];not null" json:"roles"`
	AzureADRoles  pq.StringArray `gorm:"type:text[];column:azure_ad_roles" json:"azureAdRoles,omitempty"`
	Department    string         `gorm:"type:varchar(100)" json:"department,omitempty"`
	Avatar        string         `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	CompanyID     *CompanyID     `gorm:"type:varchar(50);column:company_id" json:"companyId,omitempty"`
	Company       *Company       `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	IsActive      bool           `gorm:"not null;default:true;column:is_active" json:"isActive"`
	LastIPAddress string         `gorm:"type:varchar(100);column:last_ip_address" json:"lastIpAddress,omitempty"`
	LastLoginAt   *time.Time     `gorm:"column:last_login_at" json:"lastLoginAt,omitempty"`
	CreatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// FullName returns the user's full name, or display name if first/last not set
func (u *User) FullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.DisplayName
}

// SupplierStatus represents the status of a supplier
type SupplierStatus string

const (
	SupplierStatusActive      SupplierStatus = "active"
	SupplierStatusInactive    SupplierStatus = "inactive"
	SupplierStatusPending     SupplierStatus = "pending"
	SupplierStatusBlacklisted SupplierStatus = "blacklisted"
)

// IsValid checks if the SupplierStatus is a valid enum value
func (s SupplierStatus) IsValid() bool {
	switch s {
	case SupplierStatusActive, SupplierStatusInactive, SupplierStatusPending, SupplierStatusBlacklisted:
		return true
	}
	return false
}

// OfferSupplierStatus represents the status of a supplier in an offer
type OfferSupplierStatus string

const (
	OfferSupplierStatusActive OfferSupplierStatus = "active"
	OfferSupplierStatusDone   OfferSupplierStatus = "done"
)

// IsValid checks if the OfferSupplierStatus is a valid enum value
func (s OfferSupplierStatus) IsValid() bool {
	switch s {
	case OfferSupplierStatusActive, OfferSupplierStatusDone:
		return true
	}
	return false
}

// Supplier represents an organization that provides goods/services
type Supplier struct {
	BaseModel
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	Name          string         `gorm:"type:varchar(200);not null;index"`
	OrgNumber     string         `gorm:"type:varchar(20);unique;index;column:org_number"`
	Email         string         `gorm:"type:varchar(255)"`
	Phone         string         `gorm:"type:varchar(50)"`
	Address       string         `gorm:"type:varchar(500)"`
	City          string         `gorm:"type:varchar(100)"`
	PostalCode    string         `gorm:"type:varchar(20);column:postal_code"`
	Country       string         `gorm:"type:varchar(100);not null;default:'Norway'"`
	Municipality  string         `gorm:"type:varchar(100)"`
	County        string         `gorm:"type:varchar(100)"`
	ContactPerson string         `gorm:"type:varchar(200);column:contact_person"`
	ContactEmail  string         `gorm:"type:varchar(255);column:contact_email"`
	ContactPhone  string         `gorm:"type:varchar(50);column:contact_phone"`
	Status        SupplierStatus `gorm:"type:varchar(50);not null;default:'active';index"`
	Category      string         `gorm:"type:varchar(200)"`
	Notes         string         `gorm:"type:text"`
	PaymentTerms  string         `gorm:"type:varchar(200);column:payment_terms"`
	Website       string         `gorm:"type:varchar(500)"`
	CompanyID     *CompanyID     `gorm:"type:varchar(50);column:company_id;index"`
	Company       *Company       `gorm:"foreignKey:CompanyID"`
	// User tracking fields
	CreatedByID   string `gorm:"type:varchar(100);column:created_by_id;index"`
	CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
	UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
	UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
	// Relations
	Contacts       []SupplierContact `gorm:"foreignKey:SupplierID"`
	OfferSuppliers []OfferSupplier   `gorm:"foreignKey:SupplierID"`
}

// SupplierContact represents a contact person for a supplier
type SupplierContact struct {
	BaseModel
	SupplierID uuid.UUID `gorm:"type:uuid;not null;index;column:supplier_id"`
	Supplier   *Supplier `gorm:"foreignKey:SupplierID"`
	FirstName  string    `gorm:"type:varchar(100);not null;column:first_name"`
	LastName   string    `gorm:"type:varchar(100);not null;column:last_name"`
	Title      string    `gorm:"type:varchar(200)"`
	Email      string    `gorm:"type:varchar(255)"`
	Phone      string    `gorm:"type:varchar(50)"`
	IsPrimary  bool      `gorm:"not null;default:false;column:is_primary"`
	Notes      string    `gorm:"type:text"`
}

// TableName overrides the default table name for SupplierContact
func (SupplierContact) TableName() string {
	return "supplier_contacts"
}

// FullName returns the contact's full name
func (c *SupplierContact) FullName() string {
	return c.FirstName + " " + c.LastName
}

// OfferSupplier represents the many-to-many relationship between offers and suppliers
type OfferSupplier struct {
	BaseModel
	OfferID      uuid.UUID           `gorm:"type:uuid;not null;index;column:offer_id"`
	Offer        *Offer              `gorm:"foreignKey:OfferID"`
	SupplierID   uuid.UUID           `gorm:"type:uuid;not null;index;column:supplier_id"`
	Supplier     *Supplier           `gorm:"foreignKey:SupplierID"`
	SupplierName string              `gorm:"type:varchar(200);column:supplier_name"`
	OfferTitle   string              `gorm:"type:varchar(200);column:offer_title"`
	Status       OfferSupplierStatus `gorm:"type:varchar(50);not null;default:'active'"`
	Notes        string              `gorm:"type:text"`
	// Contact person for this offer (optional - selects one of supplier's contacts)
	ContactID   *uuid.UUID       `gorm:"type:uuid;column:contact_id"`
	Contact     *SupplierContact `gorm:"foreignKey:ContactID"`
	ContactName string           `gorm:"type:varchar(200);column:contact_name"`
	// User tracking fields
	CreatedByID   string `gorm:"type:varchar(100);column:created_by_id"`
	CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
	UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
	UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
}

// TableName overrides the default table name for OfferSupplier
func (OfferSupplier) TableName() string {
	return "offer_suppliers"
}

// Assignment represents an ERP work order synced from the datawarehouse.
// This is a read-only table populated by sync operations.
// Assignments belong to ERP projects which are matched to offers via external_reference.
type Assignment struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// Datawarehouse references
	DWAssignmentID int64 `gorm:"column:dw_assignment_id;not null"`          // AssignmentId from DW
	DWProjectID    int64 `gorm:"column:dw_project_id;not null"`             // ProjectId from DW (internal reference)

	// Link to local offer (matched via external_reference = project Code)
	OfferID   *uuid.UUID `gorm:"type:uuid;column:offer_id;index"`
	Offer     *Offer     `gorm:"foreignKey:OfferID"`
	CompanyID CompanyID  `gorm:"type:varchar(50);column:company_id;not null;index"`
	Company   *Company   `gorm:"foreignKey:CompanyID"`

	// Core assignment fields
	AssignmentNumber string `gorm:"type:varchar(50);not null;index"` // e.g., "2406200"
	Description      string `gorm:"type:text"`

	// Financial field (just FixedPriceAmount for now, extensible via DWRawData)
	FixedPriceAmount float64 `gorm:"type:decimal(15,2);default:0;column:fixed_price_amount"`

	// Status tracking (enum IDs from DW)
	StatusID   *int `gorm:"column:status_id"`   // Enum_AssignmentStatusId
	ProgressID *int `gorm:"column:progress_id"` // Enum_AssignmentProgressId

	// Extensibility: store full DW row as JSONB for future use without migrations
	DWRawData string `gorm:"type:jsonb;column:dw_raw_data"`

	// Sync metadata
	DWSyncedAt time.Time `gorm:"column:dw_synced_at;not null;default:CURRENT_TIMESTAMP"`

	// Standard timestamps
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for Assignment
func (Assignment) TableName() string {
	return "assignments"
}
