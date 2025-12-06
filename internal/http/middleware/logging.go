package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logging middleware logs HTTP requests
func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.New().String()

			// Add request ID to context
			r.Header.Set("X-Request-ID", requestID)

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Color-code status
			statusColor := "\033[32m" // Green for 2xx
			if rw.statusCode >= 400 && rw.statusCode < 500 {
				statusColor = "\033[33m" // Yellow for 4xx
			} else if rw.statusCode >= 500 {
				statusColor = "\033[31m" // Red for 5xx
			}
			resetColor := "\033[0m"

			// Simple, readable log format
			msg := fmt.Sprintf("%s%-7s%s %-40s %s%3d%s  %10s",
				"\033[36m", r.Method, resetColor, // Cyan method
				r.URL.Path,
				statusColor, rw.statusCode, resetColor,
				duration.Truncate(time.Microsecond),
			)

			// Only add extra fields for errors or slow requests
			if rw.statusCode >= 400 || duration > 500*time.Millisecond {
				logger.Info(msg,
					zap.String("request_id", requestID),
					zap.String("remote_addr", r.RemoteAddr),
				)
			} else {
				logger.Info(msg)
			}
		})
	}
}
