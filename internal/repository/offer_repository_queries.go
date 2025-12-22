package repository

// This file contains complex query methods for the OfferRepository.
// Includes:
// - Inquiry (draft offer) listing
// - Field updates
// - Offer number generation
// - Project-offer relationship methods

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ============================================================================
// Inquiry (Draft Offer) Methods
// ============================================================================

// ListInquiries returns a paginated list of offers in draft phase (inquiries)
func (r *OfferRepository) ListInquiries(ctx context.Context, page, pageSize int, customerID *uuid.UUID) ([]domain.Offer, int64, error) {
	var offers []domain.Offer
	var total int64

	// Validate and normalize pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	query := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Preload("Customer").
		Where("phase = ?", domain.OfferPhaseDraft)

	query = ApplyCompanyFilter(ctx, query)

	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&offers).Error

	return offers, total, err
}

// UpdateField updates a single field on an offer
// Returns error if offer not found or user lacks access
func (r *OfferRepository) UpdateField(ctx context.Context, id uuid.UUID, field string, value interface{}) error {
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Update(field, value)
	if result.Error != nil {
		return fmt.Errorf("failed to update offer %s: %w", field, result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateFields updates multiple fields on an offer
func (r *OfferRepository) UpdateFields(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update offer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ============================================================================
// Offer Number Generation Methods
// ============================================================================

// GenerateOfferNumber generates the next unique offer number for a company
// Format: {PREFIX}-{YEAR}-{SEQUENCE} e.g., "STB-2024-001"
// Uses SELECT FOR UPDATE to ensure thread-safe sequence generation
func (r *OfferRepository) GenerateOfferNumber(ctx context.Context, companyID domain.CompanyID) (string, error) {
	year := time.Now().Year()
	prefix := domain.GetCompanyPrefix(companyID)

	var seq domain.OfferNumberSequence
	var nextSeq int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Try to get existing sequence with row lock
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("company_id = ? AND year = ?", companyID, year).
			First(&seq)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new sequence for this company/year
			seq = domain.OfferNumberSequence{
				CompanyID:    companyID,
				Year:         year,
				LastSequence: 1,
			}
			if err := tx.Create(&seq).Error; err != nil {
				return fmt.Errorf("failed to create offer number sequence: %w", err)
			}
			nextSeq = 1
		} else if result.Error != nil {
			return fmt.Errorf("failed to get offer number sequence: %w", result.Error)
		} else {
			// Increment existing sequence
			nextSeq = seq.LastSequence + 1
			if err := tx.Model(&seq).Update("last_sequence", nextSeq).Error; err != nil {
				return fmt.Errorf("failed to update offer number sequence: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Format: PREFIX-YYYY-NNN (zero-padded to 3 digits)
	offerNumber := fmt.Sprintf("%s-%d-%03d", prefix, year, nextSeq)
	return offerNumber, nil
}

// SetOfferNumber sets the offer number for an offer
func (r *OfferRepository) SetOfferNumber(ctx context.Context, id uuid.UUID, offerNumber string) error {
	return r.UpdateField(ctx, id, "offer_number", offerNumber)
}

// LinkToProject sets the project_id and project_name for an offer
func (r *OfferRepository) LinkToProject(ctx context.Context, offerID uuid.UUID, projectID uuid.UUID) error {
	// Fetch project name for denormalized field
	var projectName string
	err := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID).
		Pluck("name", &projectName).Error
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	return r.UpdateFields(ctx, offerID, map[string]interface{}{
		"project_id":   projectID,
		"project_name": projectName,
	})
}

// UnlinkFromProject removes the project link from an offer
func (r *OfferRepository) UnlinkFromProject(ctx context.Context, offerID uuid.UUID) error {
	return r.UpdateFields(ctx, offerID, map[string]interface{}{
		"project_id":   nil,
		"project_name": "",
	})
}

// OfferNumberExists checks if an offer number already exists, excluding the given offer ID
func (r *OfferRepository) OfferNumberExists(ctx context.Context, offerNumber string, excludeOfferID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("offer_number = ? AND id != ?", offerNumber, excludeOfferID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check offer number existence: %w", err)
	}
	return count > 0, nil
}

// SetExternalReference sets the external reference for an offer
func (r *OfferRepository) SetExternalReference(ctx context.Context, id uuid.UUID, externalReference string) error {
	return r.UpdateField(ctx, id, "external_reference", externalReference)
}

// ExternalReferenceExists checks if an external reference already exists within a company, excluding the given offer ID
func (r *OfferRepository) ExternalReferenceExists(ctx context.Context, externalReference string, companyID domain.CompanyID, excludeOfferID uuid.UUID) (bool, error) {
	if externalReference == "" {
		return false, nil // Empty references are allowed to not be unique
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("external_reference = ? AND company_id = ? AND id != ?", externalReference, companyID, excludeOfferID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check external reference existence: %w", err)
	}
	return count > 0, nil
}

// ============================================================================
// Project-Offer Relationship Methods (Offer Folder Model)
// ============================================================================

// ListByProject returns all offers linked to a specific project
func (r *OfferRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Items").
		Where("project_id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("updated_at DESC").Find(&offers).Error
	return offers, err
}

// GetHighestOfferValueForProject returns the highest offer value among all offers in a project
// Only considers offers that are not in terminal states (order, completed, lost, expired)
func (r *OfferRepository) GetHighestOfferValueForProject(ctx context.Context, projectID uuid.UUID) (float64, error) {
	var maxValue float64
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("COALESCE(MAX(value), 0)").
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Scan(&maxValue).Error
	return maxValue, err
}

// ExpireSiblingOffers marks all other offers in the same project as expired
// This is called when one offer becomes an order - the others become expired (NOT lost)
// Returns the IDs of the expired offers
func (r *OfferRepository) ExpireSiblingOffers(ctx context.Context, projectID uuid.UUID, winningOfferID uuid.UUID) ([]uuid.UUID, error) {
	// First get the IDs of offers that will be expired
	var offerIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("id").
		Where("project_id = ?", projectID).
		Where("id != ?", winningOfferID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		}).
		Pluck("id", &offerIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sibling offer IDs: %w", err)
	}

	if len(offerIDs) == 0 {
		return nil, nil
	}

	// Update the phase to expired for sibling offers
	result := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id IN ?", offerIDs).
		Update("phase", domain.OfferPhaseExpired)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to expire sibling offers: %w", result.Error)
	}

	return offerIDs, nil
}

// GetExpiredSiblingOffers returns offers that were expired by selecting another offer
func (r *OfferRepository) GetExpiredSiblingOffers(ctx context.Context, projectID uuid.UUID, winningOfferID uuid.UUID) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("project_id = ?", projectID).
		Where("id != ?", winningOfferID).
		Where("phase = ?", domain.OfferPhaseExpired)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&offers).Error
	return offers, err
}

// SetOfferNumberWithSuffix updates an offer's number by adding a suffix
// This is used when an offer wins to mark it with "_P" suffix
func (r *OfferRepository) SetOfferNumberWithSuffix(ctx context.Context, id uuid.UUID, suffix string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id).
		Update("offer_number", gorm.Expr("offer_number || ?", suffix)).Error
}

// CountOffersByProject returns the count of offers for a project
func (r *OfferRepository) CountOffersByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("project_id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return count, err
}

// CountActiveOffersByProject returns the count of active offers (not order/completed/lost/expired) for a project
func (r *OfferRepository) CountActiveOffersByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return count, err
}

// GetActiveOffersByProject returns all active offers (not order/completed/lost/expired) for a project
func (r *OfferRepository) GetActiveOffersByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("value DESC").Find(&offers).Error
	return offers, err
}

