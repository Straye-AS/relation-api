package mapper

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
)

// ToCustomerDTO converts Customer to CustomerDTO
func ToCustomerDTO(customer *domain.Customer, totalValueActive float64, totalValueWon float64, activeOffers int) domain.CustomerDTO {
	return domain.CustomerDTO{
		ID:               customer.ID,
		Name:             customer.Name,
		OrgNumber:        customer.OrgNumber,
		Email:            customer.Email,
		Phone:            customer.Phone,
		Address:          customer.Address,
		City:             customer.City,
		PostalCode:       customer.PostalCode,
		Country:          customer.Country,
		ContactPerson:    customer.ContactPerson,
		ContactEmail:     customer.ContactEmail,
		ContactPhone:     customer.ContactPhone,
		Status:           customer.Status,
		Tier:             customer.Tier,
		Industry:         customer.Industry,
		Notes:            customer.Notes,
		CustomerClass:    customer.CustomerClass,
		CreditLimit:      customer.CreditLimit,
		IsInternal:       customer.IsInternal,
		Municipality:     customer.Municipality,
		County:           customer.County,
		CreatedAt:        customer.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        customer.UpdatedAt.UTC().Format(time.RFC3339),
		TotalValueActive: totalValueActive,
		TotalValueWon:    totalValueWon,
		ActiveOffers:     activeOffers,
		CreatedByID:      customer.CreatedByID,
		CreatedByName:    customer.CreatedByName,
		UpdatedByID:      customer.UpdatedByID,
		UpdatedByName:    customer.UpdatedByName,
	}
}

// ToContactDTO converts Contact to ContactDTO
func ToContactDTO(contact *domain.Contact) domain.ContactDTO {
	dto := domain.ContactDTO{
		ID:                     contact.ID,
		FirstName:              contact.FirstName,
		LastName:               contact.LastName,
		FullName:               contact.FullName(),
		Email:                  contact.Email,
		Phone:                  contact.Phone,
		Mobile:                 contact.Mobile,
		Title:                  contact.Title,
		Department:             contact.Department,
		ContactType:            contact.ContactType,
		PrimaryCustomerID:      contact.PrimaryCustomerID,
		Address:                contact.Address,
		City:                   contact.City,
		PostalCode:             contact.PostalCode,
		Country:                contact.Country,
		LinkedInURL:            contact.LinkedInURL,
		PreferredContactMethod: contact.PreferredContactMethod,
		Notes:                  contact.Notes,
		IsActive:               contact.IsActive,
		CreatedAt:              contact.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:              contact.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedByID:            contact.CreatedByID,
		CreatedByName:          contact.CreatedByName,
		UpdatedByID:            contact.UpdatedByID,
		UpdatedByName:          contact.UpdatedByName,
	}

	// Map relationships if loaded
	if len(contact.Relationships) > 0 {
		dto.Relationships = make([]domain.ContactRelationshipDTO, len(contact.Relationships))
		for i, rel := range contact.Relationships {
			dto.Relationships[i] = ToContactRelationshipDTO(&rel)
		}
	}

	return dto
}

