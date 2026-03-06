package composercmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/composer"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Composer versions",
	RunE: func(c *cobra.Command, _ []string) error {
		return app.Invoke(cmd.Debug(), func(m *composer.Manager, cfg *config.Config) error {
			versions, err := m.List()
			if err != nil {
				return err
			}
			if len(versions) == 0 {
				fmt.Println("No Composer versions installed.")
				return nil
			}

			activePhar := dispatch.ReadCache(cfg, dispatch.KeyComposer)

			table := tablewriter.NewTable(os.Stdout)
			table.Header("Version", "Status")

			for _, v := range versions {
				status := ""
				if cfg.ComposerPhar(v) == activePhar {
					status = color.GreenString("active")
				}
				_ = table.Append(v, status)
			}
			_ = table.Render()
			return nil
		})
	},
}
