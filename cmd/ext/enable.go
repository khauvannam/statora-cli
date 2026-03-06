package extcmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
	"statora-cli/internal/extension"
)

var enableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a PHP extension",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		name := args[0]
		return app.Invoke(cmd.Debug(), func(cfg *config.Config, log *zap.Logger) error {
			phpVer := dispatch.ReadCache(cfg, dispatch.KeyPHPActive)
			if phpVer == "" {
				return errNoActiveVersion()
			}
			return extension.NewInstaller(cfg, log, phpVer).Enable(name)
		})
	},
}