// ToContactRelationshipDTO converts ContactRelationship to ContactRelationshipDTO
func ToContactRelationshipDTO(rel *domain.ContactRelationship) domain.ContactRelationshipDTO {
	return domain.ContactRelationshipDTO{
		ID:         rel.ID,
		ContactID:  rel.ContactID,
		EntityType: rel.EntityType,
		EntityID:   rel.EntityID,
		Role:       rel.Role,
		IsPrimary:  rel.IsPrimary,
		CreatedAt:  rel.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// ToDealDTO converts Deal to DealDTO
func ToDealDTO(deal *domain.Deal) domain.DealDTO {
	dto := domain.DealDTO{
		ID:                 deal.ID,
		Title:              deal.Title,
		Description:        deal.Description,
		CustomerID:         deal.CustomerID,
		CompanyID:          deal.CompanyID,
		CustomerName:       deal.CustomerName,
		Stage:              deal.Stage,
		Probability:        deal.Probability,
		Value:              deal.Value,
		WeightedValue:      deal.WeightedValue,
		Currency:           deal.Currency,
		OwnerID:            deal.OwnerID,
		OwnerName:          deal.OwnerName,
		Source:             deal.Source,
		Notes:              deal.Notes,
		LostReason:         deal.LostReason,
		LossReasonCategory: deal.LossReasonCategory,
		OfferID:            deal.OfferID,
		CreatedAt:          deal.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:          deal.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if deal.ExpectedCloseDate != nil {
		formatted := deal.ExpectedCloseDate.Format("2006-01-02")
		dto.ExpectedCloseDate = &formatted
	}

	if deal.ActualCloseDate != nil {
		formatted := deal.ActualCloseDate.Format("2006-01-02")
		dto.ActualCloseDate = &formatted
	}

	return dto
}

// ToDealStageHistoryDTO converts DealStageHistory to DealStageHistoryDTO
func ToDealStageHistoryDTO(history *domain.DealStageHistory) domain.DealStageHistoryDTO {
	dto := domain.DealStageHistoryDTO{
		ID:            history.ID,
		DealID:        history.DealID,
		ToStage:       history.ToStage,
		ChangedByID:   history.ChangedByID,
		ChangedByName: history.ChangedByName,
		Notes:         history.Notes,
		ChangedAt:     history.ChangedAt.UTC().Format(time.RFC3339),
	}

	if history.FromStage != nil {
		dto.FromStage = history.FromStage
	}

	return dto
}

// ToProjectDTO converts Project to ProjectDTO
// Projects are now simplified containers for offers - economic tracking moved to Offer
func ToProjectDTO(project *domain.Project) domain.ProjectDTO {
	return ToProjectDTOWithOfferCount(project, 0)
}

// ToProjectDTOWithOfferCount converts Project to ProjectDTO with offer count
// This is used when listing projects to include the count of linked offers
func ToProjectDTOWithOfferCount(project *domain.Project, offerCount int) domain.ProjectDTO {
	// Set default phase for backwards compatibility if not set
	phase := project.Phase
	if phase == "" {
		phase = domain.ProjectPhaseTilbud
	}

	dto := domain.ProjectDTO{
		ID:                project.ID,
		Name:              project.Name,
		ProjectNumber:     project.ProjectNumber,
		Summary:           project.Summary,
		Description:       project.Description,
		CustomerID:        project.CustomerID, // Now *uuid.UUID (nullable)
		CustomerName:      project.CustomerName,
		Phase:             phase,
		Location:          project.Location,
		CreatedAt:         project.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         project.UpdatedAt.UTC().Format(time.RFC3339),
		DealID:            project.DealID,
		ExternalReference: project.ExternalReference,
		OfferCount:        offerCount,
		CreatedByID:       project.CreatedByID,
		CreatedByName:     project.CreatedByName,
		UpdatedByID:       project.UpdatedByID,
		UpdatedByName:     project.UpdatedByName,
	}

	if !project.StartDate.IsZero() {
		dto.StartDate = project.StartDate.UTC().Format(time.RFC3339)
	}

	if project.EndDate != nil {
		dto.EndDate = project.EndDate.UTC().Format(time.RFC3339)
	}

	return dto
}

// ToOfferDTO converts Offer to OfferDTO
func ToOfferDTO(offer *domain.Offer) domain.OfferDTO {
	items := make([]domain.OfferItemDTO, len(offer.Items))
	for i, item := range offer.Items {
		items[i] = ToOfferItemDTO(&item)
	}

	var dueDate *string
	if offer.DueDate != nil {
		formatted := offer.DueDate.UTC().Format(time.RFC3339)
		dueDate = &formatted
	}

	var sentDate *string
	if offer.SentDate != nil {
		formatted := offer.SentDate.UTC().Format(time.RFC3339)
		sentDate = &formatted
	}

	var expirationDate *string
	if offer.ExpirationDate != nil {
		formatted := offer.ExpirationDate.UTC().Format(time.RFC3339)
		expirationDate = &formatted
	}

	// Calculate margin (Value - Cost), margin_percent is stored in DB
	margin := offer.Value - offer.Cost

	// Map execution fields (order phase)
	var health *string
	if offer.Health != nil {
		h := string(*offer.Health)
		health = &h
	}

	var startDate *string
	if offer.StartDate != nil {
		formatted := offer.StartDate.UTC().Format(time.RFC3339)
		startDate = &formatted
	}

	var endDate *string
	if offer.EndDate != nil {
		formatted := offer.EndDate.UTC().Format(time.RFC3339)
		endDate = &formatted
	}

	var estimatedCompletionDate *string
	if offer.EstimatedCompletionDate != nil {
		formatted := offer.EstimatedCompletionDate.Format("2006-01-02")
		estimatedCompletionDate = &formatted
	}

	// Convert pq.StringArray to []string for team members
	var teamMembers []string
	if len(offer.TeamMembers) > 0 {
		teamMembers = []string(offer.TeamMembers)
	}

	return domain.OfferDTO{
		ID:                    offer.ID,
		Title:                 offer.Title,
		OfferNumber:           offer.OfferNumber,
		ExternalReference:     offer.ExternalReference,
		CustomerID:            offer.CustomerID,
		CustomerName:          offer.CustomerName,
		ProjectID:             offer.ProjectID,
		ProjectName:           offer.ProjectName,
		CompanyID:             offer.CompanyID,
		Phase:                 offer.Phase,
		Probability:           offer.Probability,
		Value:                 offer.Value,
		Status:                offer.Status,
		CreatedAt:             offer.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:             offer.UpdatedAt.UTC().Format(time.RFC3339),
		ResponsibleUserID:     offer.ResponsibleUserID,
		ResponsibleUserName:   offer.ResponsibleUserName,
		Items:                 items,
		Description:           offer.Description,
		Notes:                 offer.Notes,
		DueDate:               dueDate,
		Cost:                  offer.Cost,
		Margin:                margin,
		MarginPercent:         offer.MarginPercent, // Stored in DB, auto-calculated by trigger
		Location:              offer.Location,
		SentDate:              sentDate,
		ExpirationDate:        expirationDate,
		CustomerHasWonProject: offer.CustomerHasWonProject,
		// Order phase execution fields
		ManagerID:               offer.ManagerID,
		ManagerName:             offer.ManagerName,
		TeamMembers:             teamMembers,
		Spent:                   offer.Spent,
		Invoiced:                offer.Invoiced,
		OrderReserve:            offer.OrderReserve,
		Health:                  health,
		CompletionPercent:       offer.CompletionPercent,
		StartDate:               startDate,
		EndDate:                 endDate,
		EstimatedCompletionDate: estimatedCompletionDate,
		// User tracking fields
		CreatedByID:   offer.CreatedByID,
		CreatedByName: offer.CreatedByName,
		UpdatedByID:   offer.UpdatedByID,
		UpdatedByName: offer.UpdatedByName,
		// Data Warehouse synced fields
		DWTotalIncome:   offer.DWTotalIncome,
		DWMaterialCosts: offer.DWMaterialCosts,
		DWEmployeeCosts: offer.DWEmployeeCosts,
		DWOtherCosts:    offer.DWOtherCosts,
		DWNetResult:     offer.DWNetResult,
		DWLastSyncedAt:  formatTimePointer(offer.DWLastSyncedAt),
	}
}

// ToOfferItemDTO converts OfferItem to OfferItemDTO
func ToOfferItemDTO(item *domain.OfferItem) domain.OfferItemDTO {
	return domain.OfferItemDTO{
		ID:          item.ID,
		Discipline:  item.Discipline,
		Cost:        item.Cost,
		Revenue:     item.Revenue,
		Margin:      item.Margin,
		Description: item.Description,
		Quantity:    item.Quantity,
		Unit:        item.Unit,
	}
}

// ToBudgetItemDTO converts BudgetItem to BudgetItemDTO
func ToBudgetItemDTO(item *domain.BudgetItem) domain.BudgetItemDTO {
	return domain.BudgetItemDTO{
		ID:              item.ID,
		ParentType:      item.ParentType,
		ParentID:        item.ParentID,
		Name:            item.Name,
		ExpectedCost:    item.ExpectedCost,
		ExpectedMargin:  item.ExpectedMargin,
		ExpectedRevenue: item.ExpectedRevenue,
		ExpectedProfit:  item.ExpectedProfit,
		Quantity:        item.Quantity,
		PricePerItem:    item.PricePerItem,
		Description:     item.Description,
		DisplayOrder:    item.DisplayOrder,
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ToBudgetSummaryDTO creates a summary DTO from budget items
func ToBudgetSummaryDTO(parentType domain.BudgetParentType, parentID uuid.UUID, items []domain.BudgetItem) domain.BudgetSummaryDTO {
	totalCost := 0.0
	totalRevenue := 0.0
	totalProfit := 0.0

	for _, item := range items {
		totalCost += item.ExpectedCost
		totalRevenue += item.ExpectedRevenue
		totalProfit += item.ExpectedProfit
	}

	marginPercent := 0.0
	if totalRevenue > 0 {
		marginPercent = (totalProfit / totalRevenue) * 100
	}

	return domain.BudgetSummaryDTO{
		ParentType:    parentType,
		ParentID:      parentID,
		ItemCount:     len(items),
		TotalCost:     totalCost,
		TotalRevenue:  totalRevenue,
		TotalProfit:   totalProfit,
		MarginPercent: marginPercent,
	}
}

// ToProjectActualCostDTO converts ProjectActualCost to ProjectActualCostDTO
func ToProjectActualCostDTO(cost *domain.ProjectActualCost) domain.ProjectActualCostDTO {
	dto := domain.ProjectActualCostDTO{
		ID:               cost.ID,
		ProjectID:        cost.ProjectID,
		CostType:         cost.CostType,
		Description:      cost.Description,
		Amount:           cost.Amount,
		Currency:         cost.Currency,
		CostDate:         cost.CostDate.Format("2006-01-02"),
		BudgetItemID:     cost.BudgetItemID,
		ERPSource:        cost.ERPSource,
		ERPReference:     cost.ERPReference,
		ERPTransactionID: cost.ERPTransactionID,
		IsApproved:       cost.IsApproved,
		ApprovedByID:     cost.ApprovedByID,
		Notes:            cost.Notes,
		CreatedAt:        cost.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        cost.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if cost.PostingDate != nil {
		dto.PostingDate = cost.PostingDate.Format("2006-01-02")
	}

	if cost.ERPSyncedAt != nil {
		dto.ERPSyncedAt = cost.ERPSyncedAt.UTC().Format(time.RFC3339)
	}

	if cost.ApprovedAt != nil {
		dto.ApprovedAt = cost.ApprovedAt.UTC().Format(time.RFC3339)
	}

	return dto
}

// ToProjectCostSummaryDTO creates a cost summary DTO for a project
// DEPRECATED: Project no longer has economic fields. Use Offer-based cost tracking instead.
// This function is kept for backwards compatibility but returns zeros for economic fields.
func ToProjectCostSummaryDTO(project *domain.Project, actualCosts []domain.ProjectActualCost) domain.ProjectCostSummaryDTO {
	totalActualCosts := 0.0
	for _, cost := range actualCosts {
		totalActualCosts += cost.Amount
	}

	return domain.ProjectCostSummaryDTO{
		ProjectID:        project.ID,
		ProjectName:      project.Name,
		Value:            0, // Economic tracking moved to Offer
		Cost:             0, // Economic tracking moved to Offer
		MarginPercent:    0, // Economic tracking moved to Offer
		Spent:            0, // Economic tracking moved to Offer
		ActualCosts:      totalActualCosts,
		RemainingValue:   0,
		ValueUsedPercent: 0,
		CostEntryCount:   len(actualCosts),
	}
}

// ToUserDTO converts User to UserDTO
func ToUserDTO(user *domain.User) domain.UserDTO {
	roles := user.Roles
	if roles == nil {
		roles = []string{}
	}

	return domain.UserDTO{
		ID:         user.ID,
		Name:       user.DisplayName,
		Email:      user.Email,
		Roles:      roles,
		Department: user.Department,
		Avatar:     user.Avatar,
	}
}

// ToNotificationDTO converts Notification to NotificationDTO
func ToNotificationDTO(notification *domain.Notification) domain.NotificationDTO {
	return domain.NotificationDTO{
		ID:         notification.ID,
		Type:       notification.Type,
		Title:      notification.Title,
		Message:    notification.Message,
		Read:       notification.Read,
		CreatedAt:  notification.CreatedAt.UTC().Format(time.RFC3339),
		EntityID:   notification.EntityID,
		EntityType: notification.EntityType,
	}
}

// CalculateMargin calculates margin percentage
func CalculateMargin(cost, revenue float64) float64 {
	if revenue == 0 {
		return 0
	}
	return ((revenue - cost) / revenue) * 100
}

// CalculateOfferValue calculates total offer value from items
func CalculateOfferValue(items []domain.OfferItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Revenue
	}
	return total
}

// UpdateDenormalizedFields updates denormalized fields in related entities
func UpdateOfferDenormalizedFields(offer *domain.Offer, customerName, userName string) {
	offer.CustomerName = customerName
	offer.ResponsibleUserName = userName
}

// UpdateProjectDenormalizedFields updates denormalized fields in a project
// Note: Project no longer has ManagerName - manager tracking has moved to Offer
func UpdateProjectDenormalizedFields(project *domain.Project, customerName string) {
	project.CustomerName = customerName
}

// UpdateDealDenormalizedFields updates denormalized fields in a deal
func UpdateDealDenormalizedFields(deal *domain.Deal, customerName, ownerName string) {
	deal.CustomerName = customerName
	deal.OwnerName = ownerName
}

// ToFileDTO converts File to FileDTO
func ToFileDTO(file *domain.File) domain.FileDTO {
	return domain.FileDTO{
		ID:          file.ID,
		Filename:    file.Filename,
		ContentType: file.ContentType,
		Size:        file.Size,
		OfferID:     file.OfferID,
		CreatedAt:   file.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// ToActivityDTO converts Activity to ActivityDTO
func ToActivityDTO(activity *domain.Activity) domain.ActivityDTO {
	dto := domain.ActivityDTO{
		ID:               activity.ID,
		TargetType:       activity.TargetType,
		TargetID:         activity.TargetID,
		TargetName:       activity.TargetName,
		Title:            activity.Title,
		Body:             activity.Body,
		OccurredAt:       activity.OccurredAt.UTC().Format(time.RFC3339),
		CreatorName:      activity.CreatorName,
		CreatedAt:        activity.CreatedAt.UTC().Format(time.RFC3339),
		ActivityType:     activity.ActivityType,
		Status:           activity.Status,
		DurationMinutes:  activity.DurationMinutes,
		Priority:         activity.Priority,
		IsPrivate:        activity.IsPrivate,
		CreatorID:        activity.CreatorID,
		AssignedToID:     activity.AssignedToID,
		CompanyID:        activity.CompanyID,
		ParentActivityID: activity.ParentActivityID,
	}

	// Convert pq.StringArray to []string for attendees
	if len(activity.Attendees) > 0 {
		dto.Attendees = []string(activity.Attendees)
	}

	if activity.ScheduledAt != nil {
		dto.ScheduledAt = activity.ScheduledAt.UTC().Format(time.RFC3339)
	}

	if activity.DueDate != nil {
		dto.DueDate = activity.DueDate.Format("2006-01-02")
	}

	if activity.CompletedAt != nil {
		dto.CompletedAt = activity.CompletedAt.UTC().Format(time.RFC3339)
	}

	return dto
}

// ToUserRoleDTO converts UserRole to UserRoleDTO
func ToUserRoleDTO(role *domain.UserRole) domain.UserRoleDTO {
	dto := domain.UserRoleDTO{
		ID:        role.ID,
		UserID:    role.UserID,
		Role:      role.Role,
		CompanyID: role.CompanyID,
		GrantedBy: role.GrantedBy,
		GrantedAt: role.GrantedAt.UTC().Format(time.RFC3339),
		IsActive:  role.IsActive,
	}

	if role.ExpiresAt != nil {
		dto.ExpiresAt = role.ExpiresAt.UTC().Format(time.RFC3339)
	}

	return dto
}

// ToUserPermissionDTO converts UserPermission to UserPermissionDTO
func ToUserPermissionDTO(perm *domain.UserPermission) domain.UserPermissionDTO {
	dto := domain.UserPermissionDTO{
		ID:         perm.ID,
		UserID:     perm.UserID,
		Permission: perm.Permission,
		CompanyID:  perm.CompanyID,
		IsGranted:  perm.IsGranted,
		GrantedBy:  perm.GrantedBy,
		GrantedAt:  perm.GrantedAt.UTC().Format(time.RFC3339),
		Reason:     perm.Reason,
	}

	if perm.ExpiresAt != nil {
		dto.ExpiresAt = perm.ExpiresAt.UTC().Format(time.RFC3339)
	}

	return dto
}

// ToAuditLogDTO converts AuditLog to AuditLogDTO
func ToAuditLogDTO(log *domain.AuditLog) domain.AuditLogDTO {
	return domain.AuditLogDTO{
		ID:          log.ID,
		UserID:      log.UserID,
		UserEmail:   log.UserEmail,
		UserName:    log.UserName,
		Action:      log.Action,
		EntityType:  log.EntityType,
		EntityID:    log.EntityID,
		EntityName:  log.EntityName,
		CompanyID:   log.CompanyID,
		OldValues:   log.OldValues,
		NewValues:   log.NewValues,
		Changes:     log.Changes,
		IPAddress:   log.IPAddress,
		UserAgent:   log.UserAgent,
		RequestID:   log.RequestID,
		PerformedAt: log.PerformedAt.UTC().Format(time.RFC3339),
	}
}

// ToProjectBudgetDTO converts project budget info to DTO
// DEPRECATED: Project no longer has economic fields. Use Offer-based budget tracking instead.
// This function is kept for backwards compatibility but returns zeros.
// Budget information should now be retrieved from the associated Offers.
func ToProjectBudgetDTO(project *domain.Project) domain.ProjectBudgetDTO {
	// Project no longer has economic fields - they have moved to Offer
	return domain.ProjectBudgetDTO{
		Value:         0,
		Cost:          0,
		MarginPercent: 0,
		Spent:         0,
		Remaining:     0,
		PercentUsed:   0,
	}
}

// ToProjectWithDetailsDTO converts Project with related data to ProjectWithDetailsDTO
func ToProjectWithDetailsDTO(
	project *domain.Project,
	budgetSummary *domain.BudgetSummaryDTO,
	activities []domain.Activity,
	offer *domain.Offer,
	deal *domain.Deal,
) domain.ProjectWithDetailsDTO {
	return ToProjectWithDetailsDTOWithOfferCount(project, budgetSummary, activities, offer, deal, 0)
}

// ToProjectWithDetailsDTOWithOfferCount converts Project with related data to ProjectWithDetailsDTO with offer count
func ToProjectWithDetailsDTOWithOfferCount(
	project *domain.Project,
	budgetSummary *domain.BudgetSummaryDTO,
	activities []domain.Activity,
	offer *domain.Offer,
	deal *domain.Deal,
	offerCount int,
) domain.ProjectWithDetailsDTO {
	dto := domain.ProjectWithDetailsDTO{
		ProjectDTO:    ToProjectDTOWithOfferCount(project, offerCount),
		BudgetSummary: budgetSummary,
	}

	// Map activities
	if len(activities) > 0 {
		dto.RecentActivities = make([]domain.ActivityDTO, len(activities))
		for i, act := range activities {
			dto.RecentActivities[i] = ToActivityDTO(&act)
		}
	}

	// Map related offer
	if offer != nil {
		offerDTO := ToOfferDTO(offer)
		dto.Offer = &offerDTO
	}

	// Map related deal
	if deal != nil {
		dealDTO := ToDealDTO(deal)
		dto.Deal = &dealDTO
	}

	return dto
}

// ToCompanyDetailDTO converts Company to CompanyDetailDTO
func ToCompanyDetailDTO(company *domain.Company) domain.CompanyDetailDTO {
	return domain.CompanyDetailDTO{
		ID:                          string(company.ID),
		Name:                        company.Name,
		ShortName:                   company.ShortName,
		OrgNumber:                   company.OrgNumber,
		Color:                       company.Color,
		Logo:                        company.Logo,
		IsActive:                    company.IsActive,
		DefaultOfferResponsibleID:   company.DefaultOfferResponsibleID,
		DefaultProjectResponsibleID: company.DefaultProjectResponsibleID,
		CreatedAt:                   company.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                   company.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// FormatError creates a formatted error message
func FormatError(entity, operation string, err error) error {
	return fmt.Errorf("failed to %s %s: %w", operation, entity, err)
}

// formatTimePointer formats a time pointer to ISO 8601 string in UTC, returning nil if input is nil
func formatTimePointer(t *time.Time) *string {
	if t == nil {
		return nil
	}
	// Convert to UTC before formatting to ensure correct timezone representation
	formatted := t.UTC().Format(time.RFC3339)
	return &formatted
}

// ============================================================================
// Supplier Mappers
// ============================================================================

// SupplierToDTO converts Supplier to SupplierDTO
func SupplierToDTO(supplier *domain.Supplier) domain.SupplierDTO {
	return domain.SupplierDTO{
		ID:            supplier.ID,
		Name:          supplier.Name,
		OrgNumber:     supplier.OrgNumber,
		Email:         supplier.Email,
		Phone:         supplier.Phone,
		Address:       supplier.Address,
		City:          supplier.City,
		PostalCode:    supplier.PostalCode,
		Country:       supplier.Country,
		Municipality:  supplier.Municipality,
		County:        supplier.County,
		ContactPerson: supplier.ContactPerson,
		ContactEmail:  supplier.ContactEmail,
		ContactPhone:  supplier.ContactPhone,
		Status:        supplier.Status,
		Category:      supplier.Category,
		Notes:         supplier.Notes,
		PaymentTerms:  supplier.PaymentTerms,
		Website:       supplier.Website,
		CreatedAt:     supplier.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     supplier.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedByID:   supplier.CreatedByID,
		CreatedByName: supplier.CreatedByName,
		UpdatedByID:   supplier.UpdatedByID,
		UpdatedByName: supplier.UpdatedByName,
	}
}

// SuppliersToDTO converts a slice of Suppliers to a slice of SupplierDTOs
func SuppliersToDTO(suppliers []domain.Supplier) []domain.SupplierDTO {
	dtos := make([]domain.SupplierDTO, len(suppliers))
	for i, supplier := range suppliers {
		dtos[i] = SupplierToDTO(&supplier)
	}
	return dtos
}

// SupplierContactToDTO converts SupplierContact to SupplierContactDTO
func SupplierContactToDTO(contact *domain.SupplierContact) domain.SupplierContactDTO {
	return domain.SupplierContactDTO{
		ID:         contact.ID,
		SupplierID: contact.SupplierID,
		Name:       contact.Name,
		Title:      contact.Title,
		Email:      contact.Email,
		Phone:      contact.Phone,
		IsPrimary:  contact.IsPrimary,
		Notes:      contact.Notes,
		CreatedAt:  contact.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  contact.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// OfferSupplierToDTO converts OfferSupplier to OfferSupplierDTO
func OfferSupplierToDTO(offerSupplier *domain.OfferSupplier) domain.OfferSupplierDTO {
	return domain.OfferSupplierDTO{
		ID:           offerSupplier.ID,
		OfferID:      offerSupplier.OfferID,
		OfferTitle:   offerSupplier.OfferTitle,
		SupplierID:   offerSupplier.SupplierID,
		SupplierName: offerSupplier.SupplierName,
		Status:       offerSupplier.Status,
		Notes:        offerSupplier.Notes,
		CreatedAt:    offerSupplier.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    offerSupplier.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// SupplierStatsInput holds aggregated statistics for a supplier - used for mapping
// This mirrors repository.SupplierStats to avoid circular imports
type SupplierStatsInput struct {
	TotalOffers     int
	ActiveOffers    int
	CompletedOffers int
	TotalProjects   int
}

// SupplierToWithDetailsDTO converts Supplier to SupplierWithDetailsDTO with stats and related data
// stats should be of type *repository.SupplierStats or *SupplierStatsInput
func SupplierToWithDetailsDTO(supplier *domain.Supplier, stats *SupplierStatsInput, recentOffers []domain.OfferSupplier) domain.SupplierWithDetailsDTO {
	dto := domain.SupplierWithDetailsDTO{
		SupplierDTO: SupplierToDTO(supplier),
	}

	// Map stats if provided
	if stats != nil {
		dto.Stats = &domain.SupplierStatsDTO{
			TotalOffers:     stats.TotalOffers,
			ActiveOffers:    stats.ActiveOffers,
			CompletedOffers: stats.CompletedOffers,
			TotalProjects:   stats.TotalProjects,
		}
	}

	// Map contacts if loaded
	if len(supplier.Contacts) > 0 {
		dto.Contacts = make([]domain.SupplierContactDTO, len(supplier.Contacts))
		for i, contact := range supplier.Contacts {
			dto.Contacts[i] = SupplierContactToDTO(&contact)
		}
	}

	// Map recent offers
	if len(recentOffers) > 0 {
		dto.RecentOffers = make([]domain.OfferSupplierDTO, len(recentOffers))
		for i, offerSupplier := range recentOffers {
			dto.RecentOffers[i] = OfferSupplierToDTO(&offerSupplier)
		}
	}

	return dto
}
