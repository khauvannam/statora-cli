package extcmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
	"statora-cli/internal/extension"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and enabled extensions for the active PHP version",
	RunE: func(c *cobra.Command, _ []string) error {
		return app.Invoke(cmd.Debug(), func(cfg *config.Config, log *zap.Logger) error {
			phpVer := dispatch.ReadCache(cfg, dispatch.KeyPHPActive)
			if phpVer == "" {
				return errNoActiveVersion()
			}

			ins := extension.NewInstaller(cfg, log, phpVer)
			available, err := ins.ListAvailable()
			if err != nil {
				return err
			}

			if len(available) == 0 {
				fmt.Printf("No extensions installed for PHP %s.\n", phpVer)
				return nil
			}

			enabled, err := ins.ListEnabled()
			if err != nil {
				return err
			}
			enabledSet := make(map[string]bool, len(enabled))
			for _, e := range enabled {
				enabledSet[e] = true
			}

			table := tablewriter.NewTable(os.Stdout)
			table.Header("Extension", "Status")

			for _, name := range available {
				status := color.YellowString("available")
				if enabledSet[name] {
					status = color.GreenString("enabled")
				}
				_ = table.Append(name, status)
			}
			_ = table.Render()
			return nil
		})
	},
}
