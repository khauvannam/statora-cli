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
	dir := ctx.Cfg.Paths.ErrorsDir
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		return ""
	}

	ts := time.Now().UTC()
	filename := ts.Format("2006-01-02T15-04-05Z") + "-error.log"
	path := filepath.Join(dir, filename)

	var sb strings.Builder
	fmt.Fprintf(&sb, "=== Statora Error Log ===\n")
	fmt.Fprintf(&sb, "Timestamp : %s\n", ts.Format(time.RFC3339))
	fmt.Fprintf(&sb, "Version   : %s\n", ctx.Version)
	fmt.Fprintf(&sb, "Stage     : %s\n", stage)
	fmt.Fprintf(&sb, "Error     : %s\n", err.Error())
	if ctx.CapturedOutput != "" {
		fmt.Fprintf(&sb, "\n=== Captured Output ===\n%s\n", ctx.CapturedOutput)
	}

	_ = os.WriteFile(path, []byte(sb.String()), 0o644)
	return path
}
