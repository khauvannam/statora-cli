package extension_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"statora-cli/internal/config"
	"statora-cli/internal/extension"
)

func makeInstaller(t *testing.T, phpVer string) (*extension.Installer, *config.Config) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	cfg, err := config.New(false)
	require.NoError(t, err)
	ins := extension.NewInstaller(cfg, zap.NewNop(), phpVer)
	return ins, cfg
}

func placeExtension(t *testing.T, cfg *config.Config, phpVer, name string) {
	t.Helper()
	dir := cfg.ExtAvailableDir(phpVer)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".so"), []byte("ELF"), 0o644))
}

func TestLinker_EnableDisable(t *testing.T) {
	ins, cfg := makeInstaller(t, "8.2.15")
	placeExtension(t, cfg, "8.2.15", "redis")

	require.NoError(t, ins.Enable("redis"))
	assert.True(t, ins.IsEnabled("redis"))

	require.NoError(t, ins.Disable("redis"))
	assert.False(t, ins.IsEnabled("redis"))
}

func TestLinker_Enable_NotAvailable(t *testing.T) {
	ins, _ := makeInstaller(t, "8.2.15")
	err := ins.Enable("redis")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestLinker_Disable_NotEnabled(t *testing.T) {
	ins, cfg := makeInstaller(t, "8.2.15")
	placeExtension(t, cfg, "8.2.15", "redis")
	err := ins.Disable("redis")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestLinker_Enable_Idempotent(t *testing.T) {
	ins, cfg := makeInstaller(t, "8.2.15")
	placeExtension(t, cfg, "8.2.15", "redis")
	require.NoError(t, ins.Enable("redis"))
	require.NoError(t, ins.Enable("redis")) // second call should not error
}

func TestLinker_ListAvailable(t *testing.T) {
	ins, cfg := makeInstaller(t, "8.2.15")
	placeExtension(t, cfg, "8.2.15", "redis")
	placeExtension(t, cfg, "8.2.15", "xdebug")

	names, err := ins.ListAvailable()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"redis", "xdebug"}, names)
}

func TestLinker_ListEnabled(t *testing.T) {
	ins, cfg := makeInstaller(t, "8.2.15")
	placeExtension(t, cfg, "8.2.15", "redis")
	placeExtension(t, cfg, "8.2.15", "xdebug")

	require.NoError(t, ins.Enable("redis"))

	enabled, err := ins.ListEnabled()
	require.NoError(t, err)
	assert.Equal(t, []string{"redis"}, enabled)
}
