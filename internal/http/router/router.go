package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/database"
	"github.com/straye-as/relation-api/internal/datawarehouse"
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/straye-as/relation-api/internal/http/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/straye-as/relation-api/docs" // Import generated swagger docs
)

type Router struct {
	cfg                     *config.Config
	logger                  *zap.Logger
	db                      *gorm.DB
	dwClient                *datawarehouse.Client
	authMiddleware          *auth.Middleware
	companyFilterMiddleware *middleware.CompanyFilterMiddleware
	rateLimiter             *middleware.RateLimiter
	auditMiddleware         *middleware.AuditMiddleware
	customerHandler         *handler.CustomerHandler
	projectHandler          *handler.ProjectHandler
	offerHandler            *handler.OfferHandler
	inquiryHandler          *handler.InquiryHandler
	dealHandler             *handler.DealHandler
	fileHandler             *handler.FileHandler
	dashboardHandler        *handler.DashboardHandler
	authHandler             *handler.AuthHandler
	companyHandler          *handler.CompanyHandler
	auditHandler            *handler.AuditHandler
	contactHandler          *handler.ContactHandler
	budgetItemHandler       *handler.BudgetItemHandler
	notificationHandler     *handler.NotificationHandler
	activityHandler         *handler.ActivityHandler
	supplierHandler         *handler.SupplierHandler
}

