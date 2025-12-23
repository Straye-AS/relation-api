package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
)

func TestFile_GetEntityType(t *testing.T) {
	offerID := uuid.New()
	customerID := uuid.New()
	projectID := uuid.New()
	supplierID := uuid.New()

	tests := []struct {
		name     string
		file     domain.File
		expected string
	}{
		{
			name: "returns customer when CustomerID is set",
			file: domain.File{
				CustomerID: &customerID,
			},
			expected: "customer",
		},
		{
			name: "returns project when ProjectID is set",
			file: domain.File{
				ProjectID: &projectID,
			},
			expected: "project",
		},
		{
			name: "returns offer when OfferID is set",
			file: domain.File{
				OfferID: &offerID,
			},
			expected: "offer",
		},
		{
			name: "returns supplier when SupplierID is set",
			file: domain.File{
				SupplierID: &supplierID,
			},
			expected: "supplier",
		},
		{
			name:     "returns empty string when no entity ID is set",
			file:     domain.File{},
			expected: "",
		},
		{
			name: "customer takes precedence over project",
			file: domain.File{
				CustomerID: &customerID,
				ProjectID:  &projectID,
			},
			expected: "customer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.file.GetEntityType()
			if result != tt.expected {
				t.Errorf("GetEntityType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFile_GetEntityID(t *testing.T) {
	offerID := uuid.New()
	customerID := uuid.New()
	projectID := uuid.New()
	supplierID := uuid.New()

	tests := []struct {
		name     string
		file     domain.File
		expected *uuid.UUID
	}{
		{
			name: "returns CustomerID when set",
			file: domain.File{
				CustomerID: &customerID,
			},
			expected: &customerID,
		},
		{
			name: "returns ProjectID when set",
			file: domain.File{
				ProjectID: &projectID,
			},
			expected: &projectID,
		},
		{
			name: "returns OfferID when set",
			file: domain.File{
				OfferID: &offerID,
			},
			expected: &offerID,
		},
		{
			name: "returns SupplierID when set",
			file: domain.File{
				SupplierID: &supplierID,
			},
			expected: &supplierID,
		},
		{
			name:     "returns nil when no entity ID is set",
			file:     domain.File{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.file.GetEntityID()
			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetEntityID() = %v, want nil", result)
				}
			} else if result == nil {
				t.Errorf("GetEntityID() = nil, want %v", *tt.expected)
			} else if *result != *tt.expected {
				t.Errorf("GetEntityID() = %v, want %v", *result, *tt.expected)
			}
		})
	}
}

func TestFile_HasExactlyOneEntityID(t *testing.T) {
	offerID := uuid.New()
	customerID := uuid.New()
	projectID := uuid.New()
	supplierID := uuid.New()

	tests := []struct {
		name     string
		file     domain.File
		expected bool
	}{
		{
			name: "true when only OfferID is set",
			file: domain.File{
				OfferID: &offerID,
			},
			expected: true,
		},
		{
			name: "true when only CustomerID is set",
			file: domain.File{
				CustomerID: &customerID,
			},
			expected: true,
		},
		{
			name: "true when only ProjectID is set",
			file: domain.File{
				ProjectID: &projectID,
			},
			expected: true,
		},
		{
			name: "true when only SupplierID is set",
			file: domain.File{
				SupplierID: &supplierID,
			},
			expected: true,
		},
		{
			name:     "false when no entity ID is set",
			file:     domain.File{},
			expected: false,
		},
		{
			name: "false when two entity IDs are set",
			file: domain.File{
				OfferID:    &offerID,
				CustomerID: &customerID,
			},
			expected: false,
		},
		{
			name: "false when three entity IDs are set",
			file: domain.File{
				OfferID:    &offerID,
				CustomerID: &customerID,
				ProjectID:  &projectID,
			},
			expected: false,
		},
		{
			name: "false when all entity IDs are set",
			file: domain.File{
				OfferID:    &offerID,
				CustomerID: &customerID,
				ProjectID:  &projectID,
				SupplierID: &supplierID,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.file.HasExactlyOneEntityID()
			if result != tt.expected {
				t.Errorf("HasExactlyOneEntityID() = %v, want %v", result, tt.expected)
			}
		})
	}
}
