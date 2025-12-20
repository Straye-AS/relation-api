package service

// This file contains offer lifecycle methods extracted from offer_service.go
// for better code organization. These methods handle:
// - Phase transitions (SendOffer, AcceptOffer, AcceptOrder, etc.)
// - Offer completion and rejection
// - Cloning offers
// - Legacy advance methods

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
	"go.uber.org/zap"
	"gorm.io/gorm"
)


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
	project := &domain.Project{
		Name:          projectName,
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

	// Sync project phase if offer is linked to a project
	if offer.ProjectID != nil {
		if err := s.syncProjectPhase(ctx, *offer.ProjectID); err != nil {
			s.logger.Warn("failed to sync project phase after accept order",
				zap.String("offerID", id.String()),
				zap.String("projectID", offer.ProjectID.String()),
				zap.Error(err))
		}
	}

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

// UpdateOfferSpent is deprecated - spent field is now read-only and managed by data warehouse sync.
// This method now returns an error indicating the field cannot be manually edited.
func (s *OfferService) UpdateOfferSpent(ctx context.Context, id uuid.UUID, req *domain.UpdateOfferSpentRequest) (*domain.OfferDTO, error) {
	return nil, ErrOfferFinancialFieldReadOnly
}

// UpdateOfferInvoiced is deprecated - invoiced field is now read-only and managed by data warehouse sync.
// This method now returns an error indicating the field cannot be manually edited.
func (s *OfferService) UpdateOfferInvoiced(ctx context.Context, id uuid.UUID, req *domain.UpdateOfferInvoicedRequest) (*domain.OfferDTO, error) {
	return nil, ErrOfferFinancialFieldReadOnly
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
		fmt.Sprintf("Tilbudet '%s' ble fullført (fase: %s -> completed)", offer.Title, oldPhase))

	// Sync project phase if offer is linked to a project
	if offer.ProjectID != nil {
		if err := s.syncProjectPhase(ctx, *offer.ProjectID); err != nil {
			s.logger.Warn("failed to sync project phase after complete offer",
				zap.String("offerID", id.String()),
				zap.String("projectID", offer.ProjectID.String()),
				zap.Error(err))
		}
	}

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

	// Sync project phase if offer is linked to a project
	if offer.ProjectID != nil {
		if err := s.syncProjectPhase(ctx, *offer.ProjectID); err != nil {
			s.logger.Warn("failed to sync project phase after reopen offer",
				zap.String("offerID", id.String()),
				zap.String("projectID", offer.ProjectID.String()),
				zap.Error(err))
		}
	}

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

