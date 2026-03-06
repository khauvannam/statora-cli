package extcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
	"statora-cli/internal/extension"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <name>",
	Short: "Uninstall a PHP extension",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		name := args[0]
		return app.Invoke(cmd.Debug(), func(cfg *config.Config, log *zap.Logger) error {
			phpVer := dispatch.ReadCache(cfg, dispatch.KeyPHPActive)
			if phpVer == "" {
				return errNoActiveVersion()
			}
			ins := extension.NewInstaller(cfg, log, phpVer)
			// Disable first if enabled.
			if ins.IsEnabled(name) {
				if err := ins.Disable(name); err != nil {
					return err
				}
			}
			soPath := filepath.Join(cfg.ExtAvailableDir(phpVer), name+".so")
			if err := os.Remove(soPath); err != nil && !os.IsNotExist(err) {
				return err
			}
			fmt.Printf("Extension %s uninstalled.\n", name)
			return nil
		})
	},
}
