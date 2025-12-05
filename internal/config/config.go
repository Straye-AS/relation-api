package config

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/straye-as/relation-api/internal/secrets"
	"go.uber.org/zap"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	AzureAd  AzureAdConfig
	ApiKey   ApiKeyConfig
	Storage  StorageConfig
	Secrets  SecretsConfig
	Logging  LoggingConfig
	Server   ServerConfig
}

type AppConfig struct {
	Name        string
	Environment string
	Port        int
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type AzureAdConfig struct {
	TenantId       string
	ClientId       string
	InstanceUrl    string
	RequiredScopes string
}

type ApiKeyConfig struct {
	SecretName string
	Value      string // Loaded from secrets or environment
}

type StorageConfig struct {
	Mode                  string
	LocalBasePath         string
	CloudConnectionString string
	CloudContainer        string
	MaxUploadSizeMB       int64
}

type SecretsConfig struct {
	// Source determines where secrets are loaded from: "environment", "vault", or "auto"
	// "auto" uses environment in development, vault in staging/production
	Source       string
	KeyVaultName string
	CacheEnabled bool
	CacheTTL     int // seconds
}

type LoggingConfig struct {
	Level  string
	Format string
}

type ServerConfig struct {
	ReadTimeout    int
	WriteTimeout   int
	RequestTimeout int
	EnableSwagger  bool
}

// ConnectionString builds PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// ReadTimeoutDuration returns read timeout as duration
func (s *ServerConfig) ReadTimeoutDuration() time.Duration {
	return time.Duration(s.ReadTimeout) * time.Second
}

// WriteTimeoutDuration returns write timeout as duration
func (s *ServerConfig) WriteTimeoutDuration() time.Duration {
	return time.Duration(s.WriteTimeout) * time.Second
}

// RequestTimeoutDuration returns request timeout as duration
func (s *ServerConfig) RequestTimeoutDuration() time.Duration {
	return time.Duration(s.RequestTimeout) * time.Second
}

// ConnMaxLifetimeDuration returns connection max lifetime as duration
func (d *DatabaseConfig) ConnMaxLifetimeDuration() time.Duration {
	return time.Duration(d.ConnMaxLifetime) * time.Second
}

// Load loads configuration from file and environment variables
// This is a basic load that doesn't fetch secrets from vault
// Use LoadWithSecrets for full secret resolution
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from config file
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Environment variables override config file
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load API key from environment if not in config
	if cfg.ApiKey.Value == "" {
		cfg.ApiKey.Value = v.GetString("ADMIN_API_KEY")
	}

	return &cfg, nil
}

// LoadWithSecrets loads configuration and resolves secrets from the configured source
// In development (or when secrets.source = "environment"), secrets come from env vars
// In staging/production (or when secrets.source = "vault"), secrets come from Azure Key Vault
func LoadWithSecrets(ctx context.Context, logger *zap.Logger) (*Config, error) {
	// First load basic config
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Determine secret source
	source := secrets.SecretSource(cfg.Secrets.Source)
	if source == "" {
		source = secrets.SourceAuto
	}

	// If using environment source directly, we're done (secrets already loaded via viper)
	if source == secrets.SourceEnvironment {
		logger.Info("Using environment variables for secrets")
		return cfg, nil
	}

	// For auto or vault source, initialize the provider
	provider, err := secrets.NewProvider(&secrets.ProviderConfig{
		Source:       source,
		VaultName:    cfg.Secrets.KeyVaultName,
		Environment:  cfg.App.Environment,
		CacheEnabled: cfg.Secrets.CacheEnabled,
		CacheTTL:     time.Duration(cfg.Secrets.CacheTTL) * time.Second,
	}, logger)
	if err != nil {
		// If vault initialization fails in non-development, return error
		if cfg.App.Environment != "development" && cfg.App.Environment != "local" && cfg.App.Environment != "" {
			return nil, fmt.Errorf("failed to initialize secrets provider: %w", err)
		}
		// In development, fall back to environment variables
		logger.Warn("Failed to initialize vault, falling back to environment variables",
			zap.Error(err),
		)
		return cfg, nil
	}

	// If provider is using environment source (auto-detected), we're done
	if !provider.IsVaultEnabled() {
		logger.Info("Secrets provider using environment variables")
		return cfg, nil
	}

	// Load secrets from vault
	logger.Info("Loading secrets from Azure Key Vault")

	// Database secrets - use vault secret names, allow env override
	// Vault secret naming convention: lowercase-kebab-case
	if host, err := provider.GetSecretOrEnv(ctx, "database-host", "DATABASE_HOST"); err == nil && host != "" {
		cfg.Database.Host = host
	}
	if portStr, err := provider.GetSecretOrEnv(ctx, "database-port", "DATABASE_PORT"); err == nil && portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Database.Port = port
		}
	}
	if name, err := provider.GetSecretOrEnv(ctx, "database-name", "DATABASE_NAME"); err == nil && name != "" {
		cfg.Database.Name = name
	}
	if user, err := provider.GetSecretOrEnv(ctx, "database-user", "DATABASE_USER"); err == nil && user != "" {
		cfg.Database.User = user
	}
	if password, err := provider.GetSecretOrEnv(ctx, "database-password", "DATABASE_PASSWORD"); err == nil && password != "" {
		cfg.Database.Password = password
	}

	// Azure AD secrets
	if tenantId, err := provider.GetSecretOrEnv(ctx, "azuread-tenant-id", "AZUREAD_TENANTID"); err == nil && tenantId != "" {
		cfg.AzureAd.TenantId = tenantId
	}
	if clientId, err := provider.GetSecretOrEnv(ctx, "azuread-client-id", "AZUREAD_CLIENTID"); err == nil && clientId != "" {
		cfg.AzureAd.ClientId = clientId
	}

	// API Key
	if apiKey, err := provider.GetSecretOrEnv(ctx, "admin-api-key", "ADMIN_API_KEY"); err == nil && apiKey != "" {
		cfg.ApiKey.Value = apiKey
	}

	// Storage connection string (for cloud storage)
	if connStr, err := provider.GetSecretOrEnv(ctx, "storage-connection-string", "STORAGE_CLOUDCONNECTIONSTRING"); err == nil && connStr != "" {
		cfg.Storage.CloudConnectionString = connStr
	}

	logger.Info("Secrets loaded from vault successfully")
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "Straye Relation API")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.port", 8080)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "relation")
	v.SetDefault("database.user", "relation_user")
	v.SetDefault("database.password", "relation_password")
	v.SetDefault("database.sslMode", "disable")
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 5)
	v.SetDefault("database.connMaxLifetime", 300)

	// Secrets defaults
	v.SetDefault("secrets.source", "auto")
	v.SetDefault("secrets.cacheEnabled", true)
	v.SetDefault("secrets.cacheTTL", 300) // 5 minutes

	// Storage defaults
	v.SetDefault("storage.mode", "local")
	v.SetDefault("storage.localBasePath", "./storage")
	v.SetDefault("storage.maxUploadSizeMB", 50)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")

	// Server defaults
	v.SetDefault("server.readTimeout", 30)
	v.SetDefault("server.writeTimeout", 30)
	v.SetDefault("server.requestTimeout", 60)
	v.SetDefault("server.enableSwagger", true)
}
