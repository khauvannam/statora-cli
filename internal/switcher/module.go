package switcher

import "go.uber.org/fx"

// Module provides *Switcher to the fx dependency graph.
var Module = fx.Module("switcher",
	fx.Provide(New),
)
