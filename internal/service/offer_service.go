package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/datawarehouse"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OfferService struct {
	offerRepo        *repository.OfferRepository
	offerItemRepo    *repository.OfferItemRepository
	customerRepo     *repository.CustomerRepository
	projectRepo      *repository.ProjectRepository
	budgetItemRepo   *repository.BudgetItemRepository
	fileRepo         *repository.FileRepository
	activityRepo     *repository.ActivityRepository
	userRepo         *repository.UserRepository
	companyService   *CompanyService
	numberSeqService *NumberSequenceService
	dwClient         *datawarehouse.Client
	logger           *zap.Logger
	db               *gorm.DB
}

func NewOfferService(
	offerRepo *repository.OfferRepository,
	offerItemRepo *repository.OfferItemRepository,
	customerRepo *repository.CustomerRepository,
	projectRepo *repository.ProjectRepository,
	budgetItemRepo *repository.BudgetItemRepository,
	fileRepo *repository.FileRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
	companyService *CompanyService,
	numberSeqService *NumberSequenceService,
	logger *zap.Logger,
	db *gorm.DB,
) *OfferService {
	return &OfferService{
		offerRepo:        offerRepo,
		offerItemRepo:    offerItemRepo,
		customerRepo:     customerRepo,
		projectRepo:      projectRepo,
		budgetItemRepo:   budgetItemRepo,
		fileRepo:         fileRepo,
		activityRepo:     activityRepo,
		userRepo:         userRepo,
		companyService:   companyService,
		numberSeqService: numberSeqService,
		logger:           logger,
		db:               db,
	}
}

// SetDataWarehouseClient sets the data warehouse client for syncing financial data.
// This is called after construction because the DW client is optional.
func (s *OfferService) SetDataWarehouseClient(client *datawarehouse.Client) {
	s.dwClient = client
}

// Create creates a new offer with initial items
func (s *OfferService) Create(ctx context.Context, req *domain.CreateOfferRequest) (*domain.OfferDTO, error) {
	resp, err := s.CreateWithProjectResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Offer, nil
}

