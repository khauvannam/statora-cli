package extcmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
	"statora-cli/internal/extension"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show info about an extension",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		name := args[0]
		return app.Invoke(cmd.Debug(), func(cfg *config.Config, log *zap.Logger) error {
			phpVer := dispatch.ReadCache(cfg, dispatch.KeyPHPActive)
			if phpVer == "" {
				return errNoActiveVersion()
			}

			ins := extension.NewInstaller(cfg, log, phpVer)
			fmt.Printf("Extension: %s\n", color.CyanString(name))
			fmt.Printf("PHP:       %s\n", phpVer)

			if ins.IsAvailable(name) {
				fmt.Printf("Available: %s\n", color.GreenString("yes"))
			} else {
				fmt.Printf("Available: %s\n", color.RedString("no"))
			}

			if ins.IsEnabled(name) {
				fmt.Printf("Enabled:   %s\n", color.GreenString("yes"))
			} else {
				fmt.Printf("Enabled:   %s\n", color.YellowString("no"))
			}
			return nil
		})
	},
}
