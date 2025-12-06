package service

import "errors"

// Common service errors
var (
	// ErrPermissionDenied is returned when a user doesn't have permission for an action
	ErrPermissionDenied = errors.New("permission denied")

	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrConflict is returned when there's a conflict (e.g., duplicate)
	ErrConflict = errors.New("resource conflict")

	// ErrUnauthorized is returned when user is not authenticated
	ErrUnauthorized = errors.New("unauthorized")

	// ErrRoleNotFound is returned when a role is not found
	ErrRoleNotFound = errors.New("role not found")

	// ErrRoleAlreadyAssigned is returned when trying to assign a role that's already assigned
	ErrRoleAlreadyAssigned = errors.New("role already assigned to user")

	// ErrCannotRemoveLastAdmin is returned when trying to remove the last admin
	ErrCannotRemoveLastAdmin = errors.New("cannot remove the last admin role")

	// ErrInvalidRole is returned when an invalid role type is provided
	ErrInvalidRole = errors.New("invalid role type")

	// ErrInvalidPermission is returned when an invalid permission type is provided
	ErrInvalidPermission = errors.New("invalid permission type")

	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrCompanyNotFound is returned when a company is not found
	ErrCompanyNotFound = errors.New("company not found")
)