// CreateWithProjectResponse creates a new offer and returns both the offer and any auto-created project
// Supports three scenarios:
//   - CustomerID only: Creates offer for that customer, optionally auto-creates project if CreateProject=true
//   - ProjectID only: Inherits customer from existing project
//   - Both IDs: Uses provided customer, links to specified project
func (s *OfferService) CreateWithProjectResponse(ctx context.Context, req *domain.CreateOfferRequest) (*domain.OfferWithProjectResponse, error) {
	// Validate: at least one of customerId or projectId must be provided
	if req.CustomerID == nil && req.ProjectID == nil {
		return nil, ErrMissingCustomerOrProject
	}

	var customer *domain.Customer
	var customerID *uuid.UUID
	var customerName string
	var err error

	// Handle customer - any combination is valid as long as at least one of customerID or projectID is provided
	if req.CustomerID != nil {
		// CustomerID is provided - validate and use it
		customer, err = s.customerRepo.GetByID(ctx, *req.CustomerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCustomerNotFound
			}
			return nil, fmt.Errorf("failed to verify customer: %w", err)
		}
		customerID = &customer.ID
		customerName = customer.Name
	}
	// Note: If only ProjectID is provided (no CustomerID), offer will have no customer
	// The project and customer don't need to be related - any combination is valid

	// Calculate value from items (if provided)
	totalValue := 0.0
	items := make([]domain.OfferItem, len(req.Items))
	for i, itemReq := range req.Items {
		margin := mapper.CalculateMargin(itemReq.Cost, itemReq.Revenue)
		items[i] = domain.OfferItem{
			Discipline:  itemReq.Discipline,
			Cost:        itemReq.Cost,
			Revenue:     itemReq.Revenue,
			Margin:      margin,
			Description: itemReq.Description,
			Quantity:    itemReq.Quantity,
			Unit:        itemReq.Unit,
		}
		totalValue += itemReq.Revenue
	}

	// Set defaults for optional fields
	companyID := req.CompanyID
	if companyID == "" && customer != nil && customer.CompanyID != nil {
		companyID = *customer.CompanyID
	}
	if companyID == "" {
		companyID = domain.CompanyGruppen // Default fallback
	}

	phase := req.Phase
	if phase == "" {
		// Default to in_progress phase for POST /offers
		// Draft phase is reserved for inquiries via /inquiries endpoint
		phase = domain.OfferPhaseInProgress
	}

	status := req.Status
	if status == "" {
		status = domain.OfferStatusActive
	}

	probability := 0
	if req.Probability != nil {
		probability = *req.Probability
	}

	// Auto-assign responsible user from company default if not provided
	responsibleUserID := req.ResponsibleUserID
	if responsibleUserID == "" && s.companyService != nil {
		defaultResponsible := s.companyService.GetDefaultOfferResponsible(ctx, companyID)
		if defaultResponsible != nil && *defaultResponsible != "" {
			responsibleUserID = *defaultResponsible
			s.logger.Debug("auto-assigned responsible user from company default",
				zap.String("companyID", string(companyID)),
				zap.String("responsibleUserID", responsibleUserID))
		}
	}

	// Validate expiration date if provided
	if req.ExpirationDate != nil && req.SentDate != nil {
		if req.ExpirationDate.Before(*req.SentDate) {
			return nil, ErrExpirationDateBeforeSentDate
		}
	}

	// Calculate expiration date: use provided value, or default to 60 days after sent date if sent
	var expirationDate *time.Time
	if req.ExpirationDate != nil {
		expirationDate = req.ExpirationDate
	} else if req.SentDate != nil {
		expDate := req.SentDate.AddDate(0, 0, 60)
		expirationDate = &expDate
	}

	offer := &domain.Offer{
		Title:               req.Title,
		CustomerID:          customerID,
		CustomerName:        customerName,
		CompanyID:           companyID,
		Phase:               phase,
		Probability:         probability,
		Value:               totalValue,
		Status:              status,
		ResponsibleUserID:   responsibleUserID,
		ResponsibleUserName: "", // Populated by handler/external user lookup if needed
		Description:         req.Description,
		Notes:               req.Notes,
		DueDate:             req.DueDate,
		Cost:                req.Cost,
		Location:            req.Location,
		SentDate:            req.SentDate,
		ExpirationDate:      expirationDate,
		Items:               items,
	}

	// Set user tracking fields on creation
	if userCtx, ok := auth.FromContext(ctx); ok {
		offer.CreatedByID = userCtx.UserID.String()
		offer.CreatedByName = userCtx.DisplayName
		offer.UpdatedByID = userCtx.UserID.String()
		offer.UpdatedByName = userCtx.DisplayName
	}

	// Generate offer number only for non-draft offers
	// Draft offers should NOT have offer numbers - they get one when transitioning out of draft
	if !s.isDraftPhase(phase) {
		if err := s.generateOfferNumberIfNeeded(ctx, offer); err != nil {
			return nil, err
		}
	}

	// Track project creation result
	var projectLinkRes *projectLinkResult

	// Handle project linking/creation
	if req.ProjectID != nil {
		// User provided a project ID - validate and link
		projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, req.ProjectID, false)
		if err != nil {
			return nil, err
		}
		offer.ProjectID = &projectLinkRes.ProjectID
	} else if !s.isDraftPhase(phase) {
		// No ProjectID provided and not a draft - auto-create project
		projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, nil, true)
		if err != nil {
			return nil, err
		}
		offer.ProjectID = &projectLinkRes.ProjectID
	}

	// Create the offer
	if err := s.offerRepo.Create(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	// Link offer to project if we have one
	if projectLinkRes != nil {
		if err := s.offerRepo.LinkToProject(ctx, offer.ID, projectLinkRes.ProjectID); err != nil {
			s.logger.Warn("failed to link offer to project", zap.Error(err))
		}
		// Sync project customer after linking new offer
		if err := s.syncProjectCustomer(ctx, projectLinkRes.ProjectID); err != nil {
			s.logger.Warn("failed to sync project customer after offer creation",
				zap.String("offerID", offer.ID.String()),
				zap.String("projectID", projectLinkRes.ProjectID.String()),
				zap.Error(err))
		}
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, offer.ID)
	if err != nil {
		s.logger.Warn("failed to reload offer after create", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Tilbudet '%s' ble opprettet for kunde %s", offer.Title, offer.CustomerName)
	if projectLinkRes != nil && projectLinkRes.ProjectCreated {
		activityBody = fmt.Sprintf("Tilbudet '%s' ble opprettet for kunde %s med auto-opprettet prosjekt '%s'",
			offer.Title, offer.CustomerName, projectLinkRes.Project.Name)
	}
	s.logActivity(ctx, offer.ID, offer.Title, "Tilbud opprettet", activityBody)

	offerDTO := mapper.ToOfferDTO(offer)

	response := &domain.OfferWithProjectResponse{
		Offer: &offerDTO,
	}

	// Include project in response if one was created or linked
	if projectLinkRes != nil {
		projectDTO := mapper.ToProjectDTO(projectLinkRes.Project)
		response.Project = &projectDTO
		response.ProjectCreated = projectLinkRes.ProjectCreated
	}

	return response, nil
}

// GetByID retrieves an offer by ID with items
func (s *OfferService) GetByID(ctx context.Context, id uuid.UUID) (*domain.OfferWithItemsDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Convert to DTO
	items := make([]domain.OfferItemDTO, len(offer.Items))
	for i, item := range offer.Items {
		items[i] = mapper.ToOfferItemDTO(&item)
	}

	dto := &domain.OfferWithItemsDTO{
		ID:                  offer.ID,
		Title:               offer.Title,
		OfferNumber:         offer.OfferNumber,
		ExternalReference:   offer.ExternalReference,
		CustomerID:          offer.CustomerID,
		CustomerName:        offer.CustomerName,
		CompanyID:           offer.CompanyID,
		Phase:               offer.Phase,
		Probability:         offer.Probability,
		Value:               offer.Value,
		Status:              offer.Status,
		CreatedAt:           offer.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           offer.UpdatedAt.UTC().Format(time.RFC3339),
		ResponsibleUserID:   offer.ResponsibleUserID,
		ResponsibleUserName: offer.ResponsibleUserName,
		Items:               items,
		Description:         offer.Description,
		Notes:               offer.Notes,
	}

	return dto, nil
}

// GetByIDWithBudgetItems retrieves an offer with budget items and summary
func (s *OfferService) GetByIDWithBudgetItems(ctx context.Context, id uuid.UUID) (*domain.OfferDetailDTO, error) {
	offer, items, err := s.offerRepo.GetByIDWithBudgetItems(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer with budget items: %w", err)
	}

	// Convert offer to DTO
	offerDTO := mapper.ToOfferDTO(offer)

	// Convert budget items to DTOs
	itemDTOs := make([]domain.BudgetItemDTO, len(items))
	for i, item := range items {
		itemDTOs[i] = mapper.ToBudgetItemDTO(&item)
	}

	// Get budget summary
	summary, err := s.offerRepo.GetBudgetSummary(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get budget summary", zap.Error(err))
	}

	var summaryDTO *domain.BudgetSummaryDTO
	if summary != nil {
		summaryDTO = &domain.BudgetSummaryDTO{
			ParentType:    domain.BudgetParentOffer,
			ParentID:      id,
			TotalCost:     summary.TotalCost,
			TotalRevenue:  summary.TotalRevenue,
			TotalProfit:   summary.TotalProfit,
			MarginPercent: summary.MarginPercent,
			ItemCount:     summary.ItemCount,
		}
	}

	// Get files count
	filesCount, _ := s.offerRepo.GetFilesCount(ctx, id)

	dto := &domain.OfferDetailDTO{
		OfferDTO:      offerDTO,
		BudgetItems:   itemDTOs,
		BudgetSummary: summaryDTO,
		FilesCount:    filesCount,
	}

	return dto, nil
}

// Update updates an existing offer
func (s *OfferService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateOfferRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Check if offer is in a closed state
	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	// Block updates to offers in order phase - use dedicated methods instead
	// (UpdateOfferHealth, UpdateOfferSpent, UpdateOfferInvoiced, CompleteOffer)
	if offer.Phase == domain.OfferPhaseOrder {
		return nil, ErrOfferAlreadyClosed
	}

	// Block transitions to terminal phases - must use dedicated endpoints
	// AcceptOrder, RejectOffer, or ExpireOffer handle proper phase transitions
	if s.isClosedPhase(req.Phase) || req.Phase == domain.OfferPhaseOrder {
		return nil, ErrOfferCannotAdvanceToTerminalPhase
	}

	// Track if we're transitioning from draft to non-draft
	wasInDraft := s.isDraftPhase(offer.Phase)
	willBeInDraft := s.isDraftPhase(req.Phase)
	transitioningFromDraft := wasInDraft && !willBeInDraft

	offer.Title = req.Title
	offer.Phase = req.Phase
	offer.Probability = req.Probability
	offer.Status = req.Status
	offer.ResponsibleUserID = req.ResponsibleUserID
	offer.Description = req.Description
	offer.Notes = req.Notes
	offer.DueDate = req.DueDate
	offer.Cost = req.Cost
	offer.Location = req.Location
	offer.SentDate = req.SentDate

	// Handle expiration date: validate and set
	sentDate := req.SentDate
	if sentDate == nil {
		sentDate = offer.SentDate // Use existing sent date if not being updated
	}
	if req.ExpirationDate != nil {
		// Validate expiration date is not before sent date
		if sentDate != nil && req.ExpirationDate.Before(*sentDate) {
			return nil, ErrExpirationDateBeforeSentDate
		}
		offer.ExpirationDate = req.ExpirationDate
	} else if sentDate != nil && offer.ExpirationDate == nil {
		// Default to 60 days after sent date if not set
		expDate := sentDate.AddDate(0, 0, 60)
		offer.ExpirationDate = &expDate
	}

	// Recalculate value from items
	offer.Value = mapper.CalculateOfferValue(offer.Items)

	// Handle offer number generation when transitioning from draft to non-draft
	if transitioningFromDraft {
		if err := s.generateOfferNumberIfNeeded(ctx, offer); err != nil {
			return nil, err
		}

		// Auto-create project if transitioning from draft and no project is linked
		if offer.ProjectID == nil {
			projectLinkRes, err := s.ensureProjectForOffer(ctx, offer, nil, true)
			if err != nil {
				return nil, err
			}
			offer.ProjectID = &projectLinkRes.ProjectID
			offer.ProjectName = projectLinkRes.Project.Name
		}
	}

	// Validate offer number rules:
	// - Draft offers should NOT have an offer number
	// - Non-draft offers MUST have an offer number
	if willBeInDraft && offer.OfferNumber != "" {
		// This shouldn't normally happen, but if someone tries to update to draft with a number, reject it
		return nil, ErrDraftOfferCannotHaveNumber
	}
	if !willBeInDraft && offer.OfferNumber == "" {
		// Non-draft offer without number - this is a data integrity issue
		return nil, ErrNonDraftOfferMustHaveNumber
	}

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		offer.UpdatedByID = userCtx.UserID.String()
		offer.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after update", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, offer.ID, offer.Title, "Tilbud oppdatert",
		fmt.Sprintf("Tilbudet '%s' ble oppdatert", offer.Title))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// Delete removes an offer
func (s *OfferService) Delete(ctx context.Context, id uuid.UUID) error {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOfferNotFound
		}
		return fmt.Errorf("failed to get offer: %w", err)
	}

	// Store project ID and name before deletion
	projectID := offer.ProjectID
	projectName := offer.ProjectName

	if err := s.offerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete offer: %w", err)
	}

	// If offer was linked to a project, check if project should be deleted
	if projectID != nil {
		// Check if there are any remaining offers in the project
		remainingOffers, err := s.offerRepo.ListByProject(ctx, *projectID)
		if err != nil {
			s.logger.Warn("failed to check remaining offers in project",
				zap.String("projectID", projectID.String()),
				zap.Error(err))
		} else if len(remainingOffers) == 0 {
			// No remaining offers - delete the project
			if err := s.projectRepo.Delete(ctx, *projectID); err != nil {
				s.logger.Warn("failed to delete empty project after last offer deletion",
					zap.String("projectID", projectID.String()),
					zap.Error(err))
			} else {
				s.logger.Info("deleted empty project after last offer was deleted",
					zap.String("projectID", projectID.String()),
					zap.String("projectName", projectName))
				// Log activity for project deletion (only if customer exists)
				if offer.CustomerID != nil {
					s.logActivityOnTarget(ctx, domain.ActivityTargetCustomer, *offer.CustomerID, offer.CustomerName,
						"Prosjekt slettet", fmt.Sprintf("Prosjektet '%s' ble slettet (ingen gjenværende tilbud)", projectName))
				}
			}
		} else {
			// Still has offers - sync project customer
			if err := s.syncProjectCustomer(ctx, *projectID); err != nil {
				s.logger.Warn("failed to sync project customer after offer deletion",
					zap.String("offerID", id.String()),
					zap.String("projectID", projectID.String()),
					zap.Error(err))
			}
		}
	}

	// Log activity (on customer since offer is deleted, only if customer exists)
	if offer.CustomerID != nil {
		s.logActivityOnTarget(ctx, domain.ActivityTargetCustomer, *offer.CustomerID, offer.CustomerName,
			"Tilbud slettet", fmt.Sprintf("Tilbudet '%s' ble slettet", offer.Title))
	}

	return nil
}

