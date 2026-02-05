package main

import (
	"context"
	"os"

	"github.com/quizizz/quizizz-monitoring/qlogger"
	"go.uber.org/zap"
)

func main() {
	// Set environment (this would typically come from your config)
	os.Setenv("ENV", "prod")

	// Create a new logger instance with sampling options
	logger, err := qlogger.New(
		qlogger.WithEnvironment("prod"),
		qlogger.WithInfoSampling(100, 10), // First 100/sec always logged, then 1 in 10
		qlogger.WithWarnSampling(50, 5),   // First 50/sec always logged, then 1 in 5
	)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Create a context (typically from HTTP request or similar)
	ctx := context.Background()
	ctx = qlogger.WithTraceID(ctx, "abc-123-xyz")

	// Log with context - automatically includes traceId
	logger.Info(ctx, "Application started",
		zap.String("version", "1.0.0"),
		zap.String("environment", "prod"),
	)

	// Log without context
	logger.InfoWithoutCtx("Configuration loaded",
		zap.Int("configItems", 42),
	)

	// Simulate HTTP response logging
	logger.Info(ctx, "http",
		zap.String("type", "http-response"),
		zap.Any("data", map[string]interface{}{
			"status":  200,
			"method":  "GET",
			"path":    "/api/users",
			"latency": "45ms",
		}),
	)

	// Error logging (never sampled by default)
	logger.Error(ctx, "Database connection failed",
		zap.String("host", "db.example.com"),
		zap.Int("port", 5432),
		zap.Error(nil), // Would contain actual error
	)

	logger.Info(ctx, "Application shutting down")
}
