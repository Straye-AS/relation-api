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

	// Inquiry errors

	// ErrInquiryNotFound is returned when an inquiry is not found
	ErrInquiryNotFound = errors.New("inquiry not found")

	// ErrNotAnInquiry is returned when trying to perform inquiry operations on non-draft offers
	ErrNotAnInquiry = errors.New("offer is not an inquiry (must be in draft phase)")

	// ErrInquiryMissingConversionData is returned when conversion requires data that wasn't provided
	ErrInquiryMissingConversionData = errors.New("conversion requires responsibleUserId or companyId with default responsible user")

	// ErrProjectNotFound is returned when a project is not found
	ErrProjectNotFound = errors.New("project not found")

	// ErrInvalidCompanyID is returned when an invalid company ID is provided
	ErrInvalidCompanyID = errors.New("invalid company ID")

	// ErrOfferNumberGenerationFailed is returned when offer number generation fails
	ErrOfferNumberGenerationFailed = errors.New("failed to generate offer number")

	// ErrOfferNumberConflict is returned when an offer number already exists
	ErrOfferNumberConflict = errors.New("offer number already exists")

	// ErrExternalReferenceConflict is returned when an external reference already exists within a company
	ErrExternalReferenceConflict = errors.New("external reference already exists for this company")

	// ErrDraftOfferCannotHaveNumber is returned when trying to set an offer number on a draft offer
	ErrDraftOfferCannotHaveNumber = errors.New("draft offers cannot have an offer number")

	// ErrNonDraftOfferMustHaveNumber is returned when a non-draft offer is missing an offer number
	ErrNonDraftOfferMustHaveNumber = errors.New("non-draft offers must have an offer number")

	// ErrExpirationDateBeforeSentDate is returned when the expiration date is before the sent date
	ErrExpirationDateBeforeSentDate = errors.New("expiration date cannot be before sent date")

	// Project Phase errors

	// ErrOfferNotInProject is returned when trying to win an offer that is not linked to a project
	ErrOfferNotInProject = errors.New("offer must be linked to a project to be won through this endpoint")

	// ErrOfferAlreadyWon is returned when trying to win an offer that is already won
	ErrOfferAlreadyWon = errors.New("offer is already won")

	// ErrProjectNotInTilbudPhase is returned when trying to win an offer for a project not in tilbud phase
	ErrProjectNotInTilbudPhase = errors.New("project must be in tilbud phase to win an offer")

	// ErrProjectEconomicsNotEditable is returned when trying to edit economics for a project not in active phase
	ErrProjectEconomicsNotEditable = errors.New("project economics can only be edited during active phase")

	// Offer-Project lifecycle errors

	// ErrNonDraftOfferRequiresProject is returned when trying to create/update a non-draft offer without a project
	ErrNonDraftOfferRequiresProject = errors.New("non-draft offers must be linked to a project")

	// ErrCannotAddOfferToCancelledProject is returned when trying to add an offer to a cancelled project
	ErrCannotAddOfferToCancelledProject = errors.New("cannot add offer to cancelled project")

	// ErrProjectCustomerMismatch is returned when project and offer have different customers
	ErrProjectCustomerMismatch = errors.New("project customer does not match offer customer")

	// ErrProjectCompanyMismatch is returned when project and offer have different companies
	ErrProjectCompanyMismatch = errors.New("project company does not match offer company")

	// ErrCannotChangeOfferCustomerWhenLinked is returned when trying to change customer of a linked offer
	ErrCannotChangeOfferCustomerWhenLinked = errors.New("cannot change customer of offer linked to project")

	// ErrCannotChangeOfferCompanyWhenLinked is returned when trying to change company of a linked offer
	ErrCannotChangeOfferCompanyWhenLinked = errors.New("cannot change company of offer linked to project")

	// ErrNoActiveOffersRemaining is returned when no active offers remain after losing an offer
	ErrNoActiveOffersRemaining = errors.New("no active offers remaining in project")
)
