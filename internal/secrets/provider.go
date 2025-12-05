package secrets

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// SecretSource defines where secrets are loaded from
type SecretSource string

const (
	// SourceEnvironment loads secrets from environment variables
	SourceEnvironment SecretSource = "environment"
	// SourceVault loads secrets from Azure Key Vault
	SourceVault SecretSource = "vault"
	// SourceAuto automatically determines source based on environment
	// Uses vault in staging/production, environment in development
	SourceAuto SecretSource = "auto"
)

// Provider abstracts secret retrieval from different sources
type Provider struct {
	source      SecretSource
	vaultClient *VaultClient
	logger      *zap.Logger
	environment string
}

// ProviderConfig holds configuration for the secrets provider
type ProviderConfig struct {
	Source       SecretSource
	VaultName    string
	Environment  string // "development", "staging", "production"
	CacheEnabled bool
	CacheTTL     time.Duration
}

// NewProvider creates a new secrets provider
func NewProvider(cfg *ProviderConfig, logger *zap.Logger) (*Provider, error) {
	source := cfg.Source

	// Resolve "auto" source based on environment
	if source == SourceAuto {
		switch cfg.Environment {
		case "development", "local", "":
			source = SourceEnvironment
			logger.Info("Auto-detected secret source for development",
				zap.String("source", string(source)),
				zap.String("environment", cfg.Environment),
			)
		default:
			// staging, production, or any other environment uses vault
			source = SourceVault
			logger.Info("Auto-detected secret source for non-development",
				zap.String("source", string(source)),
				zap.String("environment", cfg.Environment),
			)
		}
	}

	provider := &Provider{
		source:      source,
		logger:      logger,
		environment: cfg.Environment,
	}

	// Initialize vault client if using vault source
	if source == SourceVault {
		if cfg.VaultName == "" {
			return nil, fmt.Errorf("vault name required when using vault secret source")
		}

		vaultClient, err := NewVaultClient(&VaultConfig{
			VaultName:    cfg.VaultName,
			CacheEnabled: cfg.CacheEnabled,
			CacheTTL:     cfg.CacheTTL,
		}, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize vault client: %w", err)
		}
		provider.vaultClient = vaultClient
	}

	logger.Info("Secrets provider initialized",
		zap.String("source", string(source)),
		zap.String("environment", cfg.Environment),
	)

	return provider, nil
}

// GetSecret retrieves a secret by name
// For vault source, secretName is the Key Vault secret name
// For environment source, secretName is the environment variable name
func (p *Provider) GetSecret(ctx context.Context, secretName string) (string, error) {
	switch p.source {
	case SourceEnvironment:
		value := os.Getenv(secretName)
		if value == "" {
			return "", fmt.Errorf("environment variable '%s' not set", secretName)
		}
		return value, nil

	case SourceVault:
		if p.vaultClient == nil {
			return "", fmt.Errorf("vault client not initialized")
		}
		return p.vaultClient.GetSecret(ctx, secretName)

	default:
		return "", fmt.Errorf("unknown secret source: %s", p.source)
	}
}

// GetSecretWithDefault retrieves a secret, returning defaultValue if not found
func (p *Provider) GetSecretWithDefault(ctx context.Context, secretName, defaultValue string) string {
	value, err := p.GetSecret(ctx, secretName)
	if err != nil {
		p.logger.Debug("Using default value for secret",
			zap.String("secret_name", secretName),
			zap.String("source", string(p.source)),
		)
		return defaultValue
	}
	return value
}

// GetSecretOrEnv tries to get from configured source, falls back to environment variable
// Useful for secrets that can be overridden by environment variables even in vault mode
func (p *Provider) GetSecretOrEnv(ctx context.Context, secretName, envName string) (string, error) {
	// First check if environment variable is explicitly set (override)
	if envValue := os.Getenv(envName); envValue != "" {
		p.logger.Debug("Using environment variable override",
			zap.String("env_name", envName),
		)
		return envValue, nil
	}

	// Then try the configured source
	return p.GetSecret(ctx, secretName)
}

// GetSecretOrEnvWithDefault combines GetSecretOrEnv with a default fallback
func (p *Provider) GetSecretOrEnvWithDefault(ctx context.Context, secretName, envName, defaultValue string) string {
	value, err := p.GetSecretOrEnv(ctx, secretName, envName)
	if err != nil {
		p.logger.Debug("Using default value",
			zap.String("secret_name", secretName),
			zap.String("env_name", envName),
		)
		return defaultValue
	}
	return value
}

// Source returns the current secret source
func (p *Provider) Source() SecretSource {
	return p.source
}

// IsVaultEnabled returns true if secrets are loaded from vault
func (p *Provider) IsVaultEnabled() bool {
	return p.source == SourceVault
}