// List returns a paginated list of offers with default sorting
func (s *OfferService) List(ctx context.Context, page, pageSize int, customerID, projectID *uuid.UUID, phase *domain.OfferPhase) (*domain.PaginatedResponse, error) {
	return s.ListWithSort(ctx, page, pageSize, customerID, projectID, phase, repository.DefaultSortConfig())
}

// ListWithSort returns a paginated list of offers with custom sorting
func (s *OfferService) ListWithSort(ctx context.Context, page, pageSize int, customerID, projectID *uuid.UUID, phase *domain.OfferPhase, sort repository.SortConfig) (*domain.PaginatedResponse, error) {
	// Clamp page size
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if page < 1 {
		page = 1
	}

	filters := &repository.OfferFilters{
		CustomerID: customerID,
		ProjectID:  projectID,
		Phase:      phase,
	}

	offers, total, err := s.offerRepo.ListWithFilters(ctx, page, pageSize, filters, sort)
	if err != nil {
		return nil, fmt.Errorf("failed to list offers: %w", err)
	}

	dtos := make([]domain.OfferDTO, len(offers))
	for i, offer := range offers {
		dtos[i] = mapper.ToOfferDTO(&offer)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &domain.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ============================================================================
// Budget Methods
// ============================================================================

// GetBudgetSummary returns aggregated budget totals for an offer
func (s *OfferService) GetBudgetSummary(ctx context.Context, id uuid.UUID) (*domain.BudgetSummaryDTO, error) {
	// Verify offer exists
	_, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	summary, err := s.offerRepo.GetBudgetSummary(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget summary: %w", err)
	}

	dto := &domain.BudgetSummaryDTO{
		ParentType:    domain.BudgetParentOffer,
		ParentID:      id,
		TotalCost:     summary.TotalCost,
		TotalRevenue:  summary.TotalRevenue,
		TotalProfit:   summary.TotalProfit,
		MarginPercent: summary.MarginPercent,
		ItemCount:     summary.ItemCount,
	}

	return dto, nil
}

// RecalculateTotals recalculates the offer value from budget items
func (s *OfferService) RecalculateTotals(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	// First check if the offer exists
	_, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Calculate totals from budget items
	newValue, err := s.offerRepo.CalculateTotalsFromBudgetItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate totals: %w", err)
	}

	// Reload offer with updated values
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after recalculate", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, id, offer.Title, "Tilbudstotaler omberegnet",
		fmt.Sprintf("Tilbudet '%s' verdi oppdatert til %.2f fra budsjettposter", offer.Title, newValue))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// ============================================================================
// Item Methods
// ============================================================================

// GetItems returns all items for an offer
func (s *OfferService) GetItems(ctx context.Context, offerID uuid.UUID) ([]domain.OfferItemDTO, error) {
	items, err := s.offerItemRepo.ListByOffer(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer items: %w", err)
	}

	dtos := make([]domain.OfferItemDTO, len(items))
	for i, item := range items {
		dtos[i] = mapper.ToOfferItemDTO(&item)
	}

	return dtos, nil
}

// AddItem adds a new item to an offer
func (s *OfferService) AddItem(ctx context.Context, offerID uuid.UUID, req *domain.CreateOfferItemRequest) (*domain.OfferItemDTO, error) {
	// Verify offer exists and keep reference for recalculating totals
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	margin := mapper.CalculateMargin(req.Cost, req.Revenue)
	item := &domain.OfferItem{
		OfferID:     offerID,
		Discipline:  req.Discipline,
		Cost:        req.Cost,
		Revenue:     req.Revenue,
		Margin:      margin,
		Description: req.Description,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
	}

	if err := s.offerItemRepo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create offer item: %w", err)
	}

	// Update cached offer totals so listings and dashboards remain accurate
	offer.Items = append(offer.Items, *item)
	offer.Value = mapper.CalculateOfferValue(offer.Items)
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer totals: %w", err)
	}

	dto := mapper.ToOfferItemDTO(item)
	return &dto, nil
}

