package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/database"
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/straye-as/relation-api/internal/http/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Router struct {
	cfg                     *config.Config
	logger                  *zap.Logger
	db                      *gorm.DB
	authMiddleware          *auth.Middleware
	companyFilterMiddleware *middleware.CompanyFilterMiddleware
	rateLimiter             *middleware.RateLimiter
	auditMiddleware         *middleware.AuditMiddleware
	customerHandler         *handler.CustomerHandler
	projectHandler          *handler.ProjectHandler
	offerHandler            *handler.OfferHandler
	dealHandler             *handler.DealHandler
	fileHandler             *handler.FileHandler
	dashboardHandler        *handler.DashboardHandler
	authHandler             *handler.AuthHandler
	companyHandler          *handler.CompanyHandler
	auditHandler            *handler.AuditHandler
}

func NewRouter(
	cfg *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	authMiddleware *auth.Middleware,
	companyFilterMiddleware *middleware.CompanyFilterMiddleware,
	rateLimiter *middleware.RateLimiter,
	auditMiddleware *middleware.AuditMiddleware,
	customerHandler *handler.CustomerHandler,
	projectHandler *handler.ProjectHandler,
	offerHandler *handler.OfferHandler,
	dealHandler *handler.DealHandler,
	fileHandler *handler.FileHandler,
	dashboardHandler *handler.DashboardHandler,
	authHandler *handler.AuthHandler,
	companyHandler *handler.CompanyHandler,
	auditHandler *handler.AuditHandler,
) *Router {
	return &Router{
		cfg:                     cfg,
		logger:                  logger,
		db:                      db,
		authMiddleware:          authMiddleware,
		companyFilterMiddleware: companyFilterMiddleware,
		rateLimiter:             rateLimiter,
		auditMiddleware:         auditMiddleware,
		customerHandler:         customerHandler,
		projectHandler:          projectHandler,
		offerHandler:            offerHandler,
		dealHandler:             dealHandler,
		fileHandler:             fileHandler,
		dashboardHandler:        dashboardHandler,
		authHandler:             authHandler,
		companyHandler:          companyHandler,
		auditHandler:            auditHandler,
	}
}

