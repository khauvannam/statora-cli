# Statora CLI — Implementation Design

Date: 2026-03-06

## Overview

Statora is a PHP tooling version manager built in Go. It manages PHP versions, Composer versions, and PHP extensions per-project via a `.statora` config file.

## Approach

Top-down build order: bootstrap → config → compat → resolver → php → composer → ext → switcher → dispatch → cmd leaves.

## Libraries

| Library | Import | Purpose |
|---|---|---|
| cobra | `github.com/spf13/cobra` | CLI command tree, flags, aliases |
| fx | `go.uber.org/fx` | Dependency injection, module wiring, lifecycle |
| zap | `go.uber.org/zap` | Structured logging, management path only |
| viper | `github.com/spf13/viper` | Layered config |
| semver | `github.com/Masterminds/semver/v3` | Semver parsing, constraint matching |
| grab | `github.com/cavaliergopher/grab/v3` | Resumable downloads with progress |
| archiver | `github.com/mholt/archiver/v3` | Extract `.tar.gz`, `.tar.bz2` |
| bubbletea | `github.com/charmbracelet/bubbletea` | TUI interactive prompt for `statora switch` |
| tablewriter | `github.com/olekukonko/tablewriter` | Table output for `list` commands |
| color | `github.com/fatih/color` | Colored terminal output |
| testify | `github.com/stretchr/testify` | Assertions and mocks |

HTTP retry logic is implemented manually (no third-party retry library).

## Build Order

1. `main.go` + `cmd/root.go` + `internal/app/app.go` — fx bootstrap, cobra root, argv[0] dispatch branch
2. `internal/config` — reads `~/.statora/config.toml` and `.statora` project file via viper
3. `internal/logger` — zap logger (production vs dev based on `--debug` flag)
4. `internal/compat` — PHP/Composer compatibility matrix + `ResolveComposer`, `IsCompatible`
5. `internal/resolver` — resolves active versions (global → local → project `.statora`)
6. `internal/installer` — shared staged pipeline runner (`Stage` interface, `Pipeline.Run`)
7. `internal/php` — PHP install pipeline stages (resolve, download, verify, extract, compile, install, invalidate)
8. `internal/composer` — Composer install pipeline stages (resolve, download, GPG verify, install, invalidate)
9. `internal/extension` — binary-first (PECL → GitHub → distro), phpize-source fallback, symlink enable/disable
10. `internal/switcher` — SwitchPlan builder, Bubble Tea TUI prompt, plan executor
11. `internal/dispatch` — rescache read/write, `syscall.Exec` hot path (Unix only)
12. `cmd/php/`, `cmd/composer/`, `cmd/ext/`, `cmd/switch.go` — cobra leaf commands wired to fx services

## Architecture

### Directory Layout (runtime)

```
~/.statora/
├── runtimes/php/<version>/bin/php
├── runtimes/php/<version>/lib/extensions/available/
├── runtimes/php/<version>/lib/extensions/enabled/   ← symlinks
├── composer/<version>/composer.phar
├── cache/downloads/
├── cache/builds/
├── .rescache/php
├── .rescache/composer
├── .rescache/.php_active
├── config.toml
└── versions/global.toml
```

### Data Flow — `statora php install`

```
cmd/php/install.go → php.PHPPlugin → installer.Pipeline →
  [ResolveSource → Download → VerifyChecksum → Extract → Compile → Install → InvalidateCache]
```

### Data Flow — `statora switch`

```
cmd/switch.go → switcher.Switcher →
  resolver.Resolve()          (reads .statora + globals)
  compat.ResolveComposer()    (infers composer if absent)
  bubbletea TUI prompt        (present SwitchPlan)
  plan.Execute()              (calls php/composer/ext actions)
  dispatch.InvalidateCache()  (rebuild rescache)
```

### Data Flow — dispatch hot path

```
main.go detects argv[0] == "php" | "composer"
  dispatch.Dispatch()
    rescache.Read("php")   ← one file read, no fx startup
    syscall.Exec(phpBin, args, env)
```

## Key Design Decisions

| Decision | Choice | Reason |
|---|---|---|
| DI | `uber/fx` | Lifecycle management, module isolation |
| Config | `viper` | Layered: env → `.statora` → `config.toml` |
| Downloads | `grab` | Resumable, progress reporting |
| HTTP retry | custom | Exponential backoff, no external dependency |
| Semver | `Masterminds/semver/v3` | Constraint matching for compat matrix |
| Interactive TUI | `bubbletea` | Rich TUI for SwitchPlan confirmation |
| Dispatch | `syscall.Exec` | Zero-overhead process replacement (Unix only) |
| Ext enable/disable | symlinks | Fast, inspectable, no config rewrite |
| Logging | `zap` | Management path only, not dispatch path |

## Installer Pipeline

### PHP

`ResolveSourceStage → DownloadStage → VerifyChecksumStage → ExtractStage → CompileStage → InstallStage → InvalidateCacheStage`

### Composer

`ResolveSourceStage → DownloadStage → VerifySignatureStage → InstallStage → InvalidateCacheStage`

### Extension

`ProbeStage → BinaryInstallStage → SourceDownloadStage → PhpizeStage → ConfigureStage → MakeStage → LinkStage → InvalidateCacheStage`

Binary sources tried in order: PECL prebuilt index → GitHub releases → distro package extraction.

## Compat Matrix

```go
var ComposerMatrix = []CompatRule{
    {PHP: Range{"5.3.0", "5.6.99"}, Composer: Range{"1.0.0", "1.10.27"}},
    {PHP: Range{"7.0.0", "7.1.99"}, Composer: Range{"1.0.0", "2.2.24"}},
    {PHP: Range{"7.2.0", "7.4.99"}, Composer: Range{"2.0.0", "2.7.x"}},
    {PHP: Range{"8.0.0", "8.9.99"}, Composer: Range{"2.2.0", "2.7.x"}},
}
```

## Platform

Unix/Linux/macOS only. `syscall.Exec` used directly — no Windows shims.
