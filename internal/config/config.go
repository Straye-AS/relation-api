package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/straye-as/relation-api/internal/secrets"
	"go.uber.org/zap"
)

// Config holds all application configuration
type Config struct {
	App           AppConfig
	Database      DatabaseConfig
	DataWarehouse DataWarehouseConfig
	AzureAd       AzureAdConfig
	ApiKey        ApiKeyConfig
	Storage       StorageConfig
	Secrets       SecretsConfig
	Logging       LoggingConfig
	Server        ServerConfig
	CORS          CORSConfig
	Security      SecurityConfig
	RateLimit     RateLimitConfig
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

// DataWarehouseConfig holds configuration for the MS SQL Server data warehouse
// This connection is optional and read-only
type DataWarehouseConfig struct {
	// Enabled controls whether the data warehouse connection is attempted
	Enabled bool
	// URL is the connection URL in format host:port/database (from WAREHOUSE-URL secret)
	URL string
	// User is the database username (from WAREHOUSE-USERNAME secret)
	User string
	// Password is the database password (from WAREHOUSE-PASSWORD secret)
	Password string
	// MaxOpenConns is the maximum number of open connections to the database
	MaxOpenConns int
	// MaxIdleConns is the maximum number of connections in the idle connection pool
	MaxIdleConns int
	// ConnMaxLifetime is the maximum amount of time a connection may be reused (seconds)
	ConnMaxLifetime int
	// QueryTimeout is the default timeout for queries (seconds)
	QueryTimeout int
}

type AzureAdConfig struct {
	TenantId       string
	ClientId       string
	ClientSecret   string // Required for On-Behalf-Of flow to call MS Graph
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

// CORSConfig holds CORS configuration
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins for CORS requests
	// Use "*" to allow all origins (not recommended for production)
	AllowedOrigins []string
	// AllowedMethods is a list of allowed HTTP methods
	AllowedMethods []string
	// AllowedHeaders is a list of allowed request headers
	AllowedHeaders []string
	// ExposedHeaders is a list of headers exposed to the client
	ExposedHeaders []string
	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool
	// MaxAge is the max age (in seconds) for preflight cache
	MaxAge int
}

// SecurityConfig holds security header configuration
type SecurityConfig struct {
	// EnableHSTS enables HTTP Strict Transport Security header
	EnableHSTS bool
	// HSTSMaxAge is the max age for HSTS in seconds (default: 31536000 = 1 year)
	HSTSMaxAge int
	// HSTSIncludeSubdomains includes subdomains in HSTS
	HSTSIncludeSubdomains bool
	// HSTSPreload enables HSTS preload
	HSTSPreload bool
	// ContentSecurityPolicy sets the Content-Security-Policy header
	ContentSecurityPolicy string
	// FrameOptions sets the X-Frame-Options header (DENY, SAMEORIGIN, or empty to disable)
	FrameOptions string
	// ContentTypeNosniff enables X-Content-Type-Options: nosniff
	ContentTypeNosniff bool
	// XSSProtection sets the X-XSS-Protection header
	XSSProtection string
	// ReferrerPolicy sets the Referrer-Policy header
	ReferrerPolicy string
	// PermissionsPolicy sets the Permissions-Policy header
	PermissionsPolicy string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Enabled enables rate limiting
	Enabled bool
	// RequestsPerMinute is the default rate limit for unauthenticated requests (per IP)
	RequestsPerMinute int
	// RequestsPerMinuteAuth is the rate limit for authenticated requests (per user)
	RequestsPerMinuteAuth int
	// BurstSize is the maximum burst size allowed
	BurstSize int
	// WhitelistIPs is a list of IPs that bypass rate limiting
	WhitelistIPs []string
	// WhitelistPaths is a list of paths that bypass rate limiting (e.g., /health)
	WhitelistPaths []string
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

// ConnMaxLifetimeDuration returns connection max lifetime as duration
func (d *DataWarehouseConfig) ConnMaxLifetimeDuration() time.Duration {
	return time.Duration(d.ConnMaxLifetime) * time.Second
}

// QueryTimeoutDuration returns query timeout as duration
func (d *DataWarehouseConfig) QueryTimeoutDuration() time.Duration {
	return time.Duration(d.QueryTimeout) * time.Second
}

// Load loads configuration from file and environment variables
// This is a basic load that doesn't fetch secrets from vault
// Use LoadWithSecrets for full secret resolution
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

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

	// Load Azure AD config from environment if not in config
	if cfg.AzureAd.TenantId == "" {
		cfg.AzureAd.TenantId = v.GetString("AZURE_TENANT_ID")
	}
	if cfg.AzureAd.ClientId == "" {
		cfg.AzureAd.ClientId = v.GetString("AZURE_CLIENT_ID")
	}
	if cfg.AzureAd.ClientSecret == "" {
		cfg.AzureAd.ClientSecret = v.GetString("AZURE_CLIENT_SECRET")
	}
	if cfg.AzureAd.RequiredScopes == "" {
		cfg.AzureAd.RequiredScopes = v.GetString("AZURE_REQUIRED_SCOPES")
	}

	// Load Azure Key Vault name from environment if not in config
	if cfg.Secrets.KeyVaultName == "" {
		cfg.Secrets.KeyVaultName = v.GetString("AZURE_KEY_VAULT_NAME")
	}

	// Check for DATAWAREHOUSE_ENABLED env var override
	if v.GetBool("DATAWAREHOUSE_ENABLED") {
		cfg.DataWarehouse.Enabled = true
	}

	// NOTE: Data warehouse credentials are ONLY loaded from Azure Key Vault
	// They are never loaded from environment variables for security reasons
	// See LoadWithSecrets() for credential loading

	return &cfg, nil
}

// LoadWithSecrets loads configuration and resolves secrets from the configured source
// In development (or when secrets.source = "environment"), secrets come from env vars
// In staging/production (or when secrets.source = "vault"), secrets come from Azure Key Vault
//
// Key Vault is used when BOTH conditions are met:
// 1. USE_AZURE_KEY_VAULT environment variable is set to "true"
// 2. Environment is "staging" or "production"
//
// EXCEPTION: Data warehouse credentials are ALWAYS loaded from Key Vault when:
// - DATAWAREHOUSE_ENABLED=true AND
// - AZURE_KEY_VAULT_NAME is configured
// This allows data warehouse connectivity in any environment while keeping credentials secure.
func LoadWithSecrets(ctx context.Context, logger *zap.Logger) (*Config, error) {
	// First load basic config
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Check if Azure Key Vault should be used for main secrets
	// Requires both USE_AZURE_KEY_VAULT=true AND environment is staging/production
	useKeyVault := strings.ToLower(os.Getenv("USE_AZURE_KEY_VAULT")) == "true"
	isValidEnv := cfg.App.Environment == "staging" || cfg.App.Environment == "production"

	// Data warehouse credentials are loaded from Key Vault regardless of environment
	// when the feature is enabled and Key Vault is configured
	if cfg.DataWarehouse.Enabled && cfg.Secrets.KeyVaultName != "" {
		if err := loadDataWarehouseSecrets(ctx, cfg, logger); err != nil {
			logger.Warn("Failed to load data warehouse secrets from Key Vault",
				zap.Error(err),
				zap.String("environment", cfg.App.Environment),
			)
			// Don't fail startup - data warehouse is optional
		}
	}

	if !useKeyVault {
		logger.Info("USE_AZURE_KEY_VAULT not enabled, using environment variables for main secrets",
			zap.String("environment", cfg.App.Environment),
		)
		return cfg, nil
	}

	if !isValidEnv {
		logger.Warn("USE_AZURE_KEY_VAULT is enabled but environment is not staging or production, using environment variables for main secrets",
			zap.String("environment", cfg.App.Environment),
			zap.Bool("use_key_vault", useKeyVault),
		)
		return cfg, nil
	}

	// Validate Key Vault name is provided
	if cfg.Secrets.KeyVaultName == "" {
		return nil, fmt.Errorf("AZURE_KEY_VAULT_NAME is required when USE_AZURE_KEY_VAULT=true")
	}

	logger.Info("Azure Key Vault enabled for secrets",
		zap.String("environment", cfg.App.Environment),
		zap.String("key_vault_name", cfg.Secrets.KeyVaultName),
	)

	// Determine secret source - force vault when USE_AZURE_KEY_VAULT is true
	source := secrets.SourceVault

	// For vault source, initialize the provider
	provider, err := secrets.NewProvider(&secrets.ProviderConfig{
		Source:       source,
		VaultName:    cfg.Secrets.KeyVaultName,
		Environment:  cfg.App.Environment,
		CacheEnabled: cfg.Secrets.CacheEnabled,
		CacheTTL:     time.Duration(cfg.Secrets.CacheTTL) * time.Second,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize secrets provider (USE_AZURE_KEY_VAULT=true requires valid vault): %w", err)
	}

	// Verify vault is enabled (should always be true at this point)
	if !provider.IsVaultEnabled() {
		return nil, fmt.Errorf("vault provider not enabled despite USE_AZURE_KEY_VAULT=true")
	}

	// Load secrets from vault
	logger.Info("Loading secrets from Azure Key Vault")

	// Database secrets from Key Vault
	// Host, User, Password come from vault; Port and Database name are environment-specific
	if host, err := provider.GetSecretOrEnv(ctx, "POSTGRES-MAIN-HOST", "DATABASE_HOST"); err == nil && host != "" {
		cfg.Database.Host = host
	}
	// Database name from DEFAULT_DATABASE env var (not in vault - varies per environment)
	if defaultDB := os.Getenv("DEFAULT_DATABASE"); defaultDB != "" {
		cfg.Database.Name = defaultDB
		logger.Info("Using DEFAULT_DATABASE environment variable for database name",
			zap.String("database", defaultDB),
		)
	}
	if user, err := provider.GetSecretOrEnv(ctx, "POSTGRES-MAIN-USER", "DATABASE_USER"); err == nil && user != "" {
		cfg.Database.User = user
	}
	if password, err := provider.GetSecretOrEnv(ctx, "POSTGRES-MAIN-PASSWORD", "DATABASE_PASSWORD"); err == nil && password != "" {
		cfg.Database.Password = password
	}
	// SSL mode from env var (Azure PostgreSQL requires "require")
	if sslMode := os.Getenv("DATABASE_SSLMODE"); sslMode != "" {
		cfg.Database.SSLMode = sslMode
	}

	// Azure AD secrets
	if tenantId, err := provider.GetSecretOrEnv(ctx, "azure-tenant-id", "AZURE_TENANT_ID"); err == nil && tenantId != "" {
		cfg.AzureAd.TenantId = tenantId
	}
	if clientId, err := provider.GetSecretOrEnv(ctx, "azure-client-id", "AZURE__CLIENT_ID"); err == nil && clientId != "" {
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

	// Note: Data warehouse secrets are already loaded earlier in LoadWithSecrets
	// via loadDataWarehouseSecrets() regardless of environment

	logger.Info("Secrets loaded from vault successfully")
	return cfg, nil
}

// loadDataWarehouseSecrets loads data warehouse credentials from Azure Key Vault
// This is called regardless of environment when DATAWAREHOUSE_ENABLED=true
// Data warehouse credentials ONLY come from Key Vault, never from environment variables
func loadDataWarehouseSecrets(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	logger.Info("Loading data warehouse secrets from Key Vault",
		zap.String("key_vault_name", cfg.Secrets.KeyVaultName),
		zap.String("environment", cfg.App.Environment),
	)

	// Initialize a vault-only provider for data warehouse secrets
	provider, err := secrets.NewProvider(&secrets.ProviderConfig{
		Source:       secrets.SourceVault,
		VaultName:    cfg.Secrets.KeyVaultName,
		Environment:  cfg.App.Environment,
		CacheEnabled: cfg.Secrets.CacheEnabled,
		CacheTTL:     time.Duration(cfg.Secrets.CacheTTL) * time.Second,
	}, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize vault client for data warehouse: %w", err)
	}

	// Load credentials from Key Vault only (no env var fallback)
	url, err := provider.GetSecret(ctx, "WAREHOUSE-URL")
	if err != nil {
		return fmt.Errorf("failed to get WAREHOUSE-URL from Key Vault: %w", err)
	}
	cfg.DataWarehouse.URL = url

	user, err := provider.GetSecret(ctx, "WAREHOUSE-USERNAME")
	if err != nil {
		return fmt.Errorf("failed to get WAREHOUSE-USERNAME from Key Vault: %w", err)
	}
	cfg.DataWarehouse.User = user

	password, err := provider.GetSecret(ctx, "WAREHOUSE-PASSWORD")
	if err != nil {
		return fmt.Errorf("failed to get WAREHOUSE-PASSWORD from Key Vault: %w", err)
	}
	cfg.DataWarehouse.Password = password

	logger.Info("Data warehouse credentials loaded from Key Vault successfully")
	return nil
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

	// Data warehouse defaults (MS SQL Server - optional, read-only)
	v.SetDefault("dataWarehouse.enabled", false) // Disabled by default
	v.SetDefault("dataWarehouse.maxOpenConns", 10)
	v.SetDefault("dataWarehouse.maxIdleConns", 2)
	v.SetDefault("dataWarehouse.connMaxLifetime", 300) // 5 minutes
	v.SetDefault("dataWarehouse.queryTimeout", 30)     // 30 seconds default query timeout

	// Secrets defaults
	v.SetDefault("secrets.source", "auto")
	v.SetDefault("secrets.cacheEnabled", true)
	v.SetDefault("secrets.cacheTTL", 300) // 5 minutes

	// Azure AD defaults
	v.SetDefault("azuread.instanceUrl", "https://login.microsoftonline.com/")

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

	// CORS defaults - restrictive by default
	// In development, you may want to override with specific origins
	v.SetDefault("cors.allowedOrigins", []string{})
	v.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowedHeaders", []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Request-ID"})
	v.SetDefault("cors.exposedHeaders", []string{"Location", "X-Request-ID"})
	v.SetDefault("cors.allowCredentials", true)
	v.SetDefault("cors.maxAge", 300) // 5 minutes

	// Security header defaults - secure by default
	v.SetDefault("security.enableHSTS", false)    // Disabled by default, enable in production with HTTPS
	v.SetDefault("security.hstsMaxAge", 31536000) // 1 year
	v.SetDefault("security.hstsIncludeSubdomains", true)
	v.SetDefault("security.hstsPreload", false)
	v.SetDefault("security.contentSecurityPolicy", "default-src 'self'")
	v.SetDefault("security.frameOptions", "DENY")
	v.SetDefault("security.contentTypeNosniff", true)
	v.SetDefault("security.xssProtection", "1; mode=block")
	v.SetDefault("security.referrerPolicy", "strict-origin-when-cross-origin")
	v.SetDefault("security.permissionsPolicy", "geolocation=(), microphone=(), camera=()")

	// Rate limiting defaults
	v.SetDefault("rateLimit.enabled", true)
	v.SetDefault("rateLimit.requestsPerMinute", 60)      // 60 requests per minute for unauthenticated
	v.SetDefault("rateLimit.requestsPerMinuteAuth", 120) // 120 requests per minute for authenticated users
	v.SetDefault("rateLimit.burstSize", 10)              // Allow burst of 10 requests
	v.SetDefault("rateLimit.whitelistIPs", []string{"127.0.0.1", "::1"})
	v.SetDefault("rateLimit.whitelistPaths", []string{"/health", "/health/db", "/health/ready"})
}
