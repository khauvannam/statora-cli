package use_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"statora-cli/internal/use"
)

func TestStripStatoraDirs(t *testing.T) {
	home := "/home/user"
	path := "/home/user/.statora/runtimes/php/8.1.20/bin:/usr/bin:/home/user/.statora/composer/2.2.8/bin:/bin"
	got := use.StripStatoraDirs(path, home)
	assert.Equal(t, "/usr/bin:/bin", got)
}

func TestStripStatoraDirs_NothingToStrip(t *testing.T) {
	got := use.StripStatoraDirs("/usr/bin:/bin", "/home/user")
	assert.Equal(t, "/usr/bin:/bin", got)
}

func TestBuildPATHExport_Bash(t *testing.T) {
	out := use.BuildPATHExport("bash", "/php/bin", "/composer/bin", "/usr/bin:/bin")
	assert.Equal(t, `export PATH="/php/bin:/composer/bin:/usr/bin:/bin"`, strings.TrimSpace(out))
}

func TestBuildPATHExport_Zsh(t *testing.T) {
	out := use.BuildPATHExport("zsh", "/php/bin", "/composer/bin", "/usr/bin:/bin")
	assert.Equal(t, `export PATH="/php/bin:/composer/bin:/usr/bin:/bin"`, strings.TrimSpace(out))
}

func TestBuildPATHExport_Fish(t *testing.T) {
	out := use.BuildPATHExport("fish", "/php/bin", "/composer/bin", "/usr/bin:/bin")
	assert.Equal(t, "set -gx PATH /php/bin /composer/bin /usr/bin /bin", strings.TrimSpace(out))
}
