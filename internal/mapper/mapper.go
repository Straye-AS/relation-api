package mapper

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
)

// ToCustomerDTO converts Customer to CustomerDTO
func ToCustomerDTO(customer *domain.Customer, totalValue float64, activeOffers int) domain.CustomerDTO {
	return domain.CustomerDTO{
		ID:            customer.ID,
		Name:          customer.Name,
		OrgNumber:     customer.OrgNumber,
		Email:         customer.Email,
		Phone:         customer.Phone,
		Address:       customer.Address,
		City:          customer.City,
		PostalCode:    customer.PostalCode,
		Country:       customer.Country,
		ContactPerson: customer.ContactPerson,
		ContactEmail:  customer.ContactEmail,
		ContactPhone:  customer.ContactPhone,
		Status:        customer.Status,
		Tier:          customer.Tier,
		Industry:      customer.Industry,
		Notes:         customer.Notes,
		CustomerClass: customer.CustomerClass,
		CreditLimit:   customer.CreditLimit,
		IsInternal:    customer.IsInternal,
		Municipality:  customer.Municipality,
		County:        customer.County,
		CreatedAt:     customer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     customer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		TotalValue:    totalValue,
		ActiveOffers:  activeOffers,
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
		CreatedAt:              contact.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:              contact.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
		CreatedAt:  rel.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
		CreatedAt:          deal.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:          deal.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
		ChangedAt:     history.ChangedAt.Format("2006-01-02T15:04:05Z"),
	}

	if history.FromStage != nil {
		dto.FromStage = history.FromStage
	}

	return dto
}

