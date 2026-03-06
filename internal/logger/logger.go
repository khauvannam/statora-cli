package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a zap.Logger configured for production or development mode.
// Debug mode enables human-readable output with full stack traces.
func New(debug bool) (*zap.Logger, error) {
	if debug {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return cfg.Build()
	}
	return zap.NewProduction()
}

// Must is New but panics on error.
func Must(debug bool) *zap.Logger {
	l, err := New(debug)
	if err != nil {
		panic(err)
	}
	return l
}
