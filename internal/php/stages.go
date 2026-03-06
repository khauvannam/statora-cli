package php

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	return nil
}

// downloadStage downloads the PHP tarball with resume support.
type downloadStage struct{}

func (s *downloadStage) Name() string { return "Download PHP source" }
func (s *downloadStage) Run(ctx *installer.Context) error {
	url := ctx.Data["url"].(string)
	dest := filepath.Join(ctx.Cfg.Paths.DownloadDir, filepath.Base(url))
	ctx.Data["archive"] = dest

	if _, err := os.Stat(dest); err == nil {
		ctx.Log.Sugar().Infof("Already downloaded: %s", dest)
		return nil
	}

	resp, err := grab.Get(dest, url)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	ctx.Log.Sugar().Infof("Downloaded %s (%.2f MB)", resp.Filename, float64(resp.Size())/1e6)
	return nil
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
type extractStage struct{}

func (s *extractStage) Name() string { return "Extract source" }
func (s *extractStage) Run(ctx *installer.Context) error {
	archive := ctx.Data["archive"].(string)
	buildDir := filepath.Join(ctx.Cfg.Paths.BuildDir, "php-"+ctx.Version)

	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return err
	}

	if err := archiver.Unarchive(archive, buildDir); err != nil {
		return fmt.Errorf("extracting %s: %w", archive, err)
	}
	ctx.Data["srcDir"] = buildDir
	return nil
}

// compileStage runs ./configure + make + make install.
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

	run := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Dir = extracted
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if err := run("./configure", "--prefix="+prefix, "--enable-mbstring", "--with-openssl"); err != nil {
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
