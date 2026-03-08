package resolver

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

// NormalizeInstalled returns the highest installed concrete version that matches
// the given partial or exact version string.
//
// Examples:
//
//	NormalizeInstalled("8.1", []string{"8.1.20", "8.1.25", "8.2.15"}) → "8.1.25"
//	NormalizeInstalled("8.1.25", []string{"8.1.25"})                   → "8.1.25"
//	NormalizeInstalled("8.1", []string{"8.2.15"})                      → ""
func NormalizeInstalled(version string, installed []string) string {
	if len(installed) == 0 || version == "" {
		return ""
	}

	// Exact match first.
	for _, v := range installed {
		if v == version {
			return v
		}
	}

	// Partial match: version is a prefix like "8" or "8.1".
	prefix := version + "."
	var matches []*semver.Version
	var matchStrs []string
	for _, v := range installed {
		if strings.HasPrefix(v, prefix) {
			sv, err := semver.NewVersion(v)
			if err == nil {
				matches = append(matches, sv)
				matchStrs = append(matchStrs, v)
			}
		}
	}
	if len(matches) == 0 {
		return ""
	}

	// Find the highest semver match.
	best := matches[0]
	bestStr := matchStrs[0]
	for i := 1; i < len(matches); i++ {
		if matches[i].GreaterThan(best) {
			best = matches[i]
			bestStr = matchStrs[i]
		}
	}
	return bestStr
}
