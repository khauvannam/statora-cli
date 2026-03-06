package composer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"statora-cli/internal/installer"
)

// resolveSourceStage sets download URLs in the context.
type resolveSourceStage struct{}

func (s *resolveSourceStage) Name() string { return "Resolve Composer source" }
func (s *resolveSourceStage) Run(ctx *installer.Context) error {
	pharURL, sigURL := URLs(ctx.Version)
	ctx.Data["pharURL"] = pharURL
	ctx.Data["sigURL"] = sigURL
	return nil
}

// downloadStage downloads composer.phar and its GPG signature.
type downloadStage struct{}

func (s *downloadStage) Name() string { return "Download composer.phar" }
func (s *downloadStage) Run(ctx *installer.Context) error {
	destDir := filepath.Join(ctx.Cfg.Paths.ComposerDir, ctx.Version)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	pharDest := filepath.Join(destDir, "composer.phar")
	sigDest := filepath.Join(destDir, "composer.phar.asc")
	ctx.Data["pharDest"] = pharDest
	ctx.Data["sigDest"] = sigDest

	for _, pair := range []struct{ url, dest string }{
		{ctx.Data["pharURL"].(string), pharDest},
		{ctx.Data["sigURL"].(string), sigDest},
	} {
		if _, err := os.Stat(pair.dest); err == nil {
			continue
		}
		resp, err := doWithRetry(pair.url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		f, err := os.Create(pair.dest)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

// verifySignatureStage runs GPG verification on the downloaded phar.
// If GPG is unavailable, the stage is skipped with a warning.
type verifySignatureStage struct{}

func (s *verifySignatureStage) Name() string { return "Verify GPG signature" }
func (s *verifySignatureStage) Run(ctx *installer.Context) error {
	pharDest := ctx.Data["pharDest"].(string)
	sigDest := ctx.Data["sigDest"].(string)

	if _, err := exec.LookPath("gpg"); err != nil {
		ctx.Log.Sugar().Warn("gpg not found — skipping signature verification")
		return nil
	}

	cmd := exec.Command("gpg", "--verify", sigDest, pharDest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GPG verification failed: %w", err)
	}
	return nil
}

// installStage marks composer.phar executable.
type installStage struct{}

func (s *installStage) Name() string { return "Install composer.phar" }
func (s *installStage) Run(ctx *installer.Context) error {
	pharDest := ctx.Data["pharDest"].(string)
	return os.Chmod(pharDest, 0o755)
}
