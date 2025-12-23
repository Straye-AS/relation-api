package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/straye-as/relation-api/docs"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/database"
	"github.com/straye-as/relation-api/internal/datawarehouse"
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/straye-as/relation-api/internal/http/middleware"
	"github.com/straye-as/relation-api/internal/http/router"
	"github.com/straye-as/relation-api/internal/jobs"
	"github.com/straye-as/relation-api/internal/logger"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/straye-as/relation-api/internal/storage"
	"go.uber.org/zap"
)

// @title Straye Relation API
// @version 1.0
// @description CRM Relation API for customer, project, and sales pipeline management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@straye.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-api-key
// @description API Key for system operations
// @Security BearerAuth
// @Security ApiKeyAuth

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Load basic configuration first (for logging setup)
	basicCfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log, err := logger.NewLogger(&basicCfg.Logging, &basicCfg.App)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	log.Info("Starting application",
		zap.String("app", basicCfg.App.Name),
		zap.String("env", basicCfg.App.Environment),
		zap.Int("port", basicCfg.App.Port),
	)

	// Configure Swagger host based on environment
	switch basicCfg.App.Environment {
	case "staging":
		docs.SwaggerInfo.Host = "straye-relation-staging.proudsmoke-10281cc0.norwayeast.azurecontainerapps.io"
	case "production":
		docs.SwaggerInfo.Host = "api.straye.no" // TODO: Update when production URL is known
	default:
		docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", basicCfg.App.Port)
	}

	// Load full configuration with secrets
	// In development: uses environment variables
	// In staging/production: fetches from Azure Key Vault
	cfg, err := config.LoadWithSecrets(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	// Connect to database with retry logic
	db, err := database.NewDatabase(&cfg.Database, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize storage
	fileStorage, err := storage.NewStorage(&cfg.Storage, log)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	log.Info("Storage initialized", zap.String("mode", cfg.Storage.Mode))

	// Initialize data warehouse connection (optional - for reporting)
	// This connection is read-only and the app continues without it if not configured
	var dwClient *datawarehouse.Client
	if cfg.DataWarehouse.Enabled {
		dwClient, err = datawarehouse.NewClient(&cfg.DataWarehouse, log)
		if err != nil {
			// Log error but don't fail - data warehouse is optional
			log.Warn("Data warehouse connection failed, continuing without it",
				zap.Error(err),
			)
		} else if dwClient != nil {
			log.Info("Data warehouse connected successfully",
				zap.Int("max_open_conns", cfg.DataWarehouse.MaxOpenConns),
				zap.Int("query_timeout_seconds", cfg.DataWarehouse.QueryTimeout),
			)
		}
	} else {
		log.Info("Data warehouse not configured, skipping",
			zap.Bool("enabled", cfg.DataWarehouse.Enabled),
		)
	}

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(db)
	contactRepo := repository.NewContactRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	offerItemRepo := repository.NewOfferItemRepository(db)
	dealRepo := repository.NewDealRepository(db)
	dealStageHistoryRepo := repository.NewDealStageHistoryRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	fileRepo := repository.NewFileRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db, log)
	userPermissionRepo := repository.NewUserPermissionRepository(db, log)
	auditLogRepo := repository.NewAuditLogRepository(db)
	budgetItemRepo := repository.NewBudgetItemRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	numberSequenceRepo := repository.NewNumberSequenceRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)

	// Initialize services
	// Company service first (other services may depend on it)
	companyService := service.NewCompanyServiceWithRepo(companyRepo, userRepo, log)

	// Number sequence service (shared between offers and projects)
	numberSequenceService := service.NewNumberSequenceService(numberSequenceRepo, log)

	customerService := service.NewCustomerService(customerRepo, activityRepo, log)
	// Inject data warehouse client into customer service for ERP sync functionality
	if dwClient != nil {
		customerService.SetDataWarehouseClient(&customerDWAdapter{client: dwClient})
	}
	contactService := service.NewContactService(contactRepo, customerRepo, activityRepo, log)
	fileService := service.NewFileService(fileRepo, offerRepo, customerRepo, projectRepo, supplierRepo, activityRepo, fileStorage, log)
	projectService := service.NewProjectServiceWithDeps(projectRepo, offerRepo, customerRepo, activityRepo, fileService, log, db)
	offerService := service.NewOfferService(offerRepo, offerItemRepo, customerRepo, projectRepo, budgetItemRepo, fileRepo, activityRepo, userRepo, companyService, numberSequenceService, fileService, log, db)
	// Inject data warehouse client into offer service for DW sync functionality
	if dwClient != nil {
		offerService.SetDataWarehouseClient(dwClient)
	}
	inquiryService := service.NewInquiryService(offerRepo, customerRepo, activityRepo, companyService, log, db)
	dealService := service.NewDealService(dealRepo, dealStageHistoryRepo, customerRepo, projectRepo, activityRepo, offerRepo, budgetItemRepo, notificationRepo, log, db)
	dashboardService := service.NewDashboardService(customerRepo, projectRepo, offerRepo, activityRepo, notificationRepo, supplierRepo, log)
	permissionService := service.NewPermissionService(userRoleRepo, userPermissionRepo, activityRepo, log)
	auditLogService := service.NewAuditLogService(auditLogRepo, log)
	budgetItemService := service.NewBudgetItemService(budgetItemRepo, offerRepo, projectRepo, log)
	notificationService := service.NewNotificationService(notificationRepo, log)
	activityService := service.NewActivityService(activityRepo, notificationService, log)
	supplierService := service.NewSupplierService(supplierRepo, activityRepo, log)

	// Initialize middleware
	authMiddleware := auth.NewMiddleware(cfg, userRepo, log)
	companyFilterMiddleware := middleware.NewCompanyFilterMiddleware(log)
	rateLimiter := middleware.NewRateLimiter(&cfg.RateLimit, log)
	auditMiddleware := middleware.NewAuditMiddleware(auditLogService, nil, log)

	// Initialize handlers
	customerHandler := handler.NewCustomerHandler(customerService, contactService, offerService, projectService, log)
	contactHandler := handler.NewContactHandler(contactService, log)
	projectHandler := handler.NewProjectHandler(projectService, offerService, log)
	offerHandler := handler.NewOfferHandler(offerService, dwClient, log)
	inquiryHandler := handler.NewInquiryHandler(inquiryService, log)
	dealHandler := handler.NewDealHandler(dealService, log)
	fileHandler := handler.NewFileHandler(fileService, cfg.Storage.MaxUploadSizeMB, log)
	dashboardHandler := handler.NewDashboardHandler(dashboardService, log)
	graphClient := auth.NewGraphClient(&cfg.AzureAd)
	authHandler := handler.NewAuthHandler(userRepo, permissionService, graphClient, log)
	companyHandler := handler.NewCompanyHandler(companyService, log)
	auditHandler := handler.NewAuditHandler(auditLogService, log)
	budgetItemHandler := handler.NewBudgetItemHandler(budgetItemService, log)
	notificationHandler := handler.NewNotificationHandler(notificationService, log)
	activityHandler := handler.NewActivityHandler(activityService, log)
	supplierHandler := handler.NewSupplierHandler(supplierService, log)

	// Setup router
	rt := router.NewRouter(
		cfg,
		log,
		db,
		dwClient,
		authMiddleware,
		companyFilterMiddleware,
		rateLimiter,
		auditMiddleware,
		customerHandler,
		projectHandler,
		offerHandler,
		inquiryHandler,
		dealHandler,
		fileHandler,
		dashboardHandler,
		authHandler,
		companyHandler,
		auditHandler,
		contactHandler,
		budgetItemHandler,
		notificationHandler,
		activityHandler,
		supplierHandler,
	)

	// Initialize and start scheduler for background jobs
	var scheduler *jobs.Scheduler
	if cfg.DataWarehouse.Enabled && cfg.DataWarehouse.PeriodicSyncEnabled && dwClient != nil {
		scheduler = jobs.NewScheduler(log)

		// Register the data warehouse sync job
		// runStartupSync=true will sync stale offers (null or > 1 hour old) immediately
		if err := jobs.RegisterDWSyncJob(
			scheduler,
			offerService,
			log,
			cfg.DataWarehouse.PeriodicSyncCron,
			cfg.DataWarehouse.PeriodicSyncTimeoutDuration(),
			true, // run startup sync for stale offers
		); err != nil {
			log.Error("Failed to register DW sync job", zap.Error(err))
		} else {
			scheduler.Start()
			log.Info("Scheduler started with DW sync job",
				zap.String("cron_expr", cfg.DataWarehouse.PeriodicSyncCron),
				zap.Duration("timeout", cfg.DataWarehouse.PeriodicSyncTimeoutDuration()),
			)
		}
	} else {
		log.Info("DW periodic sync disabled",
			zap.Bool("dw_enabled", cfg.DataWarehouse.Enabled),
			zap.Bool("periodic_sync_enabled", cfg.DataWarehouse.PeriodicSyncEnabled),
			zap.Bool("dw_client_available", dwClient != nil),
		)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      rt.Setup(),
		ReadTimeout:  cfg.Server.ReadTimeoutDuration(),
		WriteTimeout: cfg.Server.WriteTimeoutDuration(),
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info("Server starting", zap.String("addr", srv.Addr))
		serverErrors <- srv.ListenAndServe()
	}()

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Info("Shutdown signal received", zap.String("signal", sig.String()))

		// Stop scheduler if running
		if scheduler != nil {
			ctx := scheduler.Stop()
			<-ctx.Done()
			log.Info("Scheduler stopped")
		}

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Failed to shutdown gracefully", zap.Error(err))
			return err
		}

		// Close data warehouse connection if initialized
		if dwClient != nil {
			if err := dwClient.Close(); err != nil {
				log.Warn("Error closing data warehouse connection", zap.Error(err))
			}
		}

		log.Info("Server stopped gracefully")
	}

	return nil
}

// customerDWAdapter adapts the datawarehouse.Client to the service.DataWarehouseClient interface
// This allows the customer service to use the data warehouse client without depending directly on it
type customerDWAdapter struct {
	client *datawarehouse.Client
}

func (a *customerDWAdapter) IsEnabled() bool {
	return a.client != nil && a.client.IsEnabled()
}

func (a *customerDWAdapter) GetERPCustomers(ctx context.Context) ([]service.DataWarehouseERPCustomer, error) {
	erpCustomers, err := a.client.GetERPCustomers(ctx)
	if err != nil {
		return nil, err
	}

	// Convert datawarehouse.ERPCustomer to service.DataWarehouseERPCustomer
	result := make([]service.DataWarehouseERPCustomer, len(erpCustomers))
	for i, c := range erpCustomers {
		result[i] = service.DataWarehouseERPCustomer{
			OrganizationNumber: c.OrganizationNumber,
			Name:               c.Name,
		}
	}
	return result, nil
}
