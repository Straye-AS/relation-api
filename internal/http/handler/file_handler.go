package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

type FileHandler struct {
	fileService *service.FileService
	maxUploadMB int64
	logger      *zap.Logger
}

func NewFileHandler(fileService *service.FileService, maxUploadMB int64, logger *zap.Logger) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		maxUploadMB: maxUploadMB,
		logger:      logger,
	}
}

// ============================================================================
// Entity-Specific Upload Handlers
// ============================================================================

// UploadToCustomer godoc
// @Summary Upload file to customer
// @Description Upload a file and attach it to a customer
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param file formData file true "File to upload"
// @Success 201 {object} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 413 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/files [post]
func (h *FileHandler) UploadToCustomer(w http.ResponseWriter, r *http.Request) {
	customerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid customer ID: must be a valid UUID")
		return
	}

	fileDTO, err := h.handleFileUpload(r, func(filename, contentType string, file io.Reader) (*domain.FileDTO, error) {
		return h.fileService.UploadToCustomer(r.Context(), customerID, filename, contentType, file)
	})
	if err != nil {
		h.handleUploadError(w, err, "customer")
		return
	}

	respondJSON(w, http.StatusCreated, fileDTO)
}

// UploadToProject godoc
// @Summary Upload file to project
// @Description Upload a file and attach it to a project
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Param file formData file true "File to upload"
// @Success 201 {object} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 413 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/files [post]
func (h *FileHandler) UploadToProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	fileDTO, err := h.handleFileUpload(r, func(filename, contentType string, file io.Reader) (*domain.FileDTO, error) {
		return h.fileService.UploadToProject(r.Context(), projectID, filename, contentType, file)
	})
	if err != nil {
		h.handleUploadError(w, err, "project")
		return
	}

	respondJSON(w, http.StatusCreated, fileDTO)
}

// UploadToOffer godoc
// @Summary Upload file to offer
// @Description Upload a file and attach it to an offer
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Offer ID" format(uuid)
// @Param file formData file true "File to upload"
// @Success 201 {object} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 413 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/files [post]
func (h *FileHandler) UploadToOffer(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	fileDTO, err := h.handleFileUpload(r, func(filename, contentType string, file io.Reader) (*domain.FileDTO, error) {
		return h.fileService.UploadToOffer(r.Context(), offerID, filename, contentType, file)
	})
	if err != nil {
		h.handleUploadError(w, err, "offer")
		return
	}

	respondJSON(w, http.StatusCreated, fileDTO)
}

// UploadToSupplier godoc
// @Summary Upload file to supplier
// @Description Upload a file and attach it to a supplier
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param file formData file true "File to upload"
// @Success 201 {object} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 413 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/files [post]
func (h *FileHandler) UploadToSupplier(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid supplier ID: must be a valid UUID")
		return
	}

	fileDTO, err := h.handleFileUpload(r, func(filename, contentType string, file io.Reader) (*domain.FileDTO, error) {
		return h.fileService.UploadToSupplier(r.Context(), supplierID, filename, contentType, file)
	})
	if err != nil {
		h.handleUploadError(w, err, "supplier")
		return
	}

	respondJSON(w, http.StatusCreated, fileDTO)
}

// ============================================================================
// Entity-Specific List Handlers
// ============================================================================

// ListCustomerFiles godoc
// @Summary List customer files
// @Description Get all files attached to a customer
// @Tags Files
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Success 200 {array} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/files [get]
func (h *FileHandler) ListCustomerFiles(w http.ResponseWriter, r *http.Request) {
	customerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid customer ID: must be a valid UUID")
		return
	}

	files, err := h.fileService.ListByCustomer(r.Context(), customerID)
	if err != nil {
		h.handleListError(w, err, "customer")
		return
	}

	respondJSON(w, http.StatusOK, files)
}

// ListProjectFiles godoc
// @Summary List project files
// @Description Get all files attached to a project
// @Tags Files
// @Produce json
// @Param id path string true "Project ID" format(uuid)
// @Success 200 {array} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /projects/{id}/files [get]
func (h *FileHandler) ListProjectFiles(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid project ID: must be a valid UUID")
		return
	}

	files, err := h.fileService.ListByProject(r.Context(), projectID)
	if err != nil {
		h.handleListError(w, err, "project")
		return
	}

	respondJSON(w, http.StatusOK, files)
}

// ListOfferFiles godoc
// @Summary List offer files
// @Description Get all files attached to an offer
// @Tags Files
// @Produce json
// @Param id path string true "Offer ID" format(uuid)
// @Success 200 {array} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/files [get]
func (h *FileHandler) ListOfferFiles(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	files, err := h.fileService.ListByOffer(r.Context(), offerID)
	if err != nil {
		h.handleListError(w, err, "offer")
		return
	}

	respondJSON(w, http.StatusOK, files)
}

// ListSupplierFiles godoc
// @Summary List supplier files
// @Description Get all files attached to a supplier
// @Tags Files
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Success 200 {array} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/files [get]
func (h *FileHandler) ListSupplierFiles(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid supplier ID: must be a valid UUID")
		return
	}

	files, err := h.fileService.ListBySupplier(r.Context(), supplierID)
	if err != nil {
		h.handleListError(w, err, "supplier")
		return
	}

	respondJSON(w, http.StatusOK, files)
}

