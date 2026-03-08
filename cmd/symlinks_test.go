package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"statora-cli/cmd"
)

// makeStatora creates a fake statora binary in a temp dir and returns the dir.
func makeStatora(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "statora")
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh"), 0o755))
	return dir
}

func TestToggleSymlinks_CreatesWhenNoneExist(t *testing.T) {
	dir := makeStatora(t)
	err := cmd.ToggleSymlinks(dir)
	require.NoError(t, err)
	assertSymlink(t, filepath.Join(dir, "php"), filepath.Join(dir, "statora"))
	assertSymlink(t, filepath.Join(dir, "composer"), filepath.Join(dir, "statora"))
}

func TestToggleSymlinks_RemovesWhenAllExist(t *testing.T) {
	dir := makeStatora(t)
	// Pre-create both symlinks.
	require.NoError(t, os.Symlink(filepath.Join(dir, "statora"), filepath.Join(dir, "php")))
	require.NoError(t, os.Symlink(filepath.Join(dir, "statora"), filepath.Join(dir, "composer")))

	err := cmd.ToggleSymlinks(dir)
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(dir, "php"))
	assert.NoFileExists(t, filepath.Join(dir, "composer"))
}

func TestToggleSymlinks_CompletesWhenPartial(t *testing.T) {
	dir := makeStatora(t)
	// Pre-create only php symlink.
	require.NoError(t, os.Symlink(filepath.Join(dir, "statora"), filepath.Join(dir, "php")))

	err := cmd.ToggleSymlinks(dir)
	require.NoError(t, err)
	// Both should now exist.
	assertSymlink(t, filepath.Join(dir, "php"), filepath.Join(dir, "statora"))
	assertSymlink(t, filepath.Join(dir, "composer"), filepath.Join(dir, "statora"))
}

// assertSymlink checks that path is a symlink pointing to target.
func assertSymlink(t *testing.T, path, target string) {
	t.Helper()
	dest, err := os.Readlink(path)
	require.NoError(t, err, "expected symlink at %s", path)
	assert.Equal(t, target, dest)
}
