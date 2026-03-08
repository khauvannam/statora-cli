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
// version may be a concrete version ("2.7.4") or a semver constraint (">= 2.2.0, < 3.0.0").
func (m *Manager) Install(versionOrConstraint string) error {
	concrete, sha256, err := ResolveVersion(versionOrConstraint)
	if err != nil {
		return fmt.Errorf("resolving Composer version: %w", err)
	}
	if concrete != versionOrConstraint {
		fmt.Printf("  Resolved %q → %s\n", versionOrConstraint, concrete)
	}

	if m.IsInstalled(concrete) {
		fmt.Printf("Composer %s is already installed.\n", concrete)
		return nil
	}

	pipeline := installer.New(
		&downloadStage{},
		&verifyChecksumStage{},
		&installStage{},
	)

	ctx := &installer.Context{
		Version:  concrete,
		Category: "composer",
		Cfg:      m.cfg,
		Log:      m.log,
		Data: map[string]any{
			"pharURL":    DownloadURL(concrete),
			"sha256URL":  SHA256URL(concrete),
			"sha256hint": sha256, // from getcomposer.org/versions (may be empty)
		},
	}
	if err := pipeline.Run(ctx); err != nil {
		return err
	}

	return m.CreateWrapperScript(concrete)
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

// CreateWrapperScript writes an executable bin/composer shell script that
// invokes the PHAR for the given version. This allows ~/.statora/composer/<v>/bin/
// to be placed on PATH for fnm-style dispatch.
func (m *Manager) CreateWrapperScript(version string) error {
	binDir := filepath.Dir(m.cfg.ComposerBin(version))
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("creating composer bin dir: %w", err)
	}

	phar := m.cfg.ComposerPhar(version)
	script := fmt.Sprintf("#!/bin/sh\nexec php %s \"$@\"\n", phar)

	bin := m.cfg.ComposerBin(version)
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		return fmt.Errorf("writing composer wrapper script: %w", err)
	}
	return nil
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
