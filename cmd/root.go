package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var debug bool

// Root is the top-level cobra command.
var Root = &cobra.Command{
	Use:   "statora",
	Short: "PHP version manager",
	Long:  "Statora manages PHP versions, Composer versions, and PHP extensions per project.",
}

func init() {
	Root.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
}

// Debug returns the global debug flag value.
func Debug() bool { return debug }

// Execute runs the root command.
func Execute() {
	if err := Root.Execute(); err != nil {
		os.Exit(1)
	}
}
