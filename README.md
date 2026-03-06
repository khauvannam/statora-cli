# Statora

PHP version manager for developers who work across multiple projects with different PHP, Composer, and extension requirements.

---

## Install

```bash
make build
# or
go install .
```

---

## Quick Start

Create a `.statora` file in your project root:

```toml
php = "8.2.15"
composer = "2.7.1"        # optional — auto-resolved if absent
extensions = ["redis", "xdebug"]
```

Then switch to those versions:

```bash
statora switch
```

Statora detects the mismatch, shows you the plan, and applies it on confirmation.

---

## PHP

```bash
# Install a version (compiles from source)
statora php install 8.2.15

# Remove a version
statora php uninstall 8.2.15

# List installed versions
statora php list

# Set the global default
statora php global 8.2.15

# Set version for current project (.statora)
statora php local 8.2.15

# Show active version
statora php current

# Show binary path for a version
statora php which 8.2.15

# Rebuild dispatch cache after manual changes
statora php rehash
```

---

## Composer

```bash
# Install a version
statora composer install 2.7.1

# Remove a version
statora composer uninstall 2.7.1

# List installed versions
statora composer list

# Set the global default
statora composer global 2.7.1

# Set version for current project (.statora)
statora composer local 2.7.1

# Show active version
statora composer current

# Show which Composer versions are compatible with a PHP version
statora composer compat 5.6.40
# PHP 5.6.40 → Composer >= 1.0.0, <= 1.10.27
```

---

## Extensions

Extensions are scoped per PHP version. Enable/disable is a symlink operation — fast and inspectable.

```bash
# Install an extension for the active PHP version
statora ext install redis

# Uninstall
statora ext uninstall redis

# Enable / disable
statora ext enable xdebug
statora ext disable xdebug

# List extensions and their status
statora ext list

# Show info for a specific extension
statora ext info redis
```

Install tries PECL binary first, falls back to compiling from source with `phpize`.

---

## Switch

```bash
statora switch
```

Reads `.statora`, compares against active versions, builds a plan, and shows an interactive prompt:

```
  Switch Plan
  ──────────
  PHP       8.2.15 → 5.6.40
  Composer  → 1.10.27
  Enable    redis

  Apply? [y/N]
```

---

## Dispatch

Statora doubles as a `php` and `composer` shim. Create symlinks:

```bash
ln -s $(which statora) /usr/local/bin/php
ln -s $(which statora) /usr/local/bin/composer
```

When invoked as `php` or `composer`, statora reads the rescache and exec's the correct binary with zero overhead (no fx startup, one file read).

---

## Global Flags

```bash
statora --debug [command]   # enable debug logging
```

---

## Development

```bash
make test          # run all tests
make dev           # watch + rebuild on file change (requires watchexec)
make lint          # go vet
```