func NewRouter(
	cfg *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	dwClient *datawarehouse.Client,
	authMiddleware *auth.Middleware,
	companyFilterMiddleware *middleware.CompanyFilterMiddleware,
	rateLimiter *middleware.RateLimiter,
	auditMiddleware *middleware.AuditMiddleware,
	customerHandler *handler.CustomerHandler,
	projectHandler *handler.ProjectHandler,
	offerHandler *handler.OfferHandler,
	inquiryHandler *handler.InquiryHandler,
	dealHandler *handler.DealHandler,
	fileHandler *handler.FileHandler,
	dashboardHandler *handler.DashboardHandler,
	authHandler *handler.AuthHandler,
	companyHandler *handler.CompanyHandler,
	auditHandler *handler.AuditHandler,
	contactHandler *handler.ContactHandler,
	budgetItemHandler *handler.BudgetItemHandler,
	notificationHandler *handler.NotificationHandler,
	activityHandler *handler.ActivityHandler,
	supplierHandler *handler.SupplierHandler,
) *Router {
	return &Router{
		cfg:                     cfg,
		logger:                  logger,
		db:                      db,
		dwClient:                dwClient,
		authMiddleware:          authMiddleware,
		companyFilterMiddleware: companyFilterMiddleware,
		rateLimiter:             rateLimiter,
		auditMiddleware:         auditMiddleware,
		customerHandler:         customerHandler,
		projectHandler:          projectHandler,
		offerHandler:            offerHandler,
		inquiryHandler:          inquiryHandler,
		dealHandler:             dealHandler,
		fileHandler:             fileHandler,
		dashboardHandler:        dashboardHandler,
		authHandler:             authHandler,
		companyHandler:          companyHandler,
		auditHandler:            auditHandler,
		contactHandler:          contactHandler,
		budgetItemHandler:       budgetItemHandler,
		notificationHandler:     notificationHandler,
		activityHandler:         activityHandler,
		supplierHandler:         supplierHandler,
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
		_, _ = w.Write([]byte("OK"))
	})

	// Database health check (readiness probe with detailed stats)
	r.Get("/health/db", func(w http.ResponseWriter, r *http.Request) {
		stats, err := database.HealthCheckWithStats(rt.db)
		if err != nil {
			rt.logger.Error("Database health check failed", zap.Error(err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "unhealthy",
				"error":   err.Error(),
				"service": "database",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
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

	// Data warehouse health check (optional service)
	r.Get("/health/datawarehouse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if rt.dwClient == nil || !rt.dwClient.IsEnabled() {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "disabled",
				"service": "datawarehouse",
				"message": "Data warehouse connection is not configured",
			})
			return
		}

		status := rt.dwClient.HealthCheck(r.Context())
		if status.Status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  status.Status,
			"service": "datawarehouse",
			"stats": map[string]interface{}{
				"latency_ms":           status.Latency.Milliseconds(),
				"max_open_connections": status.MaxOpen,
				"open_connections":     status.Open,
				"in_use":               status.InUse,
				"idle":                 status.Idle,
				"wait_count":           status.WaitCount,
				"wait_time_ms":         status.WaitTimeMs,
			},
			"error": status.Error,
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

		// Check data warehouse (optional - doesn't affect overall health)
		if rt.dwClient != nil && rt.dwClient.IsEnabled() {
			dwStatus := rt.dwClient.HealthCheck(r.Context())
			checks["datawarehouse"] = map[string]interface{}{
				"status":     dwStatus.Status,
				"latency_ms": dwStatus.Latency.Milliseconds(),
			}
			if dwStatus.Error != "" {
				checks["datawarehouse"].(map[string]interface{})["error"] = dwStatus.Error
			}
			// Note: Data warehouse issues don't make the app unhealthy
			// as it's an optional service for reporting only
		} else {
			checks["datawarehouse"] = map[string]interface{}{
				"status": "disabled",
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if allHealthy {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "healthy",
				"checks": checks,
			})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
		r.Get("/customers/search", rt.customerHandler.FuzzySearch) // Fuzzy customer search (no auth)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)
			r.Use(rt.companyFilterMiddleware.Filter)
			r.Use(rt.auditMiddleware.Audit) // Audit all modifications

			// Companies (protected routes for details and updates)
			r.Route("/companies", func(r chi.Router) {
				r.Get("/{id}", rt.companyHandler.GetByID)
				r.Put("/{id}", rt.companyHandler.Update)
			})

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
				r.Get("/erp-differences", rt.customerHandler.GetERPDifferences) // ERP sync endpoint
				r.Get("/{id}", rt.customerHandler.GetByID)
				r.Put("/{id}", rt.customerHandler.Update)
				r.Delete("/{id}", rt.customerHandler.Delete)
				r.Get("/{id}/contacts", rt.contactHandler.GetContactsForEntity)
				r.Post("/{id}/contacts", rt.customerHandler.CreateContact)
				r.Get("/{id}/offers", rt.customerHandler.ListOffers)
				r.Get("/{id}/projects", rt.customerHandler.ListProjects)

				// Individual property update endpoints
				r.Put("/{id}/status", rt.customerHandler.UpdateStatus)
				r.Put("/{id}/tier", rt.customerHandler.UpdateTier)
				r.Put("/{id}/industry", rt.customerHandler.UpdateIndustry)
				r.Put("/{id}/notes", rt.customerHandler.UpdateNotes)
				r.Put("/{id}/company", rt.customerHandler.UpdateCompany)
				r.Put("/{id}/customer-class", rt.customerHandler.UpdateCustomerClass)
				r.Put("/{id}/credit-limit", rt.customerHandler.UpdateCreditLimit)
				r.Put("/{id}/is-internal", rt.customerHandler.UpdateIsInternal)
				r.Put("/{id}/address", rt.customerHandler.UpdateAddress)
				r.Put("/{id}/contact-info", rt.customerHandler.UpdateContactInfo)
				r.Put("/{id}/postal-code", rt.customerHandler.UpdatePostalCode)
				r.Put("/{id}/city", rt.customerHandler.UpdateCity)
			})

			// Contacts
			r.Route("/contacts", func(r chi.Router) {
				r.Get("/", rt.contactHandler.ListContacts)
				r.Post("/", rt.contactHandler.CreateContact)
				r.Get("/{id}", rt.contactHandler.GetContact)
				r.Put("/{id}", rt.contactHandler.UpdateContact)
				r.Delete("/{id}", rt.contactHandler.DeleteContact)
				r.Post("/{id}/relationships", rt.contactHandler.AddRelationship)
				r.Delete("/{id}/relationships/{relationshipId}", rt.contactHandler.RemoveRelationship)
			})

			// Projects (simplified containers for offers)
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", rt.projectHandler.List)
				r.Post("/", rt.projectHandler.Create)
				r.Get("/{id}", rt.projectHandler.GetByID)
				r.Put("/{id}", rt.projectHandler.Update)
				r.Delete("/{id}", rt.projectHandler.Delete)
				r.Post("/{id}/reopen", rt.projectHandler.ReopenProject) // Reopen completed/cancelled project
				r.Get("/{id}/activities", rt.projectHandler.GetActivities)
				r.Get("/{id}/contacts", rt.contactHandler.GetContactsForEntity)
				r.Get("/{id}/offers", rt.projectHandler.GetProjectOffers) // List offers in project (offer folder model)

				// Individual property update endpoints
				r.Put("/{id}/name", rt.projectHandler.UpdateName)
				r.Put("/{id}/description", rt.projectHandler.UpdateDescription)
				r.Put("/{id}/phase", rt.projectHandler.UpdatePhase)
				r.Put("/{id}/dates", rt.projectHandler.UpdateDates)
				r.Put("/{id}/project-number", rt.projectHandler.UpdateProjectNumber)
			})

			// Inquiries (draft offers)
			r.Route("/inquiries", func(r chi.Router) {
				r.Get("/", rt.inquiryHandler.List)
				r.Post("/", rt.inquiryHandler.Create)
				r.Get("/{id}", rt.inquiryHandler.GetByID)
				r.Delete("/{id}", rt.inquiryHandler.Delete)
				r.Post("/{id}/convert", rt.inquiryHandler.Convert)
			})

			// Offers
			r.Route("/offers", func(r chi.Router) {
				r.Get("/", rt.offerHandler.List)
				r.Post("/", rt.offerHandler.Create)
				r.Get("/next-number", rt.offerHandler.GetNextNumber) // Must be before /{id} to avoid path conflict
				r.Get("/{id}", rt.offerHandler.GetByID)
				r.Put("/{id}", rt.offerHandler.Update)
				r.Delete("/{id}", rt.offerHandler.Delete)

				// Lifecycle endpoints
				r.Post("/{id}/advance", rt.offerHandler.Advance)
				r.Post("/{id}/send", rt.offerHandler.Send)
				r.Post("/{id}/accept", rt.offerHandler.Accept)
				r.Post("/{id}/reject", rt.offerHandler.Reject)
				r.Post("/{id}/win", rt.offerHandler.Win) // Win offer within project (offer folder model)
				r.Post("/{id}/clone", rt.offerHandler.Clone)

				// Order phase lifecycle endpoints
				r.Post("/{id}/accept-order", rt.offerHandler.AcceptOrder)    // Transition to order phase
				r.Put("/{id}/health", rt.offerHandler.UpdateOfferHealth)     // Update completion percentage
				r.Put("/{id}/spent", rt.offerHandler.UpdateOfferSpent)       // Update spent amount
				r.Put("/{id}/invoiced", rt.offerHandler.UpdateOfferInvoiced) // Update invoiced amount
				r.Post("/{id}/complete", rt.offerHandler.CompleteOffer)      // Transition to completed phase
				r.Post("/{id}/reopen", rt.offerHandler.ReopenOffer)          // Reopen completed offer back to order

				// Individual property update endpoints
				r.Put("/{id}/probability", rt.offerHandler.UpdateProbability)
				r.Put("/{id}/title", rt.offerHandler.UpdateTitle)
				r.Put("/{id}/responsible", rt.offerHandler.UpdateResponsible)
				r.Put("/{id}/customer", rt.offerHandler.UpdateCustomer)
				r.Put("/{id}/value", rt.offerHandler.UpdateValue)
				r.Put("/{id}/cost", rt.offerHandler.UpdateCost)
				r.Put("/{id}/due-date", rt.offerHandler.UpdateDueDate)
				r.Put("/{id}/expiration-date", rt.offerHandler.UpdateExpirationDate)
				r.Put("/{id}/sent-date", rt.offerHandler.UpdateSentDate)
				r.Put("/{id}/start-date", rt.offerHandler.UpdateStartDate)
				r.Put("/{id}/end-date", rt.offerHandler.UpdateEndDate)
				r.Put("/{id}/description", rt.offerHandler.UpdateDescription)
				r.Put("/{id}/notes", rt.offerHandler.UpdateNotes)
				r.Put("/{id}/project", rt.offerHandler.LinkToProject)
				r.Delete("/{id}/project", rt.offerHandler.UnlinkFromProject)
				r.Put("/{id}/customer-has-won-project", rt.offerHandler.UpdateCustomerHasWonProject)
				r.Put("/{id}/offer-number", rt.offerHandler.UpdateOfferNumber)
				r.Put("/{id}/external-reference", rt.offerHandler.UpdateExternalReference)

				// Sub-resources
				r.Get("/{id}/items", rt.offerHandler.GetItems)
				r.Post("/{id}/items", rt.offerHandler.AddItem)
				r.Get("/{id}/files", rt.offerHandler.GetFiles)
				r.Get("/{id}/activities", rt.offerHandler.GetActivities)

				// Budget endpoints
				r.Get("/{id}/detail", rt.offerHandler.GetWithBudgetItems)
				r.Get("/{id}/budget", rt.budgetItemHandler.GetOfferBudgetWithDimensions)
				r.Post("/{id}/recalculate", rt.offerHandler.RecalculateTotals)

				// Budget item sub-resources
				r.Get("/{id}/budget/dimensions", rt.budgetItemHandler.ListOfferDimensions)
				r.Post("/{id}/budget/dimensions", rt.budgetItemHandler.AddToOffer)
				r.Put("/{id}/budget/dimensions/{dimensionId}", rt.budgetItemHandler.UpdateOfferDimension)
				r.Delete("/{id}/budget/dimensions/{dimensionId}", rt.budgetItemHandler.DeleteOfferDimension)
				r.Put("/{id}/budget/reorder", rt.budgetItemHandler.ReorderOfferDimensions)

				// Data warehouse sync (POC)
				r.Get("/{id}/external-sync", rt.offerHandler.GetExternalSync)

				// Admin endpoints
				r.Post("/admin/trigger-dw-sync", rt.offerHandler.TriggerBulkDWSync)
			})

			// Deals
			r.Route("/deals", func(r chi.Router) {
				r.Get("/", rt.dealHandler.List)
				r.Post("/", rt.dealHandler.Create)
				r.Get("/analytics", rt.dealHandler.GetPipelineAnalytics)
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
				r.Post("/{id}/create-offer", rt.dealHandler.CreateOffer)
				r.Get("/{id}/history", rt.dealHandler.GetStageHistory)
				r.Get("/{id}/activities", rt.dealHandler.GetActivities)
				r.Get("/{id}/contacts", rt.contactHandler.GetContactsForEntity)
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

			// Notifications
			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", rt.notificationHandler.List)
				r.Get("/count", rt.notificationHandler.GetUnreadCount)
				r.Put("/read-all", rt.notificationHandler.MarkAllAsRead)
				r.Get("/{id}", rt.notificationHandler.GetByID)
				r.Put("/{id}/read", rt.notificationHandler.MarkAsRead)
			})

			// Activities
			r.Route("/activities", func(r chi.Router) {
				r.Get("/", rt.activityHandler.List)
				r.Post("/", rt.activityHandler.Create)
				r.Get("/my-tasks", rt.activityHandler.GetMyTasks)
				r.Get("/upcoming", rt.activityHandler.GetUpcoming)
				r.Get("/stats", rt.activityHandler.GetStats)
				r.Get("/{id}", rt.activityHandler.GetByID)
				r.Put("/{id}", rt.activityHandler.Update)
				r.Delete("/{id}", rt.activityHandler.Delete)
				r.Post("/{id}/complete", rt.activityHandler.Complete)
				r.Post("/{id}/follow-up", rt.activityHandler.CreateFollowUp)
				r.Post("/{id}/attendees", rt.activityHandler.AddAttendee)
				r.Delete("/{id}/attendees/{userId}", rt.activityHandler.RemoveAttendee)
			})

			// Suppliers
			r.Route("/suppliers", func(r chi.Router) {
				r.Get("/", rt.supplierHandler.List)
				r.Post("/", rt.supplierHandler.Create)
				r.Get("/{id}", rt.supplierHandler.GetByID)
				r.Put("/{id}", rt.supplierHandler.Update)
				r.Delete("/{id}", rt.supplierHandler.Delete)

				// Individual property update endpoints
				r.Put("/{id}/status", rt.supplierHandler.UpdateStatus)
				r.Put("/{id}/notes", rt.supplierHandler.UpdateNotes)
				r.Put("/{id}/category", rt.supplierHandler.UpdateCategory)
				r.Put("/{id}/payment-terms", rt.supplierHandler.UpdatePaymentTerms)
			})
		})
	})

	return r
}
