package composer

import (
	"fmt"
	"net/http"
	"time"
)

const (
	composerDownloadURL = "https://getcomposer.org/download/%s/composer.phar"
	composerSigURL      = "https://getcomposer.org/download/%s/composer.phar.asc"
)

// URLs returns the download and GPG signature URLs for a Composer version.
func URLs(version string) (pharURL, sigURL string) {
	return fmt.Sprintf(composerDownloadURL, version),
		fmt.Sprintf(composerSigURL, version)
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
