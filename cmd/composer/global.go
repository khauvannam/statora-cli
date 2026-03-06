package composercmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/composer"
	"statora-cli/internal/config"
)

var globalCmd = &cobra.Command{
	Use:   "global <version>",
	Short: "Set the global Composer version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(cfg *config.Config, m *composer.Manager) error {
			if !m.IsInstalled(version) {
				fmt.Printf("Composer %s is not installed. Installing...\n", version)
				if err := m.Install(version); err != nil {
					return fmt.Errorf("auto-install Composer %s: %w", version, err)
				}
			}

			g, err := cfg.LoadGlobal()
			if err != nil {
				return err
			}
			g.Composer = version
			return cfg.WriteGlobal(g)
		})
	},
}
