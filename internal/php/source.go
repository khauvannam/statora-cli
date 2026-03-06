package php

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const phpNetReleases = "https://www.php.net/releases/index.php?json&version=%s"

type releaseInfo struct {
	Source []struct {
		Filename string `json:"filename"`
		SHA256   string `json:"sha256"`
	} `json:"source"`
}

// ResolveSource fetches the download URL and SHA256 for a PHP version from php.net.
func ResolveSource(version string) (url, sha256 string, err error) {
	apiURL := fmt.Sprintf(phpNetReleases, version)
	client := &http.Client{Timeout: 15 * time.Second}

	var resp *http.Response
	for attempt := range 3 {
		resp, err = client.Get(apiURL)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(1<<attempt) * time.Second)
	}
	if err != nil {
		return "", "", fmt.Errorf("fetching php.net releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("php.net returned %d for version %s", resp.StatusCode, version)
	}

	var info releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", "", fmt.Errorf("parsing php.net response: %w", err)
	}

	for _, s := range info.Source {
		if len(s.Filename) > 7 && s.Filename[len(s.Filename)-7:] == ".tar.gz" {
			return fmt.Sprintf("https://www.php.net/distributions/%s", s.Filename), s.SHA256, nil
		}
	}
	return "", "", fmt.Errorf("no .tar.gz source found for PHP %s", version)
}
