package composercmd

import (
	"os"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/composer"
	"statora-cli/internal/config"
)

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a Composer version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(m *composer.Manager, cfg *config.Config) error {
			concrete, err := m.Install(version)
			if err != nil {
				return err
			}
			if concrete != version {
				// Update global config if it referenced the partial version.
				if g, loadErr := cfg.LoadGlobal(); loadErr == nil && g.Composer == version {
					g.Composer = concrete
					_ = cfg.WriteGlobal(g)
				}
				// Update .statora in cwd if it referenced the partial version.
				cwd, wdErr := os.Getwd()
				if wdErr == nil {
					if proj, found, loadErr := config.LoadProject(cwd); loadErr == nil && found && proj.Composer == version {
						proj.Composer = concrete
						_ = config.WriteProject(cwd, proj)
					}
				}
			}
			return nil
		})
	},
}
