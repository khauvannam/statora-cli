package installer

import (
	"go.uber.org/zap"

	"statora-cli/internal/config"
)

// Context carries shared state through pipeline stages.
type Context struct {
	// Version being installed (e.g. "8.2.15").
	Version string
	// Cfg holds resolved application paths.
	Cfg *config.Config
	// Log is the logger for the management path.
	Log *zap.Logger
	// Data allows stages to pass arbitrary values to downstream stages.
	Data map[string]any
	// CapturedOutput holds stderr/stdout captured by stages for error reporting.
	CapturedOutput string
}

// Stage is a single step in an install pipeline.
type Stage interface {
	Name() string
	Run(ctx *Context) error
}
