package service

import "errors"

// Common service errors
var (
	// ErrPermissionDenied is returned when user doesn't have required permission
	ErrPermissionDenied = errors.New("permission denied")

	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrConflict is returned when an operation conflicts with existing state
	ErrConflict = errors.New("operation conflicts with existing state")

	// ErrUnauthorized is returned when user is not authenticated
	ErrUnauthorized = errors.New("unauthorized")
)
