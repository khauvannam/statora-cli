package extcmd

import (
	"github.com/spf13/cobra"

	"statora-cli/cmd"
)

// Cmd is the `statora ext` parent command.
var Cmd = &cobra.Command{
	Use:   "ext",
	Short: "Manage PHP extensions",
}

func init() {
	Cmd.AddCommand(
		installCmd,
		uninstallCmd,
		enableCmd,
		disableCmd,
		listCmd,
		infoCmd,
	)
	cmd.Root.AddCommand(Cmd)
}
