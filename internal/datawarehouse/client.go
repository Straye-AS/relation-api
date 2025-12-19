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

	// Account number ranges for classifying general ledger entries.
	IncomeAccountMin = 3000
	IncomeAccountMax = 3999
	MaterialCostMin  = 4000
	MaterialCostMax  = 4999
	EmployeeCostMin  = 5000
	EmployeeCostMax  = 5999
	OtherCostMin     = 6000
)

// CompanyMapping maps Straye company identifiers to data warehouse table name prefixes
var CompanyMapping = map[string]string{
	"tak":        "strayetak",
	"stalbygg":   "strayestaal",
	"montasje":   "strayemontasje",
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
// Also handles URLs with https:// or http:// prefix which will be stripped.
func buildConnectionString(cfg *config.DataWarehouseConfig) (string, error) {
	// Strip https:// or http:// prefix if present
	urlStr := cfg.URL
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")

	// Parse URL format: host:port/database or host:port
	urlParts := strings.SplitN(urlStr, "/", 2)
	hostPort := urlParts[0]
	database := "STR_BI" // Default database for Straye data warehouse
	if len(urlParts) > 1 && urlParts[1] != "" {
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
	return fmt.Sprintf("nxt_%s_generalledgertransaction", prefix), nil
}

// IsEnabled returns true if the client is initialized and ready for queries.
func (c *Client) IsEnabled() bool {
	return c != nil && c.db != nil
}

// IsIncomeAccount returns true if the account number falls within the income/revenue range (3000-3999).
func IsIncomeAccount(accountNo int) bool {
	return accountNo >= IncomeAccountMin && accountNo <= IncomeAccountMax
}

// IsCostAccount returns true if the account number is NOT within the income/revenue range.
// Cost accounts are any accounts outside the 3000-3999 range.
func IsCostAccount(accountNo int) bool {
	return !IsIncomeAccount(accountNo)
}

// ProjectFinancials represents aggregated income and costs for a project from the general ledger.
type ProjectFinancials struct {
	ExternalReference string  `json:"externalReference"`
	TotalIncome       float64 `json:"totalIncome"`
	MaterialCosts     float64 `json:"materialCosts"`
	EmployeeCosts     float64 `json:"employeeCosts"`
	OtherCosts        float64 `json:"otherCosts"`
	NetResult         float64 `json:"netResult"`
}

// GetProjectIncome queries the general ledger for total income (accounts 3000-3999) for a project.
// The project is identified by the externalRef which matches the OrgUnit2 column in the GL table.
// Returns the sum of PostedAmountDomestic for all income accounts.
func (c *Client) GetProjectIncome(ctx context.Context, companyID, externalRef string) (float64, error) {
	if c == nil || c.db == nil {
		return 0, fmt.Errorf("data warehouse client not initialized")
	}

	tableName, err := GetGeneralLedgerTableName(companyID)
	if err != nil {
		return 0, fmt.Errorf("get table name: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(PostedAmountDomestic), 0) as total_income
		FROM %s
		WHERE OrgUnit8 = @p1
		  AND AccountNo >= @p2
		  AND AccountNo <= @p3
	`, tableName)

	row, err := c.QueryRow(ctx, query, externalRef, IncomeAccountMin, IncomeAccountMax)
	if err != nil {
		return 0, fmt.Errorf("query project income: %w", err)
	}

	if row == nil {
		return 0, nil
	}

	totalIncome, err := parseFloat64(row["total_income"])
	if err != nil {
		return 0, fmt.Errorf("parse income result: %w", err)
	}

	// Income is stored as negative (credit) in accounting, so negate to show as positive
	return -totalIncome, nil
}

// GetProjectCosts queries the general ledger for total costs (accounts outside 3000-3999) for a project.
// The project is identified by the externalRef which matches the OrgUnit2 column in the GL table.
// Returns the sum of PostedAmountDomestic for all non-income accounts.
func (c *Client) GetProjectCosts(ctx context.Context, companyID, externalRef string) (float64, error) {
	if c == nil || c.db == nil {
		return 0, fmt.Errorf("data warehouse client not initialized")
	}

	tableName, err := GetGeneralLedgerTableName(companyID)
	if err != nil {
		return 0, fmt.Errorf("get table name: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(PostedAmountDomestic), 0) as total_costs
		FROM %s
		WHERE OrgUnit8 = @p1
		  AND (AccountNo < @p2 OR AccountNo > @p3)
	`, tableName)

	row, err := c.QueryRow(ctx, query, externalRef, IncomeAccountMin, IncomeAccountMax)
	if err != nil {
		return 0, fmt.Errorf("query project costs: %w", err)
	}

	if row == nil {
		return 0, nil
	}

	totalCosts, err := parseFloat64(row["total_costs"])
	if err != nil {
		return 0, fmt.Errorf("parse costs result: %w", err)
	}

	return totalCosts, nil
}

// GetProjectFinancials queries the general ledger for income and costs for a project.
// Returns aggregated financials including:
// - Income (3000-3999)
// - Material costs (4000-4999)
// - Employee costs (5000-5999)
// - Other costs (>=6000)
// - Net result (income - all costs)
func (c *Client) GetProjectFinancials(ctx context.Context, companyID, externalRef string) (*ProjectFinancials, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("data warehouse client not initialized")
	}

	tableName, err := GetGeneralLedgerTableName(companyID)
	if err != nil {
		return nil, fmt.Errorf("get table name: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(CASE WHEN AccountNo >= @p2 AND AccountNo <= @p3 THEN PostedAmountDomestic ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN AccountNo >= @p4 AND AccountNo <= @p5 THEN PostedAmountDomestic ELSE 0 END), 0) as material_costs,
			COALESCE(SUM(CASE WHEN AccountNo >= @p6 AND AccountNo <= @p7 THEN PostedAmountDomestic ELSE 0 END), 0) as employee_costs,
			COALESCE(SUM(CASE WHEN AccountNo >= @p8 THEN PostedAmountDomestic ELSE 0 END), 0) as other_costs
		FROM %s
		WHERE OrgUnit8 = @p1
	`, tableName)

	row, err := c.QueryRow(ctx, query, externalRef,
		IncomeAccountMin, IncomeAccountMax,
		MaterialCostMin, MaterialCostMax,
		EmployeeCostMin, EmployeeCostMax,
		OtherCostMin)
	if err != nil {
		return nil, fmt.Errorf("query project financials: %w", err)
	}

	if row == nil {
		return &ProjectFinancials{
			ExternalReference: externalRef,
			TotalIncome:       0,
			MaterialCosts:     0,
			EmployeeCosts:     0,
			OtherCosts:        0,
			NetResult:         0,
		}, nil
	}

	totalIncome, err := parseFloat64(row["total_income"])
	if err != nil {
		return nil, fmt.Errorf("parse income result: %w", err)
	}

	materialCosts, err := parseFloat64(row["material_costs"])
	if err != nil {
		return nil, fmt.Errorf("parse material costs result: %w", err)
	}

	employeeCosts, err := parseFloat64(row["employee_costs"])
	if err != nil {
		return nil, fmt.Errorf("parse employee costs result: %w", err)
	}

	otherCosts, err := parseFloat64(row["other_costs"])
	if err != nil {
		return nil, fmt.Errorf("parse other costs result: %w", err)
	}

	// Income is stored as negative (credit) in accounting, so negate to show as positive
	// Costs are stored as positive (debit), so keep as-is
	return &ProjectFinancials{
		ExternalReference: externalRef,
		TotalIncome:       -totalIncome,
		MaterialCosts:     materialCosts,
		EmployeeCosts:     employeeCosts,
		OtherCosts:        otherCosts,
		NetResult:         -totalIncome - materialCosts - employeeCosts - otherCosts,
	}, nil
}

// parseFloat64 safely extracts a float64 from a database result value.
// Handles various numeric types returned by the SQL driver.
func parseFloat64(val interface{}) (float64, error) {
	if val == nil {
		return 0, nil
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case []byte:
		var f float64
		_, err := fmt.Sscanf(string(v), "%f", &f)
		return f, err
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		return f, err
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}

// truncateQuery truncates a query string for logging purposes
func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

// ERPCustomer represents a customer record from the ERP data warehouse.
type ERPCustomer struct {
	OrganizationNumber string `json:"organizationNumber"`
	Name               string `json:"name"`
}

// GetERPCustomers retrieves all customers from the ERP data warehouse.
// Uses the table dbo.Kunde which contains customers across all companies.
// Returns customers with their organization numbers and names for matching.
func (c *Client) GetERPCustomers(ctx context.Context) ([]ERPCustomer, error) {
	if c == nil || c.db == nil {
		return nil, fmt.Errorf("data warehouse client not initialized")
	}

	// Query all customers from dbo.Kunde
	// Columns: Firmanr, KundeId, Kundenr, Kundenavn, Organisasjonsnr
	query := `
		SELECT
			ISNULL(Organisasjonsnr, '') as Organisasjonsnr,
			ISNULL(Kundenavn, '') as Kundenavn
		FROM dbo.Kunde
	`

	c.logger.Debug("Querying ERP customers from dbo.Kunde")

	rows, err := c.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query ERP customers: %w", err)
	}

	customers := make([]ERPCustomer, 0, len(rows))
	for _, row := range rows {
		customer := ERPCustomer{
			OrganizationNumber: parseString(row["Organisasjonsnr"]),
			Name:               parseString(row["Kundenavn"]),
		}
		customers = append(customers, customer)
	}

	c.logger.Info("Retrieved ERP customers",
		zap.Int("count", len(customers)),
	)

	return customers, nil
}

// parseString safely extracts a string from a database result value.
func parseString(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