// ============================================================================
// File Methods
// ============================================================================

// GetFiles returns all files for an offer
func (s *OfferService) GetFiles(ctx context.Context, offerID uuid.UUID) ([]domain.FileDTO, error) {
	files, err := s.fileRepo.ListByOffer(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer files: %w", err)
	}

	dtos := make([]domain.FileDTO, len(files))
	for i, file := range files {
		dtos[i] = mapper.ToFileDTO(&file)
	}

	return dtos, nil
}

// ============================================================================
// Activity Methods
// ============================================================================

// GetActivities returns activities for an offer
func (s *OfferService) GetActivities(ctx context.Context, id uuid.UUID, limit int) ([]domain.ActivityDTO, error) {
	activities, err := s.activityRepo.ListByTarget(ctx, domain.ActivityTargetOffer, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
	}

	return dtos, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// projectLinkResult contains the result of ensureProjectForOffer
type projectLinkResult struct {
	ProjectID      uuid.UUID
	Project        *domain.Project
	ProjectCreated bool
}

// ensureProjectForOffer validates an existing project or creates a new one for the offer.
// If requestedProjectID is provided, validates it exists and matches company.
// If requestedProjectID is nil and createProject is true, creates a new project automatically.
// If requestedProjectID is nil and createProject is false, returns nil (no project needed).
// Returns the project link result and whether a new project was created.
func (s *OfferService) ensureProjectForOffer(ctx context.Context, offer *domain.Offer, requestedProjectID *uuid.UUID, createProject bool) (*projectLinkResult, error) {
	// If a project ID is provided, validate and use it
	if requestedProjectID != nil {
		project, err := s.projectRepo.GetByID(ctx, *requestedProjectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProjectNotFound
			}
			return nil, fmt.Errorf("failed to get project: %w", err)
		}

		// Validate project is not cancelled
		if project.Phase == domain.ProjectPhaseCancelled {
			return nil, ErrCannotAddOfferToCancelledProject
		}

		// Projects are now cross-company, so no company validation needed

		// Log when linking offer with different customer for audit purposes
		// Check if both have the same customer (both nil or both point to same ID)
		offerCustomerMatches := (project.CustomerID == nil && offer.CustomerID == nil) ||
			(project.CustomerID != nil && offer.CustomerID != nil && *project.CustomerID == *offer.CustomerID)
		if !offerCustomerMatches {
			s.logger.Info("linking offer to project with different customer",
				zap.String("offerID", offer.ID.String()),
				zap.String("offerCustomerID", func() string {
					if offer.CustomerID != nil {
						return offer.CustomerID.String()
					}
					return "nil"
				}()),
				zap.String("offerCustomerName", offer.CustomerName),
				zap.String("projectID", project.ID.String()),
				zap.String("projectCustomerID", func() string {
					if project.CustomerID != nil {
						return project.CustomerID.String()
					}
					return "nil"
				}()),
				zap.String("projectCustomerName", project.CustomerName))
		}

		return &projectLinkResult{
			ProjectID:      project.ID,
			Project:        project,
			ProjectCreated: false,
		}, nil
	}

	// No project ID provided - check if we should create one
	if !createProject {
		return nil, nil // No project creation requested
	}

	// Create project automatically
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Auto-create project (as a folder/container for offers)
	// Projects are now simplified - no manager or company fields
	project := &domain.Project{
		Name:          fmt.Sprintf("[AUTO] %s", offer.Title),
		CustomerID:    offer.CustomerID, // Already a pointer, use directly
		CustomerName:  offer.CustomerName,
		Phase:         domain.ProjectPhaseTilbud,
		StartDate:     time.Now(),
		Description:   offer.Description,
		CreatedByID:   userCtx.UserID.String(),
		CreatedByName: userCtx.DisplayName,
		UpdatedByID:   userCtx.UserID.String(),
		UpdatedByName: userCtx.DisplayName,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProjectCreationFailed, err)
	}

	s.logger.Info("auto-created project for offer",
		zap.String("offerID", offer.ID.String()),
		zap.String("projectID", project.ID.String()),
		zap.String("projectName", project.Name))

	// Log activity on the new project
	s.logActivityOnTarget(ctx, domain.ActivityTargetProject, project.ID, project.Name,
		"Prosjekt auto-opprettet",
		fmt.Sprintf("Prosjektet '%s' ble automatisk opprettet for tilbud '%s'", project.Name, offer.Title))

	return &projectLinkResult{
		ProjectID:      project.ID,
		Project:        project,
		ProjectCreated: true,
	}, nil
}

