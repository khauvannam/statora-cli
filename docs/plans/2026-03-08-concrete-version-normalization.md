# Concrete Version Normalization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** After resolving a partial version (e.g. `8.1`, `2`) to a concrete patch version (e.g. `8.1.25`, `2.7.4`), use the concrete version for all paths — cache, build dirs, runtime dirs, extension dirs, and .statora config.

**Architecture:** Two independent changes. PHP normalizes `ctx.Version` inside `resolveSourceStage` by extracting the version from the resolved URL filename. Composer normalizes in `ResolveVersion()` by adding partial-version prefix matching before the existing exact-match lookup.

**Tech Stack:** Go 1.22+, testify (`github.com/stretchr/testify`), existing `internal/installer.Context`

---

### Task 1: PHP — extract and normalize concrete version from resolved URL

**Files:**
- Modify: `internal/php/source.go`
- Modify: `internal/php/stages.go`
- Create: `internal/php/source_test.go`

**Context:**
`ResolveSource("8.1")` hits `https://www.php.net/releases/index.php?json&version=8.1`
and returns a URL like `https://www.php.net/distributions/php-8.1.25.tar.gz`.
The concrete version `8.1.25` is embedded in the filename. We need to extract it and
overwrite `ctx.Version` so all downstream stages (extract, compile, install) use `8.1.25`.

---

**Step 1: Write the failing test for `ExtractPHPVersion`**

Add a new file `internal/php/source_test.go`:

```go
package php_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"statora-cli/internal/php"
)

func TestExtractPHPVersion(t *testing.T) {
	cases := []struct {
		url      string
		expected string
	}{
		{"https://www.php.net/distributions/php-8.1.25.tar.gz", "8.1.25"},
		{"https://www.php.net/distributions/php-8.2.15.tar.gz", "8.2.15"},
		{"https://www.php.net/distributions/php-8.3.0.tar.gz", "8.3.0"},
	}
	for _, tc := range cases {
		got := php.ExtractPHPVersion(tc.url)
		assert.Equal(t, tc.expected, got, "url: %s", tc.url)
	}
}

func TestExtractPHPVersion_Empty(t *testing.T) {
	assert.Equal(t, "", php.ExtractPHPVersion("https://example.com/unknown.tar.gz"))
	assert.Equal(t, "", php.ExtractPHPVersion(""))
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/php/... -run TestExtractPHPVersion -v
```

Expected: `FAIL — php.ExtractPHPVersion undefined`

**Step 3: Add `ExtractPHPVersion` to `internal/php/source.go`**

Append after the existing `ResolveSource` function:

```go
// ExtractPHPVersion parses the concrete version from a php.net distribution URL.
// E.g. "https://www.php.net/distributions/php-8.1.25.tar.gz" → "8.1.25".
// Returns "" if the filename does not match the expected pattern.
func ExtractPHPVersion(url string) string {
	base := filepath.Base(url)                    // "php-8.1.25.tar.gz"
	base = strings.TrimSuffix(base, ".tar.gz")    // "php-8.1.25"
	base = strings.TrimPrefix(base, "php-")       // "8.1.25"
	if !IsVersionString(base) {
		return ""
	}
	return base
}
```

