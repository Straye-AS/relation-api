package service

import (
	"context"
	"errors"
	"fmt"
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
func (s *OfferService) CreateWithProjectResponse(ctx context.Context, req *domain.CreateOfferRequest) (*domain.OfferWithProjectResponse, error) {
	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to verify customer: %w", err)
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
		phase = domain.OfferPhaseDraft
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
		CustomerID:          req.CustomerID,
		CustomerName:        customer.Name,
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

	// Generate offer number only for non-draft offers
	// Draft offers should NOT have offer numbers - they get one when transitioning out of draft
	if !s.isDraftPhase(phase) {
		if err := s.generateOfferNumberIfNeeded(ctx, offer); err != nil {
			return nil, err
		}
	}

	// Track project creation result
	var projectLinkRes *projectLinkResult

	// Handle project linking/creation for non-draft offers
	if !s.isDraftPhase(phase) {
		// For non-draft offers, ensure project exists (either provided or auto-created)
		// But first create the offer so we have its ID for linking
		if err := s.offerRepo.Create(ctx, offer); err != nil {
			return nil, fmt.Errorf("failed to create offer: %w", err)
		}

		// Now handle project linking
		if req.ProjectID != nil {
			// User provided a project ID - validate and link
			projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, req.ProjectID)
			if err != nil {
				// Rollback offer creation
				_ = s.offerRepo.Delete(ctx, offer.ID)
				return nil, err
			}
		} else {
			// Auto-create project
			projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, nil)
			if err != nil {
				// Rollback offer creation
				_ = s.offerRepo.Delete(ctx, offer.ID)
				return nil, err
			}
		}

		// Link offer to project
		offer.ProjectID = &projectLinkRes.ProjectID
		if err := s.offerRepo.LinkToProject(ctx, offer.ID, projectLinkRes.ProjectID); err != nil {
			return nil, fmt.Errorf("failed to link offer to project: %w", err)
		}

		// Update project economics (CalculatedOfferValue)
		if projectLinkRes.Project != nil && projectLinkRes.Project.Phase == domain.ProjectPhaseTilbud {
			if err := s.projectRepo.RecalculateBestOfferEconomics(ctx, projectLinkRes.ProjectID); err != nil {
				s.logger.Warn("failed to recalculate project economics", zap.Error(err))
			}
		}
	} else {
		// Draft offer - optionally link to provided project (but don't auto-create)
		if req.ProjectID != nil {
			// Validate the project if provided
			projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, req.ProjectID)
			if err != nil {
				return nil, err
			}
			offer.ProjectID = &projectLinkRes.ProjectID
		}

		if err := s.offerRepo.Create(ctx, offer); err != nil {
			return nil, fmt.Errorf("failed to create offer: %w", err)
		}
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, offer.ID)
	if err != nil {
		s.logger.Warn("failed to reload offer after create", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Offer '%s' was created for customer %s", offer.Title, offer.CustomerName)
	if projectLinkRes != nil && projectLinkRes.ProjectCreated {
		activityBody = fmt.Sprintf("Offer '%s' was created for customer %s with auto-created project '%s'",
			offer.Title, offer.CustomerName, projectLinkRes.Project.Name)
	}
	s.logActivity(ctx, offer.ID, "Offer created", activityBody)

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

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after update", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, offer.ID, "Offer updated",
		fmt.Sprintf("Offer '%s' was updated", offer.Title))

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

	if err := s.offerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete offer: %w", err)
	}

	// Log activity (on customer since offer is deleted)
	s.logActivityOnTarget(ctx, domain.ActivityTargetCustomer, offer.CustomerID,
		"Offer deleted", fmt.Sprintf("Offer '%s' was deleted", offer.Title))

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
	s.logActivity(ctx, offer.ID, "Offer sent",
		fmt.Sprintf("Offer '%s' was sent to customer (phase: %s -> sent)", offer.Title, oldPhase))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// AcceptOffer transitions an offer to won phase, optionally creating a project
