package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AssignmentRepository handles database operations for assignments
type AssignmentRepository struct {
	db *gorm.DB
}

// NewAssignmentRepository creates a new assignment repository
func NewAssignmentRepository(db *gorm.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// AssignmentUpsertInput represents data for upserting an assignment from DW
type AssignmentUpsertInput struct {
	DWAssignmentID   int64
	DWProjectID      int64
	OfferID          *uuid.UUID
	CompanyID        domain.CompanyID
	AssignmentNumber string
	Description      string
	FixedPriceAmount float64
	StatusID         *int
	ProgressID       *int
	RawData          map[string]interface{}
}

// UpsertFromDW performs a bulk upsert of assignments from the datawarehouse.
// Uses ON CONFLICT to update existing records or insert new ones.
// Returns counts of created and updated records.
func (r *AssignmentRepository) UpsertFromDW(ctx context.Context, inputs []AssignmentUpsertInput) (created int, updated int, err error) {
	if len(inputs) == 0 {
		return 0, 0, nil
	}

	now := time.Now()
	assignments := make([]domain.Assignment, 0, len(inputs))

	for _, input := range inputs {
		rawDataJSON := ""
		if input.RawData != nil {
			data, err := json.Marshal(input.RawData)
			if err == nil {
				rawDataJSON = string(data)
			}
		}

		assignments = append(assignments, domain.Assignment{
			DWAssignmentID:   input.DWAssignmentID,
			DWProjectID:      input.DWProjectID,
			OfferID:          input.OfferID,
			CompanyID:        input.CompanyID,
			AssignmentNumber: input.AssignmentNumber,
			Description:      input.Description,
			FixedPriceAmount: input.FixedPriceAmount,
			StatusID:         input.StatusID,
			ProgressID:       input.ProgressID,
			DWRawData:        rawDataJSON,
			DWSyncedAt:       now,
		})
	}

	// Count existing before upsert
	var existingIDs []int64
	for _, a := range assignments {
		existingIDs = append(existingIDs, a.DWAssignmentID)
	}

	var existingCount int64
	err = r.db.WithContext(ctx).
		Model(&domain.Assignment{}).
		Where("company_id = ? AND dw_assignment_id IN ?", assignments[0].CompanyID, existingIDs).
		Count(&existingCount).Error
	if err != nil {
		return 0, 0, fmt.Errorf("count existing assignments: %w", err)
	}

	// Perform upsert using ON CONFLICT
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "company_id"}, {Name: "dw_assignment_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"dw_project_id",
				"offer_id",
				"assignment_number",
				"description",
				"fixed_price_amount",
				"status_id",
				"progress_id",
				"dw_raw_data",
				"dw_synced_at",
				"updated_at",
			}),
		}).
		Create(&assignments)

	if result.Error != nil {
		return 0, 0, fmt.Errorf("upsert assignments: %w", result.Error)
	}

	totalAffected := int(result.RowsAffected)
	updated = int(existingCount)
	created = totalAffected - updated
	if created < 0 {
		created = 0
	}

	return created, updated, nil
}

// GetByOfferID retrieves all assignments for an offer
func (r *AssignmentRepository) GetByOfferID(ctx context.Context, offerID uuid.UUID) ([]domain.Assignment, error) {
	var assignments []domain.Assignment
	err := r.db.WithContext(ctx).
		Where("offer_id = ?", offerID).
		Order("assignment_number ASC").
		Find(&assignments).Error
	if err != nil {
		return nil, fmt.Errorf("get assignments by offer ID: %w", err)
	}
	return assignments, nil
}

// GetAggregatedByOfferID returns aggregated financial data for an offer's assignments
func (r *AssignmentRepository) GetAggregatedByOfferID(ctx context.Context, offerID uuid.UUID) (totalFixedPriceAmount float64, count int, lastSyncedAt *time.Time, err error) {
	var result struct {
		TotalFixedPriceAmount float64
		Count                 int64
		LastSyncedAt          *time.Time
	}

	err = r.db.WithContext(ctx).
		Model(&domain.Assignment{}).
		Select("COALESCE(SUM(fixed_price_amount), 0) as total_fixed_price_amount, COUNT(*) as count, MAX(dw_synced_at) as last_synced_at").
		Where("offer_id = ?", offerID).
		Scan(&result).Error

	if err != nil {
		return 0, 0, nil, fmt.Errorf("aggregate assignments by offer ID: %w", err)
	}

	return result.TotalFixedPriceAmount, int(result.Count), result.LastSyncedAt, nil
}

// DeleteStaleByOfferID removes assignments that are no longer in the DW.
// validDWIDs contains the current DW assignment IDs that should be kept.
// Returns the count of deleted records.
func (r *AssignmentRepository) DeleteStaleByOfferID(ctx context.Context, offerID uuid.UUID, validDWIDs []int64) (deleted int, err error) {
	if len(validDWIDs) == 0 {
		// Delete all assignments for this offer
		result := r.db.WithContext(ctx).
			Where("offer_id = ?", offerID).
			Delete(&domain.Assignment{})
		return int(result.RowsAffected), result.Error
	}

	// Delete assignments not in the valid list
	result := r.db.WithContext(ctx).
		Where("offer_id = ? AND dw_assignment_id NOT IN ?", offerID, validDWIDs).
		Delete(&domain.Assignment{})

	if result.Error != nil {
		return 0, fmt.Errorf("delete stale assignments: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// GetByCompanyAndDWID retrieves an assignment by company and DW assignment ID
func (r *AssignmentRepository) GetByCompanyAndDWID(ctx context.Context, companyID domain.CompanyID, dwAssignmentID int64) (*domain.Assignment, error) {
	var assignment domain.Assignment
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND dw_assignment_id = ?", companyID, dwAssignmentID).
		First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// CountByOfferID returns the count of assignments for an offer
func (r *AssignmentRepository) CountByOfferID(ctx context.Context, offerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Assignment{}).
		Where("offer_id = ?", offerID).
		Count(&count).Error
	return count, err
}

