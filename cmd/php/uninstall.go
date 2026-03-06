package phpcmd

import (
	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/php"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <version>",
	Short: "Uninstall a PHP version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(p *php.Plugin) error {
			return p.Uninstall(version)
		})
	},
}
