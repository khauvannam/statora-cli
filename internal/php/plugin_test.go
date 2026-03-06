package php_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"statora-cli/internal/config"
	"statora-cli/internal/php"
)

func makeConfig(t *testing.T) *config.Config {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	cfg, err := config.New(false)
	require.NoError(t, err)
	return cfg
}

func TestPlugin_ListEmpty(t *testing.T) {
	cfg := makeConfig(t)
	p := php.NewPlugin(cfg, zap.NewNop())

	versions, err := p.List()
	require.NoError(t, err)
	assert.Empty(t, versions)
}

func TestPlugin_IsInstalled_False(t *testing.T) {
	cfg := makeConfig(t)
	p := php.NewPlugin(cfg, zap.NewNop())
	assert.False(t, p.IsInstalled("8.2.15"))
}

func TestPlugin_IsInstalled_True(t *testing.T) {
	cfg := makeConfig(t)
	p := php.NewPlugin(cfg, zap.NewNop())

	// Simulate an installed PHP binary.
	bin := cfg.PHPBin("8.2.15")
	require.NoError(t, os.MkdirAll(filepath.Dir(bin), 0o755))
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh"), 0o755))

	assert.True(t, p.IsInstalled("8.2.15"))
}

func TestPlugin_List(t *testing.T) {
	cfg := makeConfig(t)
	p := php.NewPlugin(cfg, zap.NewNop())

	for _, ver := range []string{"8.1.0", "8.2.15"} {
		bin := cfg.PHPBin(ver)
		require.NoError(t, os.MkdirAll(filepath.Dir(bin), 0o755))
		require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh"), 0o755))
	}

	versions, err := p.List()
	require.NoError(t, err)
	assert.Len(t, versions, 2)
}

func TestPlugin_Uninstall(t *testing.T) {
	cfg := makeConfig(t)
	p := php.NewPlugin(cfg, zap.NewNop())

	bin := cfg.PHPBin("8.2.15")
	require.NoError(t, os.MkdirAll(filepath.Dir(bin), 0o755))
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh"), 0o755))

	require.NoError(t, p.Uninstall("8.2.15"))
	assert.False(t, p.IsInstalled("8.2.15"))
}

func TestPlugin_Uninstall_NotInstalled(t *testing.T) {
	cfg := makeConfig(t)
	p := php.NewPlugin(cfg, zap.NewNop())
	err := p.Uninstall("8.2.15")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestPlugin_Which(t *testing.T) {
	cfg := makeConfig(t)
	p := php.NewPlugin(cfg, zap.NewNop())

	bin := cfg.PHPBin("8.2.15")
	require.NoError(t, os.MkdirAll(filepath.Dir(bin), 0o755))
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh"), 0o755))

	path, err := p.Which("8.2.15")
	require.NoError(t, err)
	assert.Equal(t, bin, path)
}

func TestIsVersionString(t *testing.T) {
	assert.True(t, php.IsVersionString("8.2.15"))
	assert.False(t, php.IsVersionString("8.2"))
	assert.False(t, php.IsVersionString("latest"))
}
