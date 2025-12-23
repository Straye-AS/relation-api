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

// Project-specific service errors
var (
	// ErrInvalidPhaseTransition is returned when trying to make an invalid phase transition
	ErrInvalidPhaseTransition = errors.New("invalid project phase transition")

	// ErrCannotReopenProject is returned when a project cannot be reopened
	ErrCannotReopenProject = errors.New("project cannot be reopened from its current state")

	// ErrWorkingPhaseRequiresStartDate is returned when transitioning to working without a start date
	ErrWorkingPhaseRequiresStartDate = errors.New("working phase requires a start date")
)

// ProjectService handles business logic for projects
// Projects are now simplified containers/folders for offers. Economic tracking lives on Offer.
type ProjectService struct {
	projectRepo  *repository.ProjectRepository
	offerRepo    *repository.OfferRepository
	customerRepo *repository.CustomerRepository
	activityRepo *repository.ActivityRepository
	fileService  *FileService
	logger       *zap.Logger
	db           *gorm.DB
}

// NewProjectService creates a new ProjectService with basic dependencies
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

// NewProjectServiceWithDeps creates a ProjectService with all dependencies for full feature support
func NewProjectServiceWithDeps(
	projectRepo *repository.ProjectRepository,
	offerRepo *repository.OfferRepository,
	customerRepo *repository.CustomerRepository,
	activityRepo *repository.ActivityRepository,
	fileService *FileService,
	logger *zap.Logger,
	db *gorm.DB,
) *ProjectService {
	return &ProjectService{
		projectRepo:  projectRepo,
		offerRepo:    offerRepo,
		customerRepo: customerRepo,
		activityRepo: activityRepo,
		fileService:  fileService,
		logger:       logger,
		db:           db,
	}
}

// Create creates a new project with activity logging
// Projects are now simplified containers for offers - no economics or management fields
func (s *ProjectService) Create(ctx context.Context, req *domain.CreateProjectRequest) (*domain.ProjectDTO, error) {
	// Verify customer exists and get name for denormalized field (only if CustomerID provided)
	var customerName string
	if req.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *req.CustomerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCustomerNotFound
			}
			return nil, fmt.Errorf("failed to verify customer: %w", err)
		}
		customerName = customer.Name
	}

	// Set default phase if not provided
	phase := req.Phase
	if phase == "" {
		phase = domain.ProjectPhaseTilbud
	}

	project := &domain.Project{
		Name:              req.Name,
		ProjectNumber:     req.ProjectNumber,
		Summary:           req.Summary,
		Description:       req.Description,
		CustomerID:        req.CustomerID,
		CustomerName:      customerName,
		Phase:             phase,
		EndDate:           req.EndDate,
		Location:          req.Location,
		DealID:            req.DealID,
		ExternalReference: req.ExternalReference,
	}

	// Set StartDate if provided
	if req.StartDate != nil {
		project.StartDate = *req.StartDate
	}

	// Set user tracking fields on creation
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.CreatedByID = userCtx.UserID.String()
		project.CreatedByName = userCtx.DisplayName
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Reload with customer relation
	project, err := s.projectRepo.GetByID(ctx, project.ID)
	if err != nil {
		s.logger.Warn("failed to reload project after create", zap.Error(err))
	}

	// Log activity with project number if available
	activityBody := fmt.Sprintf("Prosjektet '%s' ble opprettet", project.Name)
	if project.CustomerName != "" {
		activityBody = fmt.Sprintf("Prosjektet '%s' ble opprettet for kunde %s", project.Name, project.CustomerName)
	}
	if project.ProjectNumber != "" {
		activityBody = fmt.Sprintf("Prosjektet '%s' (%s) ble opprettet", project.Name, project.ProjectNumber)
		if project.CustomerName != "" {
			activityBody = fmt.Sprintf("Prosjektet '%s' (%s) ble opprettet for kunde %s", project.Name, project.ProjectNumber, project.CustomerName)
		}
	}
	s.logActivity(ctx, project.ID, project.Name, "Prosjekt opprettet", activityBody)

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get offer count for this project
	offerCount := 0
	if s.offerRepo != nil {
		counts, err := s.offerRepo.CountOffersByProjectIDs(ctx, []uuid.UUID{id})
		if err != nil {
			s.logger.Warn("failed to get offer count for project", zap.Error(err), zap.String("project_id", id.String()))
		} else {
			offerCount = counts[id]
		}
	}

	dto := mapper.ToProjectDTOWithOfferCount(project, offerCount)
	return &dto, nil
}

