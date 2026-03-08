package php

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/mholt/archiver/v3"

	"statora-cli/internal/installer"
)

// resolveSourceStage fetches the download URL and checksum from php.net.
type resolveSourceStage struct{}

func (s *resolveSourceStage) Name() string { return "Resolve PHP source" }
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

// downloadStage downloads the PHP tarball with a live progress bar.
type downloadStage struct{}

func (s *downloadStage) Name() string { return "Download PHP source" }
func (s *downloadStage) Run(ctx *installer.Context) error {
	url := ctx.Data["url"].(string)
	dest := filepath.Join(ctx.Cfg.Paths.DownloadDir, filepath.Base(url))
	ctx.Data["archive"] = dest

	if _, err := os.Stat(dest); err == nil {
		ctx.Log.Sugar().Infof("Already downloaded: %s", dest)
		fmt.Printf("  Already downloaded: %s\n", filepath.Base(dest))
		return nil
	}

	fmt.Printf("  → %s\n", filepath.Base(url))

	client := grab.NewClient()
	req, err := grab.NewRequest(dest, url)
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
				return fmt.Errorf("downloading %s: %w", url, err)
			}
			fmt.Printf("\r  Downloaded %.1f MB                                    \n", float64(resp.Size())/1e6)
			return nil
		}
	}
}

// verifyChecksumStage validates the SHA256 of the downloaded archive.
type verifyChecksumStage struct{}

func (s *verifyChecksumStage) Name() string { return "Verify checksum" }
func (s *verifyChecksumStage) Run(ctx *installer.Context) error {
	archive := ctx.Data["archive"].(string)
	expected := ctx.Data["sha256"].(string)

	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expected) {
		return fmt.Errorf("SHA256 mismatch: got %s, want %s", got, expected)
	}
	return nil
}

// extractStage unpacks the tarball into the build cache.
// If the directory already contains a valid extraction (configure script present),
// it is reused. Any partial extraction is cleaned before re-extracting.
type extractStage struct{}

func (s *extractStage) Name() string { return "Extract source" }
func (s *extractStage) Run(ctx *installer.Context) error {
	archive := ctx.Data["archive"].(string)
	buildDir := filepath.Join(ctx.Cfg.Paths.BuildDir, "php-"+ctx.Version)

	if isExtracted(buildDir) {
		fmt.Println("  Already extracted, reusing build directory.")
		ctx.Data["srcDir"] = buildDir
		return nil
	}

	// Remove any partial or failed previous extraction.
	_ = os.RemoveAll(buildDir)
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return err
	}

	if err := archiver.Unarchive(archive, buildDir); err != nil {
		_ = os.RemoveAll(buildDir) // clean up partial extraction
		return fmt.Errorf("extracting %s: %w", archive, err)
	}
	ctx.Data["srcDir"] = buildDir
	return nil
}

// isExtracted returns true when buildDir contains a subdirectory with a configure script.
func isExtracted(buildDir string) bool {
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(buildDir, e.Name(), "configure")); err == nil {
				return true
			}
		}
	}
	return false
}

// compileStage runs ./configure + make + make install.
// Stderr is captured and stored in ctx.CapturedOutput for error logs.
type compileStage struct{}

func (s *compileStage) Name() string { return "Compile PHP" }
func (s *compileStage) Run(ctx *installer.Context) error {
	srcDir := ctx.Data["srcDir"].(string)
	prefix := ctx.Cfg.PHPRuntimeDir(ctx.Version)

	// Find the actual extracted directory (e.g. php-8.2.15/)
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	extracted := srcDir
	for _, e := range entries {
		if e.IsDir() {
			extracted = filepath.Join(srcDir, e.Name())
			break
		}
	}

	var outputBuf bytes.Buffer

	run := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Dir = extracted
		cmd.Stdout = io.MultiWriter(os.Stdout, &outputBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &outputBuf)
		if err := cmd.Run(); err != nil {
			ctx.CapturedOutput = outputBuf.String()
			// Remove the build dir so the next run starts with a clean extraction.
			_ = os.RemoveAll(ctx.Data["srcDir"].(string))
			return err
		}
		return nil
	}

	configureArgs := []string{
		"--prefix=" + prefix,
		"--enable-mbstring",
		"--with-openssl",
		"--disable-phpdbg",  // phpdbg has linker issues on macOS ARM64
		"--without-pear",    // skip PEAR download during install
	}
	if runtime.GOOS == "darwin" {
		if p := darwinIconvPrefix(); p != "" {
			configureArgs = append(configureArgs, "--with-iconv="+p)
		}
	}

	if err := run("./configure", configureArgs...); err != nil {
		return fmt.Errorf("configure: %w", err)
	}
	if err := run("make", "-j4"); err != nil {
		return fmt.Errorf("make: %w", err)
	}
	if err := run("make", "install"); err != nil {
		return fmt.Errorf("make install: %w", err)
	}
	return nil
}

// darwinIconvPrefix returns the iconv prefix directory on macOS.
// It tries multiple locations in order and verifies iconv.h exists before returning.
func darwinIconvPrefix() string {
	hasHeader := func(prefix string) bool {
		_, err := os.Stat(filepath.Join(prefix, "include", "iconv.h"))
		return err == nil
	}

	// 1. Homebrew-managed libiconv (brew --prefix libiconv).
	if out, err := exec.Command("brew", "--prefix", "libiconv").Output(); err == nil {
		if p := strings.TrimSpace(string(out)); p != "" && hasHeader(p) {
			return p
		}
	}

	// 2. Hardcoded Homebrew opt paths (avoids PATH issues in exec env).
	for _, p := range []string{
		"/opt/homebrew/opt/libiconv",  // Apple Silicon
		"/usr/local/opt/libiconv",     // Intel
	} {
		if hasHeader(p) {
			return p
		}
	}

	// 3. Xcode / Command Line Tools SDK.
	if out, err := exec.Command("xcrun", "--show-sdk-path").Output(); err == nil {
		if p := filepath.Join(strings.TrimSpace(string(out)), "usr"); hasHeader(p) {
			return p
		}
	}

	return ""
}

// installStage ensures extension directories exist after compilation.
type installStage struct{}

func (s *installStage) Name() string { return "Finalize install" }
func (s *installStage) Run(ctx *installer.Context) error {
	for _, dir := range []string{
		ctx.Cfg.ExtAvailableDir(ctx.Version),
		ctx.Cfg.ExtEnabledDir(ctx.Version),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
