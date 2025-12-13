package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NumberSequenceRepository handles database operations for number sequences.
// Number sequences are SHARED between offers and projects to ensure unique
// numbers across both entity types within a company/year combination.
type NumberSequenceRepository struct {
	db *gorm.DB
}

// NewNumberSequenceRepository creates a new NumberSequenceRepository
func NewNumberSequenceRepository(db *gorm.DB) *NumberSequenceRepository {
	return &NumberSequenceRepository{db: db}
}

// GetNextNumber atomically retrieves and increments the sequence for a company/year.
// This method is thread-safe and uses SELECT FOR UPDATE to prevent race conditions.
// If no sequence exists for the company/year, it creates one starting at 1.
//
// Returns the next sequence number to use (already incremented in DB).
func (r *NumberSequenceRepository) GetNextNumber(ctx context.Context, companyID domain.CompanyID, year int) (int, error) {
	var seq domain.NumberSequence
	var nextSeq int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Try to get existing sequence with row lock for atomicity
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("company_id = ? AND year = ?", companyID, year).
			First(&seq)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new sequence for this company/year combination
			seq = domain.NumberSequence{
				CompanyID:    companyID,
				Year:         year,
				LastSequence: 1,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			if err := tx.Create(&seq).Error; err != nil {
				return fmt.Errorf("failed to create number sequence: %w", err)
			}
			nextSeq = 1
		} else if result.Error != nil {
			return fmt.Errorf("failed to get number sequence: %w", result.Error)
		} else {
			// Increment existing sequence
			nextSeq = seq.LastSequence + 1
			if err := tx.Model(&seq).Updates(map[string]interface{}{
				"last_sequence": nextSeq,
				"updated_at":    time.Now(),
			}).Error; err != nil {
				return fmt.Errorf("failed to update number sequence: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return nextSeq, nil
}

// GetCurrentSequence retrieves the current sequence value without incrementing.
// Returns 0 if no sequence exists for the company/year.
func (r *NumberSequenceRepository) GetCurrentSequence(ctx context.Context, companyID domain.CompanyID, year int) (int, error) {
	var seq domain.NumberSequence
	result := r.db.WithContext(ctx).
		Where("company_id = ? AND year = ?", companyID, year).
		First(&seq)

	if result.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if result.Error != nil {
		return 0, fmt.Errorf("failed to get number sequence: %w", result.Error)
	}

	return seq.LastSequence, nil
}

// SetSequence sets the sequence to a specific value.
// This is useful for data migrations when we need to set the sequence
// to account for existing numbered offers/projects.
// The value should be the LAST USED sequence number (next number will be value+1).
func (r *NumberSequenceRepository) SetSequence(ctx context.Context, companyID domain.CompanyID, year int, value int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var seq domain.NumberSequence
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("company_id = ? AND year = ?", companyID, year).
			First(&seq)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new sequence with the specified value
			seq = domain.NumberSequence{
				CompanyID:    companyID,
				Year:         year,
				LastSequence: value,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			if err := tx.Create(&seq).Error; err != nil {
				return fmt.Errorf("failed to create number sequence: %w", err)
			}
		} else if result.Error != nil {
			return fmt.Errorf("failed to get number sequence: %w", result.Error)
		} else {
			// Update existing sequence only if new value is higher
			// This prevents accidental reduction of the sequence
			if value > seq.LastSequence {
				if err := tx.Model(&seq).Updates(map[string]interface{}{
					"last_sequence": value,
					"updated_at":    time.Now(),
				}).Error; err != nil {
					return fmt.Errorf("failed to update number sequence: %w", err)
				}
			}
		}

		return nil
	})
}

// ListSequences returns all sequences (useful for debugging/admin)
func (r *NumberSequenceRepository) ListSequences(ctx context.Context) ([]domain.NumberSequence, error) {
	var sequences []domain.NumberSequence
	err := r.db.WithContext(ctx).
		Order("company_id ASC, year DESC").
		Find(&sequences).Error
	return sequences, err
}
