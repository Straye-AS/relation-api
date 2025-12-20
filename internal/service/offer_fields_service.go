package service

// This file contains individual field update methods extracted from offer_service.go
// for better code organization. These methods handle:
// - Individual property updates (UpdateProbability, UpdateTitle, etc.)
// - Project linking/unlinking
// - Offer number management

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

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
// If expirationDate is nil, clears the expiration date
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

	// Validate expiration date is not before sent date (if provided)
	if expirationDate != nil && offer.SentDate != nil && expirationDate.Before(*offer.SentDate) {
		return nil, ErrExpirationDateBeforeSentDate
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"expiration_date": expirationDate,
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

	expirationDateStr := "fjernet"
	if expirationDate != nil {
		expirationDateStr = expirationDate.Format("2006-01-02")
	}
	s.logActivity(ctx, id, offer.Title, "Tilbud utløpsdato oppdatert",
		fmt.Sprintf("Utløpsdato satt til %s", expirationDateStr))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateSentDate updates only the sent date field of an offer.
// The sent date records when the offer was sent to the customer.
func (s *OfferService) UpdateSentDate(ctx context.Context, id uuid.UUID, sentDate *time.Time) (*domain.OfferDTO, error) {
	// Check offer exists
	if _, err := s.offerRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"sent_date": sentDate,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update sent date: %w", err)
	}

	// Reload and return
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	sentDateStr := "fjernet"
	if sentDate != nil {
		sentDateStr = sentDate.Format("2006-01-02")
	}
	s.logActivity(ctx, id, offer.Title, "Tilbud sendedato oppdatert",
		fmt.Sprintf("Sendedato satt til %s", sentDateStr))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateStartDate updates only the start date field of an offer.
// The start date records when work on the offer/order is expected to begin.
// Can be edited at any phase.
func (s *OfferService) UpdateStartDate(ctx context.Context, id uuid.UUID, startDate *time.Time) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate that start date is not after end date (if both are set)
	if startDate != nil && offer.EndDate != nil && startDate.After(*offer.EndDate) {
		return nil, ErrEndDateBeforeStartDate
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"start_date": startDate,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update start date: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	startDateStr := "fjernet"
	if startDate != nil {
		startDateStr = startDate.Format("2006-01-02")
	}
	s.logActivity(ctx, id, offer.Title, "Tilbud startdato oppdatert",
		fmt.Sprintf("Startdato satt til %s", startDateStr))

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// UpdateEndDate updates only the end date field of an offer.
// The end date records when work on the offer/order is expected to complete.
// End date cannot be before start date. Can be edited at any phase.
func (s *OfferService) UpdateEndDate(ctx context.Context, id uuid.UUID, endDate *time.Time) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Validate that end date is not before start date (if both are set)
	if endDate != nil && offer.StartDate != nil && endDate.Before(*offer.StartDate) {
		return nil, ErrEndDateBeforeStartDate
	}

	// Build updates including user tracking fields
	updates := map[string]interface{}{
		"end_date": endDate,
	}
	if userCtx, ok := auth.FromContext(ctx); ok {
		updates["updated_by_id"] = userCtx.UserID.String()
		updates["updated_by_name"] = userCtx.DisplayName
	}
	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update end date: %w", err)
	}

	// Reload and return
	offer, err = s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer: %w", err)
	}

	endDateStr := "fjernet"
	if endDate != nil {
		endDateStr = endDate.Format("2006-01-02")
	}
	s.logActivity(ctx, id, offer.Title, "Tilbud sluttdato oppdatert",
		fmt.Sprintf("Sluttdato satt til %s", endDateStr))

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
	externalRefChanged := oldRef != externalReference

	offer.ExternalReference = externalReference

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		offer.UpdatedByID = userCtx.UserID.String()
		offer.UpdatedByName = userCtx.DisplayName
	}

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Handle DW financial data based on external_reference change
	if externalRefChanged {
		if externalReference == "" {
			// Clearing external_reference - reset all DW financial fields to 0
			if err := s.offerRepo.ClearDWFinancials(ctx, offerID); err != nil {
				s.logger.Error("failed to clear DW financials after removing external reference",
					zap.Error(err),
					zap.String("offer_id", offerID.String()))
				// Don't fail the request, just log the error
			}
		} else {
			// Setting or changing external_reference - sync from DW if available
			if s.dwClient != nil && s.dwClient.IsEnabled() {
				// Trigger async sync in background so we don't block the response
				go func() {
					syncCtx := context.Background()
					_, syncErr := s.SyncFromDataWarehouse(syncCtx, offerID)
					if syncErr != nil {
						s.logger.Error("failed to sync from DW after external reference change",
							zap.Error(syncErr),
							zap.String("offer_id", offerID.String()),
							zap.String("new_external_reference", externalReference))
					} else {
						s.logger.Info("synced from DW after external reference change",
							zap.String("offer_id", offerID.String()),
							zap.String("new_external_reference", externalReference))
					}
				}()
			}
		}
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

	// Reload offer to get updated values (especially after ClearDWFinancials)
	offer, err = s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		s.logger.Warn("failed to reload offer after external reference update", zap.Error(err))
	}

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

