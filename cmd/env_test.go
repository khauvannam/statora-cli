package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"statora-cli/cmd"
)

func TestDetectShell(t *testing.T) {
	cases := []struct {
		shellEnv string
		flag     string
		expected string
	}{
		{"/bin/zsh", "", "zsh"},
		{"/usr/bin/bash", "", "bash"},
		{"/usr/bin/fish", "", "fish"},
		{"/opt/homebrew/bin/zsh", "", "zsh"},
		// flag overrides env
		{"/bin/zsh", "bash", "bash"},
		{"/bin/bash", "fish", "fish"},
	}
	for _, tc := range cases {
		got, err := cmd.DetectShell(tc.shellEnv, tc.flag)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, got, "SHELL=%s flag=%s", tc.shellEnv, tc.flag)
	}
}

func TestDetectShell_Unknown(t *testing.T) {
	_, err := cmd.DetectShell("/bin/sh", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported shell")
}

func TestDetectShell_UnknownFlag(t *testing.T) {
	_, err := cmd.DetectShell("", "powershell")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported shell")
}

func TestEnvOutput_Zsh(t *testing.T) {
	var buf bytes.Buffer
	err := cmd.PrintEnv(&buf, "zsh")
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "add-zsh-hook")
	assert.Contains(t, out, "statora_auto_switch")
	assert.Contains(t, out, "statora switch")
	assert.Contains(t, out, "chpwd")
}

func TestEnvOutput_Bash(t *testing.T) {
	var buf bytes.Buffer
	err := cmd.PrintEnv(&buf, "bash")
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "PROMPT_COMMAND")
	assert.Contains(t, out, "statora_auto_switch")
	assert.Contains(t, out, "statora switch")
}

func TestEnvOutput_Fish(t *testing.T) {
	var buf bytes.Buffer
	err := cmd.PrintEnv(&buf, "fish")
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "--on-variable PWD")
	assert.Contains(t, out, "statora switch")
}

func TestEnvOutput_StartsWithNewlineForEval(t *testing.T) {
	// Each snippet should end with a newline so eval works cleanly.
	for _, shell := range []string{"zsh", "bash", "fish"} {
		var buf bytes.Buffer
		require.NoError(t, cmd.PrintEnv(&buf, shell))
		assert.True(t, strings.HasSuffix(buf.String(), "\n"), "shell=%s", shell)
	}
}
