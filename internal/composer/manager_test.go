package composer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"statora-cli/internal/composer"
	"statora-cli/internal/config"
)

func makeConfig(t *testing.T) *config.Config {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	cfg, err := config.New(false)
	require.NoError(t, err)
	return cfg
}

func fakeInstall(t *testing.T, cfg *config.Config, version string) {
	t.Helper()
	phar := cfg.ComposerPhar(version)
	require.NoError(t, os.MkdirAll(filepath.Dir(phar), 0o755))
	require.NoError(t, os.WriteFile(phar, []byte("#!/usr/bin/env php"), 0o755))
}

func TestManager_ListEmpty(t *testing.T) {
	cfg := makeConfig(t)
	m := composer.NewManager(cfg, zap.NewNop())
	versions, err := m.List()
	require.NoError(t, err)
	assert.Empty(t, versions)
}

func TestManager_IsInstalled_False(t *testing.T) {
	cfg := makeConfig(t)
	m := composer.NewManager(cfg, zap.NewNop())
	assert.False(t, m.IsInstalled("2.7.1"))
}

func TestManager_IsInstalled_True(t *testing.T) {
	cfg := makeConfig(t)
	m := composer.NewManager(cfg, zap.NewNop())
	fakeInstall(t, cfg, "2.7.1")
	assert.True(t, m.IsInstalled("2.7.1"))
}

func TestManager_List(t *testing.T) {
	cfg := makeConfig(t)
	m := composer.NewManager(cfg, zap.NewNop())

	fakeInstall(t, cfg, "2.5.0")
	fakeInstall(t, cfg, "2.7.1")

	versions, err := m.List()
	require.NoError(t, err)
	assert.Len(t, versions, 2)
}

func TestManager_Uninstall(t *testing.T) {
	cfg := makeConfig(t)
	m := composer.NewManager(cfg, zap.NewNop())
	fakeInstall(t, cfg, "2.7.1")

	require.NoError(t, m.Uninstall("2.7.1"))
	assert.False(t, m.IsInstalled("2.7.1"))
}

func TestManager_Uninstall_NotInstalled(t *testing.T) {
	cfg := makeConfig(t)
	m := composer.NewManager(cfg, zap.NewNop())
	err := m.Uninstall("2.7.1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestManager_Phar(t *testing.T) {
	cfg := makeConfig(t)
	m := composer.NewManager(cfg, zap.NewNop())
	fakeInstall(t, cfg, "2.7.1")

	path, err := m.Phar("2.7.1")
	require.NoError(t, err)
	assert.Equal(t, cfg.ComposerPhar("2.7.1"), path)
}

func TestManager_CreateWrapperScript(t *testing.T) {
	cfg := makeConfig(t)
	m := composer.NewManager(cfg, zap.NewNop())
	fakeInstall(t, cfg, "2.9.5")

	err := m.CreateWrapperScript("2.9.5")
	require.NoError(t, err)

	bin := cfg.ComposerBin("2.9.5")
	info, err := os.Stat(bin)
	require.NoError(t, err)
	assert.True(t, info.Mode()&0o111 != 0, "wrapper must be executable")

	content, err := os.ReadFile(bin)
	require.NoError(t, err)
	assert.Contains(t, string(content), "composer.phar")
	assert.Contains(t, string(content), "#!/bin/sh")
}
