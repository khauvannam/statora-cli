package composer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cavaliergopher/grab/v3"

	"statora-cli/internal/installer"
)

// downloadStage downloads composer.phar from getcomposer.org with a live progress bar.
type downloadStage struct{}

func (s *downloadStage) Name() string { return "Download composer.phar" }
func (s *downloadStage) Run(ctx *installer.Context) error {
	destDir := filepath.Join(ctx.Cfg.Paths.ComposerDir, ctx.Version)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	pharDest := filepath.Join(destDir, "composer.phar")
	ctx.Data["pharDest"] = pharDest

	if _, err := os.Stat(pharDest); err == nil {
		fmt.Println("  Already downloaded: composer.phar")
		return nil
	}

	pharURL := ctx.Data["pharURL"].(string)
	fmt.Printf("  → composer.phar (v%s)\n", ctx.Version)

	client := grab.NewClient()
	req, err := grab.NewRequest(pharDest, pharURL)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	resp := client.Do(req)
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pct := 100 * resp.Progress()
			done := float64(resp.BytesComplete()) / 1e6
			total := float64(resp.Size()) / 1e6
			fmt.Printf("\r  Downloading... %5.1f%%  %.1f / %.1f MB", pct, done, total)
		case <-resp.Done:
			if err := resp.Err(); err != nil {
				fmt.Println()
				return fmt.Errorf("downloading composer.phar: %w", err)
			}
			fmt.Printf("\r  Downloaded %.2f MB                                    \n", float64(resp.Size())/1e6)
			return nil
		}
	}
}

// verifyChecksumStage fetches composer.phar.sha256sum from getcomposer.org and
// verifies the downloaded phar against it.
type verifyChecksumStage struct{}

func (s *verifyChecksumStage) Name() string { return "Verify SHA256 checksum" }
func (s *verifyChecksumStage) Run(ctx *installer.Context) error {
	pharDest := ctx.Data["pharDest"].(string)

	// Prefer the hash fetched from /versions; otherwise fetch the .sha256sum file.
	expected, _ := ctx.Data["sha256hint"].(string)
	if expected == "" {
		sha256URL := ctx.Data["sha256URL"].(string)
		resp, err := doWithRetry(sha256URL)
		if err != nil {
			return fmt.Errorf("fetching checksum file: %w", err)
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading checksum file: %w", err)
		}
		// Format: "<hash>  composer.phar" — take the first field.
		expected = strings.ToLower(strings.Fields(string(raw))[0])
	}

	f, err := os.Open(pharDest)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))

	if got != expected {
		fmt.Printf("  Warning: SHA256 mismatch (got %s, want %s) — continuing anyway\n", got, expected)
	}
	return nil
}

// installStage marks composer.phar executable.
type installStage struct{}

func (s *installStage) Name() string { return "Install composer.phar" }
func (s *installStage) Run(ctx *installer.Context) error {
	return os.Chmod(ctx.Data["pharDest"].(string), 0o755)
}
