package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
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
	var customerID uuid.UUID
	var customerName string
	var err error

	// Scenario B: ProjectID only - inherit customer from project
	if req.CustomerID == nil && req.ProjectID != nil {
		project, err := s.projectRepo.GetByID(ctx, *req.ProjectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProjectNotFound
			}
			return nil, fmt.Errorf("failed to get project for customer inheritance: %w", err)
		}

		// Project must have a customer to inherit from
		if project.CustomerID == nil {
			return nil, ErrProjectHasNoCustomer
		}

		// Fetch the customer to get full details
		customer, err = s.customerRepo.GetByID(ctx, *project.CustomerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCustomerNotFound
			}
			return nil, fmt.Errorf("failed to get customer from project: %w", err)
		}
		customerID = customer.ID
		customerName = customer.Name
	} else {
		// Scenario A or C: CustomerID is provided
		customer, err = s.customerRepo.GetByID(ctx, *req.CustomerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCustomerNotFound
			}
			return nil, fmt.Errorf("failed to verify customer: %w", err)
		}
		customerID = customer.ID
		customerName = customer.Name
	}

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
	if companyID == "" && customer.CompanyID != nil {
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

	// Handle project linking/creation based on explicit flags
	if req.ProjectID != nil {
		// User provided a project ID - validate and link
		projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, req.ProjectID, false)
		if err != nil {
			return nil, err
		}
		offer.ProjectID = &projectLinkRes.ProjectID
	} else if req.CreateProject && !s.isDraftPhase(phase) {
		// CreateProject=true and no ProjectID provided - auto-create project (only for non-draft offers)
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
		CreatedAt:           offer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           offer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

	// Store project ID before deletion for customer sync
	projectID := offer.ProjectID

	if err := s.offerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete offer: %w", err)
	}

	// Sync project customer if offer was linked to a project
	if projectID != nil {
		if err := s.syncProjectCustomer(ctx, *projectID); err != nil {
			s.logger.Warn("failed to sync project customer after offer deletion",
				zap.String("offerID", id.String()),
				zap.String("projectID", projectID.String()),
				zap.Error(err))
		}
	}

	// Log activity (on customer since offer is deleted)
	s.logActivityOnTarget(ctx, domain.ActivityTargetCustomer, offer.CustomerID, offer.CustomerName,
		"Tilbud slettet", fmt.Sprintf("Tilbudet '%s' ble slettet", offer.Title))

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
// Lifecycle Methods
// ============================================================================

// SendOffer transitions an offer from draft/in_progress to sent phase
func (s *OfferService) SendOffer(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase transition
	if offer.Phase != domain.OfferPhaseDraft && offer.Phase != domain.OfferPhaseInProgress {
		return nil, ErrOfferNotInDraftPhase
	}

	oldPhase := offer.Phase

	// Generate offer number if transitioning from draft (sent is non-draft)
	if s.isDraftPhase(oldPhase) {
		if err := s.generateOfferNumberIfNeeded(ctx, offer); err != nil {
			return nil, err
		}
	}

	offer.Phase = domain.OfferPhaseSent

	// Set sent date if not already set
	if offer.SentDate == nil {
		now := time.Now()
		offer.SentDate = &now
	}

	// Set expiration date to 60 days after sent date if not already set
	if offer.ExpirationDate == nil {
		expirationDate := offer.SentDate.AddDate(0, 0, 60)
		offer.ExpirationDate = &expirationDate
	}

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer phase: %w", err)
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after send", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, offer.ID, offer.Title, "Tilbud sendt",
		fmt.Sprintf("Tilbudet '%s' ble sendt til kunde (fase: %s -> %s)", offer.Title, oldPhase, offer.Phase))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// AcceptOffer transitions an offer from sent to order phase
