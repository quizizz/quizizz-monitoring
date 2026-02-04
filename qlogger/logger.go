package qlogger

import (
	"context"
	"os"

	"go.quizizz.org/monitoring/qlogger/sampler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field is an alias for zap.Field for convenience.
type Field = zap.Field

// Type aliases for zap field constructors for convenience.
var (
	String  = zap.String
	Int     = zap.Int
	Int64   = zap.Int64
	Float64 = zap.Float64
	Bool    = zap.Bool
	Any     = zap.Any
	Err     = zap.Error
	Duration  = zap.Duration
    Time      = zap.Time
    Stringer  = zap.Stringer
    Int32     = zap.Int32
    Uint      = zap.Uint
    Uint64    = zap.Uint64
    Binary    = zap.Binary
    ByteString = zap.ByteString
    Namespace = zap.Namespace
    Stack     = zap.Stack
    Object    = zap.Object
    Array     = zap.Array
    Reflect   = zap.Reflect
)

// Logger wraps a zap.Logger with additional functionality.
type Logger struct {
	log         *zap.Logger
	environment string
}

// New creates a new Logger instance with the given options.
//
// Example:
//
//	logger, err := wlogger.New(
//	    wlogger.WithEnvironment("prod"),
//	    wlogger.WithInfoSampling(100, 10),
//	    wlogger.WithWarnSampling(50, 5),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer logger.Sync()
//	logger.Info(ctx, "message", zap.String("key", "value"))
func New(opts ...LoggerOption) (*Logger, error) {
	cfg := newLoggerConfig()

	// Apply environment from ENV var if not explicitly set
	if env := os.Getenv("ENV"); env != "" {
		cfg.environment = env
		cfg.development = env == "local" || env == "dev" || env == "development"
	}

	// Apply user options
	for _, opt := range opts {
		opt(cfg)
	}

	var zapLogger *zap.Logger
	var err error

	// If we have sampling options, use the sampler builder
	if len(cfg.samplerOptions) > 0 && cfg.environment == "prod" {
		// Add caller configuration to sampler options
		cfg.samplerOptions = append(cfg.samplerOptions,
			sampler.WithCaller(cfg.addCaller),
			sampler.WithCallerSkip(cfg.callerSkip),
		)

		zapLogger, err = sampler.New(cfg.samplerOptions...)
		if err != nil {
			return nil, err
		}
	} else {
		// Use standard zap configuration
		if cfg.development {
			config := zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			zapLogger, err = config.Build()
		} else {
			zapLogger, err = zap.NewProduction()
		}
		if err != nil {
			return nil, err
		}

		// Add caller if enabled
		if cfg.addCaller {
			zapLogger = zapLogger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(cfg.callerSkip))
		}
	}

	return &Logger{
		log:         zapLogger,
		environment: cfg.environment,
	}, nil
}

// MustNew creates a new Logger instance, panicking on error.
// Use this only when you're certain the configuration is valid.
func MustNew(opts ...LoggerOption) *Logger {
	logger, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return logger
}

// Zap returns the underlying zap.Logger for advanced use cases.
func (l *Logger) Zap() *zap.Logger {
	return l.log
}

// Environment returns the configured environment.
func (l *Logger) Environment() string {
	return l.environment
}

// Sync flushes any buffered log entries.
// Applications should call this before exiting.
func (l *Logger) Sync() error {
	return l.log.Sync()
}

// --- Context-aware logging methods ---

// Info logs a message at Info level with context.
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Info(msg, fields...)
}

// Error logs a message at Error level with context.
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Error(msg, fields...)
}

// Warn logs a message at Warn level with context.
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Warn(msg, fields...)
}

// Debug logs a message at Debug level with context.
// In production, debug logs are suppressed unless explicitly enabled.
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if l.environment == "prod" || l.environment == "production" {
		return
	}
	fields = addDefaultFields(ctx, fields...)
	l.log.Debug(msg, fields...)
}

// Fatal logs a message at Fatal level with context, then calls os.Exit(1).
func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Fatal(msg, fields...)
}

// Panic logs a message at Panic level with context, then panics.
func (l *Logger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Panic(msg, fields...)
}

// --- Context-free logging methods ---

// InfoWithoutCtx logs a message at Info level without context.
func (l *Logger) InfoWithoutCtx(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}

// ErrorWithoutCtx logs a message at Error level without context.
func (l *Logger) ErrorWithoutCtx(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}

// WarnWithoutCtx logs a message at Warn level without context.
func (l *Logger) WarnWithoutCtx(msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
}

// DebugWithoutCtx logs a message at Debug level without context.
// In production, debug logs are suppressed.
func (l *Logger) DebugWithoutCtx(msg string, fields ...zap.Field) {
	if l.environment == "prod" || l.environment == "production" {
		return
	}
	l.log.Debug(msg, fields...)
}

// FatalWithoutCtx logs a message at Fatal level without context, then exits.
func (l *Logger) FatalWithoutCtx(msg string, fields ...zap.Field) {
	l.log.Fatal(msg, fields...)
}

// PanicWithoutCtx logs a message at Panic level without context, then panics.
func (l *Logger) PanicWithoutCtx(msg string, fields ...zap.Field) {
	l.log.Panic(msg, fields...)
}

// --- Utility methods ---

// With creates a child logger with the given fields attached.
func (l *Logger) With(fields ...zap.Field) *Logger {
	childLog := l.log.With(fields...)
	return &Logger{
		log:         childLog,
		environment: l.environment,
	}
}

// WithContext creates a child logger with context fields attached.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := addDefaultFields(ctx)
	return l.With(fields...)
}

// --- Internal helper functions ---

// addDefaultFields adds standard fields like traceId to log entries.
func addDefaultFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	traceID := GetTraceIDFromCtx(ctx)
	if traceID != "" {
		fields = append(fields, zap.String("traceId", traceID))
	}
	return fields
}
