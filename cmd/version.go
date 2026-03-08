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
		fmt.Println(Version)
	},
}

func init() {
	Root.AddCommand(versionCmd)
}
