package installer

import (
	"fmt"

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
// Progress is printed to stdout so the user can see what's happening.
func (p *Pipeline) Run(ctx *Context) error {
	for i, s := range p.stages {
		label := color.CyanString("[%d/%d]", i+1, len(p.stages))
		fmt.Printf("%s %s\n", label, s.Name())

		if err := s.Run(ctx); err != nil {
			return fmt.Errorf("stage %q failed: %w", s.Name(), err)
		}

		color.Green("     ✓ %s", s.Name())
	}
	return nil
}
