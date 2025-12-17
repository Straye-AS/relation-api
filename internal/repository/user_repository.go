package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	// Use case-insensitive lookup since Azure AD may return different casing
	err := r.db.WithContext(ctx).First(&user, "LOWER(email) = LOWER(?)", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByStringID retrieves a user by their string ID (e.g., Azure AD object ID)
func (r *UserRepository) GetByStringID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Upsert(ctx context.Context, user *domain.User) error {
	var existing domain.User
	// Use case-insensitive email lookup since Azure AD may return different casing
	err := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?)", user.Email).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// New user - set azure_ad_oid to match ID (the Azure AD OID from JWT)
		if user.AzureADOID == "" {
			user.AzureADOID = user.ID
		}
		// Normalize email to lowercase to prevent case-sensitivity issues
		user.Email = strings.ToLower(user.Email)
		return r.db.WithContext(ctx).Create(user).Error
	}

	if err != nil {
		return err
	}

	// Update only specific fields from login, preserving manually-assigned roles and company
	updates := map[string]interface{}{
		"name":            user.DisplayName,
		"last_login_at":   user.LastLoginAt,
		"last_ip_address": user.LastIPAddress,
		"azure_ad_roles":  user.AzureADRoles,
	}

	// Set azure_ad_oid if not already set (the user.ID contains the Azure AD OID from JWT)
	if existing.AzureADOID == "" && user.ID != "" {
		updates["azure_ad_oid"] = user.ID
	}

	// Only update these fields if they have values (don't overwrite with empty)
	if user.Department != "" {
		updates["department"] = user.Department
	}
	if user.FirstName != "" {
		updates["first_name"] = user.FirstName
	}
	if user.LastName != "" {
		updates["last_name"] = user.LastName
	}

	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", existing.ID).Updates(updates).Error
}
