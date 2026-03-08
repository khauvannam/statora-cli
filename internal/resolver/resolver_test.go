package resolver_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"statora-cli/internal/compat"
	"statora-cli/internal/config"
	"statora-cli/internal/resolver"
)

func makeConfig(t *testing.T) *config.Config {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg, err := config.New(false)
	require.NoError(t, err)
	return cfg
}

func TestResolve_ProjectFile(t *testing.T) {
	cfg := makeConfig(t)
	checker := compat.NewChecker()
	r := resolver.New(cfg, checker)

	dir := t.TempDir()
	content := `php = "8.2.15"
composer = "2.7.1"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".statora"), []byte(content), 0o644))

	res, err := r.Resolve(dir)
	require.NoError(t, err)
	assert.Equal(t, "8.2.15", res.PHP)
	assert.Equal(t, "2.7.1", res.Composer)
	assert.Equal(t, "project", res.Source)
}

func TestResolve_ProjectFileInfersComposer(t *testing.T) {
	cfg := makeConfig(t)
	checker := compat.NewChecker()
	r := resolver.New(cfg, checker)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".statora"), []byte(`php = "8.2.15"`), 0o644))

	res, err := r.Resolve(dir)
	require.NoError(t, err)
	assert.Equal(t, "8.2.15", res.PHP)
	assert.NotEmpty(t, res.Composer)
}

func TestResolve_GlobalFallback(t *testing.T) {
	cfg := makeConfig(t)
	checker := compat.NewChecker()
	r := resolver.New(cfg, checker)

	require.NoError(t, cfg.WriteGlobal(config.GlobalConfig{PHP: "8.1.0", Composer: "2.5.0"}))

	res, err := r.Resolve(t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, "8.1.0", res.PHP)
	assert.Equal(t, "2.5.0", res.Composer)
	assert.Equal(t, "global", res.Source)
}

func TestResolve_NoConfig(t *testing.T) {
	cfg := makeConfig(t)
	checker := compat.NewChecker()
	r := resolver.New(cfg, checker)

	res, err := r.Resolve(t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, "none", res.Source)
	assert.Empty(t, res.PHP)
}

func TestNearestProjectFile(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".statora"), []byte(`php = "8.2.15"`), 0o644))

	found := resolver.NearestProjectFile(sub)
	assert.Equal(t, root, found)
}

func TestResolve_NormalizesPartialPHPToInstalled(t *testing.T) {
	cfg := makeConfig(t)

	// Fake-install PHP 8.1.25
	phpBin := cfg.PHPBin("8.1.25")
	require.NoError(t, os.MkdirAll(filepath.Dir(phpBin), 0o755))
	require.NoError(t, os.WriteFile(phpBin, []byte("#!/bin/sh"), 0o755))

	// Fake-install composer 2.9.5 wrapper
	composerBin := cfg.ComposerBin("2.9.5")
	require.NoError(t, os.MkdirAll(filepath.Dir(composerBin), 0o755))
	require.NoError(t, os.WriteFile(composerBin, []byte("#!/bin/sh"), 0o755))

	checker := compat.NewChecker()
	r := resolver.New(cfg, checker)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".statora"), []byte("php = \"8.1\"\ncomposer = \"2.9.5\"\n"), 0o644))

	res, err := r.Resolve(dir)
	require.NoError(t, err)
	assert.Equal(t, "8.1.25", res.PHP)     // normalized from "8.1"
	assert.Equal(t, "2.9.5", res.Composer) // already concrete, unchanged
}

func TestResolve_NormalizesPartialComposerToInstalled(t *testing.T) {
	cfg := makeConfig(t)

	// Fake-install PHP 8.1.25
	phpBin := cfg.PHPBin("8.1.25")
	require.NoError(t, os.MkdirAll(filepath.Dir(phpBin), 0o755))
	require.NoError(t, os.WriteFile(phpBin, []byte("#!/bin/sh"), 0o755))

	// Fake-install composer 2.2.8 wrapper
	composerBin := cfg.ComposerBin("2.2.8")
	require.NoError(t, os.MkdirAll(filepath.Dir(composerBin), 0o755))
	require.NoError(t, os.WriteFile(composerBin, []byte("#!/bin/sh"), 0o755))

	checker := compat.NewChecker()
	r := resolver.New(cfg, checker)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".statora"), []byte("php = \"8.1\"\ncomposer = \"2.2\"\n"), 0o644))

	res, err := r.Resolve(dir)
	require.NoError(t, err)
	assert.Equal(t, "8.1.25", res.PHP)    // normalized
	assert.Equal(t, "2.2.8", res.Composer) // normalized from "2.2"
}

func TestResolve_KeepsPartialIfNothingInstalled(t *testing.T) {
	cfg := makeConfig(t)
	checker := compat.NewChecker()
	r := resolver.New(cfg, checker)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".statora"), []byte("php = \"8.1\"\ncomposer = \"2.2.0\"\n"), 0o644))

	res, err := r.Resolve(dir)
	require.NoError(t, err)
	// Nothing installed → keep as-is so caller can prompt to install
	assert.Equal(t, "8.1", res.PHP)
	assert.Equal(t, "2.2.0", res.Composer)
}
