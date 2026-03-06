Here is the full redesign, narrowed to PHP tooling only with the compatibility-aware switching system.

---

# Statora — PHP Tooling Architecture

## 1. Scope

Three managed concerns, nothing else:

```
PHP versions
Composer versions
PHP extensions (per PHP version)
```

---

## 2. Project File Format

The `.statora` file is the source of truth for a project:

```toml
php = "5.6.40"
composer = "1.10.27"       # optional — auto-resolved if absent
extensions = ["redis", "imagick", "xdebug"]
```

If `composer` is not declared, the resolver infers a compatible version automatically using the compatibility matrix described in section 4.

---

## 3. Directory Layout

```
~/.statora/
│
├── runtimes/
│   └── php/
│       ├── 5.6.40/
│       │   ├── bin/php
│       │   ├── bin/phpize
│       │   ├── etc/php.ini
│       │   └── lib/extensions/
│       │       ├── enabled/
│       │       │   └── redis.so -> ../available/redis.so
│       │       └── available/
│       │           ├── redis.so
│       │           └── xdebug.so
│       └── 8.2.15/
│           └── ...
│
├── composer/
│   ├── 1.10.27/
│   │   └── composer.phar
│   └── 2.7.4/
│       └── composer.phar
│
├── cache/
│   ├── downloads/
│   └── builds/
│
├── .rescache/             ← hot dispatch cache
│   ├── php
│   └── composer
│
├── config.toml
└── versions/
    └── global.toml
```

Extensions live **inside** the PHP version directory. Each PHP version has its own isolated extension set. No cross-version bleed.

---

## 4. Composer Compatibility Matrix

This is a first-class subsystem, not a comment in the code.

```go
// internal/compat/matrix.go

type Range struct {
    Min string
    Max string
}

var ComposerMatrix = []CompatRule{
    {PHP: Range{"5.3.0", "5.6.99"}, Composer: Range{"1.0.0", "1.10.27"}},
    {PHP: Range{"7.0.0", "7.1.99"}, Composer: Range{"1.0.0", "2.2.24"}},  // 2.2.x is LTS
    {PHP: Range{"7.2.0", "7.4.99"}, Composer: Range{"2.0.0", "2.7.x"}},
    {PHP: Range{"8.0.0", "8.9.99"}, Composer: Range{"2.2.0", "2.7.x"}},
}

func ResolveComposer(phpVersion string) (string, error) {
    // walk matrix, return highest compatible Composer version
}

func IsCompatible(phpVersion, composerVersion string) (bool, string) {
    // returns bool + human reason if incompatible
}
```

This matrix is the authority. The resolver and the switch prompt both consult it.

---

## 5. Smart Switch — The Core UX Feature

When a project has a `.statora` file, running any PHP command (or `statora switch`) triggers a compatibility check against the active global versions.

Flow:

```
enter project directory
  │
  ▼
statora detects .statora
  │
  ▼
compare .statora php vs current global php
  │
  ├── match → no prompt, proceed
  │
  └── mismatch
        │
        ▼
      prompt user
        │
        ├── php version change required?
        ├── composer version compatible with new php?
        └── extensions declared — are they installed?
```

Example prompt output:

```
.statora detected — version mismatch

  php:       global=8.2.15  project=5.6.40
  composer:  global=2.7.4   required=1.10.27 (PHP 5.x constraint)
  extensions: redis [not installed for php 5.6.40]

Actions required:
  [1] Switch php to 5.6.40
  [2] Switch composer to 1.10.27
  [3] Install redis for php 5.6.40

Apply all? [Y/n]  or select individually [1/2/3]:
```

The prompt is driven by a `SwitchPlan` struct:

```go
// internal/switcher/plan.go

type SwitchPlan struct {
    PHPChange      *VersionChange
    ComposerChange *VersionChange
    ExtActions     []ExtAction
}

type VersionChange struct {
    From   string
    To     string
    Reason string
}

type ExtAction struct {
    Name   string
    Action string  // "install" | "enable" | "missing_build_deps"
}
```

The switcher assembles the plan, presents it, waits for confirmation, then executes each action in order.

---

## 6. Extension Management

### Architecture

Extensions are scoped per PHP version:

```
~/.statora/runtimes/php/8.2.15/lib/extensions/
    available/     ← installed .so files live here
    enabled/       ← symlinks to available/
```

Enable/disable is a symlink operation, not a config rewrite. Fast and inspectable.

### Install Strategy — Binary First, Source Fallback

```go
// internal/extension/installer.go

type ExtInstaller struct {
    log *zap.Logger
}

func (e *ExtInstaller) Install(ext, phpVersion string) error {
    // Step 1: try binary
    if err := e.installBinary(ext, phpVersion); err == nil {
        return nil
    }

    e.log.Info("binary install failed, falling back to source compile",
        zap.String("ext", ext),
        zap.String("php", phpVersion),
    )

    // Step 2: compile via phpize
    return e.compileFromSource(ext, phpVersion)
}

func (e *ExtInstaller) compileFromSource(ext, phpVersion string) error {
    phpizeBin := phpizePath(phpVersion)
    // run: phpize, ./configure, make, make install
    // into ~/.statora/runtimes/php/<version>/lib/extensions/available/
}
```

