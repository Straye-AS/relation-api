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
	logger         *zap.Logger
}

func NewProjectHandler(projectService *service.ProjectService, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
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
// @Param status query string false "Filter by status" Enums(planning, active, on_hold, completed, cancelled)
// @Param health query string false "Filter by health" Enums(on_track, at_risk, over_budget)
// @Param managerId query string false "Filter by manager ID"
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

	// Parse status filter
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.ProjectStatus(s)
		filters.Status = &st
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

	result, err := h.projectService.ListWithFilters(r.Context(), page, pageSize, filters)
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

// UpdateStatus godoc
// @Summary Update project status
// @Description Update project status with optional health override. Requires manager or admin permissions.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param request body domain.UpdateProjectStatusRequest true "Status update data"
// @Success 200 {object} domain.ProjectDTO
// @Failure 400 {object} domain.APIError "Invalid status transition"
// @Failure 401 {object} domain.APIError
// @Failure 403 {object} domain.APIError "User is not the project manager"
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/status [put]
func (h *ProjectHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	var req domain.UpdateProjectStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.UpdateStatusAndHealth(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update project status", zap.Error(err), zap.String("project_id", id.String()))
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

// handleProjectError maps service errors to HTTP status codes
func (h *ProjectHandler) handleProjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrProjectNotFound):
		respondWithError(w, http.StatusNotFound, "Project not found")
	case errors.Is(err, service.ErrCustomerNotFound):
		respondWithError(w, http.StatusBadRequest, "Customer not found")
	case errors.Is(err, service.ErrProjectNotManager):
		respondWithError(w, http.StatusForbidden, "User is not the project manager")
	case errors.Is(err, service.ErrInvalidStatusTransition):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidCompletionPercent):
		respondWithError(w, http.StatusBadRequest, "Completion percent must be between 0 and 100")
	case errors.Is(err, service.ErrUnauthorized):
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	default:
		h.logger.Error("project handler error", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}
