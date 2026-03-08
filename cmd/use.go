package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"statora-cli/internal/app"
	"statora-cli/internal/composer"
	"statora-cli/internal/config"
	"statora-cli/internal/php"
	"statora-cli/internal/resolver"
	"statora-cli/internal/use"
)

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Output PATH export for the active PHP/Composer versions (eval this in your shell)",
	Long: `Resolve the active PHP/Composer versions and print a shell PATH export.

Designed to be eval'd by the shell hook set up by 'statora env'.

  zsh/bash:  eval "$(statora use)"
  fish:      statora use | source`,
	RunE: func(c *cobra.Command, _ []string) error {
		shellFlag, _ := c.Flags().GetString("shell")

		shell, err := DetectShell(os.Getenv("SHELL"), shellFlag)
		if err != nil {
			return err
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		return app.Invoke(Debug(), func(
			res *resolver.Resolver,
			phpPlugin *php.Plugin,
			composerMgr *composer.Manager,
			cfg *config.Config,
		) error {
			resolution, err := res.Resolve(dir)
			if err != nil {
				return err
			}
			if resolution.Source == "none" {
				return nil
			}

			phpVersion := resolution.PHP
			if !phpPlugin.IsInstalled(phpVersion) {
				if !promptInstall(fmt.Sprintf("PHP %s", phpVersion)) {
					fmt.Fprintf(os.Stderr, "statora: skipping PATH update (PHP %s not installed)\n", phpVersion)
					return nil
				}
				if err := phpPlugin.Install(phpVersion); err != nil {
					return fmt.Errorf("installing PHP %s: %w", phpVersion, err)
				}
				resolution, err = res.Resolve(dir)
				if err != nil {
					return err
				}
				phpVersion = resolution.PHP
			}

			composerVersion := resolution.Composer
			if !composerMgr.IsInstalled(composerVersion) {
				if !promptInstall(fmt.Sprintf("Composer %s", composerVersion)) {
					fmt.Fprintf(os.Stderr, "statora: skipping PATH update (Composer %s not installed)\n", composerVersion)
					return nil
				}
				if err := composerMgr.Install(composerVersion); err != nil {
					return fmt.Errorf("installing Composer %s: %w", composerVersion, err)
				}
				resolution, err = res.Resolve(dir)
				if err != nil {
					return err
				}
				composerVersion = resolution.Composer
			}

			phpBinDir := cfg.PHPRuntimeDir(phpVersion) + "/bin"
			composerBinDir := cfg.Paths.ComposerDir + "/" + composerVersion + "/bin"

			return use.PrintUse(os.Stdout, shell, phpBinDir, composerBinDir, os.Getenv("PATH"), homeDir)
		})
	},
}

// promptInstall asks the user interactively whether to install a tool.
// Returns true if yes, false if no or non-interactive.
func promptInstall(label string) bool {
	if !use.IsTerminal(os.Stdin) {
		return false
	}
	fmt.Printf("%s is not installed. Install now? [y/N] ", label)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}

func init() {
	useCmd.Flags().String("shell", "", "Shell to target (zsh, bash, fish)")
	Root.AddCommand(useCmd)
}
