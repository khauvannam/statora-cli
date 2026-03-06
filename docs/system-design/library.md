## Core Framework

| Library | Import | Purpose |
|---|---|---|
| cobra | `github.com/spf13/cobra` | CLI command tree, flags, aliases |
| fx | `go.uber.org/fx` | Dependency injection, module wiring, lifecycle |
| zap | `go.uber.org/zap` | Structured logging, management path only |
| viper | `github.com/spf13/viper` | Layered config, `.statora` + `config.toml` + env vars |

---

## Version & Semver

| Library | Import | Purpose |
|---|---|---|
| semver | `github.com/Masterminds/semver/v3` | Semver parsing, constraint matching for compat matrix |

---

## Download & HTTP

| Library | Import | Purpose |
|---|---|---|
| grab | `github.com/cavaliergopher/grab/v3` | Resumable downloads with progress, used in installer pipeline |
| go-retryablehttp | `github.com/hashicorp/go-retryablehttp` | Retry logic for PHP.net, getcomposer.org, PECL requests |

---

## Archive & Filesystem

| Library | Import | Purpose |
|---|---|---|
| archiver | `github.com/mholt/archiver/v3` | Extract `.tar.gz`, `.tar.bz2` for PHP source and PECL tarballs |

---

## Interactive CLI

| Library | Import | Purpose |
|---|---|---|
| survey | `github.com/AlecAiven/survey/v2` | Interactive prompts for `statora switch` SwitchPlan confirmation |
| tablewriter | `github.com/olekukonko/tablewriter` | Table output for `list` commands |
| color | `github.com/fatih/color` | Colored terminal output for status, warnings, errors |

---

## Process Execution

| Library | Import | Purpose |
|---|---|---|
| stdlib `syscall` | built-in | `syscall.Exec` for zero-overhead dispatch |
| stdlib `os/exec` | built-in | phpize, configure, make stages in extension compiler |

---

## Testing

| Library | Import | Purpose |
|---|---|---|
| testify | `github.com/stretchr/testify` | Assertions and mocks across all internal packages |

