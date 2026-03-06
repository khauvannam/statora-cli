//go:build !windows

package dispatch

import (
	"fmt"
	"os"
	"syscall"

	"statora-cli/internal/config"
)

// Dispatch replaces the current process with the binary resolved from rescache.
// binary is one of "php" or "composer".
// This function never returns on success (syscall.Exec replaces the process).
func Dispatch(cfg *config.Config, binary string) error {
	binPath := ReadCache(cfg, binary)
	if binPath == "" {
		return fmt.Errorf("no %s version active — run `statora switch` first", binary)
	}

	if _, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("%s binary not found at %s: run `statora switch` to refresh", binary, binPath)
	}

	args := append([]string{binPath}, os.Args[1:]...)
	return syscall.Exec(binPath, args, os.Environ())
}
