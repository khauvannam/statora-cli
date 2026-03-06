package main

import (
	"fmt"
	"os"
	"path/filepath"

	"statora-cli/cmd"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"

	// Register all cobra subcommands via init().
	_ "statora-cli/cmd/php"
	_ "statora-cli/cmd/composer"
	_ "statora-cli/cmd/ext"
)

func main() {
	// Fast path: if invoked as "php" or "composer" (via symlink),
	// exec the active binary directly without starting the fx app.
	base := filepath.Base(os.Args[0])
	if base == "php" || base == "composer" {
		cfg, err := config.New(false)
		if err != nil {
			fatalf("statora: failed to load config: %v\n", err)
		}
		if err := dispatch.Dispatch(cfg, base); err != nil {
			fatalf("statora: %v\n", err)
		}
		return
	}

	// Normal path: run the cobra CLI.
	cmd.Execute()
}

func fatalf(format string, args ...any) {
	//nolint:forbidigo
	_, _ = os.Stderr.WriteString(fmt.Sprintf(format, args...))
	os.Exit(1)
}
