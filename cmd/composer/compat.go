package composercmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/compat"
)

var compatCmd = &cobra.Command{
	Use:   "compat <php-version>",
	Short: "Show the compatible Composer constraint for a PHP version",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		phpVersion := args[0]
		return app.Invoke(cmd.Debug(), func(ch *compat.Checker) error {
			constraint, err := ch.ResolveComposer(phpVersion)
			if err != nil {
				return err
			}
			if constraint == "" {
				fmt.Printf("No Composer compatibility rule for PHP %s\n", phpVersion)
				return nil
			}
			fmt.Printf("PHP %s → Composer %s\n", phpVersion, constraint)
			return nil
		})
	},
}
