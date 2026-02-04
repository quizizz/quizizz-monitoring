package sampler

import "go.uber.org/zap/zapcore"

// Rule defines sampling behavior for a specific log level.
// It determines how many logs are allowed through before sampling kicks in.
type Rule struct {
	// Level is the log level this rule applies to
	Level zapcore.Level

	// Initial is the number of logs per second that are always logged
	// before sampling begins
	Initial int

	// Thereafter determines the sampling rate after Initial logs.
	// After Initial logs per second, only every Nth log is recorded.
	// Set to 0 to drop all logs after Initial.
	Thereafter int

	// Enabled indicates whether this sampling rule is active.
	// When false, all logs at this level pass through without sampling.
	Enabled bool
}

// NewRule creates a new sampling rule for the given level.
func NewRule(level zapcore.Level, initial, thereafter int) *Rule {
	return &Rule{
		Level:      level,
		Initial:    initial,
		Thereafter: thereafter,
		Enabled:    true,
	}
}

// Disable returns a disabled rule for the given level.
// Disabled rules allow all logs to pass through without sampling.
func Disable(level zapcore.Level) *Rule {
	return &Rule{
		Level:   level,
		Enabled: false,
	}
}

// IsValid checks if the rule has valid configuration values.
func (r *Rule) IsValid() bool {
	if r == nil {
		return false
	}
	if !r.Enabled {
		return true // Disabled rules are always valid
	}
	return r.Initial >= 0 && r.Thereafter >= 0
}
