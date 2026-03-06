package composercmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/composer"
)

var localCmd = &cobra.Command{
	Use:   "local <version>",
	Short: "Set the Composer version for the current project (.statora)",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(m *composer.Manager) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			v := viper.New()
			v.SetConfigFile(dir + "/.statora")
			v.SetConfigType("toml")
			_ = v.ReadInConfig()

			current := v.GetString("composer")
			if current == version {
				fmt.Printf("Composer is already set to %s (no change).\n", version)
				return nil
			}

			if !m.IsInstalled(version) {
				fmt.Printf("Composer %s is not installed. Installing...\n", version)
				if err := m.Install(version); err != nil {
					return fmt.Errorf("auto-install Composer %s: %w", version, err)
				}
			}

			v.Set("composer", version)
			if err := v.WriteConfigAs(dir + "/.statora"); err != nil {
				return err
			}

			if current == "" {
				fmt.Printf("Set local Composer to %s.\n", version)
			} else {
				fmt.Printf("Updated local Composer: %s → %s.\n", current, version)
			}
			return nil
		})
	},
}
