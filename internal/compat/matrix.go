package compat

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Rule maps a PHP semver range to compatible Composer semver range.
type Rule struct {
	PHPConstraint      string
	ComposerConstraint string
}

// matrix defines which Composer versions are compatible with each PHP range.
var matrix = []Rule{
	{PHPConstraint: ">= 5.3.0, <= 5.6.99", ComposerConstraint: ">= 1.0.0, <= 1.10.27"},
	{PHPConstraint: ">= 7.0.0, <= 7.1.99", ComposerConstraint: ">= 1.0.0, <= 2.2.24"},
	{PHPConstraint: ">= 7.2.0, <= 7.4.99", ComposerConstraint: ">= 2.0.0, < 3.0.0"},
	{PHPConstraint: ">= 8.0.0, < 9.0.0", ComposerConstraint: ">= 2.2.0, < 3.0.0"},
}

// ResolveComposer returns the latest compatible Composer constraint for phpVersion.
// Returns an empty string if no rule matches.
func ResolveComposer(phpVersion string) (string, error) {
	php, err := semver.NewVersion(phpVersion)
	if err != nil {
		return "", fmt.Errorf("invalid PHP version %q: %w", phpVersion, err)
	}
	for _, r := range matrix {
		c, err := semver.NewConstraint(r.PHPConstraint)
		if err != nil {
			return "", fmt.Errorf("invalid PHP constraint %q: %w", r.PHPConstraint, err)
		}
		if c.Check(php) {
			return r.ComposerConstraint, nil
		}
	}
	return "", nil
}

// IsCompatible reports whether the given PHP and Composer versions are compatible.
func IsCompatible(phpVersion, composerVersion string) (bool, error) {
	composerConstraint, err := ResolveComposer(phpVersion)
	if err != nil {
		return false, err
	}
	if composerConstraint == "" {
		return false, fmt.Errorf("no compat rule for PHP %s", phpVersion)
	}
	composer, err := semver.NewVersion(composerVersion)
	if err != nil {
		return false, fmt.Errorf("invalid Composer version %q: %w", composerVersion, err)
	}
	c, err := semver.NewConstraint(composerConstraint)
	if err != nil {
		return false, err
	}
	return c.Check(composer), nil
}

// Rules returns the full compatibility matrix (read-only copy).
func Rules() []Rule {
	out := make([]Rule, len(matrix))
	copy(out, matrix)
	return out
}
