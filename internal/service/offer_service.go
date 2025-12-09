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
	offerRepo      *repository.OfferRepository
	offerItemRepo  *repository.OfferItemRepository
	customerRepo   *repository.CustomerRepository
	projectRepo    *repository.ProjectRepository
	dimensionRepo  *repository.BudgetDimensionRepository
	fileRepo       *repository.FileRepository
	activityRepo   *repository.ActivityRepository
	companyService *CompanyService
	logger         *zap.Logger
	db             *gorm.DB
}

func NewOfferService(
	offerRepo *repository.OfferRepository,
	offerItemRepo *repository.OfferItemRepository,
	customerRepo *repository.CustomerRepository,
	projectRepo *repository.ProjectRepository,
	dimensionRepo *repository.BudgetDimensionRepository,
	fileRepo *repository.FileRepository,
	activityRepo *repository.ActivityRepository,
	companyService *CompanyService,
	logger *zap.Logger,
	db *gorm.DB,
) *OfferService {
	return &OfferService{
		offerRepo:      offerRepo,
		offerItemRepo:  offerItemRepo,
		customerRepo:   customerRepo,
		projectRepo:    projectRepo,
		dimensionRepo:  dimensionRepo,
		fileRepo:       fileRepo,
		activityRepo:   activityRepo,
		companyService: companyService,
		logger:         logger,
		db:             db,
	}
}

