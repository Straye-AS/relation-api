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
		ID:            deal.ID,
		Title:         deal.Title,
		Description:   deal.Description,
		CustomerID:    deal.CustomerID,
		CompanyID:     deal.CompanyID,
		CustomerName:  deal.CustomerName,
		Stage:         deal.Stage,
		Probability:   deal.Probability,
		Value:         deal.Value,
		WeightedValue: deal.WeightedValue,
		Currency:      deal.Currency,
		OwnerID:       deal.OwnerID,
		OwnerName:     deal.OwnerName,
		Source:        deal.Source,
		Notes:         deal.Notes,
		LostReason:    deal.LostReason,
		CreatedAt:     deal.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     deal.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
	dto := domain.ProjectDTO{
		ID:                project.ID,
		Name:              project.Name,
		Summary:           project.Summary,
		Description:       project.Description,
		CustomerID:        project.CustomerID,
		CustomerName:      project.CustomerName,
		CompanyID:         project.CompanyID,
		Status:            project.Status,
		StartDate:         project.StartDate.Format("2006-01-02T15:04:05Z"),
		Budget:            project.Budget,
		Spent:             project.Spent,
		ManagerID:         project.ManagerID,
		ManagerName:       project.ManagerName,
		TeamMembers:       project.TeamMembers,
		CreatedAt:         project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         project.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		OfferID:           project.OfferID,
		DealID:            project.DealID,
		HasDetailedBudget: project.HasDetailedBudget,
		Health:            project.Health,
		CompletionPercent: project.CompletionPercent,
	}

	if project.EndDate != nil {
		dto.EndDate = project.EndDate.Format("2006-01-02T15:04:05Z")
	}

	if project.EstimatedCompletionDate != nil {
		dto.EstimatedCompletionDate = project.EstimatedCompletionDate.Format("2006-01-02")
	}

	return dto
}

// ToOfferDTO converts Offer to OfferDTO
func ToOfferDTO(offer *domain.Offer) domain.OfferDTO {
	items := make([]domain.OfferItemDTO, len(offer.Items))
	for i, item := range offer.Items {
		items[i] = ToOfferItemDTO(&item)
	}

	return domain.OfferDTO{
		ID:                  offer.ID,
		Title:               offer.Title,
		CustomerID:          offer.CustomerID,
		CustomerName:        offer.CustomerName,
		CompanyID:           offer.CompanyID,
		Phase:               offer.Phase,
		Probability:         offer.Probability,
		Value:               offer.Value,
		Status:              offer.Status,
		CreatedAt:           offer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           offer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		ResponsibleUserID:   offer.ResponsibleUserID,
		ResponsibleUserName: offer.ResponsibleUserName,
		Items:               items,
		Description:         offer.Description,
		Notes:               offer.Notes,
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

// ToBudgetDimensionCategoryDTO converts BudgetDimensionCategory to BudgetDimensionCategoryDTO
func ToBudgetDimensionCategoryDTO(cat *domain.BudgetDimensionCategory) domain.BudgetDimensionCategoryDTO {
	return domain.BudgetDimensionCategoryDTO{
		ID:           cat.ID,
		Name:         cat.Name,
		Description:  cat.Description,
		DisplayOrder: cat.DisplayOrder,
		IsActive:     cat.IsActive,
	}
}

// ToBudgetDimensionDTO converts BudgetDimension to BudgetDimensionDTO
func ToBudgetDimensionDTO(dim *domain.BudgetDimension) domain.BudgetDimensionDTO {
	dto := domain.BudgetDimensionDTO{
		ID:                  dim.ID,
		ParentType:          dim.ParentType,
		ParentID:            dim.ParentID,
		CategoryID:          dim.CategoryID,
		CustomName:          dim.CustomName,
		Name:                dim.GetName(),
		Cost:                dim.Cost,
		Revenue:             dim.Revenue,
		TargetMarginPercent: dim.TargetMarginPercent,
		MarginOverride:      dim.MarginOverride,
		MarginPercent:       dim.MarginPercent,
		Description:         dim.Description,
		Quantity:            dim.Quantity,
		Unit:                dim.Unit,
		DisplayOrder:        dim.DisplayOrder,
		CreatedAt:           dim.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           dim.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if dim.Category != nil {
		catDTO := ToBudgetDimensionCategoryDTO(dim.Category)
		dto.Category = &catDTO
	}

	return dto
}

// ToBudgetSummaryDTO creates a summary DTO from budget dimensions
func ToBudgetSummaryDTO(parentType domain.BudgetParentType, parentID uuid.UUID, dimensions []domain.BudgetDimension) domain.BudgetSummaryDTO {
	totalCost := 0.0
	totalRevenue := 0.0

	for _, dim := range dimensions {
		totalCost += dim.Cost
		totalRevenue += dim.Revenue
	}

	overallMargin := 0.0
	if totalRevenue > 0 {
		overallMargin = ((totalRevenue - totalCost) / totalRevenue) * 100
	}

	return domain.BudgetSummaryDTO{
		ParentType:           parentType,
		ParentID:             parentID,
		DimensionCount:       len(dimensions),
		TotalCost:            totalCost,
		TotalRevenue:         totalRevenue,
		OverallMarginPercent: overallMargin,
		TotalProfit:          totalRevenue - totalCost,
	}
}

// ToProjectActualCostDTO converts ProjectActualCost to ProjectActualCostDTO
func ToProjectActualCostDTO(cost *domain.ProjectActualCost) domain.ProjectActualCostDTO {
	dto := domain.ProjectActualCostDTO{
		ID:                cost.ID,
		ProjectID:         cost.ProjectID,
		CostType:          cost.CostType,
		Description:       cost.Description,
		Amount:            cost.Amount,
		Currency:          cost.Currency,
		CostDate:          cost.CostDate.Format("2006-01-02"),
		BudgetDimensionID: cost.BudgetDimensionID,
		ERPSource:         cost.ERPSource,
		ERPReference:      cost.ERPReference,
		ERPTransactionID:  cost.ERPTransactionID,
		IsApproved:        cost.IsApproved,
		ApprovedByID:      cost.ApprovedByID,
		Notes:             cost.Notes,
		CreatedAt:         cost.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         cost.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
	return domain.ActivityDTO{
		ID:          activity.ID,
		TargetType:  activity.TargetType,
		TargetID:    activity.TargetID,
		Title:       activity.Title,
		Body:        activity.Body,
		OccurredAt:  activity.OccurredAt.Format("2006-01-02T15:04:05Z"),
		CreatorName: activity.CreatorName,
		CreatedAt:   activity.CreatedAt.Format("2006-01-02T15:04:05Z"),
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

// FormatError creates a formatted error message
func FormatError(entity, operation string, err error) error {
	return fmt.Errorf("failed to %s %s: %w", operation, entity, err)
}
