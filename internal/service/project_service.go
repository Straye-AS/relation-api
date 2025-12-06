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

type ProjectService struct {
	projectRepo  *repository.ProjectRepository
	customerRepo *repository.CustomerRepository
	activityRepo *repository.ActivityRepository
	logger       *zap.Logger
}

func NewProjectService(
	projectRepo *repository.ProjectRepository,
	customerRepo *repository.CustomerRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *ProjectService {
	return &ProjectService{
		projectRepo:  projectRepo,
		customerRepo: customerRepo,
		activityRepo: activityRepo,
		logger:       logger,
	}
}

func (s *ProjectService) Create(ctx context.Context, req *domain.CreateProjectRequest) (*domain.ProjectDTO, error) {
	// Verify customer exists
	if _, err := s.customerRepo.GetByID(ctx, req.CustomerID); err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	project := &domain.Project{
		CustomerID:    req.CustomerID,
		Name:          req.Name,
		ProjectNumber: req.ProjectNumber,
		Summary:       req.Summary,
		Description:   req.Description,
		Budget:        req.Budget,
		Spent:         req.Spent,
		Status:        req.Status,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		CompanyID:     req.CompanyID,
		ManagerID:     req.ManagerID,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Reload with customer
	project, _ = s.projectRepo.GetByID(ctx, project.ID)

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetProject,
			TargetID:    project.ID,
			Title:       "Project created",
			Body:        fmt.Sprintf("Project '%s' was created", project.Name),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateProjectRequest) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	project.Name = req.Name
	project.ProjectNumber = req.ProjectNumber
	project.Summary = req.Summary
	project.Description = req.Description
	project.Budget = req.Budget
	project.Spent = req.Spent
	project.Status = req.Status
	project.StartDate = req.StartDate
	project.EndDate = req.EndDate
	project.CompanyID = req.CompanyID
	project.ManagerID = req.ManagerID

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Reload with customer
	project, _ = s.projectRepo.GetByID(ctx, id)

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetProject,
			TargetID:    project.ID,
			Title:       "Project updated",
			Body:        fmt.Sprintf("Project '%s' was updated", project.Name),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

func (s *ProjectService) List(ctx context.Context, page, pageSize int, customerID *uuid.UUID, status *domain.ProjectStatus) (*domain.PaginatedResponse, error) {
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

	projects, total, err := s.projectRepo.List(ctx, page, pageSize, customerID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	dtos := make([]domain.ProjectDTO, len(projects))
	for i, project := range projects {
		dtos[i] = mapper.ToProjectDTO(&project)
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

func (s *ProjectService) GetBudget(ctx context.Context, id uuid.UUID) (*domain.ProjectBudgetDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	dto := mapper.ToProjectBudgetDTO(project)
	return &dto, nil
}

func (s *ProjectService) GetActivities(ctx context.Context, id uuid.UUID, limit int) ([]domain.ActivityDTO, error) {
	activities, err := s.activityRepo.ListByTarget(ctx, domain.ActivityTargetProject, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
	}

	return dtos, nil
}
