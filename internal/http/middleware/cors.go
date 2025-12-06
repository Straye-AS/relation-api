package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
	"github.com/straye-as/relation-api/internal/config"
	"go.uber.org/zap"
)

// CORS returns a CORS middleware configured from the application config
func CORS(cfg *config.CORSConfig, environment string, logger *zap.Logger) func(http.Handler) http.Handler {
	options := cors.Options{
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}

	// Handle allowed origins
	if len(cfg.AllowedOrigins) > 0 {
		// Check if wildcard is specified
		for _, origin := range cfg.AllowedOrigins {
			if origin == "*" {
				// Wildcard mode - allow all origins
				// Note: This is NOT recommended for production
				if environment != "development" && environment != "local" {
					logger.Warn("CORS configured with wildcard origin in non-development environment",
						zap.String("environment", environment))
				}
				options.AllowOriginFunc = func(r *http.Request, origin string) bool {
					return origin != ""
				}
				break
			}
		}

		// If no wildcard, use the explicit list
		if options.AllowOriginFunc == nil {
			options.AllowedOrigins = cfg.AllowedOrigins
			logger.Info("CORS configured with explicit origins",
				zap.Strings("origins", cfg.AllowedOrigins))
		}
	} else {
		// No origins configured - in development allow all, in production deny all
		if environment == "development" || environment == "local" || environment == "" {
			options.AllowOriginFunc = func(r *http.Request, origin string) bool {
				return origin != ""
			}
			logger.Info("CORS configured to allow all origins in development mode")
		} else {
			// In production, if no origins configured, explicitly deny all
			// Note: We must use AllowOriginFunc because empty AllowedOrigins defaults to "*"
			options.AllowOriginFunc = func(r *http.Request, origin string) bool {
				return false
			}
			logger.Warn("CORS configured with no allowed origins - all cross-origin requests will be denied",
				zap.String("environment", environment))
		}
	}

	return cors.Handler(options)
}
