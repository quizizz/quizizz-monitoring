package qlogger

import (
	"github.com/quizizz/quizizz-monitoring/qlogger/sampler"
	"go.uber.org/zap/zapcore"
)

// LoggerOption is a functional option for configuring the logger.
type LoggerOption func(*loggerConfig)

// loggerConfig holds the logger configuration.
type loggerConfig struct {
	environment    string
	samplerOptions []sampler.Option
	addCaller      bool
	callerSkip     int
	development    bool
}

// newLoggerConfig creates a new config with defaults.
func newLoggerConfig() *loggerConfig {
	return &loggerConfig{
		environment: "prod",
		addCaller:   true,
		callerSkip:  1, // Skip the wrapper functions
		development: false,
	}
}

// --- Environment Options ---

// WithEnvironment sets the environment (e.g., "prod", "staging", "local").
// In "local" or "dev" environments, development-friendly formatting is used.
func WithEnvironment(env string) LoggerOption {
	return func(c *loggerConfig) {
		c.environment = env
		c.development = env == "local" || env == "dev" || env == "development"
	}
}

// WithDevelopmentMode enables development mode formatting.
// This is automatically enabled for "local", "dev", and "development" environments.
func WithDevelopmentMode(enabled bool) LoggerOption {
	return func(c *loggerConfig) {
		c.development = enabled
	}
}

// --- Sampling Options ---

// WithSamplerOptions sets the sampling options for the logger.
// These options are passed to the sampler builder.
func WithSamplerOptions(opts ...sampler.Option) LoggerOption {
	return func(c *loggerConfig) {
		c.samplerOptions = append(c.samplerOptions, opts...)
	}
}

// WithSampling adds a sampling rule for a specific log level.
// Convenience method that wraps sampler.WithSampling.
func WithSampling(level zapcore.Level, initial, thereafter int) LoggerOption {
	return func(c *loggerConfig) {
		c.samplerOptions = append(c.samplerOptions, sampler.WithSampling(level, initial, thereafter))
	}
}

// WithInfoSampling adds sampling for Info level logs.
func WithInfoSampling(initial, thereafter int) LoggerOption {
	return func(c *loggerConfig) {
		c.samplerOptions = append(c.samplerOptions, sampler.WithInfoSampling(initial, thereafter))
	}
}

// WithWarnSampling adds sampling for Warn level logs.
func WithWarnSampling(initial, thereafter int) LoggerOption {
	return func(c *loggerConfig) {
		c.samplerOptions = append(c.samplerOptions, sampler.WithWarnSampling(initial, thereafter))
	}
}

// WithDebugSampling adds sampling for Debug level logs.
func WithDebugSampling(initial, thereafter int) LoggerOption {
	return func(c *loggerConfig) {
		c.samplerOptions = append(c.samplerOptions, sampler.WithDebugSampling(initial, thereafter))
	}
}

// WithErrorSampling adds sampling for Error level logs.
// Use with caution - sampling errors may hide issues.
func WithErrorSampling(initial, thereafter int) LoggerOption {
	return func(c *loggerConfig) {
		c.samplerOptions = append(c.samplerOptions, sampler.WithErrorSampling(initial, thereafter))
	}
}

// --- Caller Options ---

// WithCaller enables or disables caller information in logs.
func WithCaller(enabled bool) LoggerOption {
	return func(c *loggerConfig) {
		c.addCaller = enabled
	}
}

// WithCallerSkip sets the number of stack frames to skip for caller info.
// Default is 1 to skip the qlogger wrapper functions.
func WithCallerSkip(skip int) LoggerOption {
	return func(c *loggerConfig) {
		c.callerSkip = skip
	}
}

