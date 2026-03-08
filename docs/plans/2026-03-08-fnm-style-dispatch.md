# FNM-Style PATH Dispatch Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace symlink + `.rescache` dispatch with per-terminal `PATH` manipulation (like fnm/nvm), fixing the parallel-project conflict and stale `.rescache` version bugs.

**Architecture:** A new `statora use` command resolves the current dir's version, normalizes partial versions (e.g. `"8.1"`) to the highest concrete installed match, and outputs a shell-specific `PATH` export. The shell hook (`statora env`) is updated to `eval` this output on every directory change. Symlinks, the dispatcher, and the `main.go` fast-path are removed. Composer gets a thin wrapper script so its bin dir can be placed on PATH like PHP.

**Tech Stack:** Go, Cobra, go.uber.org/fx, github.com/Masterminds/semver/v3, github.com/stretchr/testify

---

### Task 1: Add `ComposerBin` path helper to config

**Files:**
- Modify: `internal/config/config.go:137`
- Modify: `internal/config/config_test.go`

**Step 1: Write the failing test**

In `internal/config/config_test.go`, add:
```go
func TestComposerBin(t *testing.T) {
    t.Setenv("HOME", t.TempDir())
    cfg, err := config.New(false)
    require.NoError(t, err)
    got := cfg.ComposerBin("2.9.5")
    assert.True(t, strings.HasSuffix(got, "composer/2.9.5/bin/composer"), got)
}
```

**Step 2: Run to verify it fails**
```
go test ./internal/config/... -run TestComposerBin -v
```
Expected: FAIL — method does not exist yet.

**Step 3: Implement**

In `internal/config/config.go`, after `ComposerPhar`:
```go
// ComposerBin returns the wrapper script path for a given Composer version.
func (c *Config) ComposerBin(version string) string {
    return filepath.Join(c.Paths.ComposerDir, version, "bin", "composer")
}
```

**Step 4: Run to verify it passes**
```
go test ./internal/config/... -run TestComposerBin -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add ComposerBin path helper"
```

---

### Task 2: Add `NormalizeInstalled` helper

This is the core fix for the `.rescache` stale version bug. Given a partial version like `"8.1"` and a list of installed concrete versions, it returns the highest matching one.

**Files:**
- Create: `internal/resolver/normalize.go`
- Create: `internal/resolver/normalize_test.go`

**Step 1: Write the failing tests**

Create `internal/resolver/normalize_test.go`:
```go
package resolver_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "statora-cli/internal/resolver"
)

func TestNormalizeInstalled_ExactMatch(t *testing.T) {
    installed := []string{"8.1.20", "8.1.25", "8.2.15"}
    got := resolver.NormalizeInstalled("8.1.25", installed)
    assert.Equal(t, "8.1.25", got)
}

func TestNormalizeInstalled_PartialMinor(t *testing.T) {
    installed := []string{"8.1.20", "8.1.25", "8.2.15"}
    got := resolver.NormalizeInstalled("8.1", installed)
    assert.Equal(t, "8.1.25", got) // highest 8.1.x
}

func TestNormalizeInstalled_PartialMajor(t *testing.T) {
    installed := []string{"8.1.20", "8.2.15"}
    got := resolver.NormalizeInstalled("8", installed)
    assert.Equal(t, "8.2.15", got) // highest 8.x.x
}

func TestNormalizeInstalled_NoMatch(t *testing.T) {
    installed := []string{"8.2.15"}
    got := resolver.NormalizeInstalled("8.1", installed)
    assert.Equal(t, "", got) // nothing installed for 8.1.x
}

func TestNormalizeInstalled_Empty(t *testing.T) {
    got := resolver.NormalizeInstalled("8.1", nil)
    assert.Equal(t, "", got)
}
```

**Step 2: Run to verify it fails**
```
go test ./internal/resolver/... -run TestNormalizeInstalled -v
```
Expected: FAIL — function not defined.

**Step 3: Implement**

