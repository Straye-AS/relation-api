package datawarehouse_test

import (
	"context"
	"testing"

	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/datawarehouse"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewClient_DisabledConfig(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	client, err := datawarehouse.NewClient(nil, logger)
	assert.NoError(t, err)
	assert.Nil(t, client)

	// Test with disabled config
	cfg := &config.DataWarehouseConfig{
		Enabled: false,
	}
	client, err = datawarehouse.NewClient(cfg, logger)
	assert.NoError(t, err)
	assert.Nil(t, client)
}

func TestNewClient_MissingCredentials(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		cfg     *config.DataWarehouseConfig
		wantNil bool
		wantErr bool
	}{
		{
			name: "missing URL",
			cfg: &config.DataWarehouseConfig{
				Enabled:  true,
				URL:      "",
				User:     "user",
				Password: "pass",
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "missing user",
			cfg: &config.DataWarehouseConfig{
				Enabled:  true,
				URL:      "host:1433/db",
				User:     "",
				Password: "pass",
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "missing password",
			cfg: &config.DataWarehouseConfig{
				Enabled:  true,
				URL:      "host:1433/db",
				User:     "user",
				Password: "",
			},
			wantNil: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := datawarehouse.NewClient(tt.cfg, logger)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.Nil(t, client)
			}
		})
	}
}

func TestGetGeneralLedgerTableName(t *testing.T) {
	tests := []struct {
		name      string
		companyID string
		wantTable string
		wantErr   bool
	}{
		{
			name:      "tak company",
			companyID: "tak",
			wantTable: "nxt_strayetak_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "stalbygg company",
			companyID: "stalbygg",
			wantTable: "nxt_strayestaal_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "montasje company",
			companyID: "montasje",
			wantTable: "nxt_strayemontasje_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "hybridbygg company",
			companyID: "hybridbygg",
			wantTable: "nxt_strayehybridbygg_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "industri company",
			companyID: "industri",
			wantTable: "nxt_strayeindustri_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "unknown company",
			companyID: "unknown",
			wantTable: "",
			wantErr:   true,
		},
		{
			name:      "gruppen company (not in warehouse)",
			companyID: "gruppen",
			wantTable: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, err := datawarehouse.GetGeneralLedgerTableName(tt.companyID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, tableName)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTable, tableName)
			}
		})
	}
}

func TestCompanyMapping(t *testing.T) {
	// Verify all expected mappings exist
	expectedMappings := map[string]string{
		"tak":        "strayetak",
		"stalbygg":   "strayestaal",
		"montasje":   "strayemontasje",
		"hybridbygg": "strayehybridbygg",
		"industri":   "strayeindustri",
	}

	for companyID, expectedPrefix := range expectedMappings {
		prefix, ok := datawarehouse.CompanyMapping[companyID]
		assert.True(t, ok, "company %s should exist in mapping", companyID)
		assert.Equal(t, expectedPrefix, prefix, "company %s should map to %s", companyID, expectedPrefix)
	}

	// Verify there are exactly 5 mappings
	assert.Len(t, datawarehouse.CompanyMapping, 5)
}

func TestClient_IsEnabled(t *testing.T) {
	// Nil client should return false
	var nilClient *datawarehouse.Client
	assert.False(t, nilClient.IsEnabled())
}

func TestClient_Close_NilClient(t *testing.T) {
	// Nil client close should not panic
	var nilClient *datawarehouse.Client
	err := nilClient.Close()
	assert.NoError(t, err)
}

func TestClient_HealthCheck_NilClient(t *testing.T) {
	// Nil client health check should return disabled status
	var nilClient *datawarehouse.Client
	status := nilClient.HealthCheck(context.Background())
	assert.NotNil(t, status)
	assert.Equal(t, "disabled", status.Status)
}

func TestAccountConstants(t *testing.T) {
	// Verify the account range constants are set correctly
	assert.Equal(t, 3000, datawarehouse.IncomeAccountMin)
	assert.Equal(t, 3999, datawarehouse.IncomeAccountMax)
}

func TestIsIncomeAccount(t *testing.T) {
	tests := []struct {
		name      string
		accountNo int
		want      bool
	}{
		{
			name:      "below income range",
			accountNo: 2999,
			want:      false,
		},
		{
			name:      "at income range minimum",
			accountNo: 3000,
			want:      true,
		},
		{
			name:      "within income range",
			accountNo: 3500,
			want:      true,
		},
		{
			name:      "at income range maximum",
			accountNo: 3999,
			want:      true,
		},
		{
			name:      "above income range",
			accountNo: 4000,
			want:      false,
		},
		{
			name:      "typical cost account (low)",
			accountNo: 1000,
			want:      false,
		},
		{
			name:      "typical cost account (high)",
			accountNo: 5000,
			want:      false,
		},
		{
			name:      "zero account",
			accountNo: 0,
			want:      false,
		},
		{
			name:      "negative account",
			accountNo: -1,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := datawarehouse.IsIncomeAccount(tt.accountNo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsCostAccount(t *testing.T) {
	tests := []struct {
		name      string
		accountNo int
		want      bool
	}{
		{
			name:      "below income range - is cost",
			accountNo: 2999,
			want:      true,
		},
		{
			name:      "at income range minimum - not cost",
			accountNo: 3000,
			want:      false,
		},
		{
			name:      "within income range - not cost",
			accountNo: 3500,
			want:      false,
		},
		{
			name:      "at income range maximum - not cost",
			accountNo: 3999,
			want:      false,
		},
		{
			name:      "above income range - is cost",
			accountNo: 4000,
			want:      true,
		},
		{
			name:      "typical cost account (low)",
			accountNo: 1000,
			want:      true,
		},
		{
			name:      "typical cost account (high)",
			accountNo: 7000,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := datawarehouse.IsCostAccount(tt.accountNo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsIncomeAccount_IsCostAccount_Inverse(t *testing.T) {
	// Verify that IsIncomeAccount and IsCostAccount are always inverses
	testAccounts := []int{0, 1000, 2999, 3000, 3500, 3999, 4000, 5000, 10000}

	for _, accountNo := range testAccounts {
		isIncome := datawarehouse.IsIncomeAccount(accountNo)
		isCost := datawarehouse.IsCostAccount(accountNo)
		assert.NotEqual(t, isIncome, isCost,
			"Account %d: IsIncomeAccount(%v) and IsCostAccount(%v) should be inverses",
			accountNo, isIncome, isCost)
	}
}

func TestClient_GetProjectIncome_NilClient(t *testing.T) {
	var nilClient *datawarehouse.Client
	_, err := nilClient.GetProjectIncome(context.Background(), "tak", "PROJECT-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestClient_GetProjectCosts_NilClient(t *testing.T) {
	var nilClient *datawarehouse.Client
	_, err := nilClient.GetProjectCosts(context.Background(), "tak", "PROJECT-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestClient_GetProjectFinancials_NilClient(t *testing.T) {
	var nilClient *datawarehouse.Client
	_, err := nilClient.GetProjectFinancials(context.Background(), "tak", "PROJECT-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}
