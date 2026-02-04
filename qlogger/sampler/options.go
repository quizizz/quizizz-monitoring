package sampler

import "go.uber.org/zap/zapcore"

// Option is a functional option for configuring the sampler.
// It follows the functional options pattern for flexible configuration.
type Option func(*Config)

// --- Core Sampling Options ---

// WithSampling adds or updates a sampling rule for any log level.
// This is the core extensible method that works for ANY zapcore.Level.
//
// Parameters:
//   - level: The log level to apply sampling to
//   - initial: Number of logs per second that are always logged
//   - thereafter: After initial, log every Nth message (0 = drop all after initial)
//
// Example:
//
//	sampler.WithSampling(zapcore.InfoLevel, 100, 10)  // Log first 100/sec, then 1 in 10
func WithSampling(level zapcore.Level, initial, thereafter int) Option {
	return func(c *Config) {
		c.SetRule(NewRule(level, initial, thereafter))
	}
}

// WithoutSampling explicitly disables sampling for a level.
// All logs at this level will pass through without any sampling.
func WithoutSampling(level zapcore.Level) Option {
	return func(c *Config) {
		c.SetRule(Disable(level))
	}
}

// --- Output Configuration ---

// WithOutput sets the output writer for the logger.
// Default is os.Stdout.
func WithOutput(output zapcore.WriteSyncer) Option {
	return func(c *Config) {
		c.SetOutput(output)
	}
}

// WithEncoderConfig sets a custom encoder configuration.
// If not set, production encoder config is used by default.
func WithEncoderConfig(cfg zapcore.EncoderConfig) Option {
	return func(c *Config) {
		c.SetEncoderConfig(&cfg)
	}
}

// --- Caller Information ---

// WithCaller enables or disables caller information in logs.
// Default is enabled (true).
func WithCaller(enabled bool) Option {
	return func(c *Config) {
		c.SetAddCaller(enabled)
	}
}

// WithCallerSkip sets the number of stack frames to skip when determining caller.
// Useful when wrapping the logger in your own functions.
func WithCallerSkip(skip int) Option {
	return func(c *Config) {
		c.SetCallerSkip(skip)
	}
}

// --- Convenience Options (Syntactic Sugar) ---
// These don't violate OCP as they delegate to WithSampling

// WithDebugSampling is a convenience option for Debug level sampling.
func WithDebugSampling(initial, thereafter int) Option {
	return WithSampling(zapcore.DebugLevel, initial, thereafter)
}

// WithInfoSampling is a convenience option for Info level sampling.
func WithInfoSampling(initial, thereafter int) Option {
	return WithSampling(zapcore.InfoLevel, initial, thereafter)
}

// WithWarnSampling is a convenience option for Warn level sampling.
func WithWarnSampling(initial, thereafter int) Option {
	return WithSampling(zapcore.WarnLevel, initial, thereafter)
}

// WithErrorSampling is a convenience option for Error level sampling.
// Use with caution - sampling errors may hide issues.
func WithErrorSampling(initial, thereafter int) Option {
	return WithSampling(zapcore.ErrorLevel, initial, thereafter)
}

