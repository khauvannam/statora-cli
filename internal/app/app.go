package app

import (
	"go.uber.org/fx"

	"statora-cli/internal/compat"
	"statora-cli/internal/composer"
	"statora-cli/internal/config"
	"statora-cli/internal/extension"
	"statora-cli/internal/installer"
	"statora-cli/internal/logger"
	"statora-cli/internal/php"
	"statora-cli/internal/resolver"
	"statora-cli/internal/switcher"
)

// Options builds the fx.Option set for the full application.
func Options(debug bool) fx.Option {
	return fx.Options(
		fx.Supply(fx.Annotate(debug, fx.ResultTags(`name:"debug"`))),
		config.Module,
		logger.Module,
		compat.Module,
		resolver.Module,
		installer.Module,
		php.Module,
		composer.Module,
		extension.Module,
		switcher.Module,
		fx.NopLogger,
	)
}

// Invoke constructs the dependency graph and calls fn.
// Use this for CLI subcommands that need injected services.
func Invoke(debug bool, fn any, extra ...fx.Option) error {
	opts := []fx.Option{Options(debug), fx.Invoke(fn)}
	opts = append(opts, extra...)
	app := fx.New(opts...)
	if err := app.Err(); err != nil {
		return err
	}
	return nil
}
