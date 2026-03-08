# `statora symlinks` Command Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a `statora symlinks` toggle command that creates or removes `php` and `composer` shim symlinks next to the statora binary, replacing the manual `ln -sf` / `rm` instructions in the README.

**Architecture:** The core toggle logic lives in an exported `ToggleSymlinks(binDir string) error` function so it is unit-testable with a temp dir. The cobra command in `cmd/symlinks.go` resolves the binary directory via `os.Executable()` + `filepath.EvalSymlinks()` and delegates to `ToggleSymlinks`. The README Setup §3 and Uninstall §1 sections are updated to reference the new command.

**Tech Stack:** Go stdlib (`os`, `path/filepath`), `github.com/spf13/cobra`, `github.com/stretchr/testify`

---

### Task 1: Core toggle logic + tests

**Files:**
- Create: `cmd/symlinks.go`
- Create: `cmd/symlinks_test.go`

---

**Step 1: Write the failing test**

Create `cmd/symlinks_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

```bash
go test ./cmd/... -run TestToggleSymlinks -v
```

Expected: `FAIL — cmd.ToggleSymlinks undefined`

**Step 3: Implement `ToggleSymlinks` and the cobra command**

Create `cmd/symlinks.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// shims are the symlink names managed by ToggleSymlinks.
var shims = []string{"php", "composer"}

// ToggleSymlinks creates or removes php/composer symlinks in binDir.
// If all shims exist → removes them.
// If any are missing → creates the missing ones.
func ToggleSymlinks(binDir string) error {
	statoraBin := filepath.Join(binDir, "statora")

	// Check which shims already exist.
	var missing, present []string
	for _, name := range shims {
		target := filepath.Join(binDir, name)
		if _, err := os.Lstat(target); os.IsNotExist(err) {
			missing = append(missing, name)
		} else {
			present = append(present, name)
		}
	}

	// All present → remove all.
	if len(missing) == 0 {
		for _, name := range shims {
			if err := os.Remove(filepath.Join(binDir, name)); err != nil {
				return fmt.Errorf("removing symlink %s: %w", name, err)
			}
		}
		fmt.Printf("Removed symlinks: %s\n", strings.Join(shims, ", "))
		return nil
	}

	// Any missing → create only the missing ones.
	for _, name := range missing {
		target := filepath.Join(binDir, name)
		if err := os.Symlink(statoraBin, target); err != nil {
			return fmt.Errorf("creating symlink %s: %w", name, err)
		}
	}
	fmt.Printf("Created symlinks: %s\n", strings.Join(missing, ", "))
	_ = present // already exist, no action needed
	return nil
}

var symlinksCmd = &cobra.Command{
	Use:   "symlinks",
	Short: "Toggle php and composer shim symlinks next to the statora binary",
	Long: `Toggle php and composer shim symlinks.

If neither symlink exists, both are created.
If both exist, both are removed.
If only one exists, the missing one is created.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolving statora binary: %w", err)
		}
		// EvalSymlinks resolves the real path (e.g. Homebrew cellar).
		real, err := filepath.EvalSymlinks(exe)
		if err != nil {
			return fmt.Errorf("resolving real binary path: %w", err)
		}
		return ToggleSymlinks(filepath.Dir(real))
	},
}

func init() {
	Root.AddCommand(symlinksCmd)
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./cmd/... -run TestToggleSymlinks -v
```

Expected: all 3 tests PASS

**Step 5: Run full test suite to catch regressions**

```bash
go test ./... 2>&1
```

Expected: all packages PASS

**Step 6: Commit**

```bash
git add cmd/symlinks.go cmd/symlinks_test.go
git commit -m "feat: add statora symlinks toggle command"
```

---

### Task 2: Update README

**Files:**
- Modify: `README.md`

**Context:**
Two sections need updating:
- **Setup §3** ("Set up shims so php and composer dispatch to the active version") — replace the `ln -sf` blocks
- **Uninstall §1** ("Remove the shim symlinks") — replace the `rm` blocks

---

**Step 1: Update Setup §3 in README.md**

Find this section (lines ~73–88):

```markdown
### 3. Set up shims so `php` and `composer` dispatch to the active version

Statora doubles as a `php` and `composer` shim. Create symlinks once:

```bash
# bash / zsh
ln -sf "$(which statora)" "$(dirname "$(which statora)")/php"
ln -sf "$(which statora)" "$(dirname "$(which statora)")/composer"

# fish
ln -sf (which statora) (path dirname (which statora))/php
ln -sf (which statora) (path dirname (which statora))/composer
```

> This places the symlinks next to the `statora` binary, so they are guaranteed to be on your `$PATH`.
```

Replace with:

```markdown
### 3. Set up shims so `php` and `composer` dispatch to the active version

Statora doubles as a `php` and `composer` shim. Create symlinks once:

```bash
statora symlinks
```

> Run `statora symlinks` again at any time to remove the symlinks.
```

**Step 2: Update Uninstall §1 in README.md**

Find this section (lines ~112–122):

```markdown
### 1. Remove the shim symlinks

```bash
# bash / zsh
rm "$(dirname "$(which statora)")/php"
rm "$(dirname "$(which statora)")/composer"

# fish
rm (path dirname (which statora))/php
rm (path dirname (which statora))/composer
```
```

Replace with:

```markdown
### 1. Remove the shim symlinks

```bash
statora symlinks
```
```

**Step 3: Verify the README renders correctly**

Read through `README.md` and confirm the two sections look clean and consistent.

**Step 4: Commit**

```bash
git add README.md
git commit -m "docs: replace manual ln/rm symlink instructions with statora symlinks"
```
