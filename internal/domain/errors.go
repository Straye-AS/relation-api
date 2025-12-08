package domain

// APIError represents a standardized API error with HTTP status code
type APIError struct {
	Type   string            `json:"type"`
	Title  string            `json:"title"`
	Status int               `json:"status"`
	Detail string            `json:"detail,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return e.Title
}

// ValidationFieldError maps a field name to its validation error message
type ValidationFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationMessages provides human-readable validation error messages
// These map validator tags to user-friendly messages
var ValidationMessages = map[string]string{
	"required":   "This field is required",
	"email":      "Must be a valid email address",
	"max":        "Exceeds maximum length",
	"min":        "Below minimum length",
	"gte":        "Must be greater than or equal to minimum value",
	"gt":         "Must be greater than minimum value",
	"lte":        "Must be less than or equal to maximum value",
	"lt":         "Must be less than maximum value",
	"uuid":       "Must be a valid UUID",
	"url":        "Must be a valid URL",
	"oneof":      "Must be one of the allowed values",
	"alphanum":   "Must contain only alphanumeric characters",
	"numeric":    "Must be a numeric value",
	"alpha":      "Must contain only alphabetic characters",
	"len":        "Must be exactly the specified length",
	"eq":         "Must equal the specified value",
	"ne":         "Must not equal the specified value",
	"contains":   "Must contain the specified value",
	"excludes":   "Must not contain the specified value",
	"startswith": "Must start with the specified value",
	"endswith":   "Must end with the specified value",
}

// GetValidationMessage returns a human-readable message for a validation tag
func GetValidationMessage(tag string) string {
	if msg, ok := ValidationMessages[tag]; ok {
		return msg
	}
	return "Validation failed: " + tag
}

// Common error types for RFC 7807 Problem Details
const (
	ErrorTypeValidation   = "validation_error"
	ErrorTypeNotFound     = "not_found"
	ErrorTypeBadRequest   = "bad_request"
	ErrorTypeConflict     = "conflict"
	ErrorTypeUnauthorized = "unauthorized"
	ErrorTypeForbidden    = "forbidden"
	ErrorTypeInternal     = "internal_error"
)
