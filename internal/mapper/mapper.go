package mapper

import (
	"fmt"

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
	return domain.ContactDTO{
		ID:           contact.ID,
		Name:         contact.Name,
		Email:        contact.Email,
		Phone:        contact.Phone,
		Role:         contact.Role,
		CustomerID:   contact.CustomerID,
		CustomerName: contact.CustomerName,
		ProjectID:    contact.ProjectID,
		ProjectName:  contact.ProjectName,
		CreatedAt:    contact.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    contact.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToProjectDTO converts Project to ProjectDTO
func ToProjectDTO(project *domain.Project) domain.ProjectDTO {
	dto := domain.ProjectDTO{
		ID:              project.ID,
		Name:            project.Name,
		Summary:         project.Summary,
		Description:     project.Description,
		CustomerID:      project.CustomerID,
		CustomerName:    project.CustomerName,
		CompanyID:       project.CompanyID,
		Status:          project.Status,
		StartDate:       project.StartDate.Format("2006-01-02T15:04:05Z"),
		Budget:          project.Budget,
		Spent:           project.Spent,
		ManagerID:       project.ManagerID,
		ManagerName:     project.ManagerName,
		TeamMembers:     project.TeamMembers,
		TeamsChannelID:  project.TeamsChannelID,
		TeamsChannelURL: project.TeamsChannelURL,
		CreatedAt:       project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       project.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		OfferID:         project.OfferID,
	}

	if project.EndDate != nil {
		dto.EndDate = project.EndDate.Format("2006-01-02T15:04:05Z")
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

func UpdateContactDenormalizedFields(contact *domain.Contact, customerName, projectName string) {
	if contact.CustomerID != nil && customerName != "" {
		contact.CustomerName = customerName
	}
	if contact.ProjectID != nil && projectName != "" {
		contact.ProjectName = projectName
	}
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
