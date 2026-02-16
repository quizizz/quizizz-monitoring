package sampler

import (
	"os"

	"go.uber.org/zap/zapcore"
)

// Config holds the sampler configuration.
// It is used by the Builder to construct a sampled logger.
type Config struct {
	rules         map[zapcore.Level]*Rule
	output        zapcore.WriteSyncer
	encoderConfig *zapcore.EncoderConfig
	addCaller     bool
	callerSkip    int
}

// NewConfig creates a new Config with sensible defaults.
func NewConfig() *Config {
	return &Config{
		rules:      make(map[zapcore.Level]*Rule),
		output:     zapcore.AddSync(os.Stdout),
		addCaller:  true,
		callerSkip: 1,
	}
}

// GetRule returns the sampling rule for a given level, if configured.
func (c *Config) GetRule(level zapcore.Level) (*Rule, bool) {
	rule, exists := c.rules[level]
	return rule, exists
}

// SetRule sets or updates a sampling rule for a level.
func (c *Config) SetRule(rule *Rule) {
	if rule != nil {
		c.rules[rule.Level] = rule
	}
}

// GetOutput returns the configured output writer.
func (c *Config) GetOutput() zapcore.WriteSyncer {
	return c.output
}

// SetOutput sets the output writer.
func (c *Config) SetOutput(output zapcore.WriteSyncer) {
	c.output = output
}

// GetEncoderConfig returns the encoder configuration, if set.
func (c *Config) GetEncoderConfig() *zapcore.EncoderConfig {
	return c.encoderConfig
}

// SetEncoderConfig sets the encoder configuration.
func (c *Config) SetEncoderConfig(cfg *zapcore.EncoderConfig) {
	c.encoderConfig = cfg
}

// AddCallerEnabled returns whether caller information should be added to logs.
func (c *Config) AddCallerEnabled() bool {
	return c.addCaller
}

// SetAddCaller enables or disables caller information in logs.
func (c *Config) SetAddCaller(enabled bool) {
	c.addCaller = enabled
}

// GetCallerSkip returns the number of stack frames to skip for caller info.
func (c *Config) GetCallerSkip() int {
	return c.callerSkip
}

// SetCallerSkip sets the number of stack frames to skip for caller info.
func (c *Config) SetCallerSkip(skip int) {
	c.callerSkip = skip
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	newConfig := &Config{
		rules:      make(map[zapcore.Level]*Rule),
		output:     c.output,
		addCaller:  c.addCaller,
		callerSkip: c.callerSkip,
	}

	if c.encoderConfig != nil {
		encoderCopy := *c.encoderConfig
		newConfig.encoderConfig = &encoderCopy
	}

	for level, rule := range c.rules {
		if rule != nil {
			ruleCopy := *rule
			newConfig.rules[level] = &ruleCopy
		}
	}

	return newConfig
}
