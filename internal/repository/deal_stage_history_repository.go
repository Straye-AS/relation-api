package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type DealStageHistoryRepository struct {
	db *gorm.DB
}

func NewDealStageHistoryRepository(db *gorm.DB) *DealStageHistoryRepository {
	return &DealStageHistoryRepository{db: db}
}

// Create records a new stage transition
func (r *DealStageHistoryRepository) Create(ctx context.Context, history *domain.DealStageHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetByDealID returns all stage history for a deal, ordered by change time
func (r *DealStageHistoryRepository) GetByDealID(ctx context.Context, dealID uuid.UUID) ([]domain.DealStageHistory, error) {
	var history []domain.DealStageHistory
	err := r.db.WithContext(ctx).
		Where("deal_id = ?", dealID).
		Order("changed_at DESC").
		Find(&history).Error
	return history, err
}

// GetLatestByDealID returns the most recent stage change for a deal
func (r *DealStageHistoryRepository) GetLatestByDealID(ctx context.Context, dealID uuid.UUID) (*domain.DealStageHistory, error) {
	var history domain.DealStageHistory
	err := r.db.WithContext(ctx).
		Where("deal_id = ?", dealID).
		Order("changed_at DESC").
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// GetByUserID returns all stage changes made by a specific user
func (r *DealStageHistoryRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]domain.DealStageHistory, error) {
	var history []domain.DealStageHistory
	err := r.db.WithContext(ctx).
		Preload("Deal").
		Where("changed_by_id = ?", userID).
		Order("changed_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// GetTransitionsToStage returns all transitions to a specific stage within a date range
func (r *DealStageHistoryRepository) GetTransitionsToStage(ctx context.Context, stage domain.DealStage, from, to time.Time) ([]domain.DealStageHistory, error) {
	var history []domain.DealStageHistory
	err := r.db.WithContext(ctx).
		Preload("Deal").
		Where("to_stage = ?", stage).
		Where("changed_at >= ? AND changed_at <= ?", from, to).
		Order("changed_at DESC").
		Find(&history).Error
	return history, err
}

// CountTransitionsByStage returns the count of transitions to each stage within a date range
func (r *DealStageHistoryRepository) CountTransitionsByStage(ctx context.Context, from, to time.Time) (map[domain.DealStage]int64, error) {
	type result struct {
		ToStage domain.DealStage
		Count   int64
	}
	var results []result

	err := r.db.WithContext(ctx).Model(&domain.DealStageHistory{}).
		Select("to_stage, COUNT(*) as count").
		Where("changed_at >= ? AND changed_at <= ?", from, to).
		Group("to_stage").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[domain.DealStage]int64)
	for _, r := range results {
		counts[r.ToStage] = r.Count
	}
	return counts, nil
}

// RecordTransition is a convenience method to create a stage history record
func (r *DealStageHistoryRepository) RecordTransition(
	ctx context.Context,
	dealID uuid.UUID,
	fromStage *domain.DealStage,
	toStage domain.DealStage,
	changedByID string,
	changedByName string,
	notes string,
) error {
	history := &domain.DealStageHistory{
		DealID:        dealID,
		FromStage:     fromStage,
		ToStage:       toStage,
		ChangedByID:   changedByID,
		ChangedByName: changedByName,
		Notes:         notes,
		ChangedAt:     time.Now(),
	}
	return r.Create(ctx, history)
}

// GetAverageTimeInStage calculates the average time deals spend in each stage
func (r *DealStageHistoryRepository) GetAverageTimeInStage(ctx context.Context) (map[domain.DealStage]time.Duration, error) {
	// This is a complex query - we need to calculate time between consecutive stage changes
	// For simplicity, we'll do this in Go rather than complex SQL

	var allHistory []domain.DealStageHistory
	err := r.db.WithContext(ctx).
		Order("deal_id, changed_at ASC").
		Find(&allHistory).Error
	if err != nil {
		return nil, err
	}

	stageDurations := make(map[domain.DealStage][]time.Duration)

	// Group by deal and calculate durations
	var currentDealID uuid.UUID
	var lastChange *domain.DealStageHistory

	for i := range allHistory {
		h := &allHistory[i]
		if h.DealID != currentDealID {
			currentDealID = h.DealID
			lastChange = h
			continue
		}

		if lastChange != nil && lastChange.FromStage != nil {
			duration := h.ChangedAt.Sub(lastChange.ChangedAt)
			stageDurations[*lastChange.FromStage] = append(stageDurations[*lastChange.FromStage], duration)
		}
		lastChange = h
	}

	// Calculate averages
	averages := make(map[domain.DealStage]time.Duration)
	for stage, durations := range stageDurations {
		if len(durations) == 0 {
			continue
		}
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		averages[stage] = total / time.Duration(len(durations))
	}

	return averages, nil
}

// DeleteByDealID removes all history for a deal (used when deal is deleted)
func (r *DealStageHistoryRepository) DeleteByDealID(ctx context.Context, dealID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("deal_id = ?", dealID).
		Delete(&domain.DealStageHistory{}).Error
}
