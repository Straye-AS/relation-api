package handler

import (
	"net/http"

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

// @Summary Get all companies
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