// GetBestActiveOfferForProject returns the highest value active offer for a project
// Returns nil if no active offers exist
func (r *OfferRepository) GetBestActiveOfferForProject(ctx context.Context, projectID uuid.UUID) (*domain.Offer, error) {
	var offer domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		}).
		Order("value DESC").
		Limit(1)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&offer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No active offers found
		}
		return nil, err
	}
	return &offer, nil
}

// GetDistinctCustomerIDsForActiveOffers returns the distinct customer IDs from active offers
// (in_progress or sent phase) for a project. Used to determine if project customer can be inferred.
// Returns empty slice if no active offers exist.
func (r *OfferRepository) GetDistinctCustomerIDsForActiveOffers(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	var customerIDs []uuid.UUID
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("DISTINCT customer_id").
		Where("project_id = ?", projectID).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Pluck("customer_id", &customerIDs).Error
	if err != nil {
		return nil, err
	}
	return customerIDs, nil
}

// UpdateProjectNameByProjectID updates the project_name for all offers linked to a project
func (r *OfferRepository) UpdateProjectNameByProjectID(ctx context.Context, projectID uuid.UUID, projectName string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("project_id = ?", projectID).
		Update("project_name", projectName)
	if result.Error != nil {
		return fmt.Errorf("failed to update project_name for offers: %w", result.Error)
	}
	return nil
}

