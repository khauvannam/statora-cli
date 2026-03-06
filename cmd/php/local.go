package phpcmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
)

var localCmd = &cobra.Command{
	Use:   "local <version>",
	Short: "Set the PHP version for the current project (.statora)",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		version := args[0]
		return app.Invoke(cmd.Debug(), func(_ *config.Config) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			v := viper.New()
			v.SetConfigFile(dir + "/.statora")
			v.SetConfigType("toml")
			_ = v.ReadInConfig()
			v.Set("php", version)
			return v.WriteConfigAs(dir + "/.statora")
		})
	},
}