// ============================================================================
// Generic File Operations (existing)
// ============================================================================

// Upload godoc
// @Summary Upload file (legacy)
// @Description Upload a file with optional offer attachment. Deprecated: Use entity-specific upload endpoints instead.
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param offerId formData string false "Offer ID to attach file to"
// @Success 201 {object} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 413 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /files/upload [post]
// @Deprecated
func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadMB*1024*1024)

	if err := r.ParseMultipartForm(h.maxUploadMB * 1024 * 1024); err != nil {
		respondWithError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("File too large: maximum size is %dMB", h.maxUploadMB))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid file upload: file field is required")
		return
	}
	defer file.Close()

	var offerID *uuid.UUID
	if oid := r.FormValue("offerId"); oid != "" {
		id, err := uuid.Parse(oid)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid offerId: must be a valid UUID")
			return
		}
		offerID = &id
	}

	fileDTO, err := h.fileService.Upload(r.Context(), header.Filename, header.Header.Get("Content-Type"), file, offerID)
	if err != nil {
		h.logger.Error("failed to upload file", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	respondJSON(w, http.StatusCreated, fileDTO)
}

// GetByID godoc
// @Summary Get file metadata
// @Description Get file metadata by ID
// @Tags Files
// @Produce json
// @Param id path string true "File ID" format(uuid)
// @Success 200 {object} domain.FileDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /files/{id} [get]
func (h *FileHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid file ID: must be a valid UUID")
		return
	}

	fileDTO, err := h.fileService.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get file", zap.Error(err), zap.String("file_id", id.String()))
		respondWithError(w, http.StatusNotFound, "File not found")
		return
	}

	respondJSON(w, http.StatusOK, fileDTO)
}

// Download godoc
// @Summary Download file
// @Description Download file content by ID
// @Tags Files
// @Produce application/octet-stream
// @Param id path string true "File ID" format(uuid)
// @Success 200
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /files/{id}/download [get]
func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid file ID: must be a valid UUID")
		return
	}

	reader, filename, contentType, err := h.fileService.Download(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to download file", zap.Error(err), zap.String("file_id", id.String()))
		respondWithError(w, http.StatusNotFound, "File not found")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	// Use actual content type from file, fallback to octet-stream if empty
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	_, _ = io.Copy(w, reader)
}

// Delete godoc
// @Summary Delete file
// @Description Delete a file from both storage and database
// @Tags Files
// @Produce json
// @Param id path string true "File ID" format(uuid)
// @Success 204
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /files/{id} [delete]
func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid file ID: must be a valid UUID")
		return
	}

	err = h.fileService.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "failed to get file") {
			respondWithError(w, http.StatusNotFound, "File not found")
			return
		}
		h.logger.Error("failed to delete file", zap.Error(err), zap.String("file_id", id.String()))
		respondWithError(w, http.StatusInternalServerError, "Failed to delete file")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Helper Methods
// ============================================================================

// handleFileUpload is a helper that handles common file upload logic
func (h *FileHandler) handleFileUpload(r *http.Request, uploadFn func(filename, contentType string, file io.Reader) (*domain.FileDTO, error)) (*domain.FileDTO, error) {
	// Limit request size
	r.Body = http.MaxBytesReader(nil, r.Body, h.maxUploadMB*1024*1024)

	if err := r.ParseMultipartForm(h.maxUploadMB * 1024 * 1024); err != nil {
		return nil, fmt.Errorf("file too large: maximum size is %dMB", h.maxUploadMB)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("invalid file upload: file field is required")
	}
	defer file.Close()

	return uploadFn(header.Filename, header.Header.Get("Content-Type"), file)
}

// handleUploadError handles common error cases for file uploads
func (h *FileHandler) handleUploadError(w http.ResponseWriter, err error, entityType string) {
	errMsg := err.Error()

	// Check for file too large error
	if strings.Contains(errMsg, "file too large") {
		respondWithError(w, http.StatusRequestEntityTooLarge, errMsg)
		return
	}

	// Check for invalid file upload
	if strings.Contains(errMsg, "invalid file upload") {
		respondWithError(w, http.StatusBadRequest, errMsg)
		return
	}

	// Check for entity not found
	if strings.Contains(errMsg, fmt.Sprintf("%s not found", entityType)) {
		respondWithError(w, http.StatusNotFound, capitalizeFirst(entityType)+" not found")
		return
	}

	// Generic error
	h.logger.Error("failed to upload file", zap.Error(err), zap.String("entity_type", entityType))
	respondWithError(w, http.StatusInternalServerError, "Failed to upload file")
}

// handleListError handles common error cases for listing files
func (h *FileHandler) handleListError(w http.ResponseWriter, err error, entityType string) {
	errMsg := err.Error()

	// Check for entity not found
	if strings.Contains(errMsg, fmt.Sprintf("%s not found", entityType)) {
		respondWithError(w, http.StatusNotFound, capitalizeFirst(entityType)+" not found")
		return
	}

	// Generic error
	h.logger.Error("failed to list files", zap.Error(err), zap.String("entity_type", entityType))
	respondWithError(w, http.StatusInternalServerError, "Failed to list files")
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
