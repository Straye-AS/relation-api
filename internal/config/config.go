package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
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
	KeyVaultName string
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
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("app.name", "Straye Relation API")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.port", 8080)
	v.SetDefault("database.sslMode", "disable")
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 5)
	v.SetDefault("database.connMaxLifetime", 300)
	v.SetDefault("storage.mode", "local")
	v.SetDefault("storage.localBasePath", "./storage")
	v.SetDefault("storage.maxUploadSizeMB", 50)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")
	v.SetDefault("server.readTimeout", 30)
	v.SetDefault("server.writeTimeout", 30)
	v.SetDefault("server.requestTimeout", 60)
	v.SetDefault("server.enableSwagger", true)

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