Create `internal/resolver/normalize.go`:
```go
package resolver

import (
    "strings"

    "github.com/Masterminds/semver/v3"
)

// NormalizeInstalled returns the highest installed concrete version that matches
// the given partial or exact version string.
//
// Examples:
//   NormalizeInstalled("8.1", ["8.1.20", "8.1.25", "8.2.15"]) → "8.1.25"
//   NormalizeInstalled("8.1.25", ["8.1.25"])                   → "8.1.25"
//   NormalizeInstalled("8.1", ["8.2.15"])                       → ""
func NormalizeInstalled(version string, installed []string) string {
    if len(installed) == 0 || version == "" {
        return ""
    }

    // Exact match: version is already a full X.Y.Z
    for _, v := range installed {
        if v == version {
            return v
        }
    }

    // Partial match: version is a prefix like "8" or "8.1"
    prefix := version + "."
    var matches []*semver.Version
    var matchStrs []string
    for _, v := range installed {
        if strings.HasPrefix(v, prefix) {
            sv, err := semver.NewVersion(v)
            if err == nil {
                matches = append(matches, sv)
                matchStrs = append(matchStrs, v)
            }
        }
    }
    if len(matches) == 0 {
        return ""
    }

    // Find the highest semver match.
    best := matches[0]
    bestStr := matchStrs[0]
    for i := 1; i < len(matches); i++ {
        if matches[i].GreaterThan(best) {
            best = matches[i]
            bestStr = matchStrs[i]
        }
    }
    return bestStr
}
```

**Step 4: Run to verify it passes**
```
go test ./internal/resolver/... -run TestNormalizeInstalled -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/resolver/normalize.go internal/resolver/normalize_test.go
git commit -m "feat(resolver): add NormalizeInstalled helper for partial version matching"
```

---

### Task 3: Update resolver to normalize PHP and Composer versions using installed dirs

After resolution, if the version is partial, replace it with the highest concrete installed version.

**Files:**
- Modify: `internal/resolver/resolver.go`
- Modify: `internal/resolver/resolver_test.go`

**Step 1: Write the failing tests**

Add to `internal/resolver/resolver_test.go`:
```go
func TestResolve_NormalizesPartialPHPToInstalled(t *testing.T) {
    cfg := makeConfig(t)

    // Fake-install PHP 8.1.25
    phpBin := cfg.PHPBin("8.1.25")
    require.NoError(t, os.MkdirAll(filepath.Dir(phpBin), 0o755))
    require.NoError(t, os.WriteFile(phpBin, []byte("#!/bin/sh"), 0o755))

    checker := compat.NewChecker()
    r := resolver.New(cfg, checker)

    dir := t.TempDir()
    require.NoError(t, os.WriteFile(filepath.Join(dir, ".statora"), []byte(`php = "8.1"
composer = "2.9.5"
`), 0o644))

    // Also fake-install composer 2.9.5
    composerBin := cfg.ComposerBin("2.9.5")
    require.NoError(t, os.MkdirAll(filepath.Dir(composerBin), 0o755))
    require.NoError(t, os.WriteFile(composerBin, []byte("#!/bin/sh"), 0o755))

    res, err := r.Resolve(dir)
    require.NoError(t, err)
    assert.Equal(t, "8.1.25", res.PHP)      // normalized from "8.1"
    assert.Equal(t, "2.9.5", res.Composer)  // already concrete, unchanged
}

func TestResolve_NormalizesPartialComposerToInstalled(t *testing.T) {
    cfg := makeConfig(t)
    checker := compat.NewChecker()
    r := resolver.New(cfg, checker)

    // Fake-install composer 2.2.8
    composerBin := cfg.ComposerBin("2.2.8")
    require.NoError(t, os.MkdirAll(filepath.Dir(composerBin), 0o755))
    require.NoError(t, os.WriteFile(composerBin, []byte("#!/bin/sh"), 0o755))

    // Fake-install PHP 8.1.25
    phpBin := cfg.PHPBin("8.1.25")
    require.NoError(t, os.MkdirAll(filepath.Dir(phpBin), 0o755))
    require.NoError(t, os.WriteFile(phpBin, []byte("#!/bin/sh"), 0o755))

    dir := t.TempDir()
    require.NoError(t, os.WriteFile(filepath.Join(dir, ".statora"), []byte(`php = "8.1"
