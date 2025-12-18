// Package datawarehouse provides read-only connectivity to the MS SQL Server data warehouse.
// This package is used for querying general ledger data and other reporting information
// from the company's data warehouse.
package datawarehouse

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/microsoft/go-mssqldb" // MS SQL Server driver
	"github.com/straye-as/relation-api/internal/config"
	"go.uber.org/zap"
)

const (
	// Default retry configuration for connection attempts
	defaultMaxRetries     = 3
	defaultInitialBackoff = 1 * time.Second
	defaultMaxBackoff     = 10 * time.Second
	defaultBackoffFactor  = 2.0

	// Default health check timeout
	defaultHealthCheckTimeout = 5 * time.Second
)

// CompanyMapping maps Straye company identifiers to data warehouse table name prefixes
var CompanyMapping = map[string]string{
	"tak":        "strayetak",
	"stalbygg":   "strayestaal",
	"montasje":   "strayemontage",
	"hybridbygg": "strayehybridbygg",
	"industri":   "strayeindustri",
}

// Client provides read-only access to the MS SQL Server data warehouse.
// It manages connection pooling and provides methods for executing queries.
type Client struct {
	db           *sql.DB
	config       *config.DataWarehouseConfig
	logger       *zap.Logger
	queryTimeout time.Duration
}

// HealthStatus represents the health check result for the data warehouse connection
type HealthStatus struct {
	Status     string        `json:"status"`
	Latency    time.Duration `json:"latency_ms"`
	Error      string        `json:"error,omitempty"`
	MaxOpen    int           `json:"max_open_connections"`
	Open       int           `json:"open_connections"`
	InUse      int           `json:"in_use"`
	Idle       int           `json:"idle"`
	WaitCount  int64         `json:"wait_count"`
	WaitTimeMs int64         `json:"wait_time_ms"`
}

// NewClient creates a new data warehouse client with the given configuration.
// Returns nil if the data warehouse is not enabled or not configured.
// The client establishes a connection pool with retry logic for transient failures.
func NewClient(cfg *config.DataWarehouseConfig, logger *zap.Logger) (*Client, error) {
	if cfg == nil || !cfg.Enabled {
		logger.Info("Data warehouse connection disabled")
		return nil, nil
	}

	// Validate required configuration
	if cfg.URL == "" || cfg.User == "" || cfg.Password == "" {
		logger.Warn("Data warehouse enabled but missing credentials, skipping connection",
			zap.Bool("url_present", cfg.URL != ""),
			zap.Bool("user_present", cfg.User != ""),
			zap.Bool("password_present", cfg.Password != ""),
		)
		return nil, nil
	}

	logger.Info("Initializing data warehouse connection",
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Int("conn_max_lifetime_seconds", cfg.ConnMaxLifetime),
		zap.Int("query_timeout_seconds", cfg.QueryTimeout),
	)

	// Build connection string
	connStr, err := buildConnectionString(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build connection string: %w", err)
	}

	// Attempt connection with retry logic
	var db *sql.DB
	backoff := defaultInitialBackoff

	for attempt := 1; attempt <= defaultMaxRetries; attempt++ {
		logger.Info("Attempting data warehouse connection",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", defaultMaxRetries),
		)

		db, err = sql.Open("sqlserver", connStr)
		if err != nil {
			logger.Warn("Failed to open data warehouse connection",
				zap.Error(err),
				zap.Int("attempt", attempt),
			)
			if attempt < defaultMaxRetries {
				time.Sleep(backoff)
				backoff = min(time.Duration(float64(backoff)*defaultBackoffFactor), defaultMaxBackoff)
			}
			continue
		}

		// Configure connection pool
		db.SetMaxOpenConns(cfg.MaxOpenConns)
		db.SetMaxIdleConns(cfg.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.ConnMaxLifetimeDuration())

		// Test connection with ping
		ctx, cancel := context.WithTimeout(context.Background(), defaultHealthCheckTimeout)
		err = db.PingContext(ctx)
		cancel()

		if err != nil {
			logger.Warn("Data warehouse ping failed",
				zap.Error(err),
				zap.Int("attempt", attempt),
			)
			_ = db.Close()
			if attempt < defaultMaxRetries {
				time.Sleep(backoff)
				backoff = min(time.Duration(float64(backoff)*defaultBackoffFactor), defaultMaxBackoff)
			}
			continue
		}

		// Connection successful
		logger.Info("Data warehouse connection established successfully",
			zap.Int("attempts_taken", attempt),
		)

		return &Client{
			db:           db,
			config:       cfg,
			logger:       logger,
			queryTimeout: cfg.QueryTimeoutDuration(),
		}, nil
	}

	return nil, fmt.Errorf("failed to connect to data warehouse after %d attempts: %w", defaultMaxRetries, err)
}