func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery(rt.logger))
	r.Use(middleware.Logging(rt.logger))
	r.Use(middleware.SecurityHeaders(&rt.cfg.Security))
	r.Use(middleware.CORS(&rt.cfg.CORS, rt.cfg.App.Environment, rt.logger))
	r.Use(rt.rateLimiter.LimitByIP) // Apply IP-based rate limiting globally

	// Health check (basic liveness probe)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Database health check (readiness probe with detailed stats)
	r.Get("/health/db", func(w http.ResponseWriter, r *http.Request) {
		stats, err := database.HealthCheckWithStats(rt.db)
		if err != nil {
			rt.logger.Error("Database health check failed", zap.Error(err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "unhealthy",
				"error":   err.Error(),
				"service": "database",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"service": "database",
			"stats": map[string]interface{}{
				"max_open_connections": stats.MaxOpenConnections,
				"open_connections":     stats.OpenConnections,
				"in_use":               stats.InUse,
				"idle":                 stats.Idle,
				"wait_count":           stats.WaitCount,
				"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
				"max_idle_closed":      stats.MaxIdleClosed,
				"max_lifetime_closed":  stats.MaxLifetimeClosed,
			},
		})
	})

	// Combined readiness check (checks all dependencies)
	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]interface{})
		allHealthy := true

		// Check database
		if err := database.HealthCheck(rt.db); err != nil {
			rt.logger.Error("Database health check failed", zap.Error(err))
			checks["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			allHealthy = false
		} else {
			checks["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if allHealthy {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "healthy",
				"checks": checks,
			})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "unhealthy",
				"checks": checks,
			})
		}
	})

	// Swagger documentation
	if rt.cfg.Server.EnableSwagger {
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		r.Get("/companies", rt.companyHandler.List)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)
			r.Use(rt.companyFilterMiddleware.Filter)
			r.Use(rt.auditMiddleware.Audit) // Audit all modifications

			// Auth
			r.Get("/auth/me", rt.authHandler.Me)
			r.Get("/auth/permissions", rt.authHandler.Permissions)
			r.Get("/users", rt.authHandler.ListUsers)

			// Audit logs (requires system:audit_logs permission)
			r.Route("/audit", func(r chi.Router) {
				r.Get("/", rt.auditHandler.List)
				r.Get("/stats", rt.auditHandler.GetStats)
				r.Get("/export", rt.auditHandler.Export)
				r.Get("/entity/{entityType}/{entityId}", rt.auditHandler.GetByEntity)
				r.Get("/{id}", rt.auditHandler.GetByID)
			})

			// Customers
			r.Route("/customers", func(r chi.Router) {
				r.Get("/", rt.customerHandler.List)
				r.Post("/", rt.customerHandler.Create)
				r.Get("/{id}", rt.customerHandler.GetByID)
				r.Put("/{id}", rt.customerHandler.Update)
				r.Delete("/{id}", rt.customerHandler.Delete)
				r.Get("/{id}/contacts", rt.customerHandler.ListContacts)
				r.Post("/{id}/contacts", rt.customerHandler.CreateContact)
			})

			// Projects
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", rt.projectHandler.List)
				r.Post("/", rt.projectHandler.Create)
				r.Get("/{id}", rt.projectHandler.GetByID)
				r.Put("/{id}", rt.projectHandler.Update)
				r.Get("/{id}/budget", rt.projectHandler.GetBudget)
				r.Get("/{id}/activities", rt.projectHandler.GetActivities)
			})

			// Offers
			r.Route("/offers", func(r chi.Router) {
				r.Get("/", rt.offerHandler.List)
				r.Post("/", rt.offerHandler.Create)
				r.Get("/{id}", rt.offerHandler.GetByID)
				r.Put("/{id}", rt.offerHandler.Update)
				r.Post("/{id}/advance", rt.offerHandler.Advance)
				r.Get("/{id}/items", rt.offerHandler.GetItems)
				r.Post("/{id}/items", rt.offerHandler.AddItem)
				r.Get("/{id}/files", rt.offerHandler.GetFiles)
				r.Get("/{id}/activities", rt.offerHandler.GetActivities)
			})

			// Deals
			r.Route("/deals", func(r chi.Router) {
				r.Get("/", rt.dealHandler.List)
				r.Post("/", rt.dealHandler.Create)
				r.Get("/pipeline", rt.dealHandler.GetPipelineOverview)
				r.Get("/stats", rt.dealHandler.GetPipelineStats)
				r.Get("/forecast", rt.dealHandler.GetForecast)
				r.Get("/{id}", rt.dealHandler.GetByID)
				r.Put("/{id}", rt.dealHandler.Update)
				r.Delete("/{id}", rt.dealHandler.Delete)
				r.Post("/{id}/advance", rt.dealHandler.AdvanceStage)
				r.Post("/{id}/win", rt.dealHandler.WinDeal)
				r.Post("/{id}/lose", rt.dealHandler.LoseDeal)
				r.Post("/{id}/reopen", rt.dealHandler.ReopenDeal)
				r.Get("/{id}/history", rt.dealHandler.GetStageHistory)
				r.Get("/{id}/activities", rt.dealHandler.GetActivities)
			})

			// Files
			r.Route("/files", func(r chi.Router) {
				r.Post("/upload", rt.fileHandler.Upload)
				r.Get("/{id}", rt.fileHandler.GetByID)
				r.Get("/{id}/download", rt.fileHandler.Download)
			})

			// Dashboard & Search
			r.Get("/dashboard/metrics", rt.dashboardHandler.GetMetrics)
			r.Get("/search", rt.dashboardHandler.Search)
		})
	})

	return r
}
