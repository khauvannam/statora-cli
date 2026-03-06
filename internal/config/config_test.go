package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"statora-cli/internal/config"
)

func TestNew_CreatesDirectories(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.New(false)
	require.NoError(t, err)

	assert.DirExists(t, cfg.Paths.RuntimesDir)
	assert.DirExists(t, cfg.Paths.ComposerDir)
	assert.DirExists(t, cfg.Paths.DownloadDir)
	assert.DirExists(t, cfg.Paths.RescacheDir)
}

func TestNew_PathsLayout(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.New(false)
	require.NoError(t, err)

	expected := filepath.Join(home, ".statora")
	assert.Equal(t, expected, cfg.Paths.Home)
	assert.Equal(t, filepath.Join(expected, "runtimes"), cfg.Paths.RuntimesDir)
	assert.Equal(t, filepath.Join(expected, ".rescache"), cfg.Paths.RescacheDir)
}

func TestLoadProject_Found(t *testing.T) {
	dir := t.TempDir()
	content := `php = "8.2.15"
composer = "2.7.1"
extensions = ["redis", "xdebug"]
`
	err := os.WriteFile(filepath.Join(dir, ".statora"), []byte(content), 0o644)
	require.NoError(t, err)

	proj, found, err := config.LoadProject(dir)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "8.2.15", proj.PHP)
	assert.Equal(t, "2.7.1", proj.Composer)
	assert.Equal(t, []string{"redis", "xdebug"}, proj.Extensions)
}

func TestLoadProject_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, found, err := config.LoadProject(dir)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestConfig_PHPBin(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := config.New(false)
	require.NoError(t, err)

	bin := cfg.PHPBin("8.2.15")
	assert.Contains(t, bin, "8.2.15")
	assert.Contains(t, bin, "bin/php")
}
