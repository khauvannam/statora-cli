package compat

import "go.uber.org/fx"

// Module exposes compat functions — no stateful struct needed,
// so we provide a *Checker wrapper for fx-friendly injection.
var Module = fx.Module("compat",
	fx.Provide(NewChecker),
)

// Checker wraps the package-level compat functions for DI.
type Checker struct{}

func NewChecker() *Checker { return &Checker{} }

func (c *Checker) ResolveComposer(phpVersion string) (string, error) {
	return ResolveComposer(phpVersion)
}

func (c *Checker) IsCompatible(phpVersion, composerVersion string) (bool, error) {
	return IsCompatible(phpVersion, composerVersion)
}
