package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httprate"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"go.uber.org/zap"
)

// RateLimiter holds rate limiting middleware and configuration
type RateLimiter struct {
	cfg            *config.RateLimitConfig
	logger         *zap.Logger
	ipLimiter      func(http.Handler) http.Handler
	userLimiter    func(http.Handler) http.Handler
	whitelistIPs   map[string]bool
	whitelistPaths map[string]bool
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(cfg *config.RateLimitConfig, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		cfg:            cfg,
		logger:         logger,
		whitelistIPs:   make(map[string]bool),
		whitelistPaths: make(map[string]bool),
	}

	// Build whitelist maps for O(1) lookup
	for _, ip := range cfg.WhitelistIPs {
		rl.whitelistIPs[ip] = true
	}
	for _, path := range cfg.WhitelistPaths {
		rl.whitelistPaths[path] = true
	}

	// Create IP-based rate limiter for unauthenticated requests
	rl.ipLimiter = httprate.Limit(
		cfg.RequestsPerMinute,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(rl.rateLimitExceededHandler),
	)

	// Create user-based rate limiter for authenticated requests
	rl.userLimiter = httprate.Limit(
		cfg.RequestsPerMinuteAuth,
		time.Minute,
		httprate.WithKeyFuncs(rl.keyByUserOrIP),
		httprate.WithLimitHandler(rl.rateLimitExceededHandler),
	)

	logger.Info("Rate limiter initialized",
		zap.Int("requests_per_minute", cfg.RequestsPerMinute),
		zap.Int("requests_per_minute_auth", cfg.RequestsPerMinuteAuth),
		zap.Int("burst_size", cfg.BurstSize),
		zap.Strings("whitelist_ips", cfg.WhitelistIPs),
		zap.Strings("whitelist_paths", cfg.WhitelistPaths),
	)

	return rl
}

// Limit returns the rate limiting middleware
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	if !rl.cfg.Enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is whitelisted
		if rl.isPathWhitelisted(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Check if IP is whitelisted
		clientIP := rl.getClientIP(r)
		if rl.isIPWhitelisted(clientIP) {
			next.ServeHTTP(w, r)
			return
		}

		// Check if user is authenticated
		userCtx, ok := auth.FromContext(r.Context())
		if ok && userCtx != nil {
			// Use user-based rate limiting for authenticated requests
			rl.userLimiter(next).ServeHTTP(w, r)
		} else {
			// Use IP-based rate limiting for unauthenticated requests
			rl.ipLimiter(next).ServeHTTP(w, r)
		}
	})
}

// LimitByIP returns IP-based rate limiting middleware (for use before auth)
func (rl *RateLimiter) LimitByIP(next http.Handler) http.Handler {
	if !rl.cfg.Enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is whitelisted
		if rl.isPathWhitelisted(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Check if IP is whitelisted
		clientIP := rl.getClientIP(r)
		if rl.isIPWhitelisted(clientIP) {
			next.ServeHTTP(w, r)
			return
		}

		rl.ipLimiter(next).ServeHTTP(w, r)
	})
}

// keyByUserOrIP returns user ID for authenticated requests, or IP for unauthenticated
func (rl *RateLimiter) keyByUserOrIP(r *http.Request) (string, error) {
	if userCtx, ok := auth.FromContext(r.Context()); ok && userCtx != nil {
		return "user:" + userCtx.UserID.String(), nil
	}
	return "ip:" + rl.getClientIP(r), nil
}

// getClientIP extracts the client IP from the request
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// isIPWhitelisted checks if the IP is in the whitelist
func (rl *RateLimiter) isIPWhitelisted(ip string) bool {
	return rl.whitelistIPs[ip]
}

// isPathWhitelisted checks if the path is in the whitelist
func (rl *RateLimiter) isPathWhitelisted(path string) bool {
	// Check exact match
	if rl.whitelistPaths[path] {
		return true
	}

	// Check prefix match for paths ending with /*
	for wp := range rl.whitelistPaths {
		if strings.HasSuffix(wp, "/*") {
			prefix := strings.TrimSuffix(wp, "/*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}

// rateLimitExceededHandler handles rate limit exceeded responses
func (rl *RateLimiter) rateLimitExceededHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := rl.getClientIP(r)
	userID := ""
	if userCtx, ok := auth.FromContext(r.Context()); ok && userCtx != nil {
		userID = userCtx.UserID.String()
	}

	rl.logger.Warn("rate limit exceeded",
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.String("client_ip", clientIP),
		zap.String("user_id", userID),
	)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(`{"error":"rate limit exceeded","message":"Too many requests. Please try again later."}`))
}
