package switcher

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"statora-cli/internal/composer"
	"statora-cli/internal/config"
	"statora-cli/internal/dispatch"
	"statora-cli/internal/extension"
	"statora-cli/internal/php"
	"statora-cli/internal/resolver"
)

// Switcher builds and executes SwitchPlans.
type Switcher struct {
	cfg      *config.Config
	resolver *resolver.Resolver
	php      *php.Plugin
	composer *composer.Manager
	log      *zap.Logger
}

func New(
	cfg *config.Config,
	res *resolver.Resolver,
	phpPlugin *php.Plugin,
	composerMgr *composer.Manager,
	log *zap.Logger,
) *Switcher {
	return &Switcher{
		cfg:      cfg,
		resolver: res,
		php:      phpPlugin,
		composer: composerMgr,
		log:      log,
	}
}

// BuildPlan resolves the target versions and computes what changes are needed.
func (s *Switcher) BuildPlan(dir string) (*Plan, error) {
	res, err := s.resolver.Resolve(dir)
	if err != nil {
		return nil, err
	}
	if res.Source == "none" {
		return nil, fmt.Errorf("no PHP version configured — set one with `statora php global <version>` or create a .statora file")
	}

	plan := &Plan{Resolution: res}

	// PHP action
	currentActive := dispatch.ReadCache(s.cfg, dispatch.KeyPHPActive)
	if currentActive != res.PHP {
		plan.Actions = append(plan.Actions, Action{
			Kind:    "php",
			Name:    res.PHP,
			Current: currentActive,
			Next:    res.PHP,
		})
	}

	// Composer action
	currentComposer := dispatch.ReadCache(s.cfg, dispatch.KeyComposer)
	composerPhar := s.cfg.ComposerPhar(res.Composer)
	if currentComposer != composerPhar {
		plan.Actions = append(plan.Actions, Action{
			Kind:    "composer",
			Name:    res.Composer,
			Current: currentComposer,
			Next:    composerPhar,
		})
	}

	return plan, nil
}

// Execute applies the plan: installs missing versions and updates rescache.
func (s *Switcher) Execute(plan *Plan) error {
	res := plan.Resolution

	// Ensure PHP is installed.
	if !s.php.IsInstalled(res.PHP) {
		fmt.Printf("Installing PHP %s...\n", res.PHP)
		if err := s.php.Install(res.PHP); err != nil {
			return fmt.Errorf("installing PHP %s: %w", res.PHP, err)
		}
	}

	// Ensure Composer is installed.
	if !s.composer.IsInstalled(res.Composer) {
		fmt.Printf("Installing Composer %s...\n", res.Composer)
		if err := s.composer.Install(res.Composer); err != nil {
			return fmt.Errorf("installing Composer %s: %w", res.Composer, err)
		}
	}

	// Enable declared extensions.
	if proj, found, err := config.LoadProject(currentDir()); err == nil && found {
		ins := extension.NewInstaller(s.cfg, s.log, res.PHP)
		for _, extName := range proj.Extensions {
			if !ins.IsEnabled(extName) {
				if err := ins.Enable(extName); err != nil {
					s.log.Sugar().Warnf("Could not enable extension %s: %v", extName, err)
				}
			}
		}
	}

	// Rebuild rescache.
	return dispatch.InvalidateCache(s.cfg, res)
}

func currentDir() string {
	dir, _ := os.Getwd()
	return dir
}
