package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-api-key
// @description API Key for system operations

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
	activityRepo := repository.NewActivityRepository(db)
	fileRepo := repository.NewFileRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	customerService := service.NewCustomerService(customerRepo, activityRepo, log)
	contactService := service.NewContactService(contactRepo, customerRepo, activityRepo, log)
	projectService := service.NewProjectService(projectRepo, customerRepo, activityRepo, log)
	offerService := service.NewOfferService(offerRepo, offerItemRepo, customerRepo, projectRepo, fileRepo, activityRepo, log)
	fileService := service.NewFileService(fileRepo, offerRepo, activityRepo, fileStorage, log)
	dashboardService := service.NewDashboardService(customerRepo, projectRepo, offerRepo, notificationRepo, log)
	companyService := service.NewCompanyService(log)

	// Initialize middleware
	authMiddleware := auth.NewMiddleware(cfg, log)
	companyFilterMiddleware := middleware.NewCompanyFilterMiddleware(log)

	// Initialize handlers
	customerHandler := handler.NewCustomerHandler(customerService, contactService, log)
	projectHandler := handler.NewProjectHandler(projectService, log)
	offerHandler := handler.NewOfferHandler(offerService, log)
	fileHandler := handler.NewFileHandler(fileService, cfg.Storage.MaxUploadSizeMB, log)
	dashboardHandler := handler.NewDashboardHandler(dashboardService, log)
	authHandler := handler.NewAuthHandler(userRepo, log)
	companyHandler := handler.NewCompanyHandler(companyService, log)

	// Setup router
	rt := router.NewRouter(
		cfg,
		log,
		db,
		authMiddleware,
		companyFilterMiddleware,
		customerHandler,
		projectHandler,
		offerHandler,
		fileHandler,
		dashboardHandler,
		authHandler,
		companyHandler,
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
