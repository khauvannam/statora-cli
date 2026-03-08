package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Paths holds all ~/.statora directory paths.
type Paths struct {
	Home        string // ~/.statora
	RuntimesDir string // ~/.statora/runtimes
	ComposerDir string // ~/.statora/composer
	CacheDir    string // ~/.statora/cache
	DownloadDir string // ~/.statora/cache/downloads
	BuildDir    string // ~/.statora/cache/builds
	RescacheDir string // ~/.statora/.rescache
	VersionsDir string // ~/.statora/versions
	GlobalFile  string // ~/.statora/versions/global.toml
	ConfigFile  string // ~/.statora/config.toml
	ErrorsDir   string // ~/.statora/errors
}

// ProjectConfig holds the parsed .statora project file.
type ProjectConfig struct {
	PHP        string   `mapstructure:"php"`
	Composer   string   `mapstructure:"composer"`
	Extensions []string `mapstructure:"extensions"`
}

// Config is the top-level application configuration.
type Config struct {
	Paths   Paths
	Global  GlobalConfig
	Debug   bool
}

// GlobalConfig holds ~/.statora/versions/global.toml.
type GlobalConfig struct {
	PHP      string `mapstructure:"php"`
	Composer string `mapstructure:"composer"`
}

// New builds a Config by resolving ~/.statora paths and loading config.toml.
func New(debug bool) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	statoraHome := filepath.Join(home, ".statora")
	paths := Paths{
		Home:        statoraHome,
		RuntimesDir: filepath.Join(statoraHome, "runtimes"),
		ComposerDir: filepath.Join(statoraHome, "composer"),
		CacheDir:    filepath.Join(statoraHome, "cache"),
		DownloadDir: filepath.Join(statoraHome, "cache", "downloads"),
		BuildDir:    filepath.Join(statoraHome, "cache", "builds"),
		RescacheDir: filepath.Join(statoraHome, ".rescache"),
		VersionsDir: filepath.Join(statoraHome, "versions"),
		GlobalFile:  filepath.Join(statoraHome, "versions", "global.toml"),
		ConfigFile:  filepath.Join(statoraHome, "config.toml"),
		ErrorsDir:   filepath.Join(statoraHome, "errors"),
	}

	if err := ensureDirs(paths); err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigFile(paths.ConfigFile)
	v.SetConfigType("toml")
	v.AutomaticEnv()
	// Ignore missing config file — first run has none.
	_ = v.ReadInConfig()

	cfg := &Config{Paths: paths, Debug: debug}
	return cfg, nil
}

// LoadGlobal reads ~/.statora/versions/global.toml.
func (c *Config) LoadGlobal() (GlobalConfig, error) {
	v := viper.New()
	v.SetConfigFile(c.Paths.GlobalFile)
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		return GlobalConfig{}, err
	}
	var g GlobalConfig
	_ = v.Unmarshal(&g)
	return g, nil
}

// LoadProject reads the nearest .statora file walking up from dir.
func LoadProject(dir string) (ProjectConfig, bool, error) {
	v := viper.New()
	v.SetConfigName(".statora")
	v.SetConfigType("toml")
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		return ProjectConfig{}, false, nil //nolint:nilerr
	}
	var p ProjectConfig
	if err := v.Unmarshal(&p); err != nil {
		return ProjectConfig{}, false, err
	}
	return p, true, nil
}

// WriteGlobal persists GlobalConfig to ~/.statora/versions/global.toml.
func (c *Config) WriteGlobal(g GlobalConfig) error {
	v := viper.New()
	v.SetConfigFile(c.Paths.GlobalFile)
	v.SetConfigType("toml")
	if g.PHP != "" {
		v.Set("php", g.PHP)
	}
	if g.Composer != "" {
		v.Set("composer", g.Composer)
	}
	return v.WriteConfigAs(c.Paths.GlobalFile)
}

// PHPRuntimeDir returns the install path for a given PHP version.
func (c *Config) PHPRuntimeDir(version string) string {
	return filepath.Join(c.Paths.RuntimesDir, "php", version)
}

// PHPBin returns the php binary path for a given version.
func (c *Config) PHPBin(version string) string {
	return filepath.Join(c.PHPRuntimeDir(version), "bin", "php")
}

// ComposerPhar returns the composer.phar path for a given version.
func (c *Config) ComposerPhar(version string) string {
	return filepath.Join(c.Paths.ComposerDir, version, "composer.phar")
}

// ComposerBin returns the wrapper script path for a given Composer version.
func (c *Config) ComposerBin(version string) string {
	return filepath.Join(c.Paths.ComposerDir, version, "bin", "composer")
}

// ExtAvailableDir returns the extensions/available dir for a PHP version.
func (c *Config) ExtAvailableDir(phpVersion string) string {
	return filepath.Join(c.PHPRuntimeDir(phpVersion), "lib", "extensions", "available")
}

// ExtEnabledDir returns the extensions/enabled dir for a PHP version.
func (c *Config) ExtEnabledDir(phpVersion string) string {
	return filepath.Join(c.PHPRuntimeDir(phpVersion), "lib", "extensions", "enabled")
}

func ensureDirs(p Paths) error {
	dirs := []string{
		p.RuntimesDir,
		p.ComposerDir,
		p.DownloadDir,
		p.BuildDir,
		p.RescacheDir,
		p.VersionsDir,
		p.ErrorsDir,
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}