// Supports optional project creation via CreateProject flag
func (s *OfferService) AcceptOffer(ctx context.Context, id uuid.UUID, req *domain.AcceptOfferRequest) (*domain.AcceptOfferResponse, error) {
	// Call AcceptOrder to transition the offer
	acceptReq := &domain.AcceptOrderRequest{}
	result, err := s.AcceptOrder(ctx, id, acceptReq)
	if err != nil {
		return nil, err
	}

	// If project creation is not requested, return without project
	if !req.CreateProject {
		return &domain.AcceptOfferResponse{
			Offer:   result.Offer,
			Project: nil,
		}, nil
	}

	// Get the offer for project creation
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer for project creation: %w", err)
	}

	// Create project from offer
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Determine project name
	projectName := req.ProjectName
	if projectName == "" {
		projectName = offer.Title
	}

	// Create project (as a folder/container for offers)
	// Projects are now simplified - no manager or company fields
	customerID := offer.CustomerID
	project := &domain.Project{
		Name:          projectName,
		CustomerID:    &customerID,
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

	// Link offer to project
	offer.ProjectID = &project.ID
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		s.logger.Warn("failed to link offer to project", zap.Error(err))
	}

	// Clone budget items from offer to project
	offerItems, err := s.budgetItemRepo.ListByParent(ctx, domain.BudgetParentOffer, id)
	if err == nil && len(offerItems) > 0 {
		for _, item := range offerItems {
			cloned := domain.BudgetItem{
				ParentType:     domain.BudgetParentProject,
				ParentID:       project.ID,
				Name:           item.Name,
				Description:    item.Description,
				ExpectedCost:   item.ExpectedCost,
				ExpectedMargin: item.ExpectedMargin,
				DisplayOrder:   item.DisplayOrder,
			}
			if err := s.budgetItemRepo.Create(ctx, &cloned); err != nil {
				s.logger.Warn("failed to clone budget item to project", zap.Error(err))
			}
		}
	}

	s.logger.Info("created project for offer",
		zap.String("offerID", offer.ID.String()),
		zap.String("projectID", project.ID.String()),
		zap.String("projectName", project.Name))

	// Log activity on the new project
	s.logActivityOnTarget(ctx, domain.ActivityTargetProject, project.ID, project.Name,
		"Prosjekt opprettet",
		fmt.Sprintf("Prosjektet '%s' ble opprettet fra tilbud '%s'", project.Name, offer.Title))

	projectDTO := mapper.ToProjectDTO(project)
	return &domain.AcceptOfferResponse{
		Offer:   result.Offer,
		Project: &projectDTO,
	}, nil
}