// buildConnectionString constructs a SQL Server connection string from the config.
// URL format expected: host:port/database or host:port (uses default database)
func buildConnectionString(cfg *config.DataWarehouseConfig) (string, error) {
	// Parse URL format: host:port/database or host:port
	urlParts := strings.SplitN(cfg.URL, "/", 2)
	hostPort := urlParts[0]
	database := ""
	if len(urlParts) > 1 {
		database = urlParts[1]
	}

	// Parse host and port
	hostParts := strings.SplitN(hostPort, ":", 2)
	host := hostParts[0]
	port := "1433" // Default SQL Server port
	if len(hostParts) > 1 {
		port = hostParts[1]
	}

	// Build connection string using URL format
	query := url.Values{}
	query.Add("encrypt", "true")
	query.Add("TrustServerCertificate", "false")
	query.Add("connection timeout", "30")
	if database != "" {
		query.Add("database", database)
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%s", host, port),
		RawQuery: query.Encode(),
	}

	return u.String(), nil
}

// Close gracefully closes the data warehouse connection.
// Should be called during application shutdown.
func (c *Client) Close() error {
	if c == nil || c.db == nil {
		return nil
	}

	c.logger.Info("Closing data warehouse connection")

	if err := c.db.Close(); err != nil {
		c.logger.Error("Failed to close data warehouse connection", zap.Error(err))
		return fmt.Errorf("failed to close data warehouse connection: %w", err)
	}

	c.logger.Info("Data warehouse connection closed successfully")
	return nil
}

// HealthCheck performs a health check on the data warehouse connection.
// Returns detailed status including connection pool statistics.
func (c *Client) HealthCheck(ctx context.Context) *HealthStatus {
	if c == nil || c.db == nil {
		return &HealthStatus{
			Status: "disabled",
		}
	}

	start := time.Now()

	// Use provided context or create one with default timeout
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultHealthCheckTimeout)
		defer cancel()
	}

	err := c.db.PingContext(ctx)
	latency := time.Since(start)

	stats := c.db.Stats()
	status := &HealthStatus{
		Latency:    latency,
		MaxOpen:    stats.MaxOpenConnections,
		Open:       stats.OpenConnections,
		InUse:      stats.InUse,
		Idle:       stats.Idle,
		WaitCount:  stats.WaitCount,
		WaitTimeMs: stats.WaitDuration.Milliseconds(),
	}

	if err != nil {
		c.logger.Warn("Data warehouse health check failed",
			zap.Error(err),
			zap.Duration("latency", latency),
		)
		status.Status = "unhealthy"
		status.Error = err.Error()
	} else {
		status.Status = "healthy"
	}

	return status
}

// ExecuteQuery executes a read-only query and returns all rows.
// The context is used for cancellation and timeout control.
// Results are returned as a slice of maps, where each map represents a row
// with column names as keys.
func (c *Client) ExecuteQuery(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("data warehouse client not initialized")
	}

	// Apply default query timeout if context doesn't have a deadline
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.queryTimeout)
		defer cancel()
	}

	c.logger.Debug("Executing data warehouse query",
		zap.String("query", truncateQuery(query, 200)),
		zap.Int("args_count", len(args)),
	)

	start := time.Now()

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		c.logger.Error("Data warehouse query failed",
			zap.Error(err),
			zap.String("query", truncateQuery(query, 200)),
			zap.Duration("duration", time.Since(start)),
		)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}

	var results []map[string]interface{}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	c.logger.Debug("Data warehouse query completed",
		zap.Int("rows_returned", len(results)),
		zap.Duration("duration", time.Since(start)),
	)

	return results, nil
}

// QueryRow executes a query that is expected to return at most one row.
// Returns the row as a map with column names as keys, or nil if no rows found.
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("data warehouse client not initialized")
	}

	// Apply default query timeout if context doesn't have a deadline
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.queryTimeout)
		defer cancel()
	}

	c.logger.Debug("Executing data warehouse single-row query",
		zap.String("query", truncateQuery(query, 200)),
		zap.Int("args_count", len(args)),
	)

	start := time.Now()

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		c.logger.Error("Data warehouse query failed",
			zap.Error(err),
			zap.String("query", truncateQuery(query, 200)),
			zap.Duration("duration", time.Since(start)),
		)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}

	if !rows.Next() {
		// No rows found
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error reading row: %w", err)
		}
		c.logger.Debug("Data warehouse query returned no rows",
			zap.Duration("duration", time.Since(start)),
		)
		return nil, nil
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	row := make(map[string]interface{})
	for i, col := range columns {
		row[col] = values[i]
	}

	c.logger.Debug("Data warehouse single-row query completed",
		zap.Duration("duration", time.Since(start)),
	)

	return row, nil
}

// GetGeneralLedgerTableName returns the fully qualified table name for a company's
// general ledger transactions table.
// Uses the company mapping to convert Straye company IDs to data warehouse table prefixes.
func GetGeneralLedgerTableName(companyID string) (string, error) {
	prefix, ok := CompanyMapping[companyID]
	if !ok {
		return "", fmt.Errorf("unknown company ID: %s", companyID)
	}
	return fmt.Sprintf("dbo.nxt_%s_generalledgertransaction", prefix), nil
}

// IsEnabled returns true if the client is initialized and ready for queries.
func (c *Client) IsEnabled() bool {
	return c != nil && c.db != nil
}

// truncateQuery truncates a query string for logging purposes
func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}
