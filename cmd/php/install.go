package phpcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/php"
)

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a PHP version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(p *php.Plugin) error {
			return p.Install(version)
		})
	},
}

var _ = fmt.Sprintf // suppress unused import
