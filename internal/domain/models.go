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
	Name         string     `gorm:"type:varchar(200);not null;index"`
	Email        string     `gorm:"type:varchar(255);not null"`
	Phone        string     `gorm:"type:varchar(50);not null"`
	Role         string     `gorm:"type:varchar(120)"`
	CustomerID   *uuid.UUID `gorm:"type:uuid;index"`
	Customer     *Customer  `gorm:"foreignKey:CustomerID"`
	CustomerName string     `gorm:"type:varchar(200)"`
	ProjectID    *uuid.UUID `gorm:"type:uuid;index"`
	Project      *Project   `gorm:"foreignKey:ProjectID"`
	ProjectName  string     `gorm:"type:varchar(200)"`
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

// OfferItem represents a line item in an offer
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

// ActivityType represents the type of entity an activity is associated with
type ActivityType string

const (
	ActivityTypeCustomer     ActivityType = "Customer"
	ActivityTypeContact      ActivityType = "Contact"
	ActivityTypeProject      ActivityType = "Project"
	ActivityTypeOffer        ActivityType = "Offer"
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
