package middleware

import (
	"fmt"
	"net/http"

	"github.com/straye-as/relation-api/internal/config"
)

// SecurityHeaders returns a middleware that adds security headers to responses
func SecurityHeaders(cfg *config.SecurityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// X-Content-Type-Options prevents MIME type sniffing
			if cfg.ContentTypeNosniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			// X-Frame-Options prevents clickjacking
			if cfg.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.FrameOptions)
			}

			// X-XSS-Protection enables browser XSS filtering
			if cfg.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", cfg.XSSProtection)
			}

			// Content-Security-Policy controls resource loading
			if cfg.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}

			// Referrer-Policy controls referrer information
			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			// Permissions-Policy (formerly Feature-Policy) controls browser features
			if cfg.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
			}

			// Strict-Transport-Security enforces HTTPS
			if cfg.EnableHSTS {
				hstsValue := fmt.Sprintf("max-age=%d", cfg.HSTSMaxAge)
				if cfg.HSTSIncludeSubdomains {
					hstsValue += "; includeSubDomains"
				}
				if cfg.HSTSPreload {
					hstsValue += "; preload"
				}
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Remove headers that leak server information
			w.Header().Del("X-Powered-By")
			w.Header().Del("Server")

			next.ServeHTTP(w, r)
		})
	}
}
