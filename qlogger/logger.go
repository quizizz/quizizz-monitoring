package qlogger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/quizizz/quizizz-monitoring/qlogger/sampler"
)

// Field is an alias for zap.Field for convenience.
type Field = zap.Field

// Type aliases for zap field constructors for convenience.
var (
	String     = zap.String
	Int        = zap.Int
	Int64      = zap.Int64
	Float64    = zap.Float64
	Bool       = zap.Bool
	Any        = zap.Any
	Err        = zap.Error
	Duration   = zap.Duration
	Time       = zap.Time
	Stringer   = zap.Stringer
	Int32      = zap.Int32
	Uint       = zap.Uint
	Uint64     = zap.Uint64
	Binary     = zap.Binary
	ByteString = zap.ByteString
	Namespace  = zap.Namespace
	Stack      = zap.Stack
	Object     = zap.Object
	Array      = zap.Array
	Reflect    = zap.Reflect
)

// Logger wraps a zap.Logger with additional functionality.
type Logger struct {
	log         *zap.Logger
	environment string
}

// NewDevelopment creates a Logger for local/development environments.
// Uses zap.NewDevelopmentConfig() with colored, human-readable output.
// No sampling is applied - all logs are printed.
//
// Example:
//
//	logger, err := qlogger.NewDevelopment()
//	if err != nil {
//	    panic(err)
//	}
//	defer logger.Sync()
//	logger.Info(ctx, "message", zap.String("key", "value"))
func NewDevelopment() (*Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	zapLogger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{
		log:         zapLogger,
		environment: "dev",
	}, nil
}

// NewProduction creates a Logger for production environments with sampling.
// Uses sampler.New() with JSON output and configurable sampling rules.
//
// Example:
//
//	logger, err := qlogger.NewProduction(
//	    sampler.WithInfoSampling(100, 10),
//	    sampler.WithWarnSampling(50, 5),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer logger.Sync()
//	logger.Info(ctx, "message", zap.String("key", "value"))
func NewProduction(samplerOpts ...sampler.Option) (*Logger, error) {
	var zapLogger *zap.Logger
	var err error
	if len(samplerOpts) == 0 {
		zapLogger, err = zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
	// Add default caller options
		samplerOpts = append(samplerOpts,
			sampler.WithCaller(true),
			sampler.WithCallerSkip(1),
		)

		zapLogger, err = sampler.New(samplerOpts...)
	}
	if err != nil {
		return nil, err
	}

	return &Logger{
		log:         zapLogger,
		environment: "prod",
	}, nil
}

// NewNop creates a no-op Logger that discards all logs.
// Useful for testing or when you want to disable logging entirely.
//
// Example:
//
//	logger := qlogger.NewNop()
//	logger.Info(ctx, "this will be discarded")
func NewNop() *Logger {
	return &Logger{
		log:         zap.NewNop(),
		environment: "nop",
	}
}

// New creates a new Logger instance with the given options.
// For explicit control, prefer using NewDevelopment(), NewProduction(), or NewNop().
//
// Example:
//
//	logger, err := qlogger.New(
//	    qlogger.WithEnvironment("prod"),
//	    qlogger.WithInfoSampling(100, 10),
//	    qlogger.WithWarnSampling(50, 5),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer logger.Sync()
//	logger.Info(ctx, "message", zap.String("key", "value"))
func New(opts ...LoggerOption) (*Logger, error) {
	cfg := newLoggerConfig()

	// Apply user options
	for _, opt := range opts {
		opt(cfg)
	}

	// For local/dev environments, use NewDevelopment (no sampling)
	if cfg.development {
		return NewDevelopment()
	}

	// For production with sampling options, use NewProduction
	if len(cfg.samplerOptions) > 0 {
		return NewProduction(cfg.samplerOptions...)
	}

	// Default production without sampling
	zapLogger, err := zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(cfg.callerSkip))
	if err != nil {
		return nil, err
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
func (l *Logger) InfoWithTraceIdCtx(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Info(msg, fields...)
}

// Error logs a message at Error level with context.
func (l *Logger) ErrorWithTraceIdCtx(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Error(msg, fields...)
}

// Warn logs a message at Warn level with context.
func (l *Logger) WarnWithTraceIdCtx(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Warn(msg, fields...)
}

// Debug logs a message at Debug level with context.
// In production, debug logs are suppressed unless explicitly enabled.
func (l *Logger) DebugWithTraceIdCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if l.environment == "prod" || l.environment == "production" {
		return
	}
	fields = addDefaultFields(ctx, fields...)
	l.log.Debug(msg, fields...)
}

// Fatal logs a message at Fatal level with context, then calls os.Exit(1).
func (l *Logger) FatalWithTraceIdCtx(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Fatal(msg, fields...)
}

// Panic logs a message at Panic level with context, then panics.
func (l *Logger) PanicWithTraceIdCtx(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.Panic(msg, fields...)
}

func (l *Logger) DPanicWithTraceIdCtx(ctx context.Context, msg string, fields ...zap.Field) {
	fields = addDefaultFields(ctx, fields...)
	l.log.DPanic(msg, fields...)
}

// --- Context-free logging methods ---

// Info logs a message at Info level without context.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}

// Error logs a message at Error level without context.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}

// Warn logs a message at Warn level without context.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
}

// Debug logs a message at Debug level without context.
// In production, debug logs are suppressed.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	if l.environment == "prod" || l.environment == "production" {
		return
	}
	l.log.Debug(msg, fields...)
}

// Fatal logs a message at Fatal level without context, then exits.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log.Fatal(msg, fields...)
}

// Panic logs a message at Panic level without context, then panics.
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.log.Panic(msg, fields...)
}

func (l *Logger) DPanic(msg string, fields ...zap.Field) {
	l.log.DPanic(msg, fields...)
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
		fields = append(fields, zap.String("trace_id", traceID))
	}

	return fields
}
