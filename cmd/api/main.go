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
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/straye-as/relation-api/internal/http/middleware"
	"github.com/straye-as/relation-api/internal/http/router"
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
	defer log.Sync()

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
	fileStorage, err := storage.NewStorage(&cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	log.Info("Storage initialized", zap.String("mode", cfg.Storage.Mode))

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

	// Initialize services
	// Company service first (other services may depend on it)
	companyService := service.NewCompanyServiceWithRepo(companyRepo, userRepo, log)

	// Number sequence service (shared between offers and projects)
	numberSequenceService := service.NewNumberSequenceService(numberSequenceRepo, log)

	customerService := service.NewCustomerService(customerRepo, activityRepo, log)
	contactService := service.NewContactService(contactRepo, customerRepo, activityRepo, log)
	projectService := service.NewProjectServiceWithDeps(projectRepo, offerRepo, customerRepo, activityRepo, log, db)
	offerService := service.NewOfferService(offerRepo, offerItemRepo, customerRepo, projectRepo, budgetItemRepo, fileRepo, activityRepo, companyService, numberSequenceService, log, db)
	inquiryService := service.NewInquiryService(offerRepo, customerRepo, activityRepo, companyService, log, db)
	dealService := service.NewDealService(dealRepo, dealStageHistoryRepo, customerRepo, projectRepo, activityRepo, offerRepo, budgetItemRepo, notificationRepo, log, db)
	fileService := service.NewFileService(fileRepo, offerRepo, activityRepo, fileStorage, log)
	dashboardService := service.NewDashboardService(customerRepo, projectRepo, offerRepo, activityRepo, notificationRepo, log)
	permissionService := service.NewPermissionService(userRoleRepo, userPermissionRepo, activityRepo, log)
	auditLogService := service.NewAuditLogService(auditLogRepo, log)
	budgetItemService := service.NewBudgetItemService(budgetItemRepo, offerRepo, projectRepo, log)
	notificationService := service.NewNotificationService(notificationRepo, log)
	activityService := service.NewActivityService(activityRepo, notificationService, log)

	// Initialize middleware
	authMiddleware := auth.NewMiddleware(cfg, log)
	companyFilterMiddleware := middleware.NewCompanyFilterMiddleware(log)
	rateLimiter := middleware.NewRateLimiter(&cfg.RateLimit, log)
	auditMiddleware := middleware.NewAuditMiddleware(auditLogService, nil, log)

	// Initialize handlers
	customerHandler := handler.NewCustomerHandler(customerService, contactService, offerService, projectService, log)
	contactHandler := handler.NewContactHandler(contactService, log)
	projectHandler := handler.NewProjectHandler(projectService, offerService, log)
	offerHandler := handler.NewOfferHandler(offerService, log)
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

	// Setup router
	rt := router.NewRouter(
		cfg,
		log,
		db,
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
	)

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

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Failed to shutdown gracefully", zap.Error(err))
			return err
		}

		log.Info("Server stopped gracefully")
	}

	return nil
}
