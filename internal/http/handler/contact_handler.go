package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// ContactHandler handles HTTP requests for contacts
type ContactHandler struct {
	contactService *service.ContactService
	logger         *zap.Logger
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(contactService *service.ContactService, logger *zap.Logger) *ContactHandler {
	return &ContactHandler{
		contactService: contactService,
		logger:         logger,
	}
}

// ListContacts godoc
// @Summary List contacts
// @Description Get paginated list of contacts with optional filters
// @Tags Contacts
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page" default(20)
// @Param search query string false "Search by name or email"
// @Param title query string false "Filter by job title"
// @Param contactType query string false "Filter by contact type" Enums(primary, secondary, billing, technical, executive, other)
// @Param entityType query string false "Filter by related entity type" Enums(customer, deal, project)
// @Param entityId query string false "Filter by related entity ID"
// @Param sortBy query string false "Sort option" Enums(name_asc, name_desc, email_asc, created_desc)
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /contacts [get]
func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
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

	filters := &repository.ContactFilters{
		Search: r.URL.Query().Get("search"),
		Title:  r.URL.Query().Get("title"),
	}

	// Contact type filter
	if contactType := r.URL.Query().Get("contactType"); contactType != "" {
		ct := domain.ContactType(contactType)
		// Validate contact type
		validTypes := map[domain.ContactType]bool{
			domain.ContactTypePrimary:   true,
			domain.ContactTypeSecondary: true,
			domain.ContactTypeBilling:   true,
			domain.ContactTypeTechnical: true,
			domain.ContactTypeExecutive: true,
			domain.ContactTypeOther:     true,
		}
		if !validTypes[ct] {
			http.Error(w, "Invalid contactType. Must be one of: primary, secondary, billing, technical, executive, other", http.StatusBadRequest)
			return
		}
		filters.ContactType = &ct
	}

	// Entity type filter
	if entityType := r.URL.Query().Get("entityType"); entityType != "" {
		et := domain.ContactEntityType(entityType)
		if et != domain.ContactEntityCustomer && et != domain.ContactEntityDeal && et != domain.ContactEntityProject {
			http.Error(w, "Invalid entityType. Must be one of: customer, deal, project", http.StatusBadRequest)
			return
		}
		filters.EntityType = &et
	}

	// Entity ID filter
	if entityID := r.URL.Query().Get("entityId"); entityID != "" {
		id, err := uuid.Parse(entityID)
		if err != nil {
			http.Error(w, "Invalid entityId format", http.StatusBadRequest)
			return
		}
		filters.EntityID = &id
	}

	// Sort option
	sortBy := repository.ContactSortByNameAsc
	if s := r.URL.Query().Get("sortBy"); s != "" {
		sortBy = repository.ContactSortOption(s)
	}

	result, err := h.contactService.ListWithFilters(r.Context(), page, pageSize, filters, sortBy)
	if err != nil {
		h.logger.Error("failed to list contacts", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetContact godoc
// @Summary Get contact
// @Description Get a contact by ID with all relationships
// @Tags Contacts
// @Produce json
// @Param id path string true "Contact ID"
// @Success 200 {object} domain.ContactDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /contacts/{id} [get]
func (h *ContactHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	contact, err := h.contactService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, contact)
}

// CreateContact godoc
// @Summary Create contact
// @Description Create a new contact
// @Tags Contacts
// @Accept json
// @Produce json
// @Param request body domain.CreateContactRequest true "Contact data"
// @Success 201 {object} domain.ContactDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /contacts [post]
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	contact, err := h.contactService.Create(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			respondJSON(w, http.StatusConflict, map[string]string{
				"error":   "Conflict",
				"message": "A contact with this email already exists",
			})
			return
		}
		h.logger.Error("failed to create contact", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/api/v1/contacts/"+contact.ID.String())
	respondJSON(w, http.StatusCreated, contact)
}

// UpdateContact godoc
// @Summary Update contact
// @Description Update an existing contact
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Param request body domain.UpdateContactRequest true "Contact data"
// @Success 200 {object} domain.ContactDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /contacts/{id} [put]
func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req domain.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	contact, err := h.contactService.Update(r.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Contact not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "email already exists") {
			respondJSON(w, http.StatusConflict, map[string]string{
				"error":   "Conflict",
				"message": "A contact with this email already exists",
			})
			return
		}
		h.logger.Error("failed to update contact", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, contact)
}

// DeleteContact godoc
// @Summary Delete contact
// @Description Delete a contact
// @Tags Contacts
// @Param id path string true "Contact ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	if err := h.contactService.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Contact not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to delete contact", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddRelationship godoc
// @Summary Add relationship
// @Description Add a relationship between a contact and an entity (customer, deal, or project)
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path string true "Contact ID"
// @Param request body domain.AddContactRelationshipRequest true "Relationship data"
// @Success 201 {object} domain.ContactRelationshipDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /contacts/{id}/relationships [post]
func (h *ContactHandler) AddRelationship(w http.ResponseWriter, r *http.Request) {
	contactID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req domain.AddContactRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Validate entity type
	if req.EntityType != domain.ContactEntityCustomer &&
		req.EntityType != domain.ContactEntityDeal &&
		req.EntityType != domain.ContactEntityProject {
		http.Error(w, "Invalid entityType. Must be one of: customer, deal, project", http.StatusBadRequest)
		return
	}

	relationship, err := h.contactService.AddRelationship(r.Context(), contactID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Contact not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Relationship already exists", http.StatusConflict)
			return
		}
		h.logger.Error("failed to add relationship", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, relationship)
}

// RemoveRelationship godoc
// @Summary Remove relationship
// @Description Remove a relationship between a contact and an entity
// @Tags Contacts
// @Param id path string true "Contact ID"
// @Param relationshipId path string true "Relationship ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /contacts/{id}/relationships/{relationshipId} [delete]
func (h *ContactHandler) RemoveRelationship(w http.ResponseWriter, r *http.Request) {
	contactID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	relationshipID, err := uuid.Parse(chi.URLParam(r, "relationshipId"))
	if err != nil {
		http.Error(w, "Invalid relationship ID", http.StatusBadRequest)
		return
	}

	if err := h.contactService.RemoveRelationship(r.Context(), contactID, relationshipID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Relationship not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "does not belong") {
			http.Error(w, "Relationship does not belong to this contact", http.StatusBadRequest)
			return
		}
		h.logger.Error("failed to remove relationship", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetContactsForEntity godoc
// @Summary Get contacts for entity
// @Description Get all contacts related to a specific entity (customer, deal, or project)
// @Tags Contacts
// @Produce json
// @Param entityType path string true "Entity type" Enums(customers, deals, projects)
// @Param id path string true "Entity ID"
// @Success 200 {array} domain.ContactDTO
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /{entityType}/{id}/contacts [get]
func (h *ContactHandler) GetContactsForEntity(w http.ResponseWriter, r *http.Request) {
	// Parse entity ID
	entityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	// Determine entity type from URL path
	path := r.URL.Path
	var entityType domain.ContactEntityType

	switch {
	case strings.Contains(path, "/customers/"):
		entityType = domain.ContactEntityCustomer
	case strings.Contains(path, "/deals/"):
		entityType = domain.ContactEntityDeal
	case strings.Contains(path, "/projects/"):
		entityType = domain.ContactEntityProject
	default:
		http.Error(w, "Invalid entity type in URL", http.StatusBadRequest)
		return
	}

	contacts, err := h.contactService.ListByEntity(r.Context(), entityType, entityID)
	if err != nil {
		h.logger.Error("failed to get contacts for entity", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, contacts)
}
