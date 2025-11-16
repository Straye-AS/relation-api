package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

type OfferService struct {
	offerRepo     *repository.OfferRepository
	offerItemRepo *repository.OfferItemRepository
	customerRepo  *repository.CustomerRepository
	projectRepo   *repository.ProjectRepository
	fileRepo      *repository.FileRepository
	activityRepo  *repository.ActivityRepository
	logger        *zap.Logger
}

func NewOfferService(
	offerRepo *repository.OfferRepository,
	offerItemRepo *repository.OfferItemRepository,
	customerRepo *repository.CustomerRepository,
	projectRepo *repository.ProjectRepository,
	fileRepo *repository.FileRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *OfferService {
	return &OfferService{
		offerRepo:     offerRepo,
		offerItemRepo: offerItemRepo,
		customerRepo:  customerRepo,
		projectRepo:   projectRepo,
		fileRepo:      fileRepo,
		activityRepo:  activityRepo,
		logger:        logger,
	}
}

func (s *OfferService) Create(ctx context.Context, req *domain.CreateOfferRequest) (*domain.OfferDTO, error) {
	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
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
		ResponsibleUserName: "", // TODO: Get from user service
		Description:         req.Description,
		Notes:               req.Notes,
		Items:               items,
	}

	if err := s.offerRepo.Create(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	// Reload with relations
	offer, _ = s.offerRepo.GetByID(ctx, offer.ID)

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTypeOffer,
			TargetID:    offer.ID,
			Title:       "Offer created",
			Body:        fmt.Sprintf("Offer '%s' was created", offer.Title),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

func (s *OfferService) GetByID(ctx context.Context, id uuid.UUID) (*domain.OfferWithItemsDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
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

func (s *OfferService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateOfferRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer: %w", err)
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
	offer, _ = s.offerRepo.GetByID(ctx, id)

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTypeOffer,
			TargetID:    offer.ID,
			Title:       "Offer updated",
			Body:        fmt.Sprintf("Offer '%s' was updated", offer.Title),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

func (s *OfferService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.offerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete offer: %w", err)
	}

	return nil
}

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

func (s *OfferService) Advance(ctx context.Context, id uuid.UUID, req *domain.AdvanceOfferRequest) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	oldPhase := offer.Phase
	offer.Phase = req.Phase

	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTypeOffer,
			TargetID:    offer.ID,
			Title:       "Offer phase advanced",
			Body:        fmt.Sprintf("Offer '%s' advanced from %s to %s", offer.Title, oldPhase, offer.Phase),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

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

func (s *OfferService) AddItem(ctx context.Context, offerID uuid.UUID, req *domain.CreateOfferItemRequest) (*domain.OfferItemDTO, error) {
	// Verify offer exists and keep reference for recalculating totals
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("offer not found: %w", err)
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

func (s *OfferService) GetActivities(ctx context.Context, id uuid.UUID, limit int) ([]domain.ActivityDTO, error) {
	activities, err := s.activityRepo.ListByTarget(ctx, domain.ActivityTypeOffer, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
	}

	return dtos, nil
}