composer = "2.2"
`), 0o644))

    res, err := r.Resolve(dir)
    require.NoError(t, err)
    assert.Equal(t, "8.1.25", res.PHP)   // normalized
    assert.Equal(t, "2.2.8", res.Composer) // normalized from "2.2"
}

func TestResolve_KeepsPartialIfNothingInstalled(t *testing.T) {
    cfg := makeConfig(t)
    checker := compat.NewChecker()
    r := resolver.New(cfg, checker)

    dir := t.TempDir()
    require.NoError(t, os.WriteFile(filepath.Join(dir, ".statora"), []byte(`php = "8.1"
composer = "2.2.0"
`), 0o644))

    res, err := r.Resolve(dir)
    require.NoError(t, err)
    // Nothing installed → keep as-is so caller can prompt to install
    assert.Equal(t, "8.1", res.PHP)
    assert.Equal(t, "2.2.0", res.Composer)
}
```

**Step 2: Run to verify they fail**
```
go test ./internal/resolver/... -run TestResolve_Normalizes -v
go test ./internal/resolver/... -run TestResolve_KeepsPartial -v
```
Expected: FAIL

**Step 3: Implement**

Add two private scanning helpers and a normalize step to `Resolve()` in `internal/resolver/resolver.go`:

```go
package resolver

import (
    "os"
    "path/filepath"

    "statora-cli/internal/config"
    "statora-cli/internal/compat"
)

// ... existing types unchanged ...

// Resolve returns the active versions for the given working directory.
// Partial versions (e.g. "8.1") are normalized to the highest concrete installed
// version. If nothing is installed for the partial, the partial is returned as-is
// so the caller can prompt for installation.
func (r *Resolver) Resolve(dir string) (Resolution, error) {
    // 1. Try project .statora
    proj, found, err := config.LoadProject(dir)
    if err != nil {
        return Resolution{}, err
    }
    if found && proj.PHP != "" {
        composer := proj.Composer
        if composer == "" {
            composer, err = r.checker.ResolveComposer(proj.PHP)
            if err != nil {
                return Resolution{}, err
            }
        }
        res := Resolution{
            PHP:      r.normalizePHP(proj.PHP),
            Composer: r.normalizeComposer(composer),
            Source:   "project",
        }
        return res, nil
    }

    // 2. Try global
    global, err := r.cfg.LoadGlobal()
    if err != nil {
        return Resolution{}, err
    }
    if global.PHP != "" {
        composer := global.Composer
        if composer == "" {
            composer, err = r.checker.ResolveComposer(global.PHP)
            if err != nil {
                return Resolution{}, err
            }
        }
        res := Resolution{
            PHP:      r.normalizePHP(global.PHP),
            Composer: r.normalizeComposer(composer),
            Source:   "global",
        }
        return res, nil
    }

    return Resolution{Source: "none"}, nil
}

// normalizePHP normalizes a partial PHP version to the highest installed concrete.
// Returns version unchanged if nothing matching is installed.
func (r *Resolver) normalizePHP(version string) string {
    installed, err := r.installedPHPVersions()
    if err != nil || len(installed) == 0 {
        return version
    }
    if norm := NormalizeInstalled(version, installed); norm != "" {
        return norm
    }
    return version
}

// normalizeComposer normalizes a partial Composer version to the highest installed concrete.
// Returns version unchanged if nothing matching is installed.
func (r *Resolver) normalizeComposer(version string) string {
    installed, err := r.installedComposerVersions()
    if err != nil || len(installed) == 0 {
        return version
    }
    if norm := NormalizeInstalled(version, installed); norm != "" {
        return norm
    }
    return version
}

// installedPHPVersions scans ~/.statora/runtimes/php/ for concrete version dirs.
func (r *Resolver) installedPHPVersions() ([]string, error) {
    dir := filepath.Join(r.cfg.Paths.RuntimesDir, "php")
    entries, err := os.ReadDir(dir)
    if os.IsNotExist(err) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    var versions []string
    for _, e := range entries {
        if e.IsDir() {
            // Only count versions that have a real php binary.
            if _, err := os.Stat(r.cfg.PHPBin(e.Name())); err == nil {
                versions = append(versions, e.Name())
            }
        }
    }
    return versions, nil
}

// installedComposerVersions scans ~/.statora/composer/ for concrete version dirs.
func (r *Resolver) installedComposerVersions() ([]string, error) {
    entries, err := os.ReadDir(r.cfg.Paths.ComposerDir)
    if os.IsNotExist(err) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    var versions []string
    for _, e := range entries {
        if e.IsDir() {
            // Only count versions that have a wrapper script (bin/composer).
            if _, err := os.Stat(r.cfg.ComposerBin(e.Name())); err == nil {
                versions = append(versions, e.Name())
            }
        }
    }
    return versions, nil
}
```