// isClosedPhase returns true if the phase is a terminal state
func (s *OfferService) isClosedPhase(phase domain.OfferPhase) bool {
	return phase == domain.OfferPhaseCompleted ||
		phase == domain.OfferPhaseLost ||
		phase == domain.OfferPhaseExpired
}

// isDraftPhase returns true if the phase is draft
func (s *OfferService) isDraftPhase(phase domain.OfferPhase) bool {
	return phase == domain.OfferPhaseDraft
}

// generateOfferNumberIfNeeded generates an offer number for the offer if it doesn't have one.
// This should be called when transitioning from draft to any other phase.
// Returns an error if the offer number generation fails.
func (s *OfferService) generateOfferNumberIfNeeded(ctx context.Context, offer *domain.Offer) error {
	// Only generate if not already set
	if offer.OfferNumber != "" {
		return nil
	}

	// Validate company ID for offer number generation
	if !domain.IsValidCompanyID(string(offer.CompanyID)) {
		return ErrInvalidCompanyID
	}

	// Check if number sequence service is available
	if s.numberSeqService == nil {
		s.logger.Error("number sequence service not available",
			zap.String("offerID", offer.ID.String()))
		return fmt.Errorf("%w: number sequence service not configured", ErrOfferNumberGenerationFailed)
	}

	offerNumber, err := s.numberSeqService.GenerateOfferNumber(ctx, offer.CompanyID)
	if err != nil {
		s.logger.Error("failed to generate offer number",
			zap.Error(err),
			zap.String("offerID", offer.ID.String()),
			zap.String("companyID", string(offer.CompanyID)))
		return fmt.Errorf("%w: %v", ErrOfferNumberGenerationFailed, err)
	}

	offer.OfferNumber = offerNumber
	s.logger.Info("generated offer number",
		zap.String("offerID", offer.ID.String()),
		zap.String("offerNumber", offerNumber))

	return nil
}

