package main

import (
	"context"

	"go.quizizz.org/monitoring/qlogger"
	"go.quizizz.org/monitoring/qlogger/sampler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Example 1: Using predefined sampling options
	samplingOpts := []sampler.Option{
		sampler.WithInfoSampling(50, 5),   // Info: 50/sec, then 1 in 5
		sampler.WithWarnSampling(20, 2),   // Warn: 20/sec, then 1 in 2
		sampler.WithDebugSampling(10, 10), // Debug: 10/sec, then 1 in 10
		// Error is not sampled by default - all errors pass through
	}

	logger, err := qlogger.New(
		qlogger.WithEnvironment("prod"),
		qlogger.WithSamplerOptions(samplingOpts...),
	)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ctx := qlogger.WithTraceID(context.Background(), "trace-001")

	// Generate many logs to see sampling in action
	for i := 0; i < 200; i++ {
		logger.Info(ctx, "Processing request",
			zap.Int("iteration", i),
		)
	}

	// Example 2: Custom sampling for any level using zapcore.Level
	customOpts := []sampler.Option{
		sampler.WithSampling(zapcore.InfoLevel, 100, 20),
		sampler.WithSampling(zapcore.WarnLevel, 50, 10),
		// sampler.WithSampling(zapcore.ErrorLevel, 10, 2), // Not recommended to sample errors
	}

	logger2, err := qlogger.New(
		qlogger.WithEnvironment("prod"),
		qlogger.WithSamplerOptions(customOpts...),
	)
	if err != nil {
		panic(err)
	}
	defer logger2.Sync()

	logger2.Info(ctx, "Custom sampling configured")
}
