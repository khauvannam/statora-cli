package phpcmd

import (
	"os"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
	"statora-cli/internal/php"
)

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a PHP version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(p *php.Plugin, cfg *config.Config) error {
			concrete, err := p.Install(version)
			if err != nil {
				return err
			}
			if concrete != version {
				// Update global config if it referenced the partial version.
				if g, loadErr := cfg.LoadGlobal(); loadErr == nil && g.PHP == version {
					g.PHP = concrete
					_ = cfg.WriteGlobal(g)
				}
				// Update .statora in cwd if it referenced the partial version.
				cwd, wdErr := os.Getwd()
				if wdErr == nil {
					if proj, found, loadErr := config.LoadProject(cwd); loadErr == nil && found && proj.PHP == version {
						proj.PHP = concrete
						_ = config.WriteProject(cwd, proj)
					}
				}
			}
			return nil
		})
	},
}
