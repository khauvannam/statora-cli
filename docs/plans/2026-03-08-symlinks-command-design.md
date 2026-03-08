# `statora symlinks` Command Design

**Date:** 2026-03-08

## Problem

The README instructs users to manually run `ln -sf` and `rm` commands in two places
(Setup §3 and Uninstall §1) to manage `php` and `composer` shim symlinks. This is
error-prone and shell-specific (different syntax for bash/zsh vs fish).

## Solution

Add a `statora symlinks` toggle command that creates or removes the `php` and `composer`
symlinks next to the statora binary automatically.

## Behaviour

Toggle logic based on symlink state:

| State | Action |
|---|---|
| All symlinks exist | Remove all → print "Removed symlinks: php, composer" |
| Any symlink missing | Create missing ones → print "Created symlinks: php" (or whichever) |

Resolves the binary directory via `os.Executable()` + `filepath.EvalSymlinks()` to
correctly handle cases where statora itself is installed as a symlink (e.g. Homebrew).

## Error Handling

- `os.Executable()` failure → return error
- `os.Remove` failure → return error with path
- `os.Symlink` failure (e.g. permission denied) → return error with path

## Files Changed

| File | Change |
|---|---|
| `cmd/symlinks.go` | New file — cobra command + toggle logic |
| `README.md` | Replace manual ln/rm blocks with `statora symlinks` |

## Non-Goals

- No `--status` flag
- No subcommands
- No management of symlinks in arbitrary directories
