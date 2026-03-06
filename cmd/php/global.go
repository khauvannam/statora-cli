package phpcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
	"statora-cli/internal/php"
)

var globalCmd = &cobra.Command{
	Use:   "global <version>",
	Short: "Set the global PHP version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(cfg *config.Config, p *php.Plugin) error {
			if !p.IsInstalled(version) {
				fmt.Printf("PHP %s is not installed. Installing...\n", version)
				if err := p.Install(version); err != nil {
					return fmt.Errorf("auto-install PHP %s: %w", version, err)
				}
			}

			g, err := cfg.LoadGlobal()
			if err != nil {
				return err
			}
			g.PHP = version
			return cfg.WriteGlobal(g)
		})
	},
}
