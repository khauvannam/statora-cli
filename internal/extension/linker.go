package extension

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Enable creates a symlink in enabled/ pointing to available/<name>.so.
func (i *Installer) Enable(name string) error {
	soName := name + ".so"
	src := filepath.Join(i.cfg.ExtAvailableDir(i.phpVersion), soName)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("extension %s is not available (install it first)", name)
	}

	enabledDir := i.cfg.ExtEnabledDir(i.phpVersion)
	if err := os.MkdirAll(enabledDir, 0o755); err != nil {
		return err
	}

	dest := filepath.Join(enabledDir, soName)
	if _, err := os.Lstat(dest); err == nil {
		return nil // already enabled
	}

	return os.Symlink(src, dest)
}

// Disable removes the symlink from enabled/.
func (i *Installer) Disable(name string) error {
	dest := filepath.Join(i.cfg.ExtEnabledDir(i.phpVersion), name+".so")
	if _, err := os.Lstat(dest); os.IsNotExist(err) {
		return fmt.Errorf("extension %s is not enabled", name)
	}
	return os.Remove(dest)
}

// IsAvailable reports whether the .so is in available/.
func (i *Installer) IsAvailable(name string) bool {
	_, err := os.Stat(filepath.Join(i.cfg.ExtAvailableDir(i.phpVersion), name+".so"))
	return err == nil
}

// IsEnabled reports whether the symlink exists in enabled/.
func (i *Installer) IsEnabled(name string) bool {
	_, err := os.Lstat(filepath.Join(i.cfg.ExtEnabledDir(i.phpVersion), name+".so"))
	return err == nil
}

// ListAvailable returns all extension names in available/.
func (i *Installer) ListAvailable() ([]string, error) {
	return listSO(i.cfg.ExtAvailableDir(i.phpVersion))
}

// ListEnabled returns all extension names in enabled/.
func (i *Installer) ListEnabled() ([]string, error) {
	return listSO(i.cfg.ExtEnabledDir(i.phpVersion))
}

func listSO(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".so") {
			names = append(names, strings.TrimSuffix(e.Name(), ".so"))
		}
	}
	return names, nil
}
