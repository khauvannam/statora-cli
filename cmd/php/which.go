package phpcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/php"
)

var whichCmd = &cobra.Command{
	Use:   "which <version>",
	Short: "Print the path to the PHP binary for a version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(p *php.Plugin) error {
			path, err := p.Which(version)
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		})
	},
}
