package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

// @Summary Upload file
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param offerId formData string false "Offer ID to attach file to"
// @Success 201 {object} domain.FileDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /files/upload [post]
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

// @Summary Get file metadata
// @Tags Files
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} domain.FileDTO
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

// @Summary Download file
// @Tags Files
// @Produce application/octet-stream
// @Param id path string true "File ID"
// @Success 200
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /files/{id}/download [get]
func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid file ID: must be a valid UUID")
		return
	}

	reader, filename, err := h.fileService.Download(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to download file", zap.Error(err), zap.String("file_id", id.String()))
		respondWithError(w, http.StatusNotFound, "File not found")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")

	_, _ = io.Copy(w, reader)
}
