package composer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
func (m *Manager) Install(versionOrConstraint string) (string, error) {
	concrete, sha256, err := ResolveVersion(versionOrConstraint)
	if err != nil {
		return "", fmt.Errorf("resolving Composer version: %w", err)
	}
	if concrete != versionOrConstraint {
		fmt.Printf("  Resolved %q → %s\n", versionOrConstraint, concrete)
	}

	if m.IsInstalled(concrete) {
		fmt.Printf("Composer %s is already installed.\n", concrete)
		return concrete, nil
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
		return "", err
	}
	return concrete, nil
}

// ResolveInstalled returns the concrete installed version matching the given spec.
func (m *Manager) ResolveInstalled(version string) (string, bool) {
	if m.IsInstalled(version) {
		return version, true
	}
	installed, err := m.List()
	if err != nil || len(installed) == 0 {
		return "", false
	}
	prefix := version + "."
	var best string
	for _, v := range installed {
		if strings.HasPrefix(v, prefix) {
			if best == "" || compareVersionStrings(v, best) > 0 {
				best = v
			}
		}
	}
	if best != "" {
		return best, true
	}
	return "", false
}

func compareVersionStrings(a, b string) int {
	ap := strings.SplitN(a, ".", 3)
	bp := strings.SplitN(b, ".", 3)
	for i := range 3 {
		var ai, bi int
		if i < len(ap) {
			ai, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bi, _ = strconv.Atoi(bp[i])
		}
		if ai != bi {
			return ai - bi
		}
	}
	return 0
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
