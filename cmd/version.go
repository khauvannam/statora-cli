package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "1.0.4"

const versionOutput = "Statora " + Version + "\nCopyright (c) 2024-2026 haonam khauvannam (khauvannam)\n"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the statora version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Print(versionOutput)
	},
}

func init() {
	Root.Version = Version
	Root.SetVersionTemplate(versionOutput)
	Root.AddCommand(versionCmd)
}
