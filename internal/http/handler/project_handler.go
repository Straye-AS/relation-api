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
// @Param phase query string false "Filter by phase" Enums(tilbud, working, on_hold, completed, cancelled)
// @Param sortBy query string false "Sort field" Enums(createdAt, updatedAt, name, phase, startDate, endDate, customerName)
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
// @Description Create a new project (simplified container for offers). Only name, description, startDate, and endDate can be set on creation. Phase defaults to "tilbud". Location and customer are inferred from linked offers.
// @Tags Projects
// @Accept json
// @Produce json
// @Param request body domain.CreateProjectRequest true "Project data (name required, description/startDate/endDate optional)"
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
// @Description Get a project with details including recent activities
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
// @Description Update an existing project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectRequest true "Project data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError
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
// @Description Delete a project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} domain.APIError
// @Failure 401 {object} domain.APIError
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

// handleProjectError maps service errors to HTTP status codes
func (h *ProjectHandler) handleProjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrProjectNotFound):
		respondWithError(w, http.StatusNotFound, "Project not found")
	case errors.Is(err, service.ErrCustomerNotFound):
		respondWithError(w, http.StatusBadRequest, "Customer not found")
	case errors.Is(err, service.ErrOfferNotFound):
		respondWithError(w, http.StatusNotFound, "Offer not found")
	case errors.Is(err, service.ErrInvalidPhaseTransition):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrCannotReopenProject):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrWorkingPhaseRequiresStartDate):
		respondWithError(w, http.StatusBadRequest, "Working phase requires a start date")
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

// ReopenProject godoc
// @Summary Reopen a completed or cancelled project
// @Description Reopens a project that was completed or cancelled. Can reopen to tilbud or working phase.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.ReopenProjectRequest true "Reopen configuration"
// @Success 200 {object} domain.ReopenProjectResponse
// @Failure 400 {object} domain.APIError "Invalid request, phase transition, or project not in closed state"
// @Failure 401 {object} domain.APIError
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
