package composercmd

import (
	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/composer"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <version>",
	Short: "Uninstall a Composer version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(m *composer.Manager) error {
			return m.Uninstall(version)
		})
	},
}
