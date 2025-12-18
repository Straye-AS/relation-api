package datawarehouse_test

import (
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
		name     string
		cfg      *config.DataWarehouseConfig
		wantNil  bool
		wantErr  bool
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
			wantTable: "dbo.nxt_strayetak_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "stalbygg company",
			companyID: "stalbygg",
			wantTable: "dbo.nxt_strayestaal_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "montasje company",
			companyID: "montasje",
			wantTable: "dbo.nxt_strayemontage_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "hybridbygg company",
			companyID: "hybridbygg",
			wantTable: "dbo.nxt_strayehybridbygg_generalledgertransaction",
			wantErr:   false,
		},
		{
			name:      "industri company",
			companyID: "industri",
			wantTable: "dbo.nxt_strayeindustri_generalledgertransaction",
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
		"montasje":   "strayemontage",
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
	status := nilClient.HealthCheck(nil)
	assert.NotNil(t, status)
	assert.Equal(t, "disabled", status.Status)
}