**Step 4: Run all resolver tests**
```
go test ./internal/resolver/... -v
```
Expected: All PASS

**Step 5: Commit**
```bash
git add internal/resolver/resolver.go internal/resolver/resolver_test.go
git commit -m "feat(resolver): normalize partial PHP/Composer versions to highest installed concrete"
```

---

### Task 4: Create composer wrapper script on install

After `composer.phar` is downloaded, create a thin `bin/composer` shell script so the version's bin dir can be placed on PATH.

**Files:**
- Modify: `internal/composer/manager.go`
- Modify: `internal/composer/manager_test.go`

**Step 1: Write the failing test**

Add to `internal/composer/manager_test.go`:
```go
func TestFakeInstall_CreatesWrapperScript(t *testing.T) {
    cfg := makeConfig(t)
    m := composer.NewManager(cfg, zap.NewNop())
    fakeInstall(t, cfg, "2.9.5")

    // fakeInstall only writes composer.phar — wrapper does not exist yet.
    // This test documents that after a real install, the wrapper must exist.
    // We test CreateWrapperScript directly.
    err := m.CreateWrapperScript("2.9.5")
    require.NoError(t, err)

    bin := cfg.ComposerBin("2.9.5")
    info, err := os.Stat(bin)
    require.NoError(t, err)
    assert.True(t, info.Mode()&0o111 != 0, "wrapper must be executable")

    content, err := os.ReadFile(bin)
    require.NoError(t, err)
    assert.Contains(t, string(content), "composer.phar")
}
```

**Step 2: Run to verify it fails**
```
go test ./internal/composer/... -run TestFakeInstall_CreatesWrapperScript -v
```
Expected: FAIL — `CreateWrapperScript` not defined.

**Step 3: Implement**

In `internal/composer/manager.go`, add after `IsInstalled`:

```go
// CreateWrapperScript writes an executable bin/composer shell script that
// invokes the PHAR for the given version. This allows ~/.statora/composer/<v>/bin/
// to be placed on PATH for fnm-style dispatch.
func (m *Manager) CreateWrapperScript(version string) error {
    binDir := filepath.Dir(m.cfg.ComposerBin(version))
    if err := os.MkdirAll(binDir, 0o755); err != nil {
        return fmt.Errorf("creating composer bin dir: %w", err)
    }

    phar := m.cfg.ComposerPhar(version)
    script := fmt.Sprintf("#!/bin/sh\nexec php %s \"$@\"\n", phar)

    bin := m.cfg.ComposerBin(version)
    if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
        return fmt.Errorf("writing composer wrapper script: %w", err)
    }
    return nil
}
```

Also update `Install()` to call `CreateWrapperScript` after the pipeline runs successfully:

```go
func (m *Manager) Install(versionOrConstraint string) error {
    concrete, sha256, err := ResolveVersion(versionOrConstraint)
    if err != nil {
        return fmt.Errorf("resolving Composer version: %w", err)
    }
    if concrete != versionOrConstraint {
        fmt.Printf("  Resolved %q → %s\n", versionOrConstraint, concrete)
    }

    if m.IsInstalled(concrete) {
        fmt.Printf("Composer %s is already installed.\n", concrete)
        return nil
    }

    pipeline := installer.New(
        &downloadStage{},
        &verifyChecksumStage{},
        &installStage{},
    )

    ctx := &installer.Context{
        Version:  concrete,
        Category: "composer",
        Cfg:      m.cfg,
        Log:      m.log,
        Data: map[string]any{
            "pharURL":    DownloadURL(concrete),
            "sha256URL":  SHA256URL(concrete),
            "sha256hint": sha256,
        },
    }
    if err := pipeline.Run(ctx); err != nil {
        return err
    }

    return m.CreateWrapperScript(concrete)
}
```

