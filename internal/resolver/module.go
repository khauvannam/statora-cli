package resolver

import "go.uber.org/fx"

// Module provides *Resolver to the fx dependency graph.
var Module = fx.Module("resolver",
	fx.Provide(New),
)
