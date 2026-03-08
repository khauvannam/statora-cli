# `statora env` Shell Integration Design

**Date:** 2026-03-08

## Problem

Users must manually write shell hooks to auto-switch PHP/Composer when changing directories.
The README only covers the manual `statora switch` command. There is no `eval`-able
integration like `fnm env` or `direnv hook`.

## Solution

`statora env` prints shell hook code to stdout. Users add a single `eval` line to their
shell rc file. When they `cd` into a project with a `.statora` file, the active PHP,
Composer, and extensions switch automatically.

## Command

```
statora env [--shell zsh|bash|fish]
```

Auto-detects shell from `filepath.Base(os.Getenv("SHELL"))`. `--shell` overrides.
Unsupported shell → error: `unsupported shell %q — use --shell zsh|bash|fish`.

## Shell Output

### zsh — `eval "$(statora env)"` in `~/.zshrc`

```zsh
autoload -U add-zsh-hook

statora_auto_switch() {
  if command -v statora >/dev/null 2>&1; then
    statora switch >/dev/null 2>&1
  fi
}

add-zsh-hook chpwd statora_auto_switch
statora_auto_switch
```

### bash — `eval "$(statora env)"` in `~/.bashrc`

```bash
statora_auto_switch() {
  if command -v statora >/dev/null 2>&1; then
    statora switch >/dev/null 2>&1
  fi
}

if [[ "${PROMPT_COMMAND}" != *"statora_auto_switch"* ]]; then
  PROMPT_COMMAND="statora_auto_switch${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
statora_auto_switch
```

### fish — `statora env | source` in `~/.config/fish/config.fish`

```fish
function __statora_auto_switch --on-variable PWD
  if command -v statora >/dev/null 2>&1
    statora switch >/dev/null 2>&1
  end
end
__statora_auto_switch
```

## Files Changed

| File | Change |
|---|---|
| `cmd/env.go` | New — cobra command + inline shell snippets |
| `README.md` | Add Setup §4: shell integration |
| `.goreleaser.yaml` | Update `brews.caveats` to reference `statora symlinks` and `statora env` |

## Non-Goals

- No `--no-use` or `--use-on-cd` flags
- No management of `$PATH` (statora env only registers the cd hook)
- No Windows support
