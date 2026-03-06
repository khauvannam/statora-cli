package composer

import "go.uber.org/fx"

// Module provides *Manager to the fx dependency graph.
var Module = fx.Module("composer",
	fx.Provide(NewManager),
)
