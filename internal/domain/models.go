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
	Name          string     `gorm:"type:varchar(200);not null;index"`
	OrgNumber     string     `gorm:"type:varchar(20);unique;index"`
	Email         string     `gorm:"type:varchar(255);not null"`
	Phone         string     `gorm:"type:varchar(50);not null"`
	Address       string     `gorm:"type:varchar(500)"`
	City          string     `gorm:"type:varchar(100)"`
	PostalCode    string     `gorm:"type:varchar(20)"`
	Country       string     `gorm:"type:varchar(100);not null;default:'Norway'"`
	ContactPerson string     `gorm:"type:varchar(200)"`
	ContactEmail  string     `gorm:"type:varchar(255)"`
	ContactPhone  string     `gorm:"type:varchar(50)"`
	CompanyID     *CompanyID `gorm:"type:varchar(50);column:company_id;index"`
	Company       *Company   `gorm:"foreignKey:CompanyID"`
	Contacts      []Contact  `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"`
	Projects      []Project  `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"`
	Offers        []Offer    `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"`
}

// Contact represents an individual person
type Contact struct {
	BaseModel
	FirstName              string                `gorm:"type:varchar(100);not null;column:first_name"`
	LastName               string                `gorm:"type:varchar(100);not null;column:last_name"`
	Email                  string                `gorm:"type:varchar(255)"`
	Phone                  string                `gorm:"type:varchar(50)"`
	Mobile                 string                `gorm:"type:varchar(50)"`
	Title                  string                `gorm:"type:varchar(100)"`
	Department             string                `gorm:"type:varchar(100)"`
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

// Deal represents a sales opportunity in the pipeline
type Deal struct {
	BaseModel
	Title             string     `gorm:"type:varchar(200);not null"`
	Description       string     `gorm:"type:text"`
	CustomerID        uuid.UUID  `gorm:"type:uuid;not null;index;column:customer_id"`
	Customer          *Customer  `gorm:"foreignKey:CustomerID"`
	CompanyID         CompanyID  `gorm:"type:varchar(50);not null;index;column:company_id"`
	Company           *Company   `gorm:"foreignKey:CompanyID"`
	CustomerName      string     `gorm:"type:varchar(200);column:customer_name"`
	Stage             DealStage  `gorm:"type:varchar(50);not null;default:'lead'"`
	Probability       int        `gorm:"type:int;not null;default:0"`
	Value             float64    `gorm:"type:decimal(15,2);not null;default:0"`
	WeightedValue     float64    `gorm:"type:decimal(15,2);column:weighted_value;->"` // Read-only, computed by DB
	Currency          string     `gorm:"type:varchar(3);not null;default:'NOK'"`
	ExpectedCloseDate *time.Time `gorm:"type:date;column:expected_close_date"`
	ActualCloseDate   *time.Time `gorm:"type:date;column:actual_close_date"`
	OwnerID           string     `gorm:"type:varchar(100);not null;column:owner_id"`
	OwnerName         string     `gorm:"type:varchar(200);column:owner_name"`
	Source            string     `gorm:"type:varchar(100)"`
	Notes             string     `gorm:"type:text"`
	LostReason        string     `gorm:"type:varchar(500);column:lost_reason"`
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

// ProjectStatus represents the status of a project
type ProjectStatus string

const (
	ProjectStatusPlanning  ProjectStatus = "planning"
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusCancelled ProjectStatus = "cancelled"
)

// Project represents work being performed for a customer
type Project struct {
	BaseModel
	Name            string         `gorm:"type:varchar(200);not null;index"`
	Summary         string         `gorm:"type:varchar(500)"`
	Description     string         `gorm:"type:text"`
	CustomerID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	Customer        *Customer      `gorm:"foreignKey:CustomerID"`
	CustomerName    string         `gorm:"type:varchar(200)"`
	CompanyID       CompanyID      `gorm:"type:varchar(50);not null;index"`
	Status          ProjectStatus  `gorm:"type:varchar(50);not null;index"`
	StartDate       time.Time      `gorm:"type:date;not null"`
	EndDate         *time.Time     `gorm:"type:date"`
	Budget          float64        `gorm:"type:decimal(15,2);not null;default:0"`
	Spent           float64        `gorm:"type:decimal(15,2);not null;default:0"`
	ManagerID       string         `gorm:"type:varchar(100);not null"`
	ManagerName     string         `gorm:"type:varchar(200)"`
	TeamMembers     pq.StringArray `gorm:"type:text[]"`
	TeamsChannelID  string         `gorm:"type:varchar(200)"`
	TeamsChannelURL string         `gorm:"type:varchar(500)"`
	OfferID         *uuid.UUID     `gorm:"type:uuid;index"`
	Offer           *Offer         `gorm:"foreignKey:OfferID"`
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
type BudgetDimensionCategory struct {
	ID           string    `gorm:"type:varchar(50);primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(200);not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description,omitempty"`
	DisplayOrder int       `gorm:"not null;default:0;column:display_order" json:"displayOrder"`
	IsActive     bool      `gorm:"not null;default:true;column:is_active" json:"isActive"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName returns the table name for BudgetDimensionCategory
func (BudgetDimensionCategory) TableName() string {
	return "budget_dimension_categories"
}

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

// ActivityType represents the type of entity an activity is associated with
type ActivityType string

const (
	ActivityTypeCustomer     ActivityType = "Customer"
	ActivityTypeContact      ActivityType = "Contact"
	ActivityTypeProject      ActivityType = "Project"
	ActivityTypeOffer        ActivityType = "Offer"
	ActivityTypeDeal         ActivityType = "Deal"
	ActivityTypeFile         ActivityType = "File"
	ActivityTypeNotification ActivityType = "Notification"
)

// Activity represents an event log entry for any entity
type Activity struct {
	BaseModel
	TargetType  ActivityType `gorm:"type:varchar(50);not null;index"`
	TargetID    uuid.UUID    `gorm:"type:uuid;not null;index"`
	Title       string       `gorm:"type:varchar(200);not null"`
	Body        string       `gorm:"type:varchar(2000)"`
	OccurredAt  time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	CreatorName string       `gorm:"type:varchar(200)"`
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