func (s *OfferService) AcceptOffer(ctx context.Context, id uuid.UUID, req *domain.AcceptOfferRequest) (*domain.AcceptOfferResponse, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase - must be in sent phase
	if offer.Phase != domain.OfferPhaseSent {
		return nil, ErrOfferNotSent
	}

	var project *domain.Project
	var projectDTO *domain.ProjectDTO

	// Use transaction for atomicity
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update offer phase to won
		offer.Phase = domain.OfferPhaseWon
		if err := tx.Save(offer).Error; err != nil {
			return fmt.Errorf("failed to update offer: %w", err)
		}

		// Create project if requested
		if req.CreateProject {
			// Get user context for manager
			userCtx, ok := auth.FromContext(ctx)
			if !ok {
				return ErrUnauthorized
			}

			// Determine project name
			projectName := req.ProjectName
			if projectName == "" {
				projectName = offer.Title
			}

			// Determine manager
			managerID := req.ManagerID
			if managerID == "" {
				managerID = userCtx.UserID.String()
			}

			project = &domain.Project{
				Name:         projectName,
				CustomerID:   offer.CustomerID,
				CustomerName: offer.CustomerName,
				CompanyID:    offer.CompanyID,
				Status:       domain.ProjectStatusPlanning,
				StartDate:    time.Now(),
				Budget:       offer.Value,
				ManagerID:    managerID,
				OfferID:      &offer.ID,
				Description:  offer.Description,
			}

			if err := tx.Create(project).Error; err != nil {
				return fmt.Errorf("%w: %v", ErrProjectCreationFailed, err)
			}

			// Clone budget items from offer to project
			if s.budgetItemRepo != nil {
				items, err := s.budgetItemRepo.ListByParent(ctx, domain.BudgetParentOffer, offer.ID)
				if err == nil && len(items) > 0 {
					for _, item := range items {
						cloned := domain.BudgetItem{
							ParentType:     domain.BudgetParentProject,
							ParentID:       project.ID,
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
					// Mark project as having detailed budget
					project.HasDetailedBudget = true
					if err := tx.Save(project).Error; err != nil {
						s.logger.Warn("failed to update project HasDetailedBudget flag", zap.Error(err))
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload offer
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after accept", zap.Error(err))
	}
	offerDTO := mapper.ToOfferDTO(offer)

	// Convert project to DTO if created
	if project != nil {
		dto := mapper.ToProjectDTO(project)
		projectDTO = &dto
	}

	// Log activity
	activityBody := fmt.Sprintf("Offer '%s' was accepted (won)", offer.Title)
	if project != nil {
		activityBody = fmt.Sprintf("Offer '%s' was accepted and project '%s' was created", offer.Title, project.Name)
	}
	s.logActivity(ctx, offer.ID, "Offer accepted", activityBody)

	return &domain.AcceptOfferResponse{
		Offer:   &offerDTO,
		Project: projectDTO,
	}, nil
}

// RejectOffer transitions an offer to lost phase with a reason
// If the offer is linked to a project in tilbud phase:
// - If other active offers remain: recalculates project economics
// - If no other active offers: cancels the project
func (s *OfferService) RejectOffer(ctx context.Context, id uuid.UUID, req *domain.RejectOfferRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate phase - must be in sent phase
	if offer.Phase != domain.OfferPhaseSent {
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

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Handle project lifecycle if offer is linked to a project
	var projectCancelled bool
	if offer.ProjectID != nil {
		project, err := s.projectRepo.GetByID(ctx, *offer.ProjectID)
		if err == nil && project.Phase == domain.ProjectPhaseTilbud {
			// Check if there are other active offers
			activeCount, err := s.offerRepo.CountActiveOffersByProject(ctx, *offer.ProjectID)
			if err != nil {
				s.logger.Warn("failed to count active offers", zap.Error(err))
			} else if activeCount == 0 {
				// No more active offers - cancel the project
				if err := s.projectRepo.CancelProject(ctx, *offer.ProjectID); err != nil {
					s.logger.Warn("failed to cancel project", zap.Error(err))
				} else {
					projectCancelled = true
					s.logActivityOnTarget(ctx, domain.ActivityTargetProject, *offer.ProjectID,
						"Project cancelled",
						fmt.Sprintf("Project cancelled because all offers were lost (last: '%s')", offer.Title))
				}
			} else {
				// Other active offers remain - recalculate economics
				if err := s.projectRepo.RecalculateBestOfferEconomics(ctx, *offer.ProjectID); err != nil {
					s.logger.Warn("failed to recalculate project economics", zap.Error(err))
				}
			}
		}
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after reject", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Offer '%s' was rejected (phase: %s -> lost)", offer.Title, oldPhase)
	if req.Reason != "" {
		activityBody = fmt.Sprintf("%s. Reason: %s", activityBody, req.Reason)
	}
	if projectCancelled {
		activityBody = fmt.Sprintf("%s. Project was cancelled (no remaining active offers).", activityBody)
	}
	s.logActivity(ctx, offer.ID, "Offer rejected", activityBody)

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// WinOffer wins an offer within a project context (offer folder model)
// This transitions the offer to won, the project to active phase,
// and expires all sibling offers in the same project
func (s *OfferService) WinOffer(ctx context.Context, id uuid.UUID, req *domain.WinOfferRequest) (*domain.WinOfferResponse, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate offer has a project linked
	if offer.ProjectID == nil {
		return nil, ErrOfferNotInProject
	}

	// Validate offer is not already in a terminal phase
	if offer.Phase == domain.OfferPhaseWon {
		return nil, ErrOfferAlreadyWon
	}
	if s.isClosedPhase(offer.Phase) {
		return nil, ErrOfferAlreadyClosed
	}

	// Get the project and verify it's in tilbud phase
	project, err := s.projectRepo.GetByID(ctx, *offer.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.Phase != domain.ProjectPhaseTilbud {
		return nil, ErrProjectNotInTilbudPhase
	}

	// Generate offer number if transitioning from draft
	if s.isDraftPhase(offer.Phase) {
		if err := s.generateOfferNumberIfNeeded(ctx, offer); err != nil {
			return nil, err
		}
	}

	// Store the original offer number for the project
	originalOfferNumber := offer.OfferNumber
	wonAt := time.Now()

	// Use transaction for atomicity
	var expiredOfferIDs []uuid.UUID
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Mark the winning offer as won
		offer.Phase = domain.OfferPhaseWon
		if err := tx.Save(offer).Error; err != nil {
			return fmt.Errorf("failed to update winning offer: %w", err)
		}

		// 2. Update the winning offer's number with "_P" suffix
		if originalOfferNumber != "" {
			newOfferNumber := originalOfferNumber + "_P"
			if err := tx.Model(&domain.Offer{}).Where("id = ?", id).Update("offer_number", newOfferNumber).Error; err != nil {
				s.logger.Warn("failed to update offer number with suffix", zap.Error(err))
			}
		}

		// 3. Expire sibling offers
		expiredIDs, err := s.offerRepo.ExpireSiblingOffers(ctx, *offer.ProjectID, id)
		if err != nil {
			return fmt.Errorf("failed to expire sibling offers: %w", err)
		}
		expiredOfferIDs = expiredIDs

		// 4. Update the project to active phase with winning offer details
		if err := s.projectRepo.SetWinningOffer(ctx, project.ID, id, originalOfferNumber, offer.Value, wonAt); err != nil {
			return fmt.Errorf("failed to update project with winning offer: %w", err)
		}

		// 5. Clone budget items from offer to project if they exist
		if s.budgetItemRepo != nil {
			items, err := s.budgetItemRepo.ListByParent(ctx, domain.BudgetParentOffer, id)
			if err == nil && len(items) > 0 {
				for _, item := range items {
					cloned := domain.BudgetItem{
						ParentType:     domain.BudgetParentProject,
						ParentID:       project.ID,
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
				// Mark project as having detailed budget
				if err := tx.Model(&domain.Project{}).Where("id = ?", project.ID).
					Update("has_detailed_budget", true).Error; err != nil {
					s.logger.Warn("failed to update project HasDetailedBudget flag", zap.Error(err))
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload offer and project
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after win", zap.Error(err))
	}

	project, err = s.projectRepo.GetByID(ctx, *offer.ProjectID)
	if err != nil {
		s.logger.Warn("failed to reload project after win", zap.Error(err))
	}

	// Get expired offers for response
	var expiredOfferDTOs []domain.OfferDTO
	if len(expiredOfferIDs) > 0 {
		expiredOffers, err := s.offerRepo.GetExpiredSiblingOffers(ctx, *offer.ProjectID, id)
		if err == nil {
			expiredOfferDTOs = make([]domain.OfferDTO, len(expiredOffers))
			for i, o := range expiredOffers {
				expiredOfferDTOs[i] = mapper.ToOfferDTO(&o)
			}
		}
	}

	// Log activities
	activityBody := fmt.Sprintf("Offer '%s' was won, transitioning project '%s' to active phase", offer.Title, project.Name)
	if req.Notes != "" {
		activityBody = fmt.Sprintf("%s. Notes: %s", activityBody, req.Notes)
	}
	s.logActivity(ctx, offer.ID, "Offer won", activityBody)

	if len(expiredOfferIDs) > 0 {
		s.logActivityOnTarget(ctx, domain.ActivityTargetProject, project.ID,
			"Project activated",
			fmt.Sprintf("Project activated with winning offer '%s'. %d sibling offer(s) were expired.",
				offer.Title, len(expiredOfferIDs)))
	} else {
		s.logActivityOnTarget(ctx, domain.ActivityTargetProject, project.ID,
			"Project activated",
			fmt.Sprintf("Project activated with winning offer '%s'.", offer.Title))
	}

	offerDTO := mapper.ToOfferDTO(offer)
	projectDTO := mapper.ToProjectDTO(project)

	return &domain.WinOfferResponse{
		Offer:         &offerDTO,
		Project:       &projectDTO,
		ExpiredOffers: expiredOfferDTOs,
		ExpiredCount:  len(expiredOfferIDs),
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
// If the offer is linked to a project in tilbud phase:
// - If other active offers remain: recalculates project economics
// - If no other active offers: cancels the project
func (s *OfferService) ExpireOffer(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Can expire from draft, in_progress, or sent phases
	if s.isClosedPhase(offer.Phase) {
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

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Handle project lifecycle if offer is linked to a project
	var projectCancelled bool
	if offer.ProjectID != nil {
		project, err := s.projectRepo.GetByID(ctx, *offer.ProjectID)
		if err == nil && project.Phase == domain.ProjectPhaseTilbud {
			// Check if there are other active offers
			activeCount, err := s.offerRepo.CountActiveOffersByProject(ctx, *offer.ProjectID)
			if err != nil {
				s.logger.Warn("failed to count active offers", zap.Error(err))
			} else if activeCount == 0 {
				// No more active offers - cancel the project
				if err := s.projectRepo.CancelProject(ctx, *offer.ProjectID); err != nil {
					s.logger.Warn("failed to cancel project", zap.Error(err))
				} else {
					projectCancelled = true
					s.logActivityOnTarget(ctx, domain.ActivityTargetProject, *offer.ProjectID,
						"Project cancelled",
						fmt.Sprintf("Project cancelled because all offers expired (last: '%s')", offer.Title))
				}
			} else {
				// Other active offers remain - recalculate economics
				if err := s.projectRepo.RecalculateBestOfferEconomics(ctx, *offer.ProjectID); err != nil {
					s.logger.Warn("failed to recalculate project economics", zap.Error(err))
				}
			}
		}
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after expire", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Offer '%s' was marked as expired (phase: %s -> expired)", offer.Title, oldPhase)
	if projectCancelled {
		activityBody = fmt.Sprintf("%s. Project was cancelled (no remaining active offers).", activityBody)
	}
	s.logActivity(ctx, offer.ID, "Offer expired", activityBody)

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
	s.logActivity(ctx, id, "Offer cloned",
		fmt.Sprintf("Offer '%s' was cloned to create '%s'", sourceOffer.Title, newOffer.Title))

	// Log activity on new offer
	s.logActivity(ctx, newOffer.ID, "Offer created from clone",
		fmt.Sprintf("Offer '%s' was created as a clone of '%s'", newOffer.Title, sourceOffer.Title))

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
	s.logActivity(ctx, id, "Offer totals recalculated",
		fmt.Sprintf("Offer '%s' value updated to %.2f from budget items", offer.Title, newValue))

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
// - If no ProjectID and offer has no project, auto-creates one
func (s *OfferService) AdvanceWithProjectResponse(ctx context.Context, id uuid.UUID, req *domain.AdvanceOfferRequest) (*domain.OfferWithProjectResponse, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
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

	// Handle project auto-creation when transitioning to in_progress (or any non-draft phase)
	// and the offer doesn't already have a project
	if transitioningFromDraft && !s.isDraftPhase(req.Phase) {
		// Determine the project ID to use
		requestedProjectID := req.ProjectID
		if requestedProjectID == nil && offer.ProjectID != nil {
			// Offer already has a project, use that
			requestedProjectID = offer.ProjectID
		}

		if requestedProjectID != nil {
			// Validate the requested/existing project
			projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, requestedProjectID)
			if err != nil {
				return nil, err
			}
		} else {
			// No project - auto-create one
			projectLinkRes, err = s.ensureProjectForOffer(ctx, offer, nil)
			if err != nil {
				return nil, err
			}
		}

		// Link offer to project
		offer.ProjectID = &projectLinkRes.ProjectID
	}

	offer.Phase = req.Phase

	// Validate offer number rules after the phase change
	if s.isDraftPhase(req.Phase) && offer.OfferNumber != "" {
		return nil, ErrDraftOfferCannotHaveNumber
	}
	if !s.isDraftPhase(req.Phase) && offer.OfferNumber == "" {
		return nil, ErrNonDraftOfferMustHaveNumber
	}

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Recalculate project economics if linked to a tilbud phase project
	if projectLinkRes != nil && projectLinkRes.Project != nil && projectLinkRes.Project.Phase == domain.ProjectPhaseTilbud {
		if err := s.projectRepo.RecalculateBestOfferEconomics(ctx, projectLinkRes.ProjectID); err != nil {
			s.logger.Warn("failed to recalculate project economics", zap.Error(err))
		}
	}

	// Reload offer
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after advance", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Offer '%s' advanced from %s to %s", offer.Title, oldPhase, offer.Phase)
	if transitioningFromDraft && offer.OfferNumber != "" {
		activityBody = fmt.Sprintf("Offer '%s' advanced from %s to %s with offer number %s",
			offer.Title, oldPhase, offer.Phase, offer.OfferNumber)
	}
	if projectLinkRes != nil && projectLinkRes.ProjectCreated {
		activityBody = fmt.Sprintf("%s (auto-created project '%s')", activityBody, projectLinkRes.Project.Name)
	}
	s.logActivity(ctx, offer.ID, "Offer phase advanced", activityBody)

	offerDTO := mapper.ToOfferDTO(offer)

	response := &domain.OfferWithProjectResponse{
		Offer: &offerDTO,
	}

	// Include project in response if one was created
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

// ensureProjectForOffer ensures an offer has a valid project link when transitioning to a non-draft phase.
// If projectID is provided, validates it exists and matches customer/company.
// If projectID is nil, creates a new project automatically.
// Returns the project ID to link and whether a new project was created.
func (s *OfferService) ensureProjectForOffer(ctx context.Context, offer *domain.Offer, requestedProjectID *uuid.UUID) (*projectLinkResult, error) {
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

		// Validate customer match
		if project.CustomerID != offer.CustomerID {
			return nil, ErrProjectCustomerMismatch
		}

		// Validate company match
		if project.CompanyID != offer.CompanyID {
			return nil, ErrProjectCompanyMismatch
		}

		return &projectLinkResult{
			ProjectID:      project.ID,
			Project:        project,
			ProjectCreated: false,
		}, nil
	}

	// No project ID provided - create one automatically
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Determine manager: use offer's responsible user if set, otherwise fall back to company default or current user
	managerID := offer.ResponsibleUserID
	if managerID == "" {
		// Try to get company default
		if s.companyService != nil {
			defaultManager := s.companyService.GetDefaultProjectResponsible(ctx, offer.CompanyID)
			if defaultManager != nil && *defaultManager != "" {
				managerID = *defaultManager
			}
		}
	}
	if managerID == "" {
		// Fall back to current user
		managerID = userCtx.UserID.String()
	}

	// Auto-create project
	project := &domain.Project{
		Name:         fmt.Sprintf("[AUTO] %s", offer.Title),
		CustomerID:   offer.CustomerID,
		CustomerName: offer.CustomerName,
		CompanyID:    offer.CompanyID,
		Status:       domain.ProjectStatusPlanning,
		Phase:        domain.ProjectPhaseTilbud,
		StartDate:    time.Now(),
		Budget:       offer.Value, // Initial budget from offer value
		ManagerID:    managerID,
		Description:  offer.Description,
		// CalculatedOfferValue will be set when economics are synced
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProjectCreationFailed, err)
	}

	s.logger.Info("auto-created project for offer",
		zap.String("offerID", offer.ID.String()),
		zap.String("projectID", project.ID.String()),
		zap.String("projectName", project.Name))

	// Log activity on the new project
	s.logActivityOnTarget(ctx, domain.ActivityTargetProject, project.ID,
		"Project auto-created",
		fmt.Sprintf("Project '%s' was automatically created for offer '%s'", project.Name, offer.Title))

	return &projectLinkResult{
		ProjectID:      project.ID,
		Project:        project,
		ProjectCreated: true,
	}, nil
}

// isClosedPhase returns true if the phase is a terminal state
func (s *OfferService) isClosedPhase(phase domain.OfferPhase) bool {
	return phase == domain.OfferPhaseWon ||
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
func (s *OfferService) logActivity(ctx context.Context, offerID uuid.UUID, title, body string) {
	s.logActivityOnTarget(ctx, domain.ActivityTargetOffer, offerID, title, body)
}

// logActivityOnTarget creates an activity log entry for any target
func (s *OfferService) logActivityOnTarget(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, title, body string) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		s.logger.Warn("no user context for activity logging")
		return
	}

	activity := &domain.Activity{
		TargetType:  targetType,
		TargetID:    targetID,
		Title:       title,
		Body:        body,
		CreatorName: userCtx.DisplayName,
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
	if err := s.offerRepo.UpdateField(ctx, id, "probability", probability); err != nil {
		return nil, fmt.Errorf("failed to update probability: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, "Offer probability updated",
		fmt.Sprintf("Probability changed from %d%% to %d%%", oldValue, probability))

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
	if err := s.offerRepo.UpdateField(ctx, id, "title", title); err != nil {
		return nil, fmt.Errorf("failed to update title: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, "Offer title updated",
		fmt.Sprintf("Title changed from '%s' to '%s'", oldTitle, title))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateResponsible updates only the responsible user field of an offer
func (s *OfferService) UpdateResponsible(ctx context.Context, id uuid.UUID, responsibleUserID string) (*domain.OfferDTO, error) {
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

	oldResponsible := offer.ResponsibleUserID
	if err := s.offerRepo.UpdateField(ctx, id, "responsible_user_id", responsibleUserID); err != nil {
		return nil, fmt.Errorf("failed to update responsible: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, "Offer responsible updated",
		fmt.Sprintf("Responsible changed from '%s' to '%s'", oldResponsible, responsibleUserID))

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
	updates := map[string]interface{}{
		"customer_id":   customerID,
		"customer_name": customer.Name,
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, "Offer customer updated",
		fmt.Sprintf("Customer changed from '%s' to '%s'", oldCustomerName, customer.Name))

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
	if err := s.offerRepo.UpdateField(ctx, id, "value", value); err != nil {
		return nil, fmt.Errorf("failed to update value: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, "Offer value updated",
		fmt.Sprintf("Value changed from %.2f to %.2f", oldValue, value))

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

	if err := s.offerRepo.UpdateField(ctx, id, "due_date", dueDate); err != nil {
		return nil, fmt.Errorf("failed to update due date: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	dueDateStr := "cleared"
	if dueDate != nil {
		dueDateStr = dueDate.Format("2006-01-02")
	}
	s.logActivity(ctx, id, "Offer due date updated",
		fmt.Sprintf("Due date set to %s", dueDateStr))

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

	if err := s.offerRepo.UpdateField(ctx, id, "expiration_date", finalExpirationDate); err != nil {
		return nil, fmt.Errorf("failed to update expiration date: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	expirationDateStr := "60 days from sent date (default)"
	if expirationDate != nil {
		expirationDateStr = expirationDate.Format("2006-01-02")
	}
	s.logActivity(ctx, id, "Offer expiration date updated",
		fmt.Sprintf("Expiration date set to %s", expirationDateStr))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateDescription updates only the description field of an offer
func (s *OfferService) UpdateDescription(ctx context.Context, id uuid.UUID, description string) (*domain.OfferDTO, error) {
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

	if err := s.offerRepo.UpdateField(ctx, id, "description", description); err != nil {
		return nil, fmt.Errorf("failed to update description: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, id, "Offer description updated", "Description was updated")

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

	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	if err := s.offerRepo.LinkToProject(ctx, offerID, projectID); err != nil {
		return nil, fmt.Errorf("failed to link offer to project: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, offerID, "Offer linked to project",
		fmt.Sprintf("Offer linked to project '%s'", project.Name))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UnlinkFromProject removes the project link from an offer
// Note: Non-draft offers cannot be unlinked from their project (per offer-project lifecycle spec)
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

	// Non-draft offers cannot be unlinked from their project
	if !s.isDraftPhase(offer.Phase) {
		return nil, ErrCannotUnlinkNonDraftOffer
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

	// Recalculate project economics after unlinking (in case draft was contributing)
	if err := s.projectRepo.RecalculateBestOfferEconomics(ctx, oldProjectID); err != nil {
		s.logger.Warn("failed to recalculate project economics after unlink", zap.Error(err))
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	s.logActivity(ctx, offerID, "Offer unlinked from project", "Project link was removed")

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

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Log activity
	var activityBody string
	if customerHasWonProject {
		activityBody = "Customer marked as having won their project"
	} else {
		activityBody = "Customer marked as not having won their project"
	}
	s.logActivity(ctx, offerID, "Customer project status updated", activityBody)

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

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	s.logActivity(ctx, offerID, "Offer number updated",
		fmt.Sprintf("Changed from '%s' to '%s'", oldNumber, offerNumber))

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

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	var activityBody string
	if externalReference == "" {
		activityBody = fmt.Sprintf("Removed external reference '%s'", oldRef)
	} else if oldRef == "" {
		activityBody = fmt.Sprintf("Set external reference to '%s'", externalReference)
	} else {
		activityBody = fmt.Sprintf("Changed external reference from '%s' to '%s'", oldRef, externalReference)
	}
	s.logActivity(ctx, offerID, "External reference updated", activityBody)

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
