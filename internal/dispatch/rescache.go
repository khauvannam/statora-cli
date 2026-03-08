package dispatch

import (
	"os"
	"path/filepath"
	"strings"

	"statora-cli/internal/config"
	"statora-cli/internal/resolver"
)

const (
	KeyPHP      = "php"
	KeyComposer = "composer"
	KeyPHPActive = ".php_active"
)

// ReadCache reads a single rescache entry by key.
// Returns "" if the file does not exist.
func ReadCache(cfg *config.Config, key string) string {
	data, err := os.ReadFile(filepath.Join(cfg.Paths.RescacheDir, key))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// WriteCache writes a value to a rescache key atomically.
func WriteCache(cfg *config.Config, key, value string) error {
	path := filepath.Join(cfg.Paths.RescacheDir, key)
	return os.WriteFile(path, []byte(value+"\n"), 0o644)
}

// InvalidateCache rebuilds all rescache entries from the resolver.
// All values are stored as version strings (not paths) for consistency.
func InvalidateCache(cfg *config.Config, res resolver.Resolution) error {
	if res.PHP != "" {
		if err := WriteCache(cfg, KeyPHP, res.PHP); err != nil {
			return err
		}
		if err := WriteCache(cfg, KeyPHPActive, res.PHP); err != nil {
			return err
		}
	}
	if res.Composer != "" {
		if err := WriteCache(cfg, KeyComposer, res.Composer); err != nil {
			return err
		}
	}
	return nil
}
