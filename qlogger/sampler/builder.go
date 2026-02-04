package sampler

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Builder constructs sampled loggers using functional options.
// It provides a fluent API for configuring and building loggers.
type Builder struct {
	config *Config
}

// NewBuilder creates a new Builder with optional configuration options.
//
// Example:
//
//	builder := sampler.NewBuilder(
//	    sampler.WithInfoSampling(100, 10),
//	    sampler.WithWarnSampling(50, 5),
//	)
func NewBuilder(opts ...Option) *Builder {
	config := NewConfig()
	for _, opt := range opts {
		opt(config)
	}
	return &Builder{config: config}
}

// Apply adds more options to the builder, allowing chaining after creation.
//
// Example:
//
//	builder := sampler.NewBuilder().
//	    Apply(sampler.WithInfoSampling(100, 10)).
//	    Apply(sampler.WithOutput(customWriter))
func (b *Builder) Apply(opts ...Option) *Builder {
	for _, opt := range opts {
		opt(b.config)
	}
	return b
}

// Config returns the current configuration for inspection.
// The returned config is a clone to prevent external modification.
func (b *Builder) Config() *Config {
	return b.config.Clone()
}

// Build creates the configured zap.Logger.
// Returns an error if the configuration is invalid.
func (b *Builder) Build() (*zap.Logger, error) {
	if err := b.validate(); err != nil {
		return nil, fmt.Errorf("sampler config validation failed: %w", err)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	if b.config.encoderConfig != nil {
		encoderConfig = *b.config.encoderConfig
	}
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Define all standard levels (excluding Debug which requires explicit enabling)
	allLevels := []zapcore.Level{
	
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.DPanicLevel,
		zapcore.PanicLevel,
		zapcore.FatalLevel,
	}

	var cores []zapcore.Core
	for _, level := range allLevels {
		core := b.createCoreForLevel(encoder, level)
		cores = append(cores, core)
	}

	combinedCore := zapcore.NewTee(cores...)

	var zapOpts []zap.Option
	if b.config.addCaller {
		zapOpts = append(zapOpts, zap.AddCaller())
		if b.config.callerSkip > 0 {
			zapOpts = append(zapOpts, zap.AddCallerSkip(b.config.callerSkip))
		}
	}

	return zap.New(combinedCore, zapOpts...), nil
}

// validate checks if the configuration is valid.
func (b *Builder) validate() error {
	if b.config == nil {
		return errors.New("config is nil")
	}

	if b.config.output == nil {
		return errors.New("output writer is nil")
	}

	// Validate sampling rules
	for level, rule := range b.config.rules {
		if rule == nil {
			continue
		}
		if !rule.IsValid() {
			return fmt.Errorf("invalid sampling rule for level %s", level)
		}
	}

	return nil
}

// createCoreForLevel creates a core for a specific level, applying sampling if configured.
func (b *Builder) createCoreForLevel(encoder zapcore.Encoder, level zapcore.Level) zapcore.Core {
	levelFilter := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == level
	})

	core := zapcore.NewCore(encoder, b.config.output, levelFilter)

	// Check if there's a sampling rule for this level
	if rule, exists := b.config.rules[level]; exists && rule.Enabled {
		return zapcore.NewSamplerWithOptions(core, time.Second, rule.Initial, rule.Thereafter)
	}

	return core
}

// New creates a logger with the given options.
// This is a shorthand for NewBuilder(opts...).Build()
//
// Example:
//
//	logger, err := sampler.New(
//	    sampler.WithInfoSampling(100, 10),
//	    sampler.WithWarnSampling(50, 5),
//	)
func New(opts ...Option) (*zap.Logger, error) {
	return NewBuilder(opts...).Build()
}

// MustNew creates a logger with the given options, panicking on error.
// Use this only when you're certain the configuration is valid.
func MustNew(opts ...Option) *zap.Logger {
	logger, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return logger
}