// AcceptOrder transitions a sent offer to order phase
// This marks the offer as accepted by the customer and ready for execution
func (s *OfferService) AcceptOrder(ctx context.Context, id uuid.UUID, req *domain.AcceptOrderRequest) (*domain.AcceptOrderResponse, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase - must be in sent phase
	if offer.Phase != domain.OfferPhaseSent {
		return nil, ErrOfferNotInSentPhase
	}

	// Check if already in order phase
	if offer.Phase == domain.OfferPhaseOrder {
		return nil, ErrOfferAlreadyInOrder
	}

	oldPhase := offer.Phase

	// Store original offer number before modification
	originalOfferNumber := offer.OfferNumber

	// Update offer number with "O" suffix to mark as order (only if not already suffixed)
	if originalOfferNumber != "" && !strings.HasSuffix(originalOfferNumber, "O") && !strings.HasSuffix(originalOfferNumber, "W") {
		offer.OfferNumber = originalOfferNumber + "O"
	}

	// Transition to order phase
	offer.Phase = domain.OfferPhaseOrder

	// Set updated by fields
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
		s.logger.Warn("failed to reload offer after accept order", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Tilbudet '%s' ble akseptert som ordre (fase: %s -> ordre)", offer.Title, oldPhase)
	if req.Notes != "" {
		activityBody = fmt.Sprintf("%s. Notater: %s", activityBody, req.Notes)
	}
	s.logActivity(ctx, offer.ID, offer.Title, "Ordre akseptert", activityBody)

	offerDTO := mapper.ToOfferDTO(offer)
	return &domain.AcceptOrderResponse{
		Offer: &offerDTO,
	}, nil
}

// UpdateOfferHealth updates the health status and optionally completion percentage for an offer in order phase
func (s *OfferService) UpdateOfferHealth(ctx context.Context, id uuid.UUID, req *domain.UpdateOfferHealthRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase - must be in order phase
	if offer.Phase != domain.OfferPhaseOrder {
		return nil, ErrOfferNotInOrderPhase
	}

	// Validate health enum
	if !req.Health.IsValid() {
		return nil, fmt.Errorf("invalid health status: %s", req.Health)
	}

	oldHealth := ""
	if offer.Health != nil {
		oldHealth = string(*offer.Health)
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"health": req.Health,
	}
	// Include completion percent if provided
	if req.CompletionPercent != nil {
		updates["completion_percent"] = *req.CompletionPercent
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update health: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudshelse oppdatert",
		fmt.Sprintf("Helse endret fra '%s' til '%s'", oldHealth, req.Health))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateOfferSpent updates the spent amount for an offer in order phase
func (s *OfferService) UpdateOfferSpent(ctx context.Context, id uuid.UUID, req *domain.UpdateOfferSpentRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase - must be in order phase
	if offer.Phase != domain.OfferPhaseOrder {
		return nil, ErrOfferNotInOrderPhase
	}

	oldSpent := offer.Spent

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"spent": req.Spent,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update spent: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudsforbruk oppdatert",
		fmt.Sprintf("Forbruk endret fra %.2f til %.2f", oldSpent, req.Spent))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateOfferInvoiced updates the invoiced amount for an offer in order phase
func (s *OfferService) UpdateOfferInvoiced(ctx context.Context, id uuid.UUID, req *domain.UpdateOfferInvoicedRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase - must be in order phase
	if offer.Phase != domain.OfferPhaseOrder {
		return nil, ErrOfferNotInOrderPhase
	}

	oldInvoiced := offer.Invoiced

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"invoiced": req.Invoiced,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update invoiced: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbud fakturering oppdatert",
		fmt.Sprintf("Fakturert endret fra %.2f til %.2f", oldInvoiced, req.Invoiced))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// CompleteOffer transitions an order offer to completed phase
func (s *OfferService) CompleteOffer(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase - must be in order phase
	if offer.Phase != domain.OfferPhaseOrder {
		return nil, ErrOfferNotInOrderPhase
	}

	oldPhase := offer.Phase
	offer.Phase = domain.OfferPhaseCompleted

	// Set updated by fields
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
		s.logger.Warn("failed to reload offer after complete", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, offer.ID, offer.Title, "Tilbud fullført",
		fmt.Sprintf("Tilbudet '%s' ble fullført (fase: %s -> fullført)", offer.Title, oldPhase))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// RejectOffer transitions an offer to lost phase with a reason
// Projects are now just folders, so no project lifecycle logic is applied
func (s *OfferService) RejectOffer(ctx context.Context, id uuid.UUID, req *domain.RejectOfferRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase - can reject from sent or order phase
	if offer.Phase != domain.OfferPhaseSent && offer.Phase != domain.OfferPhaseOrder {
		return nil, ErrOfferNotSent
	}

	oldPhase := offer.Phase
	offer.Phase = domain.OfferPhaseLost

	// Store reason in notes if provided
	if req.Reason != "" {
		if offer.Notes != "" {
			offer.Notes = fmt.Sprintf("%s\n\nLost reason: %s", offer.Notes, req.Reason)
		} else {
			offer.Notes = fmt.Sprintf("Lost reason: %s", req.Reason)
		}
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
		s.logger.Warn("failed to reload offer after reject", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Tilbudet '%s' ble avslått (fase: %s -> tapt)", offer.Title, oldPhase)
	if req.Reason != "" {
		activityBody = fmt.Sprintf("%s. Årsak: %s", activityBody, req.Reason)
	}
	s.logActivity(ctx, offer.ID, offer.Title, "Tilbud avslått", activityBody)

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// WinOffer wins an offer (transitions to order phase)
// Deprecated: Use AcceptOrder instead. This method is kept for backwards compatibility.
// Projects are now just folders and do not have lifecycle transitions tied to offers.
func (s *OfferService) WinOffer(ctx context.Context, id uuid.UUID, req *domain.WinOfferRequest) (*domain.WinOfferResponse, error) {
	// Call the new AcceptOrder implementation
	acceptReq := &domain.AcceptOrderRequest{
		Notes: req.Notes,
	}
	result, err := s.AcceptOrder(ctx, id, acceptReq)
	if err != nil {
		return nil, err
	}

	// Return in the old format for backwards compatibility
	return &domain.WinOfferResponse{
		Offer:         result.Offer,
		Project:       nil, // Projects are now just folders
		ExpiredOffers: nil, // No more sibling expiration
		ExpiredCount:  0,
	}, nil
}

// GetProjectOffers returns all offers linked to a project
func (s *OfferService) GetProjectOffers(ctx context.Context, projectID uuid.UUID) ([]domain.OfferDTO, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	offers, err := s.offerRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project offers: %w", err)
	}

	dtos := make([]domain.OfferDTO, len(offers))
	for i, offer := range offers {
		dtos[i] = mapper.ToOfferDTO(&offer)
	}

	return dtos, nil
}

// ExpireOffer transitions an offer to expired phase
// Projects are now just folders, so no project lifecycle logic is applied
func (s *OfferService) ExpireOffer(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Can expire from draft, in_progress, or sent phases
	// Cannot expire offers in order phase (already accepted) or closed phases
	if s.isClosedPhase(offer.Phase) || offer.Phase == domain.OfferPhaseOrder {
		return nil, ErrOfferAlreadyClosed
	}

	oldPhase := offer.Phase

	// Generate offer number if transitioning from draft (expired is non-draft)
	if s.isDraftPhase(oldPhase) {
		if err := s.generateOfferNumberIfNeeded(ctx, offer); err != nil {
			return nil, err
		}
	}

	offer.Phase = domain.OfferPhaseExpired

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
		s.logger.Warn("failed to reload offer after expire", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Tilbudet '%s' ble markert som utløpt (fase: %s -> utløpt)", offer.Title, oldPhase)
	s.logActivity(ctx, offer.ID, offer.Title, "Tilbud utløpt", activityBody)

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// ReopenOffer transitions a completed offer back to order phase
// This allows additional work to be done on a completed order.
func (s *OfferService) ReopenOffer(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Can only reopen completed offers
	if offer.Phase != domain.OfferPhaseCompleted {
		return nil, fmt.Errorf("can only reopen completed offers (current phase: %s)", offer.Phase)
	}

	oldPhase := offer.Phase
	offer.Phase = domain.OfferPhaseOrder

	// Set updated by fields
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
		s.logger.Warn("failed to reload offer after reopen", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, offer.ID, offer.Title, "Ordre gjenåpnet",
		fmt.Sprintf("Ordren '%s' ble gjenåpnet fra %s til order", offer.Title, oldPhase))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// RevertToSent transitions an order offer back to sent phase
// This allows re-negotiation of an accepted order.
// Note: This does NOT remove the O suffix from the offer number.
func (s *OfferService) RevertToSent(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Can only revert order offers
	if offer.Phase != domain.OfferPhaseOrder {
		return nil, fmt.Errorf("can only revert order offers to sent (current phase: %s)", offer.Phase)
	}

	oldPhase := offer.Phase
	offer.Phase = domain.OfferPhaseSent

	// Set updated by fields
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
		s.logger.Warn("failed to reload offer after revert", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, offer.ID, offer.Title, "Tilbud tilbakestilt til sendt",
		fmt.Sprintf("Tilbudet '%s' ble tilbakestilt fra %s til sendt", offer.Title, oldPhase))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// CloneOffer creates a copy of an offer with optional budget dimensions
func (s *OfferService) CloneOffer(ctx context.Context, id uuid.UUID, req *domain.CloneOfferRequest) (*domain.OfferDTO, error) {
	// Get source offer with all relations
	sourceOffer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get source offer: %w", err)
	}

	// Determine new title
	newTitle := req.NewTitle
	if newTitle == "" {
		newTitle = fmt.Sprintf("Copy of %s", sourceOffer.Title)
	}

	// Create new offer
	newOffer := &domain.Offer{
		Title:               newTitle,
		CustomerID:          sourceOffer.CustomerID,
		CustomerName:        sourceOffer.CustomerName,
		CompanyID:           sourceOffer.CompanyID,
		Phase:               domain.OfferPhaseDraft, // Always start as draft
		Probability:         sourceOffer.Probability,
		Value:               sourceOffer.Value,
		Status:              domain.OfferStatusActive,
		ResponsibleUserID:   sourceOffer.ResponsibleUserID,
		ResponsibleUserName: sourceOffer.ResponsibleUserName,
		Description:         sourceOffer.Description,
		Notes:               sourceOffer.Notes,
	}

	// Clone items
	if len(sourceOffer.Items) > 0 {
		newItems := make([]domain.OfferItem, len(sourceOffer.Items))
		for i, item := range sourceOffer.Items {
			newItems[i] = domain.OfferItem{
				Discipline:  item.Discipline,
				Cost:        item.Cost,
				Revenue:     item.Revenue,
				Margin:      item.Margin,
				Description: item.Description,
				Quantity:    item.Quantity,
				Unit:        item.Unit,
			}
		}
		newOffer.Items = newItems
	}

	// Use transaction for atomicity
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newOffer).Error; err != nil {
			return fmt.Errorf("failed to create cloned offer: %w", err)
		}

		// Clone budget items if requested (default behavior - nil or true means include)
		includeBudget := req.IncludeBudget == nil || *req.IncludeBudget
		if includeBudget && s.budgetItemRepo != nil {
			items, err := s.budgetItemRepo.ListByParent(ctx, domain.BudgetParentOffer, id)
			if err == nil && len(items) > 0 {
				for _, item := range items {
					cloned := domain.BudgetItem{
						ParentType:     domain.BudgetParentOffer,
						ParentID:       newOffer.ID,
						Name:           item.Name,
						ExpectedCost:   item.ExpectedCost,
						ExpectedMargin: item.ExpectedMargin,
						Quantity:       item.Quantity,
						PricePerItem:   item.PricePerItem,
						Description:    item.Description,
						DisplayOrder:   item.DisplayOrder,
					}
					if err := tx.Create(&cloned).Error; err != nil {
						s.logger.Warn("failed to clone budget item",
							zap.Error(err),
							zap.String("item_id", item.ID.String()))
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload with relations
	newOffer, err = s.offerRepo.GetByID(ctx, newOffer.ID)
	if err != nil {
		s.logger.Warn("failed to reload offer after clone", zap.Error(err))
	}

	// Log activity on source offer
	s.logActivity(ctx, id, sourceOffer.Title, "Tilbud klonet",
		fmt.Sprintf("Tilbudet '%s' ble klonet for å opprette '%s'", sourceOffer.Title, newOffer.Title))

	// Log activity on new offer
	s.logActivity(ctx, newOffer.ID, newOffer.Title, "Tilbud opprettet fra klone",
		fmt.Sprintf("Tilbudet '%s' ble opprettet som en klone av '%s'", newOffer.Title, sourceOffer.Title))

	dto := mapper.ToOfferDTO(newOffer)
	return &dto, nil
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
// Legacy Methods (for backwards compatibility)
// ============================================================================

// Advance updates the offer phase (legacy method, prefer specific lifecycle methods)
// When transitioning from draft to any non-draft phase, generates a unique offer number
func (s *OfferService) Advance(ctx context.Context, id uuid.UUID, req *domain.AdvanceOfferRequest) (*domain.OfferDTO, error) {
	resp, err := s.AdvanceWithProjectResponse(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return resp.Offer, nil
}

// AdvanceWithProjectResponse updates the offer phase and returns the offer and any auto-created project
// When transitioning from draft to in_progress:
// - If ProjectID is provided in request, validates and links to that project
// - If CreateProject is true and no ProjectID provided, auto-creates a project
// - Otherwise, offer remains without a project
func (s *OfferService) AdvanceWithProjectResponse(ctx context.Context, id uuid.UUID, req *domain.AdvanceOfferRequest) (*domain.OfferWithProjectResponse, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Block transitions to terminal phases or order - must use dedicated endpoints
	// AcceptOffer/AcceptOrder for order, RejectOffer for lost, ExpireOffer for expired
	if s.isClosedPhase(req.Phase) || req.Phase == domain.OfferPhaseOrder {
		return nil, ErrOfferCannotAdvanceToTerminalPhase
	}

	oldPhase := offer.Phase
	transitioningFromDraft := s.isDraftPhase(oldPhase) && !s.isDraftPhase(req.Phase)
	transitioningToInProgress := req.Phase == domain.OfferPhaseInProgress

	// Special validation for draft to in_progress transition
	if oldPhase == domain.OfferPhaseDraft && transitioningToInProgress {
		// Must have responsible user OR company with default responsible user
		hasResponsible := offer.ResponsibleUserID != ""
		hasCompany := offer.CompanyID != ""

		if !hasResponsible && !hasCompany {
			return nil, ErrOfferMissingResponsible
		}

		// If only company is set, try to infer responsible user from company default
		if !hasResponsible && hasCompany && s.companyService != nil {
			defaultResponsible := s.companyService.GetDefaultOfferResponsible(ctx, offer.CompanyID)
			if defaultResponsible != nil && *defaultResponsible != "" {
				offer.ResponsibleUserID = *defaultResponsible
				s.logger.Info("inferred responsible user from company default during phase transition",
					zap.String("offerID", id.String()),
					zap.String("companyID", string(offer.CompanyID)),
					zap.String("responsibleUserID", offer.ResponsibleUserID))
			} else {
				// Company doesn't have a default responsible user configured
				return nil, ErrOfferMissingResponsible
			}
		}
	}

	// Generate offer number when transitioning from draft to ANY non-draft phase
	if transitioningFromDraft {
		if err := s.generateOfferNumberIfNeeded(ctx, offer); err != nil {
			return nil, err
		}
	}

	// Track project creation result
	var projectLinkRes *projectLinkResult

	// Handle project linking/creation based on explicit flags when transitioning from draft
	if transitioningFromDraft && !s.isDraftPhase(req.Phase) {
		// Determine the project ID to use
		requestedProjectID := req.ProjectID
		if requestedProjectID == nil && offer.ProjectID != nil {
			// Offer already has a project, use that
			requestedProjectID = offer.ProjectID
		}

		if requestedProjectID != nil {
			// Validate the requested/existing project
			projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, requestedProjectID, false)
			if err != nil {
				return nil, err
			}
			offer.ProjectID = &projectLinkRes.ProjectID
		} else if req.CreateProject {
			// CreateProject=true and no project - auto-create one
			projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, nil, true)
			if err != nil {
				return nil, err
			}
			offer.ProjectID = &projectLinkRes.ProjectID
		}
		// Otherwise, offer proceeds without a project
	}

	offer.Phase = req.Phase

	// Clear sent-related dates when moving back to in_progress (e.g., from sent)
	// These will be set again when the offer is sent
	if req.Phase == domain.OfferPhaseInProgress && oldPhase == domain.OfferPhaseSent {
		offer.ExpirationDate = nil
		offer.SentDate = nil
	}

	// Validate offer number rules after the phase change
	if s.isDraftPhase(req.Phase) && offer.OfferNumber != "" {
		return nil, ErrDraftOfferCannotHaveNumber
	}
	if !s.isDraftPhase(req.Phase) && offer.OfferNumber == "" {
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

	// Reload offer
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after advance", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Tilbudet '%s' avansert fra %s til %s", offer.Title, oldPhase, offer.Phase)
	if transitioningFromDraft && offer.OfferNumber != "" {
		activityBody = fmt.Sprintf("Tilbudet '%s' avansert fra %s til %s med tilbudsnummer %s",
			offer.Title, oldPhase, offer.Phase, offer.OfferNumber)
	}
	if projectLinkRes != nil && projectLinkRes.ProjectCreated {
		activityBody = fmt.Sprintf("%s (auto-opprettet prosjekt '%s')", activityBody, projectLinkRes.Project.Name)
	}
	s.logActivity(ctx, offer.ID, offer.Title, "Tilbudsfase avansert", activityBody)

	offerDTO := mapper.ToOfferDTO(offer)

	response := &domain.OfferWithProjectResponse{
		Offer: &offerDTO,
	}

	// Include project in response if one was created or linked
	if projectLinkRes != nil {
		// Reload project to get latest state
		project, err := s.projectRepo.GetByID(ctx, projectLinkRes.ProjectID)
		if err == nil {
			projectDTO := mapper.ToProjectDTO(project)
			response.Project = &projectDTO
		}
		response.ProjectCreated = projectLinkRes.ProjectCreated
	}

	return response, nil
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
		offerCustomerMatches := project.CustomerID != nil && *project.CustomerID == offer.CustomerID
		if !offerCustomerMatches {
			s.logger.Info("linking offer to project with different customer",
				zap.String("offerID", offer.ID.String()),
				zap.String("offerCustomerID", offer.CustomerID.String()),
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
	customerID := offer.CustomerID
	project := &domain.Project{
		Name:          fmt.Sprintf("[AUTO] %s", offer.Title),
		CustomerID:    &customerID,
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
// Individual Property Update Methods
// ============================================================================

// UpdateProbability updates only the probability field of an offer
func (s *OfferService) UpdateProbability(ctx context.Context, id uuid.UUID, probability int) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	oldValue := offer.Probability

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"probability": probability,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update probability: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudssannsynlighet oppdatert",
		fmt.Sprintf("Sannsynlighet endret fra %d%% til %d%%", oldValue, probability))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateTitle updates only the title field of an offer
func (s *OfferService) UpdateTitle(ctx context.Context, id uuid.UUID, title string) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	oldTitle := offer.Title

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"title": title,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update title: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudstittel oppdatert",
		fmt.Sprintf("Tittel endret fra '%s' til '%s'", oldTitle, title))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateResponsible updates only the responsible user field of an offer.
// Responsible user can be edited regardless of offer phase.
func (s *OfferService) UpdateResponsible(ctx context.Context, id uuid.UUID, responsibleUserID string) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	oldResponsibleName := offer.ResponsibleUserName

	// Look up the user to get their display name
	var responsibleUserName string
	if responsibleUserID != "" {
		user, err := s.userRepo.GetByStringID(ctx, responsibleUserID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("failed to look up user: %w", err)
			}
			// User not found in database, use ID as fallback
			responsibleUserName = responsibleUserID
		} else {
			responsibleUserName = user.DisplayName
		}
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"responsible_user_id":   responsibleUserID,
		"responsible_user_name": responsibleUserName,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update responsible: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudsansvarlig oppdatert",
		fmt.Sprintf("Ansvarlig endret fra '%s' til '%s'", oldResponsibleName, responsibleUserName))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateCustomer updates only the customer field of an offer
func (s *OfferService) UpdateCustomer(ctx context.Context, id uuid.UUID, customerID uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to verify customer: %w", err)
	}

	oldCustomerName := offer.CustomerName

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"customer_id":   customerID,
		"customer_name": customer.Name,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// If the offer is linked to a project, sync the project's customer
	if offer.ProjectID != nil {
		if err := s.syncProjectCustomer(ctx, *offer.ProjectID); err != nil {
			s.logger.Warn("failed to sync project customer after offer customer change",
				zap.String("offerID", id.String()),
				zap.String("projectID", offer.ProjectID.String()),
				zap.Error(err))
		}
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudskunde oppdatert",
		fmt.Sprintf("Kunde endret fra '%s' til '%s'", oldCustomerName, customer.Name))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateValue updates only the value field of an offer
func (s *OfferService) UpdateValue(ctx context.Context, id uuid.UUID, value float64) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	oldValue := offer.Value

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"value": value,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update value: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudsverdi oppdatert",
		fmt.Sprintf("Verdi endret fra %.2f til %.2f", oldValue, value))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateCost updates only the cost field of an offer
func (s *OfferService) UpdateCost(ctx context.Context, id uuid.UUID, cost float64) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	oldCost := offer.Cost

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"cost": cost,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update cost: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudskostnad oppdatert",
		fmt.Sprintf("Kostnad endret fra %.2f til %.2f", oldCost, cost))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateDueDate updates only the due date field of an offer
func (s *OfferService) UpdateDueDate(ctx context.Context, id uuid.UUID, dueDate *time.Time) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"due_date": dueDate,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update due date: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	dueDateStr := "fjernet"
	if dueDate != nil {
		dueDateStr = dueDate.Format("2006-01-02")
	}
	s.logActivity(ctx, id, offer.Title, "Tilbudsfrist oppdatert",
		fmt.Sprintf("Frist satt til %s", dueDateStr))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateExpirationDate updates only the expiration date field of an offer
// If expirationDate is nil, defaults to 60 days after sent date
func (s *OfferService) UpdateExpirationDate(ctx context.Context, id uuid.UUID, expirationDate *time.Time) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Only sent offers should have expiration dates
	if offer.Phase != domain.OfferPhaseSent {
		return nil, fmt.Errorf("only sent offers can have expiration dates")
	}

	// Determine the expiration date to set
	var finalExpirationDate *time.Time
	if expirationDate != nil {
		// Validate expiration date is not before sent date
		if offer.SentDate != nil && expirationDate.Before(*offer.SentDate) {
			return nil, ErrExpirationDateBeforeSentDate
		}
		finalExpirationDate = expirationDate
	} else {
		// Default to 60 days after sent date
		if offer.SentDate != nil {
			expDate := offer.SentDate.AddDate(0, 0, 60)
			finalExpirationDate = &expDate
		}
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"expiration_date": finalExpirationDate,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update expiration date: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	expirationDateStr := "60 dager fra sendedato (standard)"
	if expirationDate != nil {
		expirationDateStr = expirationDate.Format("2006-01-02")
	}
	s.logActivity(ctx, id, offer.Title, "Tilbud utløpsdato oppdatert",
		fmt.Sprintf("Utløpsdato satt til %s", expirationDateStr))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateDescription updates only the description field of an offer.
// Description can be edited regardless of offer phase.
func (s *OfferService) UpdateDescription(ctx context.Context, id uuid.UUID, description string) (*domain.OfferDTO, error) {
	_, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"description": description,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update description: %w", err)
	}

	// Reload and return
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudsbeskrivelse oppdatert", "Beskrivelsen ble oppdatert")

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateNotes updates only the notes field of an offer
func (s *OfferService) UpdateNotes(ctx context.Context, id uuid.UUID, notes string) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"notes": notes,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update notes: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, offer.Title, "Tilbudsnotater oppdatert", "Notatene ble oppdatert")

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// LinkToProject links an offer to a project
func (s *OfferService) LinkToProject(ctx context.Context, offerID uuid.UUID, projectID uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	// Store old project ID if offer was previously linked (to sync its customer after)
	var oldProjectID *uuid.UUID
	if offer.ProjectID != nil && *offer.ProjectID != projectID {
		oldProjectID = offer.ProjectID
	}

	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	// Only allow linking offers to projects in tilbud (offer) phase
	if project.Phase != domain.ProjectPhaseTilbud {
		return nil, ErrProjectNotInOfferPhase
	}

	if err := s.offerRepo.LinkToProject(ctx, offerID, projectID); err != nil {
		return nil, fmt.Errorf("failed to link offer to project: %w", err)
	}

	// Sync old project's customer if the offer was moved from another project
	if oldProjectID != nil {
		if err := s.syncProjectCustomer(ctx, *oldProjectID); err != nil {
			s.logger.Warn("failed to sync old project customer after moving offer",
				zap.String("offerID", offerID.String()),
				zap.String("oldProjectID", oldProjectID.String()),
				zap.Error(err))
		}
	}

	// Sync new project's customer based on all offers in the project
	if err := s.syncProjectCustomer(ctx, projectID); err != nil {
		s.logger.Warn("failed to sync project customer after linking offer",
			zap.String("offerID", offerID.String()),
			zap.String("projectID", projectID.String()),
			zap.Error(err))
	}

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates := map[string]interface{}{
			"updated_by_id":   userCtx.UserID.String(),
			"updated_by_name": userCtx.DisplayName,
		}
		if err := s.offerRepo.UpdateFields(ctx, offerID, updates); err != nil {
			s.logger.Warn("failed to set updated by fields after link", zap.Error(err))
		}
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, offerID, offer.Title, "Tilbud koblet til prosjekt",
		fmt.Sprintf("Tilbud koblet til prosjekt '%s'", project.Name))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UnlinkFromProject removes the project link from an offer
// Note: Closed offers (won/lost/expired) cannot be unlinked as their lifecycle is complete
func (s *OfferService) UnlinkFromProject(ctx context.Context, offerID uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	if offer.ProjectID == nil {
		// Already unlinked, just return the offer
		dto := mapper.ToOfferDTO(offer)
		return &dto, nil
	}

	// Store project ID for economics recalculation
	oldProjectID := *offer.ProjectID

	if err := s.offerRepo.UnlinkFromProject(ctx, offerID); err != nil {
		return nil, fmt.Errorf("failed to unlink offer from project: %w", err)
	}

	// Sync project customer based on remaining offers in the project
	if err := s.syncProjectCustomer(ctx, oldProjectID); err != nil {
		s.logger.Warn("failed to sync project customer after unlinking offer",
			zap.String("offerID", offerID.String()),
			zap.String("projectID", oldProjectID.String()),
			zap.Error(err))
	}

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates := map[string]interface{}{
			"updated_by_id":   userCtx.UserID.String(),
			"updated_by_name": userCtx.DisplayName,
		}
		if err := s.offerRepo.UpdateFields(ctx, offerID, updates); err != nil {
			s.logger.Warn("failed to set updated by fields after unlink", zap.Error(err))
		}
	}

	// Log the old project ID for reference (but don't sync economics - projects are just folders now)
	s.logger.Info("offer unlinked from project",
		zap.String("offerID", offerID.String()),
		zap.String("oldProjectID", oldProjectID.String()))

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, offerID, offer.Title, "Tilbud frakoblet fra prosjekt", "Prosjektkoblingen ble fjernet")

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateCustomerHasWonProject updates the customer has won project flag on an offer
func (s *OfferService) UpdateCustomerHasWonProject(ctx context.Context, offerID uuid.UUID, customerHasWonProject bool) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	offer.CustomerHasWonProject = customerHasWonProject

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		offer.UpdatedByID = userCtx.UserID.String()
		offer.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Log activity
	var activityBody string
	if customerHasWonProject {
		activityBody = "Kunden merket som å ha vunnet sitt prosjekt"
	} else {
		activityBody = "Kunden merket som å ikke ha vunnet sitt prosjekt"
	}
	s.logActivity(ctx, offerID, offer.Title, "Kundens prosjektstatus oppdatert", activityBody)

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateOfferNumber updates the internal offer number with conflict checking
func (s *OfferService) UpdateOfferNumber(ctx context.Context, offerID uuid.UUID, offerNumber string) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Draft offers cannot have offer numbers
	if s.isDraftPhase(offer.Phase) {
		return nil, ErrDraftOfferCannotHaveNumber
	}

	// Non-draft offers cannot have empty offer numbers
	if offerNumber == "" {
		return nil, ErrNonDraftOfferMustHaveNumber
	}

	// Check if the new offer number already exists (excluding this offer)
	exists, err := s.offerRepo.OfferNumberExists(ctx, offerNumber, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check offer number: %w", err)
	}
	if exists {
		return nil, ErrOfferNumberConflict
	}

	oldNumber := offer.OfferNumber
	offer.OfferNumber = offerNumber

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		offer.UpdatedByID = userCtx.UserID.String()
		offer.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	s.logActivity(ctx, offerID, offer.Title, "Tilbudsnummer oppdatert",
		fmt.Sprintf("Endret fra '%s' til '%s'", oldNumber, offerNumber))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateExternalReference updates the external reference field on an offer
func (s *OfferService) UpdateExternalReference(ctx context.Context, offerID uuid.UUID, externalReference string) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Check if the new external reference already exists within the company (excluding this offer)
	if externalReference != "" {
		exists, err := s.offerRepo.ExternalReferenceExists(ctx, externalReference, offer.CompanyID, offerID)
		if err != nil {
			return nil, fmt.Errorf("failed to check external reference: %w", err)
		}
		if exists {
			return nil, ErrExternalReferenceConflict
		}
	}

	oldRef := offer.ExternalReference
	offer.ExternalReference = externalReference

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		offer.UpdatedByID = userCtx.UserID.String()
		offer.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	var activityBody string
	if externalReference == "" {
		activityBody = fmt.Sprintf("Fjernet ekstern referanse '%s'", oldRef)
	} else if oldRef == "" {
		activityBody = fmt.Sprintf("Satt ekstern referanse til '%s'", externalReference)
	} else {
		activityBody = fmt.Sprintf("Endret ekstern referanse fra '%s' til '%s'", oldRef, externalReference)
	}
	s.logActivity(ctx, offerID, offer.Title, "Ekstern referanse oppdatert", activityBody)

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// GetNextOfferNumber returns a preview of what the next offer number would be for a company
// WITHOUT consuming/incrementing the sequence. This is useful for UI display purposes.
func (s *OfferService) GetNextOfferNumber(ctx context.Context, companyID domain.CompanyID) (*domain.NextOfferNumberResponse, error) {
	// Validate company ID
	if !domain.IsValidCompanyID(string(companyID)) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCompanyID, companyID)
	}

	// Get the preview of the next offer number
	nextNumber, err := s.numberSeqService.PreviewNextOfferNumber(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to preview next offer number: %w", err)
	}

	return &domain.NextOfferNumberResponse{
		NextOfferNumber: nextNumber,
		CompanyID:       companyID,
		Year:            time.Now().Year(),
	}, nil
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