Binary sources checked in order:
1. PECL prebuilt index
2. GitHub releases for the extension repo
3. Distro package extraction (`.deb` / `.rpm` unpack, pull `.so`)

### CLI Commands

```
statora ext install redis            # installs for currently active PHP
statora ext install redis --php 5.6.40
statora ext enable xdebug
statora ext disable xdebug
statora ext list                     # shows available + enabled state
statora ext list --php 8.2.15
```

---

## 7. Project Structure

Organized around three domains: `php`, `composer`, `extension`. Each is a self-contained package with its own fx module.

```
statora/
│
├── main.go                          ← argv[0] branch: dispatch or CLI
│
├── cmd/
│   ├── root.go
│   ├── php/
│   │   ├── install.go
│   │   ├── list.go
│   │   └── use.go
│   ├── composer/
│   │   ├── install.go
│   │   ├── list.go
│   │   └── use.go
│   ├── ext/
│   │   ├── install.go
│   │   ├── enable.go
│   │   ├── disable.go
│   │   └── list.go
│   └── switch.go                    ← statora switch (smart prompt)
│
├── internal/
│   │
│   ├── dispatch/                    ← zero-dep hot path
│   │   ├── dispatcher.go
│   │   └── rescache/
│   │       ├── cache.go
│   │       └── invalidate.go
│   │
│   ├── resolver/                    ← version resolution, reads .statora + global
│   │   ├── resolver.go
│   │   ├── project.go
│   │   ├── global.go
│   │   └── module.go
│   │
│   ├── compat/                      ← PHP/Composer compatibility matrix
│   │   ├── matrix.go
│   │   └── module.go
│   │
│   ├── switcher/                    ← builds SwitchPlan, runs prompt, executes
│   │   ├── switcher.go
│   │   ├── plan.go
│   │   ├── prompt.go
│   │   └── module.go
│   │
│   ├── php/                         ← PHP install pipeline
│   │   ├── plugin.go
│   │   ├── source.go
│   │   ├── pipeline.go
│   │   └── module.go
│   │
│   ├── composer/                    ← Composer install + compat-aware resolution
│   │   ├── manager.go
│   │   ├── source.go
│   │   └── module.go
│   │
│   ├── extension/                   ← ext install/enable/disable
│   │   ├── installer.go             ← binary-first, source fallback
│   │   ├── binary.go
│   │   ├── source.go                ← phpize-based compile
│   │   ├── linker.go                ← symlink enable/disable
│   │   └── module.go
│   │
│   ├── installer/                   ← shared staged pipeline runner
│   │   ├── pipeline.go
│   │   ├── stage.go
│   │   └── module.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── module.go
│   │
│   ├── logger/
│   │   ├── logger.go
│   │   └── module.go
│   │
│   └── app/
│       └── app.go                   ← fx.New wires all modules
│
└── go.mod
```

---

## 8. fx Wiring

```go
// internal/app/app.go

func New() *fx.App {
    return fx.New(
        logger.Module,
        config.Module,
        compat.Module,
        resolver.Module,
        php.Module,
        composer.Module,
        extension.Module,
        installer.Module,
        switcher.Module,
        fx.Invoke(registerCommands),
    )
}
```

`compat.Module` is provided early because both `resolver` and `switcher` depend on it:

```go
// internal/compat/module.go

var Module = fx.Module("compat",
    fx.Provide(NewMatrix),
)

// internal/switcher/module.go

var Module = fx.Module("switcher",
    fx.Provide(NewSwitcher),   // depends on: Matrix, Resolver, PHPPlugin, ComposerManager, ExtInstaller
)
```

---

## 9. Dispatch Path — Unchanged Principle, PHP-Specific Detail

The rescache now stores both `php` and `composer` resolved paths, plus the active PHP version (needed so `composer` dispatch knows which PHP to run under):

```
~/.statora/.rescache/
    php         → /home/user/.statora/runtimes/php/5.6.40/bin/php
    composer    → /home/user/.statora/composer/1.10.27/composer.phar
    .php_active → 5.6.40
```

`composer` dispatch:

```go
func dispatchComposer() {
    composerPhar := rescache.Read("composer")
    phpBin       := rescache.Read("php")

    args := append([]string{phpBin, composerPhar}, os.Args[1:]...)
    syscall.Exec(phpBin, args, os.Environ())
}
```

Composer always runs under the same PHP version as the project — no mismatch possible.

---

## 10. Installer Pipeline — Stage Order

Shared pipeline, specialized per runtime:

```
PHP install stages:
  ResolveSourceStage
  DownloadStage
  VerifyChecksumStage
  ExtractStage
  CompileStage           ← ./configure + make + make install
  InstallStage
  InvalidateCacheStage

Composer install stages:
  ResolveSourceStage     ← getcomposer.org or GitHub releases
  DownloadStage
  VerifySignatureStage   ← composer has GPG signature verification
  InstallStage           ← place .phar into ~/.statora/composer/<version>/
  InvalidateCacheStage

Extension install stages:
  ProbeStage             ← try binary first
  BinaryInstallStage     ← on success, skip remaining
  SourceDownloadStage    ← on binary fail
  PhpizeStage
  ConfigureStage
  MakeStage
  LinkStage              ← copy .so to available/
  InvalidateCacheStage
```
