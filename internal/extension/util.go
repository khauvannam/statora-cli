package extension

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mholt/archiver/v3"
)

// downloadFile downloads a URL to dest with basic retry.
func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	var (
		resp *http.Response
		err  error
	)
	for attempt := range 3 {
		resp, err = client.Get(url)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(1<<attempt) * time.Second)
	}
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// extractArchive unpacks archive into destDir.
func extractArchive(archive, destDir string) error {
	return archiver.Unarchive(archive, destDir)
}