// GetByIDWithRelations retrieves a project with all related data
func (s *ProjectService) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*domain.ProjectDTO, []domain.BudgetItemDTO, []domain.ActivityDTO, error) {
	project, budgetItems, activities, err := s.projectRepo.GetByIDWithRelations(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, ErrProjectNotFound
		}
		return nil, nil, nil, fmt.Errorf("failed to get project with relations: %w", err)
	}

	// Get offer count for this project
	offerCount := 0
	if s.offerRepo != nil {
		counts, err := s.offerRepo.CountOffersByProjectIDs(ctx, []uuid.UUID{id})
		if err != nil {
			s.logger.Warn("failed to get offer count for project", zap.Error(err), zap.String("project_id", id.String()))
		} else {
			offerCount = counts[id]
		}
	}

	projectDTO := mapper.ToProjectDTOWithOfferCount(project, offerCount)

	itemDTOs := make([]domain.BudgetItemDTO, len(budgetItems))
	for i, item := range budgetItems {
		itemDTOs[i] = mapper.ToBudgetItemDTO(&item)
	}

	activityDTOs := make([]domain.ActivityDTO, len(activities))
	for i, act := range activities {
		activityDTOs[i] = mapper.ToActivityDTO(&act)
	}

	return &projectDTO, itemDTOs, activityDTOs, nil
}

// GetByIDWithDetails retrieves a project with full details
func (s *ProjectService) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.ProjectWithDetailsDTO, error) {
	project, _, activities, err := s.projectRepo.GetByIDWithRelations(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project with relations: %w", err)
	}

	// Get offer count for this project
	offerCount := 0
	if s.offerRepo != nil {
		counts, err := s.offerRepo.CountOffersByProjectIDs(ctx, []uuid.UUID{id})
		if err != nil {
			s.logger.Warn("failed to get offer count for project", zap.Error(err), zap.String("project_id", id.String()))
		} else {
			offerCount = counts[id]
		}
	}

	// Note: Offers link to projects (project_id on offer), not the other way around.
	// For backward compatibility, we pass nil for offer. Use GetProjectOffers to get linked offers.
	dto := mapper.ToProjectWithDetailsDTOWithOfferCount(project, nil, activities, nil, project.Deal, offerCount)
	return &dto, nil
}

// Update updates an existing project with permission check
func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateProjectRequest) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Track changes for activity logging
	changes := s.trackChanges(project, req)

	// Track if name changed for denormalized field update
	oldName := project.Name

	// Update fields
	project.Name = req.Name
	project.ProjectNumber = req.ProjectNumber
	project.Summary = req.Summary
	project.Description = req.Description
	project.Location = req.Location
	project.ExternalReference = req.ExternalReference

	// Update customer if provided
	if req.CustomerID != nil {
		if project.CustomerID == nil || *project.CustomerID != *req.CustomerID {
			// Verify new customer exists
			customer, err := s.customerRepo.GetByID(ctx, *req.CustomerID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrCustomerNotFound
				}
				return nil, fmt.Errorf("failed to verify customer: %w", err)
			}
			project.CustomerID = req.CustomerID
			project.CustomerName = customer.Name
		}
	} else {
		// Allow clearing customer
		project.CustomerID = nil
		project.CustomerName = ""
	}

	if req.StartDate != nil {
		project.StartDate = *req.StartDate
	}
	project.EndDate = req.EndDate

	// Optional fields
	if req.DealID != nil {
		project.DealID = req.DealID
	}

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Update project_name on linked offers if name changed
	if oldName != project.Name && s.offerRepo != nil {
		if err := s.offerRepo.UpdateProjectNameByProjectID(ctx, id, project.Name); err != nil {
			s.logger.Warn("failed to update project_name on linked offers",
				zap.String("project_id", id.String()),
				zap.Error(err))
		}
	}

	// Reload project
	project, err = s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload project after update", zap.String("project_id", id.String()), zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Prosjektet '%s' ble oppdatert", project.Name)
	if changes != "" {
		activityBody = fmt.Sprintf("Prosjektet '%s' ble oppdatert: %s", project.Name, changes)
	}
	s.logActivity(ctx, project.ID, project.Name, "Prosjekt oppdatert", activityBody)

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// Delete removes a project
func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	projectName := project.Name
	customerID := project.CustomerID
	customerName := project.CustomerName

	// Delete files associated with the project
	// This ensures files are removed from both storage and database
	if s.fileService != nil {
		files, err := s.fileService.ListByProject(ctx, id)
		if err != nil {
			s.logger.Warn("failed to list files for project deletion",
				zap.String("projectID", id.String()),
				zap.Error(err))
		} else {
			for _, file := range files {
				if err := s.fileService.Delete(ctx, file.ID); err != nil {
					s.logger.Warn("failed to delete file during project deletion",
						zap.String("fileID", file.ID.String()),
						zap.String("projectID", id.String()),
						zap.Error(err))
				}
			}
		}
	}

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Log activity on customer if exists, since project is deleted
	if customerID != nil {
		s.logActivityOnTarget(ctx, domain.ActivityTargetCustomer, *customerID, customerName,
			"Prosjekt slettet", fmt.Sprintf("Prosjektet '%s' ble slettet", projectName))
	}

	return nil
}

