package composercmd

import (
	"github.com/spf13/cobra"

	"statora-cli/cmd"
)

// Cmd is the `statora composer` parent command.
var Cmd = &cobra.Command{
	Use:   "composer",
	Short: "Manage Composer versions",
}

func init() {
	Cmd.AddCommand(
		installCmd,
		uninstallCmd,
		listCmd,
		globalCmd,
		localCmd,
		currentCmd,
		compatCmd,
	)
	cmd.Root.AddCommand(Cmd)
}