// logActivity creates an activity log entry for an offer
func (s *OfferService) logActivity(ctx context.Context, offerID uuid.UUID, offerTitle, title, body string) {
	s.logActivityOnTarget(ctx, domain.ActivityTargetOffer, offerID, offerTitle, title, body)
}

// logActivityOnTarget creates an activity log entry for any target
func (s *OfferService) logActivityOnTarget(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, targetName, title, body string) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		s.logger.Warn("no user context for activity logging")
		return
	}

	activity := &domain.Activity{
		TargetType:  targetType,
		TargetID:    targetID,
		TargetName:  targetName,
		Title:       title,
		Body:        body,
		CreatorName: userCtx.DisplayName,
		CreatorID:   userCtx.UserID.String(),
		CompanyID:   &userCtx.CompanyID,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to log activity", zap.Error(err))
	}
}

// ============================================================================
// Project Customer Synchronization
// ============================================================================

// syncProjectCustomer updates the project's customer based on its offers.
// If all offers have the same customer, the project gets that customer.
// If offers have different customers (or no offers), the project customer is cleared.
// This should be called whenever:
// - An offer is created with a project link
// - An offer is linked to a project
// - An offer is unlinked from a project
// - An offer's customer is changed (if it's linked to a project)
// - An offer is deleted (if it was linked to a project)
func (s *OfferService) syncProjectCustomer(ctx context.Context, projectID uuid.UUID) error {
	// Get all unique customers for the project's offers
	customers, err := s.offerRepo.GetUniqueCustomersForProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get unique customers for project: %w", err)
	}

	var customerID *uuid.UUID
	var customerName string

	// Only set customer if exactly one unique customer exists across all offers
	if len(customers) == 1 {
		customerID = &customers[0].CustomerID
		customerName = customers[0].CustomerName
		s.logger.Debug("syncing project customer - all offers have same customer",
			zap.String("projectID", projectID.String()),
			zap.String("customerID", customerID.String()),
			zap.String("customerName", customerName))
	} else if len(customers) > 1 {
		// Multiple different customers - clear project customer
		s.logger.Info("syncing project customer - offers have different customers, clearing project customer",
			zap.String("projectID", projectID.String()),
			zap.Int("uniqueCustomerCount", len(customers)))
	} else {
		// No offers - clear project customer
		s.logger.Debug("syncing project customer - no offers, clearing project customer",
			zap.String("projectID", projectID.String()))
	}

	// Update the project's customer
	if err := s.projectRepo.UpdateCustomer(ctx, projectID, customerID, customerName); err != nil {
		return fmt.Errorf("failed to update project customer: %w", err)
	}

	return nil
}

