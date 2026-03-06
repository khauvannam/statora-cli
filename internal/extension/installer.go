package extension

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"

	"statora-cli/internal/config"
)

// Installer manages PHP extensions for a specific PHP version.
type Installer struct {
	cfg        *config.Config
	log        *zap.Logger
	phpVersion string
}

func NewInstaller(cfg *config.Config, log *zap.Logger, phpVersion string) *Installer {
	return &Installer{cfg: cfg, log: log, phpVersion: phpVersion}
}

// Install installs an extension: tries binary first, falls back to source.
func (i *Installer) Install(name string) error {
	if i.IsAvailable(name) {
		fmt.Printf("Extension %s is already available.\n", name)
		return nil
	}

	// Binary-first: PECL
	if err := i.installFromPECL(name); err == nil {
		i.log.Sugar().Infof("Installed %s from PECL", name)
		return nil
	}

	// Fallback: phpize from source
	return i.installFromSource(name)
}

// installFromPECL attempts to install using the pecl command.
func (i *Installer) installFromPECL(name string) error {
	peclBin := filepath.Join(i.cfg.PHPRuntimeDir(i.phpVersion), "bin", "pecl")
	if _, err := os.Stat(peclBin); err != nil {
		return fmt.Errorf("pecl not found at %s", peclBin)
	}

	phpIni := filepath.Join(i.cfg.PHPRuntimeDir(i.phpVersion), "lib", "php.ini")
	cmd := exec.Command(peclBin, "install", name)
	cmd.Env = append(os.Environ(), "PHP_INI_DIR="+filepath.Dir(phpIni))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pecl install %s: %w", name, err)
	}

	// Move the .so into our available dir.
	return i.linkBuilt(name)
}

// installFromSource downloads and compiles the extension with phpize.
func (i *Installer) installFromSource(name string) error {
	phpizeBin := filepath.Join(i.cfg.PHPRuntimeDir(i.phpVersion), "bin", "phpize")
	if _, err := os.Stat(phpizeBin); err != nil {
		return fmt.Errorf("phpize not found at %s — is PHP %s compiled?", phpizeBin, i.phpVersion)
	}

	// Download from PECL source tarball.
	srcURL := fmt.Sprintf("https://pecl.php.net/get/%s", name)
	srcDir := filepath.Join(i.cfg.Paths.BuildDir, "ext-"+name)
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		return err
	}

	archivePath := filepath.Join(i.cfg.Paths.DownloadDir, name+".tgz")
	if err := downloadFile(srcURL, archivePath); err != nil {
		return fmt.Errorf("downloading %s source: %w", name, err)
	}

	if err := extractArchive(archivePath, srcDir); err != nil {
		return fmt.Errorf("extracting %s: %w", name, err)
	}

	// Find extracted subdir.
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("nothing extracted for %s", name)
	}
	extSrc := filepath.Join(srcDir, entries[0].Name())

	run := func(binName string, args ...string) error {
		cmd := exec.Command(binName, args...)
		cmd.Dir = extSrc
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if err := run(phpizeBin); err != nil {
		return fmt.Errorf("phpize: %w", err)
	}
	phpConfig := filepath.Join(i.cfg.PHPRuntimeDir(i.phpVersion), "bin", "php-config")
	if err := run("./configure", "--with-php-config="+phpConfig); err != nil {
		return fmt.Errorf("configure: %w", err)
	}
	if err := run("make", "-j4"); err != nil {
		return fmt.Errorf("make: %w", err)
	}

	return i.linkBuilt(name)
}

// linkBuilt finds the compiled .so and links it into available/.
func (i *Installer) linkBuilt(name string) error {
	availDir := i.cfg.ExtAvailableDir(i.phpVersion)
	if err := os.MkdirAll(availDir, 0o755); err != nil {
		return err
	}

	// The .so may be in standard extension dir or build dir.
	extDir, err := phpExtensionDir(i.cfg.PHPRuntimeDir(i.phpVersion))
	if err != nil {
		return err
	}

	soName := name + ".so"
	soPath := filepath.Join(extDir, soName)
	if _, err := os.Stat(soPath); os.IsNotExist(err) {
		return fmt.Errorf("compiled extension not found at %s", soPath)
	}

	dest := filepath.Join(availDir, soName)
	return os.Rename(soPath, dest)
}

// phpExtensionDir returns the PHP extension dir using php-config.
func phpExtensionDir(runtimeDir string) (string, error) {
	phpConfig := filepath.Join(runtimeDir, "bin", "php-config")
	out, err := exec.Command(phpConfig, "--extension-dir").Output()
	if err != nil {
		return "", fmt.Errorf("php-config --extension-dir: %w", err)
	}
	dir := string(out)
	if len(dir) > 0 && dir[len(dir)-1] == '\n' {
		dir = dir[:len(dir)-1]
	}
	return dir, nil
}
