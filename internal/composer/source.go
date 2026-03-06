package composer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	versionsURL    = "https://getcomposer.org/versions"
	downloadURLTpl = "https://getcomposer.org/download/%s/composer.phar"
	sha256URLTpl   = "https://getcomposer.org/download/%s/composer.phar.sha256sum"
)

type versionEntry struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256sum"`
}

type versionsPayload struct {
	Stable   []versionEntry `json:"stable"`
	Preview  []versionEntry `json:"preview"`
	Snapshot []versionEntry `json:"snapshot"`
}

// IsConstraint reports whether s is a semver constraint rather than a concrete version.
func IsConstraint(s string) bool {
	return strings.ContainsAny(s, " <>~^*|")
}

// ResolveVersion resolves a concrete version string or semver constraint to a
// concrete version tag and its SHA256 hash using getcomposer.org/versions.
// SHA256 may be empty if the entry is not listed (old patch versions, etc.).
func ResolveVersion(input string) (version, sha256 string, err error) {
	resp, err := doWithRetry(versionsURL)
	if err != nil {
		return "", "", fmt.Errorf("fetching getcomposer.org/versions: %w", err)
	}
	defer resp.Body.Close()

	var payload versionsPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", fmt.Errorf("parsing versions response: %w", err)
	}

	// All candidates: stable first, then preview.
	all := append(payload.Stable, payload.Preview...)

	if !IsConstraint(input) {
		// Concrete version — look up SHA256; proceed without checksum if not listed.
		for _, e := range all {
			if e.Version == input {
				return e.Version, e.SHA256, nil
			}
		}
		return input, "", nil
	}

	// Constraint — find the highest stable release that satisfies it.
	c, err := semver.NewConstraint(input)
	if err != nil {
		return "", "", fmt.Errorf("invalid constraint %q: %w", input, err)
	}

	var best *semver.Version
	var bestEntry versionEntry
	for _, e := range payload.Stable {
		v, vErr := semver.NewVersion(e.Version)
		if vErr != nil {
			continue
		}
		if c.Check(v) && (best == nil || v.GreaterThan(best)) {
			best = v
			bestEntry = e
		}
	}

	if best == nil {
		return "", "", fmt.Errorf("no stable Composer release matches %q", input)
	}
	return bestEntry.Version, bestEntry.SHA256, nil
}

// DownloadURL returns the getcomposer.org download URL for a concrete version.
func DownloadURL(version string) string {
	return fmt.Sprintf(downloadURLTpl, version)
}

// SHA256URL returns the .sha256sum file URL for a concrete version.
func SHA256URL(version string) string {
	return fmt.Sprintf(sha256URLTpl, version)
}

// doWithRetry performs an HTTP GET with exponential backoff (max 3 attempts).
func doWithRetry(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = client.Get(url)
		if err == nil && resp.StatusCode == 200 {
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(1<<attempt) * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
}
