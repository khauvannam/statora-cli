package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Pipeline runs a sequence of Stage implementations in order.
type Pipeline struct {
	stages []Stage
}

// New creates a Pipeline from the provided stages.
func New(stages ...Stage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Run executes each stage in order, stopping on the first error.
// On failure it writes a timestamped error log to ~/.statora/errors/.
func (p *Pipeline) Run(ctx *Context) error {
	for i, s := range p.stages {
		label := color.CyanString("[%d/%d]", i+1, len(p.stages))
		fmt.Printf("%s %s\n", label, s.Name())

		if err := s.Run(ctx); err != nil {
			stageErr := fmt.Errorf("stage %q failed: %w", s.Name(), err)
			logPath := writeErrorLog(ctx, s.Name(), stageErr)
			if logPath != "" {
				color.Yellow("  Error log written: %s", logPath)
			}
			return stageErr
		}

		color.Green("     ✓ %s", s.Name())
	}
	return nil
}

func writeErrorLog(ctx *Context, stage string, err error) string {
	if ctx.Cfg == nil {
		return ""
	}

	category := ctx.Category
	if category == "" {
		category = "general"
	}

	dir := filepath.Join(ctx.Cfg.Paths.ErrorsDir, category)
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return ""
	}

	ts := time.Now().UTC()
	// One file per day: errors/php/2026-03-06.log
	path := filepath.Join(dir, ts.Format("2006-01-02")+".log")

	f, fErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if fErr != nil {
		return ""
	}
	defer f.Close()

	var sb strings.Builder
	fmt.Fprintf(&sb, "=== Error at %s ===\n", ts.Format(time.RFC3339))
	fmt.Fprintf(&sb, "Version : %s\n", ctx.Version)
	fmt.Fprintf(&sb, "Stage   : %s\n", stage)
	fmt.Fprintf(&sb, "Error   : %s\n", err.Error())
	if ctx.CapturedOutput != "" {
		fmt.Fprintf(&sb, "\n--- Captured Output ---\n%s\n", ctx.CapturedOutput)
	}
	fmt.Fprintf(&sb, "\n")

	_, _ = f.WriteString(sb.String())
	return path
}
