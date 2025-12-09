package service

import "errors"

// Common service errors
var (
	// ErrPermissionDenied is returned when a user doesn't have permission for an action
	ErrPermissionDenied = errors.New("permission denied")

	// ErrForbidden is returned when access is forbidden
	ErrForbidden = errors.New("forbidden")

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

	// Offer lifecycle errors

	// ErrOfferNotFound is returned when an offer is not found
	ErrOfferNotFound = errors.New("offer not found")

	// ErrOfferInvalidPhaseTransition is returned when trying to transition to an invalid phase
	ErrOfferInvalidPhaseTransition = errors.New("invalid offer phase transition")

	// ErrOfferNotInDraftPhase is returned when offer must be in draft phase
	ErrOfferNotInDraftPhase = errors.New("offer must be in draft or in_progress phase to be sent")

	// ErrOfferNotSent is returned when offer must be in sent phase
	ErrOfferNotSent = errors.New("offer must be in sent phase")

	// ErrOfferAlreadyClosed is returned when trying to modify a closed offer
	ErrOfferAlreadyClosed = errors.New("offer is already in a closed state (won/lost/expired)")

	// ErrOfferCannotClone is returned when an offer cannot be cloned
	ErrOfferCannotClone = errors.New("cannot clone this offer")

	// ErrOfferMissingResponsible is returned when advancing to in_progress without a responsible user or company with default
	ErrOfferMissingResponsible = errors.New("offer must have a responsible user or company with default responsible user to advance to in_progress")

	// ErrProjectCreationFailed is returned when project creation fails during offer acceptance
	ErrProjectCreationFailed = errors.New("failed to create project from offer")

	// Deal lifecycle errors

	// ErrDealNotFound is returned when a deal is not found
	ErrDealNotFound = errors.New("deal not found")

	// ErrDealAlreadyHasOffer is returned when trying to create an offer for a deal that already has one
	ErrDealAlreadyHasOffer = errors.New("deal already has a linked offer")

	// ErrDealInvalidStageForOffer is returned when deal is in invalid stage for creating an offer
	ErrDealInvalidStageForOffer = errors.New("deal must be in lead or qualified stage to create an offer")
)
