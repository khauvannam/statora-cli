# FNM-Style PATH Dispatch Design

**Date:** 2026-03-08

## Problem

The current dispatch model uses a global `~/.statora/.rescache/` directory and symlinks (`php` → `statora`) to route invocations to the active binary. This has three bugs:

1. **`.rescache` stores partial versions** — if `.statora` says `php = "8.1"`, `.rescache` writes `"8.1"` instead of the concrete installed version (`"8.1.25"`), causing stale/wrong paths.
2. **Composer mismatch** — same root cause; `.rescache/composer` stores the constraint string instead of the resolved concrete path.
3. **Parallel projects conflict** — `.rescache` is global. Two terminals with different projects (7.2 vs 8.1) overwrite each other on directory change.

## Solution: FNM-Style Per-Terminal PATH Isolation

Replace symlink + `.rescache` dispatch with per-terminal `PATH` manipulation, exactly like fnm/nvm/rbenv.

Each terminal session has its own `PATH` pointing to the concrete version dirs for that session. Two terminals with different projects are fully isolated.

## Architecture

### `statora use` (new command)

Non-interactive command called by the shell hook on directory change.

1. Reads `.statora` from current dir (or global config fallback)
2. Normalizes partial version → highest concrete installed match:
   - `"8.1"` + installed `[8.1.20, 8.1.25]` → `"8.1.25"`
   - `"2.2.0"` + installed `[2.2.6, 2.2.8]` → `"2.2.8"`
3. If no installed version matches:
   - Interactive mode: prompt "PHP 8.1.25 not installed. Install now? [y/N]" → install then continue
   - Non-interactive (piped): warn and skip PATH update
4. Strips any existing `~/.statora/...` entries from `$PATH`
5. Outputs shell-specific PATH export — shell `eval`s it

### Shell Hook (`statora env`)

Updated output per shell:

**fish:**
```fish
function __statora_use --on-variable PWD
    statora use --shell fish | source
end
statora use --shell fish | source
```

**bash:**
```bash
_statora_hook() { eval "$(statora use --shell bash)"; }
PROMPT_COMMAND="_statora_hook;$PROMPT_COMMAND"
_statora_hook
```

**zsh:**
```zsh
autoload -U add-zsh-hook
_statora_hook() { eval "$(statora use --shell zsh)"; }
add-zsh-hook chpwd _statora_hook
_statora_hook
```

### PATH Construction

For each terminal session, `statora use` outputs:

```sh
export PATH="~/.statora/runtimes/php/8.1.25/bin:~/.statora/composer/2.9.5/bin:<PATH_without_statora>"
```

Both PHP and Composer are resolved to concrete versions.

### Composer Wrapper Script

Composer is a PHAR, not a compiled binary. On `composer install`, create a thin wrapper:

`~/.statora/composer/<version>/bin/composer`:
```sh
#!/bin/sh
exec php ~/.statora/composer/<version>/composer.phar "$@"
```

This allows `~/.statora/composer/<version>/bin/` to be placed on PATH like any bin dir.

### `statora switch` (keep, updated)

Remains as the interactive command for explicitly switching versions. Internally reuses the same resolve + normalize logic as `statora use`. After switching, outputs PATH export for the current shell session.

### `.rescache` (kept for status only)

- Still written after switch/use with **concrete** versions (fixes the 8.1 stale bug)
- No longer used for dispatch
- Read only by `statora status` for display

### `.statora` file

Never rewritten by statora. Partial versions (`"8.1"`, `"2.2.0"`) are valid and resolved at runtime.

## Removed Components

- `cmd/symlinks.go` — deleted entirely
- `main.go` dispatch fast-path (`argv[0] == "php"` branch) — deleted
- `internal/dispatch/dispatcher.go` — deleted (or gutted to no-op)
- Symlink creation/removal from install flows

## File Changes

| File | Action |
|---|---|
| `cmd/use.go` | New — `statora use` command |
| `cmd/env.go` | Update hook snippets to `eval (statora use --shell ...)` |
| `cmd/switch.go` | Update to reuse `use` logic; output PATH export |
| `cmd/symlinks.go` | Delete |
| `internal/resolver/resolver.go` | Add installed-version normalization (partial → highest concrete) |
| `internal/composer/manager.go` | Create `bin/composer` wrapper script on install |
| `internal/dispatch/rescache.go` | Keep; ensure writes use concrete versions |
| `internal/dispatch/dispatcher.go` | Delete |
| `main.go` | Remove symlink dispatch fast-path |
| `README.md` | Update setup instructions; remove symlinks section; document `statora use` |
