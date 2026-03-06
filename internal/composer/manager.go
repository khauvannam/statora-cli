package composer

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"statora-cli/internal/config"
	"statora-cli/internal/installer"
)

// Manager handles Composer version installation and lifecycle.
type Manager struct {
	cfg *config.Config
	log *zap.Logger
}

func NewManager(cfg *config.Config, log *zap.Logger) *Manager {
	return &Manager{cfg: cfg, log: log}
}

// Install downloads and installs a Composer version.
func (m *Manager) Install(version string) error {
	if m.IsInstalled(version) {
		fmt.Printf("Composer %s is already installed.\n", version)
		return nil
	}

	pipeline := installer.New(
		&resolveSourceStage{},
		&downloadStage{},
		&verifySignatureStage{},
		&installStage{},
	)

	ctx := &installer.Context{
		Version: version,
		Cfg:     m.cfg,
		Log:     m.log,
		Data:    map[string]any{},
	}
	return pipeline.Run(ctx)
}

// Uninstall removes a Composer version directory.
func (m *Manager) Uninstall(version string) error {
	dir := filepath.Join(m.cfg.Paths.ComposerDir, version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("Composer %s is not installed", version)
	}
	return os.RemoveAll(dir)
}

// List returns all installed Composer versions.
func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.cfg.Paths.ComposerDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	return versions, nil
}

// IsInstalled reports whether the composer.phar exists for a version.
func (m *Manager) IsInstalled(version string) bool {
	_, err := os.Stat(m.cfg.ComposerPhar(version))
	return err == nil
}

// Phar returns the path to composer.phar for a version.
func (m *Manager) Phar(version string) (string, error) {
	phar := m.cfg.ComposerPhar(version)
	if _, err := os.Stat(phar); err != nil {
		return "", fmt.Errorf("Composer %s not found at %s", version, phar)
	}
	return phar, nil
}
