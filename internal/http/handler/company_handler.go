package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

type CompanyHandler struct {
	companyService *service.CompanyService
	logger         *zap.Logger
}

func NewCompanyHandler(companyService *service.CompanyService, logger *zap.Logger) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
		logger:         logger,
	}
}

// List godoc
// @Summary Get all companies
// @Description Returns a list of all active Straye group companies
// @Tags Companies
// @Produce json
// @Success 200 {array} domain.Company
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /companies [get]
func (h *CompanyHandler) List(w http.ResponseWriter, r *http.Request) {
	companies := h.companyService.List(r.Context())
	respondJSON(w, http.StatusOK, companies)
}

// GetByID godoc
// @Summary Get company by ID
// @Description Returns detailed information about a specific company including default responsible user settings
// @Tags Companies
// @Produce json
// @Param id path string true "Company ID (e.g., gruppen, stalbygg, hybridbygg, industri, tak, montasje)"
// @Success 200 {object} domain.CompanyDetailDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /companies/{id} [get]
func (h *CompanyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		respondWithError(w, http.StatusBadRequest, "company ID is required")
		return
	}

	company, err := h.companyService.GetByIDDetailed(r.Context(), domain.CompanyID(companyID))
	if err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			respondWithError(w, http.StatusNotFound, "company not found")
			return
		}
		h.logger.Error("failed to get company", zap.Error(err), zap.String("companyID", companyID))
		respondWithError(w, http.StatusInternalServerError, "failed to get company")
		return
	}

	respondJSON(w, http.StatusOK, company)
}

// Update godoc
// @Summary Update company settings
// @Description Updates company settings including default responsible users for offers and projects
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID (e.g., gruppen, stalbygg, hybridbygg, industri, tak, montasje)"
// @Param body body domain.UpdateCompanyRequest true "Company update request"
// @Success 200 {object} domain.CompanyDetailDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /companies/{id} [put]
func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		respondWithError(w, http.StatusBadRequest, "company ID is required")
		return
	}

	var req domain.UpdateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	company, err := h.companyService.Update(r.Context(), domain.CompanyID(companyID), &req)
	if err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			respondWithError(w, http.StatusNotFound, "company not found")
			return
		}
		if errors.Is(err, service.ErrInvalidResponsibleUser) {
			respondWithError(w, http.StatusBadRequest, "invalid responsible user ID")
			return
		}
		h.logger.Error("failed to update company", zap.Error(err), zap.String("companyID", companyID))
		respondWithError(w, http.StatusInternalServerError, "failed to update company")
		return
	}

	respondJSON(w, http.StatusOK, company)
}
