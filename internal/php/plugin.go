package php

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

// Plugin manages PHP version installation and lifecycle.
type Plugin struct {
	cfg *config.Config
	log *zap.Logger
}

func NewPlugin(cfg *config.Config, log *zap.Logger) *Plugin {
	return &Plugin{cfg: cfg, log: log}
}

// Install downloads, compiles, and installs a PHP version.
func (p *Plugin) Install(version string) (string, error) {
	if concrete, ok := p.ResolveInstalled(version); ok {
		fmt.Printf("PHP %s is already installed.\n", concrete)
		return concrete, nil
	}

	pipeline := installer.New(
		&resolveSourceStage{},
		&downloadStage{},
		&verifyChecksumStage{},
		&extractStage{},
		&compileStage{},
		&installStage{},
	)

	ctx := &installer.Context{
		Version:  version,
		Category: "php",
		Cfg:      p.cfg,
		Log:      p.log,
		Data:     map[string]any{},
	}
	if err := pipeline.Run(ctx); err != nil {
		return "", err
	}
	return ctx.Version, nil
}

// ResolveInstalled returns the concrete installed version matching the given spec.
// For a full 3-part version it checks directly. For a partial like "8.1",
// it returns the highest installed 8.1.x.
func (p *Plugin) ResolveInstalled(version string) (string, bool) {
	if p.IsInstalled(version) {
		return version, true
	}
	// Only try prefix matching for non-concrete (partial) versions.
	if IsVersionString(version) {
		return "", false
	}
	installed, err := p.InstalledVersions()
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

// compareVersionStrings compares two dot-separated version strings numerically.
// Returns positive if a > b, negative if a < b, zero if equal.
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

// Uninstall removes a PHP version's runtime directory.
func (p *Plugin) Uninstall(version string) error {
	dir := p.cfg.PHPRuntimeDir(version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("PHP %s is not installed", version)
	}
	return os.RemoveAll(dir)
}

// List returns all installed PHP versions.
func (p *Plugin) List() ([]string, error) {
	runtimesDir := filepath.Join(p.cfg.Paths.RuntimesDir, "php")
	entries, err := os.ReadDir(runtimesDir)
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

// IsInstalled reports whether a PHP version's binary exists.
func (p *Plugin) IsInstalled(version string) bool {
	_, err := os.Stat(p.cfg.PHPBin(version))
	return err == nil
}

// Which returns the binary path for a version.
func (p *Plugin) Which(version string) (string, error) {
	bin := p.cfg.PHPBin(version)
	if _, err := os.Stat(bin); err != nil {
		return "", fmt.Errorf("PHP %s not found at %s", version, bin)
	}
	return bin, nil
}

// InstalledVersions returns only versions that have a valid php binary.
func (p *Plugin) InstalledVersions() ([]string, error) {
	all, err := p.List()
	if err != nil {
		return nil, err
	}
	var installed []string
	for _, v := range all {
		if p.IsInstalled(v) {
			installed = append(installed, v)
		}
	}
	return installed, nil
}

// Rehash rebuilds shims (currently a no-op; dispatch uses rescache directly).
func (p *Plugin) Rehash() error {
	fmt.Println("Rehash complete.")
	return nil
}

// IsVersionString does a quick sanity check on a version string.
func IsVersionString(s string) bool {
	parts := strings.Split(s, ".")
	return len(parts) == 3
}