// syncProjectPhase updates the project's phase based on its offers' phases.
// Logic:
//   - If any offer is in "order" phase -> project should be "working"
//   - If all offers are "completed" -> project should be "completed"
//   - If offers are only in pre-order phases (draft/in_progress/sent) -> project stays in "tilbud"
//
// This should be called whenever:
//   - An offer transitions to order phase (AcceptOrder)
//   - An offer transitions to completed phase (CompleteOffer)
//   - An offer is reopened from completed to order (ReopenOffer)
func (s *OfferService) syncProjectPhase(ctx context.Context, projectID uuid.UUID) error {
	// Get the project
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Don't modify cancelled projects
	if project.Phase == domain.ProjectPhaseCancelled {
		return nil
	}

	// Get all offers for this project
	offers, err := s.offerRepo.ListByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get offers for project: %w", err)
	}

	if len(offers) == 0 {
		return nil // No offers, don't change phase
	}

	// Determine target phase based on offers
	var targetPhase domain.ProjectPhase
	hasOrderPhase := false
	allCompleted := true
	hasAnyOffer := false

	for _, offer := range offers {
		// Skip lost/expired offers when determining phase
		if offer.Phase == domain.OfferPhaseLost || offer.Phase == domain.OfferPhaseExpired {
			continue
		}
		hasAnyOffer = true

		if offer.Phase == domain.OfferPhaseOrder {
			hasOrderPhase = true
			allCompleted = false
		} else if offer.Phase != domain.OfferPhaseCompleted {
			allCompleted = false
		}
	}

	if !hasAnyOffer {
		return nil // Only lost/expired offers, don't change phase
	}

	if hasOrderPhase {
		targetPhase = domain.ProjectPhaseWorking
	} else if allCompleted {
		targetPhase = domain.ProjectPhaseCompleted
	} else {
		targetPhase = domain.ProjectPhaseTilbud
	}

	// Only update if phase is different and transition is valid
	if project.Phase != targetPhase && project.Phase.CanTransitionTo(targetPhase) {
		s.logger.Info("syncing project phase based on offers",
			zap.String("projectID", projectID.String()),
			zap.String("oldPhase", string(project.Phase)),
			zap.String("newPhase", string(targetPhase)))

		if err := s.projectRepo.UpdatePhase(ctx, projectID, targetPhase); err != nil {
			return fmt.Errorf("failed to update project phase: %w", err)
		}
	}

	return nil
}

// GetOfferSuppliers returns all suppliers linked to an offer with their relationship details
func (s *OfferService) GetOfferSuppliers(ctx context.Context, offerID uuid.UUID) ([]domain.OfferSupplierWithDetailsDTO, error) {
	offerSuppliers, err := s.offerRepo.GetOfferSuppliers(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer suppliers: %w", err)
	}

	return mapper.OfferSuppliersToWithDetailsDTOs(offerSuppliers), nil
}

// AddSupplierToOffer links a supplier to an offer
func (s *OfferService) AddSupplierToOffer(ctx context.Context, offerID uuid.UUID, req *domain.AddOfferSupplierRequest) (*domain.OfferSupplierWithDetailsDTO, error) {
	// Get the offer to verify it exists and get denormalized fields
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Check if supplier is already linked
	exists, err := s.offerRepo.OfferSupplierExists(ctx, offerID, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to check offer supplier existence: %w", err)
	}
	if exists {
		return nil, ErrOfferSupplierAlreadyExists
	}

	// Get supplier for denormalized name
	var supplier domain.Supplier
	if err := s.db.WithContext(ctx).Where("id = ?", req.SupplierID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	// Validate and get contact if provided
	var contactID *uuid.UUID
	var contactName string
	if req.ContactID != nil {
		var contact domain.SupplierContact
		if err := s.db.WithContext(ctx).Where("id = ? AND supplier_id = ?", req.ContactID, req.SupplierID).First(&contact).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("contact not found or does not belong to this supplier")
			}
			return nil, fmt.Errorf("failed to get contact: %w", err)
		}
		contactID = req.ContactID
		contactName = contact.FullName()
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = domain.OfferSupplierStatusActive
	}

	offerSupplier := &domain.OfferSupplier{
		OfferID:      offerID,
		SupplierID:   req.SupplierID,
		SupplierName: supplier.Name,
		OfferTitle:   offer.Title,
		Status:       status,
		Notes:        req.Notes,
		ContactID:    contactID,
		ContactName:  contactName,
	}

	// Set user tracking fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		offerSupplier.CreatedByID = userCtx.UserID.String()
		offerSupplier.CreatedByName = userCtx.DisplayName
		offerSupplier.UpdatedByID = userCtx.UserID.String()
		offerSupplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.CreateOfferSupplier(ctx, offerSupplier); err != nil {
		return nil, fmt.Errorf("failed to create offer supplier: %w", err)
	}

	// Fetch the created relationship with supplier preloaded
	created, err := s.offerRepo.GetOfferSupplier(ctx, offerID, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created offer supplier: %w", err)
	}

	// Log activity
	s.logOfferSupplierActivity(ctx, offer, supplier.Name, "Leverandør lagt til", fmt.Sprintf("Leverandør '%s' lagt til tilbudet", supplier.Name))

	dto := mapper.OfferSupplierToWithDetailsDTO(created)
	return &dto, nil
}

// UpdateOfferSupplier updates the relationship between an offer and a supplier
func (s *OfferService) UpdateOfferSupplier(ctx context.Context, offerID, supplierID uuid.UUID, req *domain.UpdateOfferSupplierRequest) (*domain.OfferSupplierWithDetailsDTO, error) {
	// Get existing relationship
	offerSupplier, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get offer supplier: %w", err)
	}

	// Update fields
	if req.Status != "" {
		offerSupplier.Status = req.Status
	}
	offerSupplier.Notes = req.Notes

	// Update contact if provided
	if req.ContactID != nil {
		var contact domain.SupplierContact
		if err := s.db.WithContext(ctx).Where("id = ? AND supplier_id = ?", req.ContactID, supplierID).First(&contact).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("contact not found or does not belong to this supplier")
			}
			return nil, fmt.Errorf("failed to get contact: %w", err)
		}
		offerSupplier.ContactID = req.ContactID
		offerSupplier.ContactName = contact.FullName()
	}

	// Set user tracking fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		offerSupplier.UpdatedByID = userCtx.UserID.String()
		offerSupplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.UpdateOfferSupplier(ctx, offerSupplier); err != nil {
		return nil, fmt.Errorf("failed to update offer supplier: %w", err)
	}

	// Refetch to get updated data
	updated, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated offer supplier: %w", err)
	}

	dto := mapper.OfferSupplierToWithDetailsDTO(updated)
	return &dto, nil
}

