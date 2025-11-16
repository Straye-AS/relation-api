package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
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

			fields := []zap.Field{
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status_code", rw.statusCode),
				zap.Int64("response_size", rw.written),
				zap.Duration("duration", duration),
			}

			// Add user context if available
			if userCtx, ok := auth.FromContext(r.Context()); ok {
				fields = append(fields,
					zap.String("user_id", userCtx.UserID.String()),
					zap.String("user_name", userCtx.DisplayName),
				)
			}

			logger.Info(
				fmt.Sprintf("%s %-30s -> %3d (%s)",
					r.Method,
					r.URL.Path,
					rw.statusCode,
					duration.Truncate(time.Microsecond),
				),
				fields...,
			)
		})
	}
}