**Step 4: Run all composer tests**
```
go test ./internal/composer/... -v
```
Expected: All PASS

**Step 5: Commit**
```bash
git add internal/composer/manager.go internal/composer/manager_test.go
git commit -m "feat(composer): create bin/composer wrapper script on install for PATH dispatch"
```

---

### Task 5: Add `statora use` command

The core new command. Resolves versions, normalizes, prompts to install if missing, outputs PATH export.

**Files:**
- Create: `internal/use/use.go`
- Create: `internal/use/use_test.go`
- Create: `cmd/use.go`

**Step 1: Write unit tests for the PATH builder**

Create `internal/use/use_test.go`:
```go
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
```

**Step 2: Run to verify they fail**
```
go test ./internal/use/... -v
```
Expected: FAIL — package not found.

**Step 3: Implement `internal/use/use.go`**

Create `internal/use/use.go`:
```go
package use

import (
    "fmt"
    "io"
    "os"
    "strings"
)

// StripStatoraDirs removes ~/.statora/... entries from a colon-separated PATH string.
func StripStatoraDirs(path, homeDir string) string {
    statoraPrefix := homeDir + "/.statora/"
    var kept []string
    for _, p := range strings.Split(path, ":") {
        if p != "" && !strings.HasPrefix(p, statoraPrefix) {
            kept = append(kept, p)
        }
    }
    return strings.Join(kept, ":")
}

// BuildPATHExport returns the shell-specific line that sets PATH to prepend
// phpBinDir and composerBinDir in front of strippedPATH.
func BuildPATHExport(shell, phpBinDir, composerBinDir, strippedPATH string) string {
    switch shell {
    case "fish":
        dirs := []string{phpBinDir, composerBinDir}
        if strippedPATH != "" {
            dirs = append(dirs, strings.Split(strippedPATH, ":")...)
        }
        return fmt.Sprintf("set -gx PATH %s\n", strings.Join(dirs, " "))
    default:
        newPath := phpBinDir + ":" + composerBinDir
        if strippedPATH != "" {
            newPath += ":" + strippedPATH
        }
        return fmt.Sprintf("export PATH=%q\n", newPath)
    }
}

// PrintUse writes the PATH export line for the given shell and resolved dirs to w.
// currentPATH is os.Getenv("PATH"); homeDir is os.UserHomeDir().
func PrintUse(w io.Writer, shell, phpBinDir, composerBinDir, currentPATH, homeDir string) error {
    stripped := StripStatoraDirs(currentPATH, homeDir)
    _, err := fmt.Fprint(w, BuildPATHExport(shell, phpBinDir, composerBinDir, stripped))
    return err
}

// IsTerminal reports whether fd is connected to a terminal (for interactive prompts).
func IsTerminal(fd *os.File) bool {
    info, err := fd.Stat()
    if err != nil {
        return false
    }
    return (info.Mode() & os.ModeCharDevice) != 0
}
```

**Step 4: Run unit tests**
```
go test ./internal/use/... -v
```
Expected: All PASS

**Step 5: Implement `cmd/use.go`**

Create `cmd/use.go`:
```go
package cmd

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/spf13/cobra"

    "statora-cli/internal/app"
    "statora-cli/internal/composer"
    "statora-cli/internal/config"
    "statora-cli/internal/php"
    "statora-cli/internal/resolver"
    "statora-cli/internal/use"
)

var useCmd = &cobra.Command{
    Use:   "use",
    Short: "Output PATH export for the active PHP/Composer versions (eval this in your shell)",
    Long: `Resolve the active PHP/Composer versions and print a shell PATH export.