// List returns a paginated list of projects with optional filters and default sorting
func (s *ProjectService) List(ctx context.Context, page, pageSize int, customerID *uuid.UUID, phase *domain.ProjectPhase) (*domain.PaginatedResponse, error) {
	filters := &repository.ProjectFilters{
		CustomerID: customerID,
		Phase:      phase,
	}
	return s.ListWithSort(ctx, page, pageSize, filters, repository.DefaultSortConfig())
}

// ListWithFilters returns a paginated list of projects with filter options and default sorting
func (s *ProjectService) ListWithFilters(ctx context.Context, page, pageSize int, filters *repository.ProjectFilters) (*domain.PaginatedResponse, error) {
	return s.ListWithSort(ctx, page, pageSize, filters, repository.DefaultSortConfig())
}

// ListWithSort returns a paginated list of projects with filter and sort options
func (s *ProjectService) ListWithSort(ctx context.Context, page, pageSize int, filters *repository.ProjectFilters, sort repository.SortConfig) (*domain.PaginatedResponse, error) {
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

	projects, total, err := s.projectRepo.ListWithFilters(ctx, page, pageSize, filters, sort)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Get offer counts for all projects in the list
	offerCounts := make(map[uuid.UUID]int)
	if len(projects) > 0 && s.offerRepo != nil {
		projectIDs := make([]uuid.UUID, len(projects))
		for i, p := range projects {
			projectIDs[i] = p.ID
		}

		counts, err := s.offerRepo.CountOffersByProjectIDs(ctx, projectIDs)
		if err != nil {
			s.logger.Warn("failed to get offer counts for projects", zap.Error(err))
			// Continue without offer counts - not a critical error
		} else {
			offerCounts = counts
		}
	}

	dtos := make([]domain.ProjectDTO, len(projects))
	for i, project := range projects {
		offerCount := offerCounts[project.ID]
		dtos[i] = mapper.ToProjectDTOWithOfferCount(&project, offerCount)
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
// Phase Transition Methods
// ============================================================================

// UpdatePhase updates only the project phase
// Validates phase transitions and enforces business rules:
// - Working phase requires StartDate
func (s *ProjectService) UpdatePhase(ctx context.Context, id uuid.UUID, phase domain.ProjectPhase) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Validate phase transition
	if !project.Phase.CanTransitionTo(phase) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s",
			ErrInvalidPhaseTransition, project.Phase, phase)
	}

	// Working phase requires StartDate
	if phase == domain.ProjectPhaseWorking && project.StartDate.IsZero() {
		return nil, ErrWorkingPhaseRequiresStartDate
	}

	oldPhase := project.Phase
	project.Phase = phase

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project phase: %w", err)
	}

	s.logActivity(ctx, project.ID, project.Name, "Prosjektfase oppdatert", fmt.Sprintf("Prosjektfase endret fra '%s' til '%s'", oldPhase, phase))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// ReopenProject reopens a completed or cancelled project
// Business rules:
// - Completed projects can be reopened to working
// - Cancelled projects can be reopened to tilbud or working
// - Working phase requires StartDate (auto-set to now if not present)
func (s *ProjectService) ReopenProject(ctx context.Context, id uuid.UUID, req *domain.ReopenProjectRequest) (*domain.ReopenProjectResponse, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Validate source phase - only completed or cancelled can be reopened
	if !project.Phase.IsClosedPhase() {
		return nil, fmt.Errorf("%w: project is in %s phase, only completed or cancelled projects can be reopened",
			ErrCannotReopenProject, project.Phase)
	}

	// Validate target phase transition
	if !project.Phase.CanTransitionTo(req.TargetPhase) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s",
			ErrInvalidPhaseTransition, project.Phase, req.TargetPhase)
	}

	previousPhase := project.Phase
	response := &domain.ReopenProjectResponse{
		PreviousPhase: string(previousPhase),
	}

	// Set StartDate if going to working and not already set
	if req.TargetPhase == domain.ProjectPhaseWorking && project.StartDate.IsZero() {
		project.StartDate = time.Now()
		s.logger.Info("auto-set StartDate for working phase",
			zap.String("projectID", id.String()),
			zap.Time("startDate", project.StartDate))
	}

	// Update phase
	project.Phase = req.TargetPhase

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	// Save project
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Reload project
	project, err = s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload project after reopen", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Prosjekt gjenåpnet fra '%s' til '%s'", previousPhase, req.TargetPhase)
	if req.Notes != "" {
		activityBody = fmt.Sprintf("%s. Notater: %s", activityBody, req.Notes)
	}
	s.logActivity(ctx, project.ID, project.Name, "Prosjekt gjenåpnet", activityBody)

	// Build response
	projectDTO := mapper.ToProjectDTO(project)
	response.Project = &projectDTO

	return response, nil
}

