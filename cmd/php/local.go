package phpcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/php"
)

var localCmd = &cobra.Command{
	Use:   "local <version>",
	Short: "Set the PHP version for the current project (.statora)",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(p *php.Plugin) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			v := viper.New()
			v.SetConfigFile(dir + "/.statora")
			v.SetConfigType("toml")
			_ = v.ReadInConfig()

			current := v.GetString("php")
			if current == version {
				fmt.Printf("PHP is already set to %s (no change).\n", version)
				return nil
			}

			if !p.IsInstalled(version) {
				fmt.Printf("PHP %s is not installed. Installing...\n", version)
				if err := p.Install(version); err != nil {
					return fmt.Errorf("auto-install PHP %s: %w", version, err)
				}
			}

			v.Set("php", version)
			if err := v.WriteConfigAs(dir + "/.statora"); err != nil {
				return err
			}

			if current == "" {
				fmt.Printf("Set local PHP to %s.\n", version)
			} else {
				fmt.Printf("Updated local PHP: %s → %s.\n", current, version)
			}
			return nil
		})
	},
}
