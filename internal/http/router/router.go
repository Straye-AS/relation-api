package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/straye-as/relation-api/internal/http/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type Router struct {
	cfg              *config.Config
	logger           *zap.Logger
	authMiddleware   *auth.Middleware
	customerHandler  *handler.CustomerHandler
	projectHandler   *handler.ProjectHandler
	offerHandler     *handler.OfferHandler
	fileHandler      *handler.FileHandler
	dashboardHandler *handler.DashboardHandler
	authHandler      *handler.AuthHandler
	companyHandler   *handler.CompanyHandler
}

func NewRouter(
	cfg *config.Config,
	logger *zap.Logger,
	authMiddleware *auth.Middleware,
	customerHandler *handler.CustomerHandler,
	projectHandler *handler.ProjectHandler,
	offerHandler *handler.OfferHandler,
	fileHandler *handler.FileHandler,
	dashboardHandler *handler.DashboardHandler,
	authHandler *handler.AuthHandler,
	companyHandler *handler.CompanyHandler,
) *Router {
	return &Router{
		cfg:              cfg,
		logger:           logger,
		authMiddleware:   authMiddleware,
		customerHandler:  customerHandler,
		projectHandler:   projectHandler,
		offerHandler:     offerHandler,
		fileHandler:      fileHandler,
		dashboardHandler: dashboardHandler,
		authHandler:      authHandler,
		companyHandler:   companyHandler,
	}
}

func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery(rt.logger))
	r.Use(middleware.Logging(rt.logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{},
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return origin != ""
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Location"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
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

			// Auth
			r.Get("/auth/me", rt.authHandler.Me)
			r.Get("/users", rt.authHandler.ListUsers)

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