// ============================================================================
// Activity Methods
// ============================================================================

// GetActivities returns activities for a project
func (s *ProjectService) GetActivities(ctx context.Context, id uuid.UUID, limit int) ([]domain.ActivityDTO, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

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

// ============================================================================
// Individual Property Update Methods
// ============================================================================

// UpdateName updates only the project name
func (s *ProjectService) UpdateName(ctx context.Context, id uuid.UUID, name string) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	oldName := project.Name
	project.Name = name

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project name: %w", err)
	}

	// Update project_name on linked offers
	if oldName != name && s.offerRepo != nil {
		if err := s.offerRepo.UpdateProjectNameByProjectID(ctx, id, name); err != nil {
			s.logger.Warn("failed to update project_name on linked offers",
				zap.String("project_id", id.String()),
				zap.Error(err))
		}
	}

	s.logActivity(ctx, project.ID, project.Name, "Prosjektnavn oppdatert", fmt.Sprintf("Prosjektnavn endret fra '%s' til '%s'", oldName, name))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateDescription updates only the project description and summary
func (s *ProjectService) UpdateDescription(ctx context.Context, id uuid.UUID, summary, description string) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	project.Summary = summary
	project.Description = description

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project description: %w", err)
	}

	s.logActivity(ctx, project.ID, project.Name, "Prosjektbeskrivelse oppdatert", "Prosjektbeskrivelsen ble oppdatert")

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateDates updates only the project start and end dates
func (s *ProjectService) UpdateDates(ctx context.Context, id uuid.UUID, startDate, endDate *time.Time) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if startDate != nil {
		project.StartDate = *startDate
	}
	// EndDate is a pointer in the model, so we can assign directly
	project.EndDate = endDate

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project dates: %w", err)
	}

	s.logActivity(ctx, project.ID, project.Name, "Prosjektdatoer oppdatert", "Prosjektets start-/sluttdatoer ble oppdatert")

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateProjectNumber updates only the project number
func (s *ProjectService) UpdateProjectNumber(ctx context.Context, id uuid.UUID, projectNumber string) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	oldNumber := project.ProjectNumber
	project.ProjectNumber = projectNumber

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project number: %w", err)
	}

	s.logActivity(ctx, project.ID, project.Name, "Prosjektnummer oppdatert", fmt.Sprintf("Prosjektnummer endret fra '%s' til '%s'", oldNumber, projectNumber))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// trackChanges creates a summary of changes between the project and update request
func (s *ProjectService) trackChanges(project *domain.Project, req *domain.UpdateProjectRequest) string {
	var changes []string

	if project.Name != req.Name {
		changes = append(changes, fmt.Sprintf("name: '%s' -> '%s'", project.Name, req.Name))
	}
	if project.ProjectNumber != req.ProjectNumber {
		changes = append(changes, fmt.Sprintf("projectNumber: '%s' -> '%s'", project.ProjectNumber, req.ProjectNumber))
	}

	if len(changes) == 0 {
		return ""
	}
	if len(changes) > 3 {
		return fmt.Sprintf("%d fields updated", len(changes))
	}
	result := ""
	for i, c := range changes {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result
}

// logActivity creates an activity log entry for a project
func (s *ProjectService) logActivity(ctx context.Context, projectID uuid.UUID, projectName, title, body string) {
	s.logActivityOnTarget(ctx, domain.ActivityTargetProject, projectID, projectName, title, body)
}

// logActivityOnTarget creates an activity log entry for any target
func (s *ProjectService) logActivityOnTarget(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, targetName, title, body string) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		s.logger.Warn("no user context for activity logging")
		return
	}

	activity := &domain.Activity{
		TargetType:  targetType,
		TargetID:    targetID,
		TargetName:  targetName,
		Title:       title,
		Body:        body,
		OccurredAt:  time.Now(),
		CreatorName: userCtx.DisplayName,
		CreatorID:   userCtx.UserID.String(),
		CompanyID:   &userCtx.CompanyID,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to log activity", zap.Error(err))
	}
}