Designed to be eval'd by the shell hook set up by 'statora env'.

  zsh/bash:  eval "$(statora use)"
  fish:      statora use | source`,
    RunE: func(c *cobra.Command, _ []string) error {
        shellFlag, _ := c.Flags().GetString("shell")

        shell, err := DetectShell(os.Getenv("SHELL"), shellFlag)
        if err != nil {
            return err
        }

        dir, err := os.Getwd()
        if err != nil {
            return err
        }

        homeDir, err := os.UserHomeDir()
        if err != nil {
            return err
        }

        return app.Invoke(Debug(), func(
            res *resolver.Resolver,
            phpPlugin *php.Plugin,
            composerMgr *composer.Manager,
            cfg *config.Config,
        ) error {
            resolution, err := res.Resolve(dir)
            if err != nil {
                return err
            }
            if resolution.Source == "none" {
                // No config found — output nothing silently.
                return nil
            }

            // Ensure PHP is installed; prompt if not.
            phpVersion := resolution.PHP
            if !phpPlugin.IsInstalled(phpVersion) {
                if !promptInstall(fmt.Sprintf("PHP %s", phpVersion)) {
                    fmt.Fprintf(os.Stderr, "statora: skipping PATH update (PHP %s not installed)\n", phpVersion)
                    return nil
                }
                if err := phpPlugin.Install(phpVersion); err != nil {
                    return fmt.Errorf("installing PHP %s: %w", phpVersion, err)
                }
                // Re-normalize after install (concrete version may have been resolved by install).
                // Re-resolve to pick up newly installed version.
                resolution, err = res.Resolve(dir)
                if err != nil {
                    return err
                }
                phpVersion = resolution.PHP
            }

            // Ensure Composer is installed; prompt if not.
            composerVersion := resolution.Composer
            if !composerMgr.IsInstalled(composerVersion) {
                if !promptInstall(fmt.Sprintf("Composer %s", composerVersion)) {
                    fmt.Fprintf(os.Stderr, "statora: skipping PATH update (Composer %s not installed)\n", composerVersion)
                    return nil
                }
                if err := composerMgr.Install(composerVersion); err != nil {
                    return fmt.Errorf("installing Composer %s: %w", composerVersion, err)
                }
                resolution, err = res.Resolve(dir)
                if err != nil {
                    return err
                }
                composerVersion = resolution.Composer
            }

            phpBinDir := cfg.PHPRuntimeDir(phpVersion) + "/bin"
            composerBinDir := cfg.Paths.ComposerDir + "/" + composerVersion + "/bin"

            return use.PrintUse(os.Stdout, shell, phpBinDir, composerBinDir, os.Getenv("PATH"), homeDir)
        })
    },
}

// promptInstall asks the user interactively whether to install a tool.
// Returns true if the user answered yes, false if no or non-interactive.
func promptInstall(label string) bool {
    if !use.IsTerminal(os.Stdin) {
        return false
    }
    fmt.Printf("%s is not installed. Install now? [y/N] ", label)
    scanner := bufio.NewScanner(os.Stdin)
    if scanner.Scan() {
        answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
        return answer == "y" || answer == "yes"
    }
    return false
}

func init() {
    useCmd.Flags().String("shell", "", "Shell to target (zsh, bash, fish)")
    Root.AddCommand(useCmd)
}
```

**Step 6: Run all tests**
```
go test ./... -v 2>&1 | tail -30
```
Expected: All PASS (no compilation errors)

**Step 7: Commit**
```bash
git add internal/use/use.go internal/use/use_test.go cmd/use.go
git commit -m "feat: add statora use command for fnm-style per-terminal PATH dispatch"
```

---

### Task 6: Update `statora env` shell hooks

Replace `statora switch` calls with `statora use --shell <shell> | source` / `eval "$(statora use --shell <shell>)"`.

**Files:**
- Modify: `cmd/env.go`
- Modify: `cmd/env_test.go`

**Step 1: Update the tests first**

In `cmd/env_test.go`, update the three hook-content tests to assert `statora use` instead of `statora switch`, and remove `>/dev/null 2>&1` assertions:

```go
func TestEnvOutput_Zsh(t *testing.T) {
    var buf bytes.Buffer
    err := cmd.PrintEnv(&buf, "zsh")
    require.NoError(t, err)
    out := buf.String()
    assert.Contains(t, out, "add-zsh-hook")
    assert.Contains(t, out, "statora use")
    assert.Contains(t, out, "chpwd")
}

func TestEnvOutput_Bash(t *testing.T) {
    var buf bytes.Buffer
    err := cmd.PrintEnv(&buf, "bash")
    require.NoError(t, err)
    out := buf.String()
    assert.Contains(t, out, "PROMPT_COMMAND")
    assert.Contains(t, out, "statora use")
}

func TestEnvOutput_Fish(t *testing.T) {
    var buf bytes.Buffer
    err := cmd.PrintEnv(&buf, "fish")
    require.NoError(t, err)
    out := buf.String()
    assert.Contains(t, out, "--on-variable PWD")
    assert.Contains(t, out, "statora use")
}
```

**Step 2: Run to verify they fail**
```
go test ./cmd/... -run TestEnvOutput -v
```
Expected: FAIL (hooks still reference `statora switch`)

**Step 3: Update hook constants in `cmd/env.go`**

Replace the three hook constants:

```go
const zshHook = `autoload -U add-zsh-hook

_statora_use() {
  if command -v statora >/dev/null 2>&1; then
    eval "$(statora use --shell zsh 2>/dev/null)"
  fi
}

add-zsh-hook chpwd _statora_use
_statora_use
`

const bashHook = `_statora_use() {
  if command -v statora >/dev/null 2>&1; then
    eval "$(statora use --shell bash 2>/dev/null)"
  fi
}

if [[ "${PROMPT_COMMAND}" != *"_statora_use"* ]]; then
  PROMPT_COMMAND="_statora_use${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
_statora_use
`

const fishHook = `function __statora_use --on-variable PWD
  if command -q statora
    statora use --shell fish 2>/dev/null | source
  end
end
__statora_use
`
```

**Step 4: Run all env tests**
```
go test ./cmd/... -run TestEnv -v
go test ./cmd/... -run TestDetect -v
```
Expected: All PASS

**Step 5: Commit**
```bash
git add cmd/env.go cmd/env_test.go
git commit -m "feat(env): update shell hooks to use statora use for fnm-style PATH dispatch"
```

---

### Task 7: Remove symlinks, dispatcher, and dispatch fast-path

**Files:**
- Delete: `cmd/symlinks.go`
- Delete: `cmd/symlinks_test.go`
- Delete: `internal/dispatch/dispatcher.go`
- Modify: `main.go` — remove fast-path block and dispatch import
- Modify: `internal/switcher/switcher.go` — remove `dispatch.InvalidateCache` import if no longer needed

**Step 1: Delete the symlinks files**
```bash
rm cmd/symlinks.go cmd/symlinks_test.go
```

**Step 2: Delete the dispatcher**
```bash
rm internal/dispatch/dispatcher.go
```

**Step 3: Update `main.go`**

Remove the fast-path block and dispatch import. New `main.go`:

```go
package main

import (
    "fmt"
    "os"

    "statora-cli/cmd"

    // Register all cobra subcommands via init().
    _ "statora-cli/cmd/php"
    _ "statora-cli/cmd/composer"
    _ "statora-cli/cmd/ext"
)

func main() {
    cmd.Execute()
}

func fatalf(format string, args ...any) {
    //nolint:forbidigo
    _, _ = os.Stderr.WriteString(fmt.Sprintf(format, args...))
    os.Exit(1)
}
```

**Step 4: Verify it compiles and tests pass**
```
go build ./...
go test ./... 2>&1 | grep -E "FAIL|ok"
```
Expected: All packages build and pass. Fix any import errors (e.g. if `switcher.go` imports `dispatch` only for `InvalidateCache` — keep `rescache.go`, just the `dispatcher.go` is removed).

**Step 5: Commit**
```bash
git add main.go
git rm cmd/symlinks.go cmd/symlinks_test.go internal/dispatch/dispatcher.go
git commit -m "remove: symlinks command, dispatcher, and main.go dispatch fast-path"
```

---

