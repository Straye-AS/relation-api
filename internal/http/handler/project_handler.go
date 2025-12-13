package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

type ProjectHandler struct {
	projectService *service.ProjectService
	offerService   *service.OfferService
	logger         *zap.Logger
}

func NewProjectHandler(projectService *service.ProjectService, offerService *service.OfferService, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		offerService:   offerService,
		logger:         logger,
	}
}

// List godoc
// @Summary List projects
// @Description Get paginated list of projects with optional filters
// @Tags Projects
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Param customerId query string false "Filter by customer ID" format(uuid)
// @Param phase query string false "Filter by phase" Enums(tilbud, working, active, completed, cancelled)
// @Param health query string false "Filter by health" Enums(on_track, at_risk, over_budget)
// @Param managerId query string false "Filter by manager ID"
// @Param sortBy query string false "Sort field" Enums(createdAt, updatedAt, name, phase, health, budget, spent, startDate, endDate, customerName, wonAt)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.ProjectDTO}
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects [get]
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	// Build filters
	filters := &repository.ProjectFilters{}

	// Parse customer ID filter
	if cid := r.URL.Query().Get("customerId"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			filters.CustomerID = &id
		}
	}

	// Parse phase filter
	if p := r.URL.Query().Get("phase"); p != "" {
		ph := domain.ProjectPhase(p)
		filters.Phase = &ph
	}

	// Parse health filter
	if healthStr := r.URL.Query().Get("health"); healthStr != "" {
		health := domain.ProjectHealth(healthStr)
		filters.Health = &health
	}

	// Parse manager ID filter
	if mid := r.URL.Query().Get("managerId"); mid != "" {
		filters.ManagerID = &mid
	}

	// Parse sort configuration
	sort := repository.DefaultSortConfig()
	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		sort.Field = sortBy
	}
	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		sort.Order = repository.ParseSortOrder(sortOrder)
	}

	result, err := h.projectService.ListWithSort(r.Context(), page, pageSize, filters, sort)
	if err != nil {
		h.logger.Error("failed to list projects", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to list projects")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary Create project
// @Description Create a new project. Requires manager or admin permissions.
// @Tags Projects
// @Accept json
// @Produce json
// @Param request body domain.CreateProjectRequest true "Project data"
// @Success 201 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects [post]
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.Create(r.Context(), &req)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	w.Header().Set("Location", "/api/v1/projects/"+project.ID.String())
	respondJSON(w, http.StatusCreated, project)
}

// GetByID godoc
// @Summary Get project by ID
// @Description Get a project with full details including budget summary and recent activities
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Success 200 {object} domain.ProjectWithDetailsDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	project, err := h.projectService.GetByIDWithDetails(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get project", zap.Error(err), zap.String("project_id", id.String()))
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// Update godoc
// @Summary Update project
// @Description Update an existing project. Requires manager or admin permissions.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectRequest true "Project data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError "User is not the project manager"
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id} [put]
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.Update(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update project", zap.Error(err), zap.String("project_id", id.String()))
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// Delete godoc
// @Summary Delete project
// @Description Delete a project. Requires manager or admin permissions.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError "User is not the project manager"
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id} [delete]
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	if err := h.projectService.Delete(r.Context(), id); err != nil {
		h.logger.Error("failed to delete project", zap.Error(err), zap.String("project_id", id.String()))
		h.handleProjectError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateHealth godoc
// @Summary Update project health
// @Description Update project health and completion percent. Requires manager or admin permissions.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectHealthRequest true "Health update data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError "Invalid request"
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError "User is not the project manager"
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/health [put]
func (h *ProjectHandler) UpdateHealth(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectHealthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateHealthAndCompletion(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update project health", zap.Error(err), zap.String("project_id", id.String()))
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// GetBudget godoc
// @Summary Get project budget
// @Description Get budget information for a project including spent amount and remaining budget
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Success 200 {object} domain.ProjectBudgetDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/budget [get]
func (h *ProjectHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	budget, err := h.projectService.GetBudget(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get project budget", zap.Error(err), zap.String("project_id", id.String()))
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, budget)
}

// GetActivities godoc
// @Summary Get project activities
// @Description Get recent activities for a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param limit query int false "Limit" default(50) maximum(200)
// @Success 200 {array} domain.ActivityDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/activities [get]
func (h *ProjectHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	activities, err := h.projectService.GetActivities(r.Context(), id, limit)
	if err != nil {
		h.logger.Error("failed to get project activities", zap.Error(err), zap.String("project_id", id.String()))
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, activities)
}

// InheritBudget godoc
// @Summary Inherit budget from offer
// @Description Inherit budget dimensions from a won offer to the project. The offer must be in 'won' phase.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.InheritBudgetRequest true "Offer ID to inherit budget from"
// @Success 200 {object} domain.InheritBudgetResponse
// @Failure 400 {object} domain.APIError "Invalid request or offer not in won phase"
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError "User is not the project manager"
// @Failure 404 {object} domain.APIError "Project or offer not found"
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/inherit-budget [post]
func (h *ProjectHandler) InheritBudget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.InheritBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	result, err := h.projectService.InheritBudgetFromOffer(r.Context(), id, req.OfferID)
	if err != nil {
		h.logger.Error("failed to inherit budget from offer",
			zap.Error(err),
			zap.String("project_id", id.String()),
			zap.String("offer_id", req.OfferID.String()))
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// handleProjectError maps service errors to HTTP status codes
func (h *ProjectHandler) handleProjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrProjectNotFound):
		respondWithError(w, http.StatusNotFound, "Project not found")
	case errors.Is(err, service.ErrCustomerNotFound):
		respondWithError(w, http.StatusBadRequest, "Customer not found")
	case errors.Is(err, service.ErrOfferNotFound):
		respondWithError(w, http.StatusNotFound, "Offer not found")
	case errors.Is(err, service.ErrOfferNotWon):
		respondWithError(w, http.StatusBadRequest, "Can only inherit budget from won offers")
	case errors.Is(err, service.ErrProjectNotManager):
		respondWithError(w, http.StatusForbidden, "User is not the project manager")
	case errors.Is(err, service.ErrInvalidStatusTransition):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidPhaseTransition):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidCompletionPercent):
		respondWithError(w, http.StatusBadRequest, "Completion percent must be between 0 and 100")
	case errors.Is(err, service.ErrProjectEconomicsNotEditable):
		respondWithError(w, http.StatusBadRequest, "Project economics can only be edited during active phase")
	case errors.Is(err, service.ErrCannotReopenProject):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrWorkingPhaseRequiresStartDate):
		respondWithError(w, http.StatusBadRequest, "Working phase requires a start date")
	case errors.Is(err, service.ErrTilbudPhaseCannotHaveWinningOffer):
		respondWithError(w, http.StatusBadRequest, "Project in tilbud phase cannot have a winning offer")
	case errors.Is(err, service.ErrUnauthorized):
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	default:
		h.logger.Error("project handler error", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// GetProjectOffers godoc
// @Summary Get offers for a project
// @Description Get all offers linked to a project (offer folder model)
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Success 200 {array} domain.OfferDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError "Project not found"
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/offers [get]
func (h *ProjectHandler) GetProjectOffers(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	offers, err := h.offerService.GetProjectOffers(r.Context(), id)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offers)
}

// ============================================================================
// Individual Property Update Handlers
// ============================================================================

// UpdateName godoc
// @Summary Update project name
// @Description Update only the name of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectNameRequest true "Name data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/name [put]
func (h *ProjectHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateName(r.Context(), id, req.Name)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateDescription godoc
// @Summary Update project description
// @Description Update the summary and description of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectDescriptionRequest true "Description data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/description [put]
func (h *ProjectHandler) UpdateDescription(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateDescription(r.Context(), id, req.Summary, req.Description)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdatePhase godoc
// @Summary Update project phase
// @Description Update only the phase of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectPhaseRequest true "Phase data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/phase [put]
func (h *ProjectHandler) UpdatePhase(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectPhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdatePhase(r.Context(), id, req.Phase)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateManager godoc
// @Summary Update project manager
// @Description Update only the manager of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectManagerRequest true "Manager data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/manager [put]
func (h *ProjectHandler) UpdateManager(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectManagerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateManager(r.Context(), id, req.ManagerID)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateDates godoc
// @Summary Update project dates
// @Description Update the start and end dates of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectDatesRequest true "Dates data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/dates [put]
func (h *ProjectHandler) UpdateDates(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectDatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	project, err := h.projectService.UpdateDates(r.Context(), id, req.StartDate, req.EndDate)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateBudget godoc
// @Summary Update project budget
// @Description Update only the budget of a project (only allowed in active phase)
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectBudgetRequest true "Budget data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError "Cannot edit budget in tilbud phase"
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/budget [put]
func (h *ProjectHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateBudget(r.Context(), id, req.Budget)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateSpent godoc
// @Summary Update project spent amount
// @Description Update only the spent amount of a project (only allowed in active phase)
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectSpentRequest true "Spent data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError "Cannot edit spent in tilbud phase"
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/spent [put]
func (h *ProjectHandler) UpdateSpent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectSpentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateSpent(r.Context(), id, req.Spent)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateTeamMembers godoc
// @Summary Update project team members
// @Description Update the team members of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectTeamMembersRequest true "Team members data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/team-members [put]
func (h *ProjectHandler) UpdateTeamMembers(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectTeamMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	project, err := h.projectService.UpdateTeamMembers(r.Context(), id, req.TeamMembers)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateCompletionPercent godoc
// @Summary Update project completion percentage
// @Description Update the completion percentage of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectCompletionPercentRequest true "Completion percent data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/completion [put]
func (h *ProjectHandler) UpdateCompletionPercent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectCompletionPercentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateCompletionPercent(r.Context(), id, req.CompletionPercent)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateEstimatedCompletionDate godoc
// @Summary Update project estimated completion date
// @Description Update the estimated completion date of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectEstimatedCompletionDateRequest true "Estimated completion date data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/estimated-completion-date [put]
func (h *ProjectHandler) UpdateEstimatedCompletionDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectEstimatedCompletionDateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	project, err := h.projectService.UpdateEstimatedCompletionDate(r.Context(), id, req.EstimatedCompletionDate)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateProjectNumber godoc
// @Summary Update project number
// @Description Update the project number of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectNumberRequest true "Project number data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/project-number [put]
func (h *ProjectHandler) UpdateProjectNumber(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectNumberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateProjectNumber(r.Context(), id, req.ProjectNumber)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateCompany godoc
// @Summary Update project company
// @Description Update the company assignment of a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectCompanyRequest true "Company data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/company [put]
func (h *ProjectHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateCompanyID(r.Context(), id, req.CompanyID)
	if err != nil {
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// ResyncFromBestOffer godoc
// @Summary Resync project from best offer
// @Description Syncs project economics (value, cost, margin) from the best connected offer
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} domain.ResyncFromOfferResponse "Updated project with synced values"
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or no offers found"
// @Failure 404 {object} domain.ErrorResponse "Project not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/resync-from-offer [post]
func (h *ProjectHandler) ResyncFromBestOffer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	project, offer, err := h.projectService.ResyncFromBestOffer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to resync project from offer", zap.Error(err), zap.String("project_id", id.String()))
		h.handleProjectError(w, err)
		return
	}

	response := domain.ResyncFromOfferResponse{
		Project:     project,
		OfferID:     offer.ID,
		OfferTitle:  offer.Title,
		OfferPhase:  string(offer.Phase),
		SyncedValue: project.Value,
		SyncedCost:  project.Cost,
	}

	respondJSON(w, http.StatusOK, response)
}

// ReopenProject godoc
// @Summary Reopen a completed or cancelled project
// @Description Reopens a project that was completed or cancelled. Can reopen to tilbud or working phase.
// @Description When reopening a project with a winning offer, that offer is reverted to sent phase.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.ReopenProjectRequest true "Reopen configuration"
// @Success 200 {object} domain.ReopenProjectResponse
// @Failure 400 {object} domain.APIError "Invalid request, phase transition, or project not in closed state"
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError "User is not the project manager"
// @Failure 404 {object} domain.APIError "Project not found"
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/reopen [post]
func (h *ProjectHandler) ReopenProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.ReopenProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	result, err := h.projectService.ReopenProject(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to reopen project", zap.Error(err), zap.String("project_id", id.String()))
		h.handleProjectError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}
