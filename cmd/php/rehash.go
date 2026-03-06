package phpcmd

import (
	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/php"
)

var rehashCmd = &cobra.Command{
	Use:   "rehash",
	Short: "Rebuild the dispatch cache",
	RunE: func(c *cobra.Command, _ []string) error {
		return app.Invoke(cmd.Debug(), func(p *php.Plugin) error {
			return p.Rehash()
		})
	},
}
