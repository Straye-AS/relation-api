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
	offerRepo     *repository.OfferRepository
	offerItemRepo *repository.OfferItemRepository
	customerRepo  *repository.CustomerRepository
	projectRepo   *repository.ProjectRepository
	dimensionRepo *repository.BudgetDimensionRepository
	fileRepo      *repository.FileRepository
	activityRepo  *repository.ActivityRepository
	logger        *zap.Logger
	db            *gorm.DB
}

func NewOfferService(
	offerRepo *repository.OfferRepository,
	offerItemRepo *repository.OfferItemRepository,
	customerRepo *repository.CustomerRepository,
	projectRepo *repository.ProjectRepository,
	dimensionRepo *repository.BudgetDimensionRepository,
	fileRepo *repository.FileRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
	db *gorm.DB,
) *OfferService {
	return &OfferService{
		offerRepo:     offerRepo,
		offerItemRepo: offerItemRepo,
		customerRepo:  customerRepo,
		projectRepo:   projectRepo,
		dimensionRepo: dimensionRepo,
		fileRepo:      fileRepo,
		activityRepo:  activityRepo,
		logger:        logger,
		db:            db,
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

	// Calculate value from items
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

	offer := &domain.Offer{
		Title:               req.Title,
		CustomerID:          req.CustomerID,
		CustomerName:        customer.Name,
		CompanyID:           req.CompanyID,
		Phase:               req.Phase,
		Probability:         req.Probability,
		Value:               totalValue,
		Status:              req.Status,
		ResponsibleUserID:   req.ResponsibleUserID,
		ResponsibleUserName: "", // Populated by handler/external user lookup if needed
		Description:         req.Description,
		Notes:               req.Notes,
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

		// Clone budget dimensions if requested (default behavior)
		if req.IncludeDimensions && s.dimensionRepo != nil {
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
	offer.Phase = req.Phase

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Log activity
	s.logActivity(ctx, offer.ID, "Offer phase advanced",
		fmt.Sprintf("Offer '%s' advanced from %s to %s", offer.Title, oldPhase, offer.Phase))

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
