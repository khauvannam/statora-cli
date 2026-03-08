# `statora env` Shell Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `statora env` command that prints shell hook code for zsh/bash/fish so users can auto-switch PHP/Composer/extensions when `cd`-ing into a project.

**Architecture:** Single `cmd/env.go` file with a cobra command. Shell is auto-detected from `$SHELL` env var (basename), with a `--shell` flag as override. Each shell's hook code is an inline Go string constant. The command just prints the appropriate snippet to stdout — no state, no side effects.

**Tech Stack:** Go stdlib (`os`, `path/filepath`), `github.com/spf13/cobra`, `github.com/stretchr/testify`

---

### Task 1: `statora env` command + tests

**Files:**
- Create: `cmd/env.go`
- Create: `cmd/env_test.go`

---

**Step 1: Write the failing tests**

Create `cmd/env_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

```bash
go test ./cmd/... -run "TestDetectShell|TestEnvOutput" -v
```

Expected: `FAIL — cmd.DetectShell undefined`

**Step 3: Implement `cmd/env.go`**

Create `cmd/env.go`:

```go
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var supportedShells = []string{"zsh", "bash", "fish"}

const zshHook = `autoload -U add-zsh-hook

statora_auto_switch() {
  if command -v statora >/dev/null 2>&1; then
    statora switch >/dev/null 2>&1
  fi
}

add-zsh-hook chpwd statora_auto_switch
statora_auto_switch
`

const bashHook = `statora_auto_switch() {
  if command -v statora >/dev/null 2>&1; then
    statora switch >/dev/null 2>&1
  fi
}

if [[ "${PROMPT_COMMAND}" != *"statora_auto_switch"* ]]; then
  PROMPT_COMMAND="statora_auto_switch${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
statora_auto_switch
`

const fishHook = `function __statora_auto_switch --on-variable PWD
  if command -v statora >/dev/null 2>&1
    statora switch >/dev/null 2>&1
  end
end
__statora_auto_switch
`

var hooks = map[string]string{
	"zsh":  zshHook,
	"bash": bashHook,
	"fish": fishHook,
}

// DetectShell resolves the shell name from the SHELL env var and optional flag override.
// Returns an error if the shell is unsupported.
func DetectShell(shellEnv, flagVal string) (string, error) {
	s := flagVal
	if s == "" {
		s = filepath.Base(shellEnv)
	}
	if _, ok := hooks[s]; ok {
		return s, nil
	}
	return "", fmt.Errorf("unsupported shell %q — use --shell zsh|bash|fish", s)
}

// PrintEnv writes the shell hook snippet for the given shell to w.
func PrintEnv(w io.Writer, shell string) error {
	hook, ok := hooks[shell]
	if !ok {
		return fmt.Errorf("unsupported shell %q", shell)
	}
	_, err := fmt.Fprint(w, hook)
	return err
}

var shellFlag string

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print shell hook for auto-switching PHP/Composer on directory change",
	Long: `Print shell integration code for auto-switching.

Add to your shell rc file:

  zsh/bash:  eval "$(statora env)"
  fish:      statora env | source`,
	RunE: func(_ *cobra.Command, _ []string) error {
		shell, err := DetectShell(os.Getenv("SHELL"), shellFlag)
		if err != nil {
			return err
		}
		return PrintEnv(os.Stdout, shell)
	},
}

func init() {
	envCmd.Flags().StringVar(&shellFlag, "shell", "", "Shell to target (zsh, bash, fish)")
	Root.AddCommand(envCmd)
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./cmd/... -run "TestDetectShell|TestEnvOutput" -v
```

Expected: all 8 tests PASS

**Step 5: Run full test suite**

```bash
go test ./... 2>&1
```

Expected: all packages PASS

**Step 6: Commit**

```bash
git add cmd/env.go cmd/env_test.go
git commit -m "feat: add statora env shell integration command"
```

---

### Task 2: Update README

**Files:**
- Modify: `README.md`

**Context:**
The README currently has these Setup sections:
1. Verify installation
2. Add statora to PATH
3. Set up shims (`statora symlinks`)
4. Activate a version
5. Verify

We need to insert a new §4 "Set up auto-switching" between the current §3 (symlinks) and §4 (activate a version). Renumber the existing §4 and §5 to §5 and §6.

---

**Step 1: Insert new Setup §4 after the `statora symlinks` section**

Find the line `### 4. Activate a version` in `README.md` and insert before it:

```markdown
### 4. Set up auto-switching (optional)

Add shell integration so `statora switch` runs automatically when you `cd` into a project:

**zsh** (`~/.zshrc`):
```zsh
eval "$(statora env)"
```

**bash** (`~/.bashrc`):
```bash
eval "$(statora env)"
```

**fish** (`~/.config/fish/config.fish`):
```fish
statora env | source
```

> Statora will silently switch PHP, Composer, and extensions whenever you enter a directory with a `.statora` file.

```

**Step 2: Renumber the remaining sections**

- `### 4. Activate a version` → `### 5. Activate a version`
- `### 5. Verify` → `### 6. Verify`

**Step 3: Read README and verify it flows correctly**

```bash
# just read the Setup section to confirm ordering
```

**Step 4: Commit**

```bash
git add README.md
git commit -m "docs: add statora env shell integration to README setup"
```

---

### Task 3: Update `.goreleaser.yaml` caveats

**Files:**
- Modify: `.goreleaser.yaml`

**Context:**
The current `brews.caveats` in `.goreleaser.yaml` still has the old manual `ln -sf` symlink instructions. Update it to reference `statora symlinks` and `statora env`.

---

**Step 1: Replace the `caveats` block**

Find the `caveats: |` block (currently lines 91–100) and replace with:

```yaml
    caveats: |
      To use statora as a transparent php/composer shim, create symlinks:
        statora symlinks

      To enable auto-switching when you cd into a project, add to your shell rc:
        zsh/bash:  eval "$(statora env)"
        fish:      statora env | source

      To uninstall, remove symlinks and data first:
        statora symlinks
        rm -rf ~/.statora
      Then run: brew uninstall statora
```

**Step 2: Verify the YAML is valid**

```bash
go run . --help   # just confirms the binary still builds
```

**Step 3: Commit**

```bash
git add .goreleaser.yaml
git commit -m "chore: update goreleaser caveats to use statora symlinks and statora env"
```
