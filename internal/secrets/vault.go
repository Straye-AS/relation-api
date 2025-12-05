package secrets

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"go.uber.org/zap"
)

// VaultClient wraps Azure Key Vault client for secret retrieval
type VaultClient struct {
	client       *azsecrets.Client
	vaultName    string
	logger       *zap.Logger
	cache        map[string]cachedSecret
	cacheTTL     time.Duration
	cacheEnabled bool
}

type cachedSecret struct {
	value     string
	expiresAt time.Time
}

// VaultConfig holds configuration for the vault client
type VaultConfig struct {
	VaultName    string
	CacheEnabled bool
	CacheTTL     time.Duration
}

// NewVaultClient creates a new Azure Key Vault client
// Uses DefaultAzureCredential which supports:
// - Environment variables (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
// - Managed Identity (when running in Azure)
// - Azure CLI credentials (for local development)
// - Visual Studio Code credentials
func NewVaultClient(cfg *VaultConfig, logger *zap.Logger) (*VaultClient, error) {
	if cfg.VaultName == "" {
		return nil, fmt.Errorf("vault name is required")
	}

	logger.Info("Initializing Azure Key Vault client",
		zap.String("vault_name", cfg.VaultName),
		zap.Bool("cache_enabled", cfg.CacheEnabled),
	)

	// Create credential using DefaultAzureCredential
	// This will try multiple authentication methods in order
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		logger.Error("Failed to create Azure credential", zap.Error(err))
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Build vault URL
	vaultURL := fmt.Sprintf("https://%s.vault.azure.net/", cfg.VaultName)

	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		logger.Error("Failed to create Key Vault client", zap.Error(err))
		return nil, fmt.Errorf("failed to create Key Vault client: %w", err)
	}

	cacheTTL := cfg.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute // Default cache TTL
	}

	logger.Info("Azure Key Vault client initialized successfully",
		zap.String("vault_url", vaultURL),
	)

	return &VaultClient{
		client:       client,
		vaultName:    cfg.VaultName,
		logger:       logger,
		cache:        make(map[string]cachedSecret),
		cacheTTL:     cacheTTL,
		cacheEnabled: cfg.CacheEnabled,
	}, nil
}

// GetSecret retrieves a secret from Azure Key Vault
func (v *VaultClient) GetSecret(ctx context.Context, secretName string) (string, error) {
	// Check cache first if enabled
	if v.cacheEnabled {
		if cached, ok := v.cache[secretName]; ok {
			if time.Now().Before(cached.expiresAt) {
				v.logger.Debug("Secret retrieved from cache", zap.String("secret_name", secretName))
				return cached.value, nil
			}
			// Cache expired, remove it
			delete(v.cache, secretName)
		}
	}

	v.logger.Debug("Fetching secret from Key Vault", zap.String("secret_name", secretName))

	resp, err := v.client.GetSecret(ctx, secretName, "", nil)
	if err != nil {
		v.logger.Error("Failed to get secret from Key Vault",
			zap.String("secret_name", secretName),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to get secret '%s': %w", secretName, err)
	}

	if resp.Value == nil {
		return "", fmt.Errorf("secret '%s' has no value", secretName)
	}

	value := *resp.Value

	// Cache the secret if caching is enabled
	if v.cacheEnabled {
		v.cache[secretName] = cachedSecret{
			value:     value,
			expiresAt: time.Now().Add(v.cacheTTL),
		}
	}

	v.logger.Debug("Secret retrieved successfully", zap.String("secret_name", secretName))
	return value, nil
}

// GetSecretWithDefault retrieves a secret, returning defaultValue if not found or on error
func (v *VaultClient) GetSecretWithDefault(ctx context.Context, secretName, defaultValue string) string {
	value, err := v.GetSecret(ctx, secretName)
	if err != nil {
		v.logger.Warn("Failed to get secret, using default",
			zap.String("secret_name", secretName),
			zap.Error(err),
		)
		return defaultValue
	}
	return value
}

// ClearCache clears all cached secrets
func (v *VaultClient) ClearCache() {
	v.cache = make(map[string]cachedSecret)
	v.logger.Debug("Secret cache cleared")
}
