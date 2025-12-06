package logger

import (
	"fmt"

	"github.com/straye-as/relation-api/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new structured logger
func NewLogger(cfg *config.LoggingConfig, appCfg *config.AppConfig) (*zap.Logger, error) {
	var zapCfg zap.Config

	if cfg.Format == "json" || appCfg.Environment == "production" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
		// Configure a cleaner console output
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
		zapCfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		zapCfg.EncoderConfig.EncodeCaller = nil // Hide caller info for cleaner output
		zapCfg.DisableCaller = true
		zapCfg.DisableStacktrace = true
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// Only add initial fields in production (JSON format)
	if cfg.Format == "json" || appCfg.Environment == "production" {
		zapCfg.InitialFields = map[string]interface{}{
			"app":         appCfg.Name,
			"environment": appCfg.Environment,
		}
	}

	logger, err := zapCfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return logger, nil
}

// WithRequest adds request context to logger
func WithRequest(logger *zap.Logger, method, path, requestID string) *zap.Logger {
	return logger.With(
		zap.String("method", method),
		zap.String("path", path),
		zap.String("request_id", requestID),
	)
}

// WithUser adds user context to logger
func WithUser(logger *zap.Logger, userID, displayName string) *zap.Logger {
	return logger.With(
		zap.String("user_id", userID),
		zap.String("user_name", displayName),
	)
}