### Task 8: Fix `.rescache` to write concrete versions

After Task 3, `resolver.Resolve()` returns concrete versions. `switcher.Execute()` passes `res` (which now has concrete versions) to `dispatch.InvalidateCache()`. Verify this is correct and that the `.rescache` bug is fixed.

**Files:**
- Modify: `internal/dispatch/rescache.go` — update `KeyComposer` to store version string (not phar path) for consistency with `KeyPHP`
- Modify: `internal/switcher/switcher.go` — update plan comparison to use version strings

**Step 1: Read the current rescache write logic**

`InvalidateCache` currently writes:
- `KeyPHP` → `cfg.PHPBin(res.PHP)` (binary path)
- `KeyPHPActive` → `res.PHP` (version string)
- `KeyComposer` → `cfg.ComposerPhar(res.Composer)` (phar path)

And `BuildPlan` compares:
- PHP: `currentActive` (`.php_active` value = version string) vs `res.PHP`
- Composer: `currentComposer` (`.rescache/composer` = phar path) vs `cfg.ComposerPhar(res.Composer)`

This is inconsistent. Standardize: store **version strings** in all keys.

**Step 2: Update `rescache.go`**

```go
// InvalidateCache rebuilds all rescache entries from the resolver.
// All values are stored as version strings (not paths) for consistency.
func InvalidateCache(cfg *config.Config, res resolver.Resolution) error {
    if res.PHP != "" {
        if err := WriteCache(cfg, KeyPHP, res.PHP); err != nil {
            return err
        }
        if err := WriteCache(cfg, KeyPHPActive, res.PHP); err != nil {
            return err
        }
    }
    if res.Composer != "" {
        if err := WriteCache(cfg, KeyComposer, res.Composer); err != nil {
            return err
        }
    }
    return nil
}
```

**Step 3: Update `switcher.BuildPlan`** to compare version strings for composer:

```go
// Composer action
currentComposer := dispatch.ReadCache(s.cfg, dispatch.KeyComposer)
if currentComposer != res.Composer {
    plan.Actions = append(plan.Actions, Action{
        Kind:    "composer",
        Name:    res.Composer,
        Current: currentComposer,
        Next:    res.Composer,
    })
}
```

**Step 4: Run all tests**
```
go test ./... 2>&1 | grep -E "FAIL|ok"
```
Expected: All PASS

**Step 5: Commit**
```bash
git add internal/dispatch/rescache.go internal/switcher/switcher.go
git commit -m "fix(rescache): store version strings consistently; fixes stale 8.1 / 2.2.0 display"
```

---

### Task 9: Update README

**Files:**
- Modify: `README.md`

**Step 1: Read the current README**
```
cat README.md
```

**Step 2: Make these changes to README.md:**

1. **Remove** the `statora symlinks` section entirely.
2. **Update Setup / Shell Integration** section — change the hook instruction from:
   ```
   eval "$(statora env)"   # adds statora switch to chpwd/PROMPT_COMMAND
   ```
   to:
   ```
   eval "$(statora env)"   # adds statora use to chpwd/PROMPT_COMMAND, sets PATH per terminal
   ```
3. **Add `statora use` to the commands reference table:**
   ```
   | statora use [--shell fish\|bash\|zsh] | Output PATH export for active PHP/Composer versions |
   ```
4. **Remove** any mention of symlinks or `statora symlinks` from the docs.
5. **Add a note** that two terminals with different projects get isolated PHP/Composer versions automatically.

**Step 3: Verify README renders correctly** (visual check)

**Step 4: Commit**
```bash
git add README.md
git commit -m "docs(readme): remove symlinks section, document statora use and per-terminal isolation"
```

---

## Final verification

```bash
go build ./...
go test ./...
```

Expected: clean build, all tests pass.

Then manually test the full flow:
1. `statora env --shell fish` — verify hook references `statora use`
2. Create a `.statora` with `php = "8.1"` (partial), `cd` into it — verify PATH is set to `8.1.25/bin`
3. Open a second terminal with a different project — verify each terminal has its own PHP in PATH
4. Verify `php --version` shows the correct version in each terminal
