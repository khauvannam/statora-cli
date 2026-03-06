package phpcmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/dispatch"
	"statora-cli/internal/php"
	"statora-cli/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed PHP versions",
	RunE: func(c *cobra.Command, _ []string) error {
		return app.Invoke(cmd.Debug(), func(p *php.Plugin, cfg *config.Config) error {
			versions, err := p.List()
			if err != nil {
				return err
			}
			if len(versions) == 0 {
				fmt.Println("No PHP versions installed.")
				return nil
			}

			active := dispatch.ReadCache(cfg, dispatch.KeyPHPActive)

			table := tablewriter.NewTable(os.Stdout)
			table.Header("Version", "Status")

			for _, v := range versions {
				status := ""
				if v == active {
					status = color.GreenString("active")
				}
				_ = table.Append(v, status)
			}
			_ = table.Render()
			return nil
		})
	},
}
