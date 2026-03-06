package resolver

import (
	"os"
	"path/filepath"

	"statora-cli/internal/config"
	"statora-cli/internal/compat"
)

// Resolution holds the resolved PHP and Composer versions for a working directory.
type Resolution struct {
	PHP      string
	Composer string
	// Source indicates where the resolution came from: "project", "global", or "default".
	Source string
}

// Resolver resolves active PHP/Composer versions using priority:
//   project .statora > global ~/.statora/versions/global.toml
type Resolver struct {
	cfg     *config.Config
	checker *compat.Checker
}

func New(cfg *config.Config, checker *compat.Checker) *Resolver {
	return &Resolver{cfg: cfg, checker: checker}
}

// Resolve returns the active versions for the given working directory.
// Composer is inferred from the compat matrix if not explicitly set.
func (r *Resolver) Resolve(dir string) (Resolution, error) {
	// 1. Try project .statora
	proj, found, err := config.LoadProject(dir)
	if err != nil {
		return Resolution{}, err
	}
	if found && proj.PHP != "" {
		composer := proj.Composer
		if composer == "" {
			composer, err = r.checker.ResolveComposer(proj.PHP)
			if err != nil {
				return Resolution{}, err
			}
		}
		return Resolution{PHP: proj.PHP, Composer: composer, Source: "project"}, nil
	}

	// 2. Try global
	global, err := r.cfg.LoadGlobal()
	if err != nil {
		return Resolution{}, err
	}
	if global.PHP != "" {
		composer := global.Composer
		if composer == "" {
			composer, err = r.checker.ResolveComposer(global.PHP)
			if err != nil {
				return Resolution{}, err
			}
		}
		return Resolution{PHP: global.PHP, Composer: composer, Source: "global"}, nil
	}

	return Resolution{Source: "none"}, nil
}

// ResolveFromCwd calls Resolve using the current working directory.
func (r *Resolver) ResolveFromCwd() (Resolution, error) {
	dir, err := os.Getwd()
	if err != nil {
		return Resolution{}, err
	}
	return r.Resolve(dir)
}

// NearestProjectFile walks up from dir looking for a .statora file.
// Returns the directory containing it, or "" if not found.
func NearestProjectFile(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".statora")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