Also add `"path/filepath"` to the import block in `source.go` (it doesn't have it yet).
`strings` is already imported in `source.go`? Check — if not, add it.

> Note: `IsVersionString` lives in `internal/php/plugin.go` (same package), so it's available.

**Step 4: Run test to verify it passes**

```bash
go test ./internal/php/... -run TestExtractPHPVersion -v
```

Expected: `PASS`

**Step 5: Write a test for `resolveSourceStage` normalizing `ctx.Version`**

Add to `internal/php/source_test.go`:

```go
func TestResolveSourceStage_NormalizesVersion(t *testing.T) {
	// ExtractPHPVersion should give back the patch version embedded in the URL.
	// We test the helper directly; integration with the live php.net API is not tested here.
	url := "https://www.php.net/distributions/php-8.1.25.tar.gz"
	assert.Equal(t, "8.1.25", php.ExtractPHPVersion(url))
}
```

(This is a lightweight proxy test — the stage itself hits the network, so we don't unit-test it end-to-end here. The helper is the critical piece.)

**Step 6: Update `resolveSourceStage.Run()` in `internal/php/stages.go`**

Find the existing `resolveSourceStage.Run()`:

```go
func (s *resolveSourceStage) Run(ctx *installer.Context) error {
	url, sha256, err := ResolveSource(ctx.Version)
	if err != nil {
		return err
	}
	ctx.Data["url"] = url
	ctx.Data["sha256"] = sha256
	return nil
}
```

Replace with:

```go
func (s *resolveSourceStage) Run(ctx *installer.Context) error {
	url, sha256, err := ResolveSource(ctx.Version)
	if err != nil {
		return err
	}
	ctx.Data["url"] = url
	ctx.Data["sha256"] = sha256

	// Normalize ctx.Version to the concrete patch version embedded in the URL
	// (e.g. "8.1" → "8.1.25") so all downstream stages use the full version.
	if concrete := ExtractPHPVersion(url); concrete != "" {
		ctx.Version = concrete
	}
	return nil
}
```

**Step 7: Run all PHP tests**

```bash
go test ./internal/php/... -v
```

Expected: all PASS

**Step 8: Commit**

```bash
git add internal/php/source.go internal/php/source_test.go internal/php/stages.go
git commit -m "fix(php): normalize ctx.Version to concrete patch version after source resolution"
```

---

### Task 2: Composer — add partial version prefix matching

**Files:**
- Modify: `internal/composer/source.go`
- Create: `internal/composer/source_test.go`

**Context:**
`ResolveVersion("2")` currently tries an exact-match against the stable list.
If `"2"` is not an exact entry, it returns `"2"` unchanged — so the directory becomes
`~/.statora/composer/2/composer.phar`. We need to detect partial versions (`"2"`, `"2.7"`)
and find the highest stable release with the matching major (and optional minor) prefix.

The existing `IsConstraint` check (for `>= 2.2` style inputs) is untouched.

---

**Step 1: Write the failing test for `isPartialVersion`**

Create `internal/composer/source_test.go`:

```go
package composer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"statora-cli/internal/composer"
)

func TestIsPartialVersion(t *testing.T) {
	assert.True(t, composer.IsPartialVersion("2"))
	assert.True(t, composer.IsPartialVersion("2.7"))
	assert.False(t, composer.IsPartialVersion("2.7.4"))   // already concrete
	assert.False(t, composer.IsPartialVersion(">= 2.2"))  // constraint
	assert.False(t, composer.IsPartialVersion(""))
	assert.False(t, composer.IsPartialVersion("abc"))
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/composer/... -run TestIsPartialVersion -v
```

Expected: `FAIL — composer.IsPartialVersion undefined`

**Step 3: Add `IsPartialVersion` to `internal/composer/source.go`**

Add after `IsConstraint`:

```go
// IsPartialVersion reports whether s is a partial version with 1 or 2 numeric parts
// (e.g. "2" or "2.7") rather than a full three-part version or a constraint.
func IsPartialVersion(s string) bool {
	if s == "" || IsConstraint(s) {
		return false
	}
	parts := strings.Split(s, ".")
	if len(parts) < 1 || len(parts) > 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/composer/... -run TestIsPartialVersion -v
```

Expected: `PASS`

**Step 5: Write the failing test for partial version resolution in `ResolveVersion`**

The full `ResolveVersion` hits the network, so we test the prefix-matching logic with a
helper. Add a `resolvePartialVersion` unexported helper and test it via an exported wrapper,
OR test the behavior end-to-end with a table of inputs. Since we can't mock the network
cleanly here without refactoring, we test `IsPartialVersion` (already done) and the
prefix-match logic via a small exported helper.

Add to `source_test.go`:

```go
func TestMatchesPartialPrefix(t *testing.T) {
	cases := []struct {
		partial  string
		version  string
		expected bool
	}{
		{"2", "2.7.4", true},
		{"2", "2.0.0", true},
		{"2", "1.10.0", false},
		{"2.7", "2.7.4", true},
		{"2.7", "2.7.0", true},
		{"2.7", "2.8.0", false},
		{"2.7", "2.70.0", false},
	}
	for _, tc := range cases {
		got := composer.MatchesPartialPrefix(tc.partial, tc.version)
		assert.Equal(t, tc.expected, got, "partial=%s version=%s", tc.partial, tc.version)
	}
}
```

**Step 6: Run test to verify it fails**

```bash
go test ./internal/composer/... -run TestMatchesPartialPrefix -v
```

Expected: `FAIL — composer.MatchesPartialPrefix undefined`

**Step 7: Add `MatchesPartialPrefix` to `internal/composer/source.go`**

```go
// MatchesPartialPrefix reports whether version starts with the partial prefix.
// E.g. partial="2.7", version="2.7.4" → true; partial="2.7", version="2.70.0" → false.
func MatchesPartialPrefix(partial, version string) bool {
	prefix := partial + "."
	return strings.HasPrefix(version, prefix)
}
```

**Step 8: Run test to verify it passes**

```bash
go test ./internal/composer/... -run TestMatchesPartialPrefix -v
```

Expected: `PASS`

**Step 9: Update `ResolveVersion` to use partial version matching**

Find the existing `ResolveVersion` in `internal/composer/source.go`. The current structure:

```go
func ResolveVersion(input string) (version, sha256 string, err error) {
    // ... fetch all entries ...

    if !IsConstraint(input) {
        // exact match
        for _, e := range all {
            if e.Version == input {
                return e.Version, e.SHA256, nil
            }
        }
        return input, "", nil   // ← falls through with unresolved partial
    }

    // constraint path ...
}
```

Replace the `if !IsConstraint(input)` block with:

```go
	if !IsConstraint(input) {
		// Partial version (e.g. "2" or "2.7"): find highest stable matching prefix.
		if IsPartialVersion(input) {
			var best *semver.Version
			var bestEntry versionEntry
			for _, e := range payload.Stable {
				if !MatchesPartialPrefix(input, e.Version) {
					continue
				}
				v, vErr := semver.NewVersion(e.Version)
				if vErr != nil {
					continue
				}
				if best == nil || v.GreaterThan(best) {
					best = v
					bestEntry = e
				}
			}
			if best == nil {
				return "", "", fmt.Errorf("no stable Composer release matches %q", input)
			}
			return bestEntry.Version, bestEntry.SHA256, nil
		}

		// Concrete version — look up SHA256; proceed without checksum if not listed.
		for _, e := range all {
			if e.Version == input {
				return e.Version, e.SHA256, nil
			}
		}
		return input, "", nil
	}
```

**Step 10: Run all composer tests**

```bash
go test ./internal/composer/... -v
```

Expected: all PASS

**Step 11: Commit**

```bash
git add internal/composer/source.go internal/composer/source_test.go
git commit -m "fix(composer): resolve partial versions (e.g. '2', '2.7') to concrete patch version"
```

---

### Task 3: Verify full build and test suite

**Step 1: Run all tests**

```bash
go test ./... -v
```

Expected: all PASS, no compilation errors

**Step 2: Build the binary**

```bash
go build ./...
```

Expected: exits 0, no errors

**Step 3: Commit if any fixups were needed, otherwise done**

```bash
git log --oneline -5
```

Verify the two fix commits are present and the build is clean.
