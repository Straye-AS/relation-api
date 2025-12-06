package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
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

// @Summary List projects
// @Tags Projects
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param customerId query string false "Filter by customer ID"
// @Param status query string false "Filter by status"
// @Success 200 {object} domain.PaginatedResponse
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

	var customerID *uuid.UUID
	if cid := r.URL.Query().Get("customerId"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			customerID = &id
		}
	}

	var status *domain.ProjectStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.ProjectStatus(s)
		status = &st
	}

	result, err := h.projectService.List(r.Context(), page, pageSize, customerID, status)
	if err != nil {
		h.logger.Error("failed to list projects", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// @Summary Create project
// @Tags Projects
// @Accept json
// @Produce json
// @Param request body domain.CreateProjectRequest true "Project data"
// @Success 201 {object} domain.ProjectDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects [post]
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create project", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/projects/"+project.ID.String())
	respondJSON(w, http.StatusCreated, project)
}

// @Summary Get project
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} domain.ProjectDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.projectService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// @Summary Update project
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body domain.UpdateProjectRequest true "Project data"
// @Success 200 {object} domain.ProjectDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id} [put]
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req domain.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	project, err := h.projectService.Update(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update project", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// @Summary Get project budget
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} domain.ProjectBudgetDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/budget [get]
func (h *ProjectHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	budget, err := h.projectService.GetBudget(r.Context(), id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, budget)
}

// @Summary Get project activities
// @Tags Projects
// @Produce json
// @Param id path string true "Project ID"
// @Param limit query int false "Limit" default(50)
// @Success 200 {array} domain.ActivityDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/activities [get]
func (h *ProjectHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}

	activities, err := h.projectService.GetActivities(r.Context(), id, limit)
	if err != nil {
		h.logger.Error("failed to get activities", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, activities)
}
