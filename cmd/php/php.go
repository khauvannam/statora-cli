package phpcmd

import (
	"github.com/spf13/cobra"

	"statora-cli/cmd"
)

// Cmd is the `statora php` parent command.
var Cmd = &cobra.Command{
	Use:   "php",
	Short: "Manage PHP versions",
}

func init() {
	Cmd.AddCommand(
		installCmd,
		uninstallCmd,
		listCmd,
		globalCmd,
		localCmd,
		currentCmd,
		whichCmd,
		rehashCmd,
	)
	cmd.Root.AddCommand(Cmd)
}
