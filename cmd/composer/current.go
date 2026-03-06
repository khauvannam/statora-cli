package composercmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"statora-cli/cmd"
	"statora-cli/internal/app"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active Composer version",
	RunE: func(c *cobra.Command, _ []string) error {
		return app.Invoke(cmd.Debug(), func(cfg *config.Config) error {
			phar := dispatch.ReadCache(cfg, dispatch.KeyComposer)
			if phar == "" {
				fmt.Println("No active Composer version. Run `statora switch`.")
				return nil
			}
			// Extract version from path: ~/.statora/composer/<version>/composer.phar
			parts := strings.Split(filepath.ToSlash(phar), "/")
			for i, p := range parts {
				if p == "composer" && i+2 < len(parts) {
					fmt.Println(parts[i+1])
					return nil
				}
			}
			fmt.Println(phar)
			return nil
		})
	},
}
