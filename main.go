package main

import (
	"statora-cli/cmd"

	// Register all cobra subcommands via init().
	_ "statora-cli/cmd/php"
	_ "statora-cli/cmd/composer"
	_ "statora-cli/cmd/ext"
)

func main() {
	cmd.Execute()
}
