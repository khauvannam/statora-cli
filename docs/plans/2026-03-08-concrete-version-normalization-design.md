# Concrete Version Normalization for PHP and Composer

**Date:** 2026-03-08

## Problem

When a user runs `statora php install 8.1`, the resolver correctly fetches PHP `8.1.25`
from php.net. However, `ctx.Version` remains `"8.1"` throughout the pipeline, so build
directories, runtime directories, and extension directories are all created under `8.1`
instead of `8.1.25`. This causes conflicts if two patch releases of the same minor are
ever installed.

Similarly, `statora composer install 2` should resolve to the latest `2.x.x` patch and
use that concrete version for all storage paths.

## Approach

Normalize `ctx.Version` to the full concrete version as early as possible, and fix
partial version resolution for Composer.

## Changes

### PHP — `internal/php/stages.go`

In `resolveSourceStage.Run()`, after the URL is fetched (e.g.
`https://www.php.net/distributions/php-8.1.25.tar.gz`), extract the concrete version
from the filename:

```
php-8.1.25.tar.gz  →  strip "php-" prefix and ".tar.gz" suffix  →  "8.1.25"
```

Set `ctx.Version = "8.1.25"` before returning. All downstream stages pick it up
automatically — no other changes needed.

### Composer — `internal/composer/source.go`

In `ResolveVersion()`, before the exact-match lookup, detect a partial version input
(1–2 dot-separated numeric parts, e.g. `"2"` or `"2.7"`). Scan the stable list and
return the highest release whose version string shares the same major (and minor, if
provided) prefix.

`manager.go` already sets `ctx.Version = concrete`, so no pipeline changes are needed.

## Files Changed

| File | Change |
|---|---|
| `internal/php/stages.go` | Extract concrete version from URL, update `ctx.Version` |
| `internal/composer/source.go` | Add partial version prefix matching before exact-match |

## Non-Goals

- No changes to `config.go`, `pipeline.go`, `manager.go`, `plugin.go`, or cmd files.
- No new abstractions or helpers outside the two files above.