// ToProjectDTO converts Project to ProjectDTO
func ToProjectDTO(project *domain.Project) domain.ProjectDTO {
	// Set default phase for backwards compatibility if not set
	phase := project.Phase
	if phase == "" {
		phase = domain.ProjectPhaseTilbud
	}

	dto := domain.ProjectDTO{
		ID:                   project.ID,
		Name:                 project.Name,
		ProjectNumber:        project.ProjectNumber,
		Summary:              project.Summary,
		Description:          project.Description,
		CustomerID:           project.CustomerID,
		CustomerName:         project.CustomerName,
		CompanyID:            project.CompanyID,
		Status:               project.Status,
		Phase:                phase,
		StartDate:            project.StartDate.Format("2006-01-02T15:04:05Z"),
		Budget:               project.Budget,
		Spent:                project.Spent,
		ManagerID:            project.ManagerID,
		ManagerName:          project.ManagerName,
		TeamMembers:          project.TeamMembers,
		CreatedAt:            project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:            project.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		OfferID:              project.OfferID,
		DealID:               project.DealID,
		HasDetailedBudget:    project.HasDetailedBudget,
		Health:               project.Health,
		CompletionPercent:    project.CompletionPercent,
		WinningOfferID:       project.WinningOfferID,
		InheritedOfferNumber: project.InheritedOfferNumber,
		CalculatedOfferValue: project.CalculatedOfferValue,
		IsEconomicsEditable:  phase.IsEditablePhase(),
	}

	if project.EndDate != nil {
		dto.EndDate = project.EndDate.Format("2006-01-02T15:04:05Z")
	}

	if project.EstimatedCompletionDate != nil {
		dto.EstimatedCompletionDate = project.EstimatedCompletionDate.Format("2006-01-02")
	}

	if project.WonAt != nil {
		dto.WonAt = project.WonAt.Format("2006-01-02T15:04:05Z")
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
		formatted := offer.DueDate.Format("2006-01-02T15:04:05Z")
		dueDate = &formatted
	}

	var sentDate *string
	if offer.SentDate != nil {
		formatted := offer.SentDate.Format("2006-01-02T15:04:05Z")
		sentDate = &formatted
	}

	var expirationDate *string
	if offer.ExpirationDate != nil {
		formatted := offer.ExpirationDate.Format("2006-01-02T15:04:05Z")
		expirationDate = &formatted
	}

	// Calculate margin (Price - Cost), margin_percent is stored in DB
	margin := offer.Price - offer.Cost

	return domain.OfferDTO{
		ID:                    offer.ID,
		Title:                 offer.Title,
		OfferNumber:           offer.OfferNumber,
		ExternalReference:     offer.ExternalReference,
		CustomerID:            offer.CustomerID,
		CustomerName:          offer.CustomerName,
		ProjectID:             offer.ProjectID,
		CompanyID:             offer.CompanyID,
		Phase:                 offer.Phase,
		Probability:           offer.Probability,
		Value:                 offer.Value,
		Status:                offer.Status,
		CreatedAt:             offer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:             offer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		ResponsibleUserID:     offer.ResponsibleUserID,
		ResponsibleUserName:   offer.ResponsibleUserName,
		Items:                 items,
		Description:           offer.Description,
		Notes:                 offer.Notes,
		DueDate:               dueDate,
		Cost:                  offer.Cost,
		Price:                 offer.Price,
		Margin:                margin,
		MarginPercent:         offer.MarginPercent, // Stored in DB, auto-calculated by trigger
		Location:              offer.Location,
		SentDate:              sentDate,
		ExpirationDate:        expirationDate,
		CustomerHasWonProject: offer.CustomerHasWonProject,
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
		CreatedAt:       item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
		CreatedAt:        cost.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        cost.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if cost.PostingDate != nil {
		dto.PostingDate = cost.PostingDate.Format("2006-01-02")
	}

	if cost.ERPSyncedAt != nil {
		dto.ERPSyncedAt = cost.ERPSyncedAt.Format("2006-01-02T15:04:05Z")
	}

	if cost.ApprovedAt != nil {
		dto.ApprovedAt = cost.ApprovedAt.Format("2006-01-02T15:04:05Z")
	}

	return dto
}

// ToProjectCostSummaryDTO creates a cost summary DTO for a project
func ToProjectCostSummaryDTO(project *domain.Project, actualCosts []domain.ProjectActualCost) domain.ProjectCostSummaryDTO {
	totalActualCosts := 0.0
	for _, cost := range actualCosts {
		totalActualCosts += cost.Amount
	}

	remainingBudget := project.Budget - totalActualCosts
	budgetUsedPercent := 0.0
	if project.Budget > 0 {
		budgetUsedPercent = (totalActualCosts / project.Budget) * 100
	}

	return domain.ProjectCostSummaryDTO{
		ProjectID:         project.ID,
		ProjectName:       project.Name,
		Budget:            project.Budget,
		Spent:             project.Spent,
		ActualCosts:       totalActualCosts,
		RemainingBudget:   remainingBudget,
		BudgetUsedPercent: budgetUsedPercent,
		CostEntryCount:    len(actualCosts),
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
		CreatedAt:  notification.CreatedAt.Format("2006-01-02T15:04:05Z"),
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

func UpdateProjectDenormalizedFields(project *domain.Project, customerName, managerName string) {
	project.CustomerName = customerName
	project.ManagerName = managerName
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
		CreatedAt:   file.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToActivityDTO converts Activity to ActivityDTO
func ToActivityDTO(activity *domain.Activity) domain.ActivityDTO {
	dto := domain.ActivityDTO{
		ID:               activity.ID,
		TargetType:       activity.TargetType,
		TargetID:         activity.TargetID,
		Title:            activity.Title,
		Body:             activity.Body,
		OccurredAt:       activity.OccurredAt.Format("2006-01-02T15:04:05Z"),
		CreatorName:      activity.CreatorName,
		CreatedAt:        activity.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
		dto.ScheduledAt = activity.ScheduledAt.Format("2006-01-02T15:04:05Z")
	}

	if activity.DueDate != nil {
		dto.DueDate = activity.DueDate.Format("2006-01-02")
	}

	if activity.CompletedAt != nil {
		dto.CompletedAt = activity.CompletedAt.Format("2006-01-02T15:04:05Z")
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
		GrantedAt: role.GrantedAt.Format("2006-01-02T15:04:05Z"),
		IsActive:  role.IsActive,
	}

	if role.ExpiresAt != nil {
		dto.ExpiresAt = role.ExpiresAt.Format("2006-01-02T15:04:05Z")
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
		GrantedAt:  perm.GrantedAt.Format("2006-01-02T15:04:05Z"),
		Reason:     perm.Reason,
	}

	if perm.ExpiresAt != nil {
		dto.ExpiresAt = perm.ExpiresAt.Format("2006-01-02T15:04:05Z")
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
		PerformedAt: log.PerformedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToProjectBudgetDTO converts project budget info to DTO
func ToProjectBudgetDTO(project *domain.Project) domain.ProjectBudgetDTO {
	remaining := project.Budget - project.Spent
	percentUsed := 0.0
	if project.Budget > 0 {
		percentUsed = (project.Spent / project.Budget) * 100
	}
	return domain.ProjectBudgetDTO{
		Budget:      project.Budget,
		Spent:       project.Spent,
		Remaining:   remaining,
		PercentUsed: percentUsed,
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
	dto := domain.ProjectWithDetailsDTO{
		ProjectDTO:    ToProjectDTO(project),
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
		CreatedAt:                   company.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:                   company.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// FormatError creates a formatted error message
func FormatError(entity, operation string, err error) error {
	return fmt.Errorf("failed to %s %s: %w", operation, entity, err)
}