// Create creates a new offer with initial items
func (s *OfferService) Create(ctx context.Context, req *domain.CreateOfferRequest) (*domain.OfferDTO, error) {
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
		Items:               items,
	}

	if err := s.offerRepo.Create(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, offer.ID)
	if err != nil {
		s.logger.Warn("failed to reload offer after create", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, offer.ID, "Offer created",
		fmt.Sprintf("Offer '%s' was created for customer %s", offer.Title, offer.CustomerName))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
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

// GetByIDWithBudgetDimensions retrieves an offer with budget dimensions and summary
func (s *OfferService) GetByIDWithBudgetDimensions(ctx context.Context, id uuid.UUID) (*domain.OfferDetailDTO, error) {
	offer, dimensions, err := s.offerRepo.GetByIDWithBudgetDimensions(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer with dimensions: %w", err)
	}

	// Convert offer to DTO
	offerDTO := mapper.ToOfferDTO(offer)

	// Convert dimensions to DTOs
	dimensionDTOs := make([]domain.BudgetDimensionDTO, len(dimensions))
	for i, dim := range dimensions {
		dimensionDTOs[i] = mapper.ToBudgetDimensionDTO(&dim)
	}

	// Get budget summary
	summary, err := s.offerRepo.GetBudgetSummary(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get budget summary", zap.Error(err))
	}

	var summaryDTO *domain.BudgetSummaryDTO
	if summary != nil {
		summaryDTO = &domain.BudgetSummaryDTO{
			ParentType:           domain.BudgetParentOffer,
			ParentID:             id,
			DimensionCount:       summary.DimensionCount,
			TotalCost:            summary.TotalCost,
			TotalRevenue:         summary.TotalRevenue,
			OverallMarginPercent: summary.MarginPercent,
			TotalProfit:          summary.TotalMargin,
		}
	}

	// Get files count
	filesCount, _ := s.offerRepo.GetFilesCount(ctx, id)

	dto := &domain.OfferDetailDTO{
		OfferDTO:         offerDTO,
		BudgetDimensions: dimensionDTOs,
		BudgetSummary:    summaryDTO,
		FilesCount:       filesCount,
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

	offer.Title = req.Title
	offer.Phase = req.Phase
	offer.Probability = req.Probability
	offer.Status = req.Status
	offer.ResponsibleUserID = req.ResponsibleUserID
	offer.Description = req.Description
	offer.Notes = req.Notes
	offer.DueDate = req.DueDate

	// Recalculate value from items
	offer.Value = mapper.CalculateOfferValue(offer.Items)

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

// List returns a paginated list of offers
func (s *OfferService) List(ctx context.Context, page, pageSize int, customerID, projectID *uuid.UUID, phase *domain.OfferPhase) (*domain.PaginatedResponse, error) {
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

	offers, total, err := s.offerRepo.List(ctx, page, pageSize, customerID, projectID, phase)
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
	offer.Phase = domain.OfferPhaseSent

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

			// Clone budget dimensions from offer to project
			if s.dimensionRepo != nil {
				dimensions, err := s.dimensionRepo.GetByParent(ctx, domain.BudgetParentOffer, offer.ID)
				if err == nil && len(dimensions) > 0 {
					for _, dim := range dimensions {
						cloned := domain.BudgetDimension{
							ParentType:          domain.BudgetParentProject,
							ParentID:            project.ID,
							CategoryID:          dim.CategoryID,
							CustomName:          dim.CustomName,
							Cost:                dim.Cost,
							Revenue:             dim.Revenue,
							TargetMarginPercent: dim.TargetMarginPercent,
							MarginOverride:      dim.MarginOverride,
							Description:         dim.Description,
							Quantity:            dim.Quantity,
							Unit:                dim.Unit,
							DisplayOrder:        dim.DisplayOrder,
						}
						if err := tx.Create(&cloned).Error; err != nil {
							s.logger.Warn("failed to clone budget dimension",
								zap.Error(err),
								zap.String("dimension_id", dim.ID.String()))
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
	s.logActivity(ctx, offer.ID, "Offer rejected", activityBody)

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// ExpireOffer transitions an offer to expired phase
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
	offer.Phase = domain.OfferPhaseExpired

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Reload with relations
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after expire", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, offer.ID, "Offer expired",
		fmt.Sprintf("Offer '%s' was marked as expired (phase: %s -> expired)", offer.Title, oldPhase))

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

		// Clone budget dimensions if requested (default behavior - nil or true means include)
		includeDimensions := req.IncludeDimensions == nil || *req.IncludeDimensions
		if includeDimensions && s.dimensionRepo != nil {
			dimensions, err := s.dimensionRepo.GetByParent(ctx, domain.BudgetParentOffer, id)
			if err == nil && len(dimensions) > 0 {
				for _, dim := range dimensions {
					cloned := domain.BudgetDimension{
						ParentType:          domain.BudgetParentOffer,
						ParentID:            newOffer.ID,
						CategoryID:          dim.CategoryID,
						CustomName:          dim.CustomName,
						Cost:                dim.Cost,
						Revenue:             dim.Revenue,
						TargetMarginPercent: dim.TargetMarginPercent,
						MarginOverride:      dim.MarginOverride,
						Description:         dim.Description,
						Quantity:            dim.Quantity,
						Unit:                dim.Unit,
						DisplayOrder:        dim.DisplayOrder,
					}
					if err := tx.Create(&cloned).Error; err != nil {
						s.logger.Warn("failed to clone budget dimension",
							zap.Error(err),
							zap.String("dimension_id", dim.ID.String()))
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
		ParentType:           domain.BudgetParentOffer,
		ParentID:             id,
		DimensionCount:       summary.DimensionCount,
		TotalCost:            summary.TotalCost,
		TotalRevenue:         summary.TotalRevenue,
		OverallMarginPercent: summary.MarginPercent,
		TotalProfit:          summary.TotalMargin,
	}

	return dto, nil
}

// RecalculateTotals recalculates the offer value from budget dimensions
func (s *OfferService) RecalculateTotals(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Calculate totals from budget dimensions
	newValue, err := s.offerRepo.CalculateTotalsFromDimensions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate totals: %w", err)
	}

	// Reload offer
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload offer after recalculate", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, id, "Offer totals recalculated",
		fmt.Sprintf("Offer '%s' value updated to %.2f from budget dimensions", offer.Title, newValue))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// ============================================================================
// Legacy Methods (for backwards compatibility)
// ============================================================================

// Advance updates the offer phase (legacy method, prefer specific lifecycle methods)
func (s *OfferService) Advance(ctx context.Context, id uuid.UUID, req *domain.AdvanceOfferRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	oldPhase := offer.Phase

	// Validate phase transition from draft to in_progress
	if oldPhase == domain.OfferPhaseDraft && req.Phase == domain.OfferPhaseInProgress {
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

		// If only responsible user is set but no company, we can proceed (company is optional for in_progress)
		// The company validation is primarily to infer the responsible user if missing
	}

	offer.Phase = req.Phase

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Log activity
	activityBody := fmt.Sprintf("Offer '%s' advanced from %s to %s", offer.Title, oldPhase, offer.Phase)
	if oldPhase == domain.OfferPhaseDraft && req.Phase == domain.OfferPhaseInProgress {
		activityBody = fmt.Sprintf("Offer '%s' advanced to in progress (responsible: %s)", offer.Title, offer.ResponsibleUserID)
	}
	s.logActivity(ctx, offer.ID, "Offer phase advanced", activityBody)

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

// isClosedPhase returns true if the phase is a terminal state
func (s *OfferService) isClosedPhase(phase domain.OfferPhase) bool {
	return phase == domain.OfferPhaseWon ||
		phase == domain.OfferPhaseLost ||
		phase == domain.OfferPhaseExpired
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

	if err := s.offerRepo.UnlinkFromProject(ctx, offerID); err != nil {
		return nil, fmt.Errorf("failed to unlink offer from project: %w", err)
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
