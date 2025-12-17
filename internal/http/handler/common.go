package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/straye-as/relation-api/internal/domain"
)

var validate = validator.New()

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// respondValidationError sends a standardized validation error response with specific field messages
func respondValidationError(w http.ResponseWriter, err error) {
	errors := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			fieldName := toJSONFieldName(fe.Field())
			errors[fieldName] = formatValidationError(fe)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(domain.APIError{
		Type:   domain.ErrorTypeValidation,
		Title:  "Validation Error",
		Status: http.StatusBadRequest,
		Detail: "One or more fields failed validation",
		Errors: errors,
	})
}

// formatValidationError creates a human-readable validation error message
func formatValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", toJSONFieldName(fe.Field()))
	case "email":
		return "Must be a valid email address"
	case "max":
		return fmt.Sprintf("Must be at most %s characters", fe.Param())
	case "min":
		return fmt.Sprintf("Must be at least %s characters", fe.Param())
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", fe.Param())
	case "gt":
		return fmt.Sprintf("Must be greater than %s", fe.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", fe.Param())
	case "lt":
		return fmt.Sprintf("Must be less than %s", fe.Param())
	case "uuid":
		return "Must be a valid UUID"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fe.Param())
	case "url":
		return "Must be a valid URL"
	default:
		return domain.GetValidationMessage(fe.Tag())
	}
}

// toJSONFieldName converts a Go struct field name to its JSON equivalent (camelCase)
func toJSONFieldName(field string) string {
	if len(field) == 0 {
		return field
	}
	// Convert first character to lowercase for camelCase
	return strings.ToLower(field[:1]) + field[1:]
}

// parseJSON parses a JSON string into the target interface
func parseJSON(data string, target interface{}) error {
	return json.Unmarshal([]byte(data), target)
}

// respondWithError sends a standardized JSON error response
func respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(domain.APIError{
		Type:   getErrorType(status),
		Title:  http.StatusText(status),
		Status: status,
		Detail: message,
	})
}

// getErrorType returns the appropriate error type for an HTTP status code
func getErrorType(status int) string {
	switch status {
	case http.StatusBadRequest:
		return domain.ErrorTypeBadRequest
	case http.StatusUnauthorized:
		return domain.ErrorTypeUnauthorized
	case http.StatusForbidden:
		return domain.ErrorTypeForbidden
	case http.StatusNotFound:
		return domain.ErrorTypeNotFound
	case http.StatusConflict:
		return domain.ErrorTypeConflict
	default:
		return domain.ErrorTypeInternal
	}
}