// RemoveSupplierFromOffer removes the relationship between an offer and a supplier
func (s *OfferService) RemoveSupplierFromOffer(ctx context.Context, offerID, supplierID uuid.UUID) error {
	// Get the relationship first for activity logging
	offerSupplier, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOfferSupplierNotFound
		}
		return fmt.Errorf("failed to get offer supplier: %w", err)
	}

	// Get the offer for activity logging
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return fmt.Errorf("failed to get offer: %w", err)
	}

	if err := s.offerRepo.DeleteOfferSupplier(ctx, offerID, supplierID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOfferSupplierNotFound
		}
		return fmt.Errorf("failed to delete offer supplier: %w", err)
	}

	// Log activity
	s.logOfferSupplierActivity(ctx, offer, offerSupplier.SupplierName, "Leverandør fjernet", fmt.Sprintf("Leverandør '%s' fjernet fra tilbudet", offerSupplier.SupplierName))

	return nil
}

// UpdateOfferSupplierStatus updates only the status of an offer-supplier relationship
func (s *OfferService) UpdateOfferSupplierStatus(ctx context.Context, offerID, supplierID uuid.UUID, status domain.OfferSupplierStatus) (*domain.OfferSupplierWithDetailsDTO, error) {
	// Validate status
	if !status.IsValid() {
		return nil, ErrInvalidOfferSupplierStatus
	}

	// Get existing relationship
	offerSupplier, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get offer supplier: %w", err)
	}

	offerSupplier.Status = status

	// Set user tracking fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		offerSupplier.UpdatedByID = userCtx.UserID.String()
		offerSupplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.UpdateOfferSupplier(ctx, offerSupplier); err != nil {
		return nil, fmt.Errorf("failed to update offer supplier status: %w", err)
	}

	// Refetch to get updated data
	updated, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated offer supplier: %w", err)
	}

	dto := mapper.OfferSupplierToWithDetailsDTO(updated)
	return &dto, nil
}

// UpdateOfferSupplierNotes updates only the notes of an offer-supplier relationship
func (s *OfferService) UpdateOfferSupplierNotes(ctx context.Context, offerID, supplierID uuid.UUID, notes string) (*domain.OfferSupplierWithDetailsDTO, error) {
	// Get existing relationship
	offerSupplier, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get offer supplier: %w", err)
	}

	offerSupplier.Notes = notes

	// Set user tracking fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		offerSupplier.UpdatedByID = userCtx.UserID.String()
		offerSupplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.UpdateOfferSupplier(ctx, offerSupplier); err != nil {
		return nil, fmt.Errorf("failed to update offer supplier notes: %w", err)
	}

	// Refetch to get updated data
	updated, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated offer supplier: %w", err)
	}

	dto := mapper.OfferSupplierToWithDetailsDTO(updated)
	return &dto, nil
}

// UpdateOfferSupplierContact updates only the contact person of an offer-supplier relationship
func (s *OfferService) UpdateOfferSupplierContact(ctx context.Context, offerID, supplierID uuid.UUID, contactID *uuid.UUID) (*domain.OfferSupplierWithDetailsDTO, error) {
	// Get existing relationship
	offerSupplier, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get offer supplier: %w", err)
	}

	// Update or clear contact
	if contactID != nil {
		var contact domain.SupplierContact
		if err := s.db.WithContext(ctx).Where("id = ? AND supplier_id = ?", contactID, supplierID).First(&contact).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("contact not found or does not belong to this supplier")
			}
			return nil, fmt.Errorf("failed to get contact: %w", err)
		}
		offerSupplier.ContactID = contactID
		offerSupplier.ContactName = contact.FullName()
	} else {
		// Clear the contact
		offerSupplier.ContactID = nil
		offerSupplier.ContactName = ""
	}

	// Set user tracking fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		offerSupplier.UpdatedByID = userCtx.UserID.String()
		offerSupplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.UpdateOfferSupplier(ctx, offerSupplier); err != nil {
		return nil, fmt.Errorf("failed to update offer supplier contact: %w", err)
	}

	// Refetch to get updated data
	updated, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated offer supplier: %w", err)
	}

	dto := mapper.OfferSupplierToWithDetailsDTO(updated)
	return &dto, nil
}

// logOfferSupplierActivity logs an activity for offer-supplier operations
func (s *OfferService) logOfferSupplierActivity(ctx context.Context, offer *domain.Offer, supplierName, title, body string) {
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetOffer,
			TargetID:    offer.ID,
			TargetName:  offer.Title,
			Title:       title,
			Body:        body,
			CreatorName: userCtx.DisplayName,
		}
		_ = s.activityRepo.Create(ctx, activity)
	}
}
