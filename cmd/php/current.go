package phpcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active PHP version",
	RunE: func(c *cobra.Command, _ []string) error {
		return app.Invoke(cmd.Debug(), func(cfg *config.Config) error {
			v := dispatch.ReadCache(cfg, dispatch.KeyPHPActive)
			if v == "" {
				fmt.Println("No active PHP version. Run `statora switch`.")
				return nil
			}
			fmt.Println(v)
			return nil
		})
	},
}
