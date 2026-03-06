package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"statora-cli/internal/app"
	"statora-cli/internal/switcher"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch to the PHP/Composer versions defined in .statora or global config",
	RunE: func(c *cobra.Command, _ []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		return app.Invoke(Debug(), func(sw *switcher.Switcher) error {
			plan, err := sw.BuildPlan(dir)
			if err != nil {
				return err
			}

			if !plan.HasChanges() {
				fmt.Println("Already up to date.")
				return nil
			}

			confirmed, err := switcher.PromptConfirm(plan)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Aborted.")
				return nil
			}

			return sw.Execute(plan)
		})
	},
}

func init() {
	Root.AddCommand(switchCmd)
}
