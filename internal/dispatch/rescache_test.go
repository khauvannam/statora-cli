package dispatch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
	"statora-cli/internal/resolver"
)

func makeConfig(t *testing.T) *config.Config {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	cfg, err := config.New(false)
	require.NoError(t, err)
	return cfg
}

func TestReadCache_Missing(t *testing.T) {
	cfg := makeConfig(t)
	val := dispatch.ReadCache(cfg, "php")
	assert.Empty(t, val)
}

func TestWriteReadCache(t *testing.T) {
	cfg := makeConfig(t)
	require.NoError(t, dispatch.WriteCache(cfg, "php", "/usr/local/bin/php"))
	val := dispatch.ReadCache(cfg, "php")
	assert.Equal(t, "/usr/local/bin/php", val)
}

func TestInvalidateCache(t *testing.T) {
	cfg := makeConfig(t)
	res := resolver.Resolution{PHP: "8.2.15", Composer: "2.7.1"}

	require.NoError(t, dispatch.InvalidateCache(cfg, res))

	phpVer := dispatch.ReadCache(cfg, dispatch.KeyPHP)
	assert.Equal(t, "8.2.15", phpVer)

	composerVer := dispatch.ReadCache(cfg, dispatch.KeyComposer)
	assert.Equal(t, "2.7.1", composerVer)

	active := dispatch.ReadCache(cfg, dispatch.KeyPHPActive)
	assert.Equal(t, "8.2.15", active)
}
