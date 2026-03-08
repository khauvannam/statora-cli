package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "1.0.3"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the statora version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Statora %s\n", Version)
		fmt.Println("Copyright (c) 2024-2026 The Statora Contributors")
	},
}

func init() {
	Root.AddCommand(versionCmd)
}
