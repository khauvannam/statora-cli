package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// shims are the symlink names managed by ToggleSymlinks.
var shims = []string{"php", "composer"}

// ToggleSymlinks creates or removes php/composer symlinks in binDir.
// If all shims exist → removes them.
// If any are missing → creates the missing ones.
func ToggleSymlinks(binDir string) error {
	statoraBin := filepath.Join(binDir, "statora")

	// Check which shims already exist.
	var missing, present []string
	for _, name := range shims {
		target := filepath.Join(binDir, name)
		if _, err := os.Lstat(target); os.IsNotExist(err) {
			missing = append(missing, name)
		} else {
			present = append(present, name)
		}
	}

	// All present → remove all.
	if len(missing) == 0 {
		for _, name := range shims {
			if err := os.Remove(filepath.Join(binDir, name)); err != nil {
				return fmt.Errorf("removing symlink %s: %w", name, err)
			}
		}
		fmt.Printf("Removed symlinks: %s\n", strings.Join(shims, ", "))
		return nil
	}

	// Any missing → create only the missing ones.
	for _, name := range missing {
		target := filepath.Join(binDir, name)
		if err := os.Symlink(statoraBin, target); err != nil {
			return fmt.Errorf("creating symlink %s: %w", name, err)
		}
	}
	fmt.Printf("Created symlinks: %s\n", strings.Join(missing, ", "))
	_ = present // already exist, no action needed
	return nil
}

var symlinksCmd = &cobra.Command{
	Use:   "symlinks",
	Short: "Toggle php and composer shim symlinks next to the statora binary",
	Long: `Toggle php and composer shim symlinks.

If neither symlink exists, both are created.
If both exist, both are removed.
If only one exists, the missing one is created.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolving statora binary: %w", err)
		}
		// EvalSymlinks resolves the real path (e.g. Homebrew cellar).
		real, err := filepath.EvalSymlinks(exe)
		if err != nil {
			return fmt.Errorf("resolving real binary path: %w", err)
		}
		return ToggleSymlinks(filepath.Dir(real))
	},
}

func init() {
	Root.AddCommand(symlinksCmd)
}