// GetBestOfferForProject returns the "best" offer for a project based on priority:
// 1. Completed offer (if any exists)
// 2. Order offer (if any exists)
// 3. Sent offer with highest value
// 4. In_progress offer with highest value
// 5. Draft offer with highest value
// Returns nil if no offers exist for the project.
func (r *OfferRepository) GetBestOfferForProject(ctx context.Context, projectID uuid.UUID) (*domain.Offer, error) {
	var offer domain.Offer

	// Priority 1: Completed offer
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseCompleted).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query completed offers: %w", err)
	}

	// Priority 2: Order offer (in execution)
	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseOrder).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query order offers: %w", err)
	}

	// Priority 3: Sent offer with highest value
	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseSent).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query sent offers: %w", err)
	}

	// Priority 4: In_progress offer with highest value
	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseInProgress).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query in_progress offers: %w", err)
	}

	// Priority 5: Draft offer with highest value
	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseDraft).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query draft offers: %w", err)
	}

	// No offers found
	return nil, nil
}

// ============================================================================
// Offer-Supplier Relationship Methods
// ============================================================================

// GetOfferSuppliers returns all suppliers linked to an offer with their relationship details
// Preloads the full Supplier entity for each relationship
func (r *OfferRepository) GetOfferSuppliers(ctx context.Context, offerID uuid.UUID) ([]domain.OfferSupplier, error) {
	// First verify the offer exists and user has access
	var count int64
	offerQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("id = ?", offerID)
	offerQuery = ApplyCompanyFilter(ctx, offerQuery)
	if err := offerQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to verify offer access: %w", err)
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var offerSuppliers []domain.OfferSupplier
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("Contact").
		Where("offer_id = ?", offerID).
		Order("updated_at DESC").
		Find(&offerSuppliers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get offer suppliers: %w", err)
	}

	return offerSuppliers, nil
}

// GetOfferSupplier returns a single offer-supplier relationship
func (r *OfferRepository) GetOfferSupplier(ctx context.Context, offerID, supplierID uuid.UUID) (*domain.OfferSupplier, error) {
	// First verify the offer exists and user has access
	var count int64
	offerQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("id = ?", offerID)
	offerQuery = ApplyCompanyFilter(ctx, offerQuery)
	if err := offerQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to verify offer access: %w", err)
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var offerSupplier domain.OfferSupplier
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("Contact").
		Where("offer_id = ? AND supplier_id = ?", offerID, supplierID).
		First(&offerSupplier).Error
	if err != nil {
		return nil, err
	}

	return &offerSupplier, nil
}

// OfferSupplierExists checks if a supplier is already linked to an offer
func (r *OfferRepository) OfferSupplierExists(ctx context.Context, offerID, supplierID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.OfferSupplier{}).
		Where("offer_id = ? AND supplier_id = ?", offerID, supplierID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check offer supplier existence: %w", err)
	}
	return count > 0, nil
}

// CreateOfferSupplier adds a supplier to an offer
func (r *OfferRepository) CreateOfferSupplier(ctx context.Context, offerSupplier *domain.OfferSupplier) error {
	return r.db.WithContext(ctx).Create(offerSupplier).Error
}

// UpdateOfferSupplier updates an offer-supplier relationship
func (r *OfferRepository) UpdateOfferSupplier(ctx context.Context, offerSupplier *domain.OfferSupplier) error {
	return r.db.WithContext(ctx).Save(offerSupplier).Error
}

// DeleteOfferSupplier removes a supplier from an offer
func (r *OfferRepository) DeleteOfferSupplier(ctx context.Context, offerID, supplierID uuid.UUID) error {
	// First verify the offer exists and user has access
	var count int64
	offerQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("id = ?", offerID)
	offerQuery = ApplyCompanyFilter(ctx, offerQuery)
	if err := offerQuery.Count(&count).Error; err != nil {
		return fmt.Errorf("failed to verify offer access: %w", err)
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	result := r.db.WithContext(ctx).
		Where("offer_id = ? AND supplier_id = ?", offerID, supplierID).
		Delete(&domain.OfferSupplier{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete offer supplier: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
