package service_test

import (
	"testing"

	"github.com/straye-as/relation-api/internal/domain"
)

// TestGetCompanyPrefix tests the company prefix mapping
func TestGetCompanyPrefix(t *testing.T) {
	tests := []struct {
		name      string
		companyID domain.CompanyID
		expected  string
	}{
		{
			name:      "stalbygg returns ST",
			companyID: domain.CompanyStalbygg,
			expected:  "ST",
		},
		{
			name:      "hybridbygg returns HB",
			companyID: domain.CompanyHybridbygg,
			expected:  "HB",
		},
		{
			name:      "industri returns IN",
			companyID: domain.CompanyIndustri,
			expected:  "IN",
		},
		{
			name:      "tak returns TK",
			companyID: domain.CompanyTak,
			expected:  "TK",
		},
		{
			name:      "montasje returns MO",
			companyID: domain.CompanyMontasje,
			expected:  "MO",
		},
		{
			name:      "gruppen returns GR",
			companyID: domain.CompanyGruppen,
			expected:  "GR",
		},
		{
			name:      "unknown company defaults to GR",
			companyID: domain.CompanyID("unknown"),
			expected:  "GR",
		},
		{
			name:      "empty company defaults to GR",
			companyID: domain.CompanyID(""),
			expected:  "GR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.GetCompanyPrefix(tc.companyID)
			if result != tc.expected {
				t.Errorf("GetCompanyPrefix(%q) = %q, want %q", tc.companyID, result, tc.expected)
			}
		})
	}
}

// TestIsValidCompanyID tests the company ID validation
func TestIsValidCompanyID(t *testing.T) {
	tests := []struct {
		name      string
		companyID string
		expected  bool
	}{
		{
			name:      "valid stalbygg",
			companyID: "stalbygg",
			expected:  true,
		},
		{
			name:      "valid hybridbygg",
			companyID: "hybridbygg",
			expected:  true,
		},
		{
			name:      "valid industri",
			companyID: "industri",
			expected:  true,
		},
		{
			name:      "valid tak",
			companyID: "tak",
			expected:  true,
		},
		{
			name:      "valid montasje",
			companyID: "montasje",
			expected:  true,
		},
		{
			name:      "valid gruppen",
			companyID: "gruppen",
			expected:  true,
		},
		{
			name:      "invalid unknown",
			companyID: "unknown",
			expected:  false,
		},
		{
			name:      "invalid empty",
			companyID: "",
			expected:  false,
		},
		{
			name:      "invalid uppercase",
			companyID: "STALBYGG",
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.IsValidCompanyID(tc.companyID)
			if result != tc.expected {
				t.Errorf("IsValidCompanyID(%q) = %v, want %v", tc.companyID, result, tc.expected)
			}
		})
	}
}

// TestNumberFormatting tests the expected number format
func TestNumberFormatting(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		year     int
		sequence int
		expected string
	}{
		{
			name:     "single digit sequence",
			prefix:   "ST",
			year:     2025,
			sequence: 1,
			expected: "ST-2025-001",
		},
		{
			name:     "double digit sequence",
			prefix:   "HB",
			year:     2025,
			sequence: 42,
			expected: "HB-2025-042",
		},
		{
			name:     "triple digit sequence",
			prefix:   "IN",
			year:     2025,
			sequence: 123,
			expected: "IN-2025-123",
		},
		{
			name:     "large sequence (no padding)",
			prefix:   "TK",
			year:     2025,
			sequence: 1000,
			expected: "TK-2025-1000",
		},
		{
			name:     "different year",
			prefix:   "MO",
			year:     2024,
			sequence: 5,
			expected: "MO-2024-005",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the format used in the service
			result := formatNumber(tc.prefix, tc.year, tc.sequence)
			if result != tc.expected {
				t.Errorf("formatNumber(%q, %d, %d) = %q, want %q", tc.prefix, tc.year, tc.sequence, result, tc.expected)
			}
		})
	}
}

// Helper function that mirrors the service's format logic
func formatNumber(prefix string, year, sequence int) string {
	return prefix + "-" + itoa(year) + "-" + padLeft(itoa(sequence), 3, '0')
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

func padLeft(s string, length int, pad rune) string {
	for len(s) < length {
		s = string(pad) + s
	}
	return s
}
