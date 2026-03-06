# Statora — Cobra Command Tree

## Root

```
statora [command] [subcommand] [args] [flags]

Aliases:  none
Flags:
  --debug          enable debug logging
  --version, -v    print statora version
  --help, -h       show help
```

---

## PHP Commands

```
statora php
```

| Command | Alias | Args | Flags | Description |
|---|---|---|---|---|
| `statora php install <version>` | `statora php i` | version | `--force`, `--no-cache` | Install a PHP version |
| `statora php uninstall <version>` | `statora php rm` | version | `--purge-ext` | Uninstall a PHP version, optionally purge its extensions |
| `statora php list` | `statora php ls` | — | `--remote`, `--installed` | List PHP versions |
| `statora php global <version>` | — | version | — | Set global PHP version |
| `statora php local <version>` | — | version | — | Write version to `.statora` in current dir |
| `statora php current` | — | — | — | Print currently active PHP version |
| `statora php which` | — | — | — | Print resolved binary path |
| `statora php rehash` | — | — | — | Rebuild rescache after manual changes |

```
statora php list --remote         # fetch available versions from php.net
statora php list --installed      # show only locally installed versions
statora php install 8.2.15 --force   # reinstall even if already present
statora php uninstall 5.6.40 --purge-ext  # also remove all extensions for this version
```

---

## Composer Commands

```
statora composer
```

| Command | Alias | Args | Flags | Description |
|---|---|---|---|---|
| `statora composer install <version>` | `statora composer i` | version | `--force` | Install a Composer version |
| `statora composer uninstall <version>` | `statora composer rm` | version | — | Remove a Composer version |
| `statora composer list` | `statora composer ls` | — | `--remote`, `--installed` | List Composer versions |
| `statora composer global <version>` | — | version | — | Set global Composer version |
| `statora composer local <version>` | — | version | — | Write to `.statora` |
| `statora composer current` | — | — | — | Print active Composer version |
| `statora composer compat <php-version>` | — | php version | — | Show compatible Composer versions for a given PHP |

```
statora composer compat 5.6.40
  # Output:
  # PHP 5.6.40 is compatible with Composer 1.0.0 – 1.10.27
  # Recommended: 1.10.27

statora composer list --remote
statora composer install 1.10.27
```

---

## Extension Commands

```
statora ext
```

| Command | Alias | Args | Flags | Description |
|---|---|---|---|---|
| `statora ext install <name>` | `statora ext i` | ext name | `--php`, `--source`, `--binary` | Install extension |
| `statora ext uninstall <name>` | `statora ext rm` | ext name | `--php` | Remove extension |
| `statora ext enable <name>` | — | ext name | `--php` | Symlink into enabled/ |
| `statora ext disable <name>` | — | ext name | `--php` | Remove symlink from enabled/ |
| `statora ext list` | `statora ext ls` | — | `--php`, `--enabled`, `--available` | List extensions |
| `statora ext info <name>` | — | ext name | `--php` | Show version, status, install method |

```
statora ext install redis                     # installs for current active PHP
statora ext install redis --php 8.2.15        # installs for specific version
statora ext install xdebug --source           # force compile from source
statora ext install imagick --binary          # force binary only, no fallback
statora ext enable xdebug
statora ext disable xdebug
statora ext list --php 8.2.15 --enabled       # only enabled extensions
statora ext list --php 8.2.15 --available     # all installed regardless of state
```

---

## Switch Command

```
statora switch
```

| Command | Alias | Args | Flags | Description |
|---|---|---|---|---|
| `statora switch` | — | — | `--dry-run`, `--yes` | Read `.statora`, build SwitchPlan, prompt |

```
statora switch              # interactive prompt
statora switch --dry-run    # print plan without executing
statora switch --yes        # apply all changes without prompting
```

This is the primary UX command. Equivalent to running `statora switch` when entering a project directory.

---

## Version File Shortcuts

These mirror `phpenv`/`goenv` top-level ergonomics for users who prefer the flat style:

| Command | Equivalent |
|---|---|
| `statora global php 8.2.15` | `statora php global 8.2.15` |
| `statora local php 5.6.40` | `statora php local 5.6.40` |
| `statora global composer 2.7.4` | `statora composer global 2.7.4` |
| `statora local composer 1.10.27` | `statora composer local 1.10.27` |

Both forms work. The grouped form (`statora php global`) is canonical. The flat form (`statora global php`) is a registered alias for discoverability.

---

## Full Tree at a Glance

```
statora
├── php
│   ├── install <version>       [-f, --force] [--no-cache]
│   ├── uninstall <version>     [--purge-ext]
│   ├── list                    [--remote] [--installed]
│   ├── global <version>
│   ├── local <version>
│   ├── current
│   ├── which
│   └── rehash
│
├── composer
│   ├── install <version>       [--force]
│   ├── uninstall <version>
│   ├── list                    [--remote] [--installed]
│   ├── global <version>
│   ├── local <version>
│   ├── current
│   └── compat <php-version>
│
├── ext
│   ├── install <name>          [--php] [--source] [--binary]
│   ├── uninstall <name>        [--php]
│   ├── enable <name>           [--php]
│   ├── disable <name>          [--php]
│   ├── list                    [--php] [--enabled] [--available]
│   └── info <name>             [--php]
│
├── switch                      [--dry-run] [--yes]
├── global <runtime> <version>  (alias)
└── local <runtime> <version>   (alias)
```

---

## cmd/ File Mapping

Each leaf command gets its own file. No command file handles more than one concern.

```
cmd/
├── root.go                  ← global flags, fx boot, cobra root
├── global.go                ← alias: statora global <runtime> <version>
├── local.go                 ← alias: statora local <runtime> <version>
├── switch.go                ← statora switch
│
├── php/
│   ├── php.go               ← parent command: statora php
│   ├── install.go
│   ├── uninstall.go
│   ├── list.go
│   ├── global.go
│   ├── local.go
│   ├── current.go
│   ├── which.go
│   └── rehash.go
│
├── composer/
│   ├── composer.go          ← parent command: statora composer
│   ├── install.go
│   ├── uninstall.go
│   ├── list.go
│   ├── global.go
│   ├── local.go
│   ├── current.go
│   └── compat.go
│
└── ext/
    ├── ext.go               ← parent command: statora ext
    ├── install.go
    ├── uninstall.go
    ├── enable.go
    ├── disable.go
    ├── list.go
    └── info.go
```

One parent `.go` file per group registers the subcommands onto itself, then `root.go` registers all three parents onto the cobra root. Clean, no circular registration.
