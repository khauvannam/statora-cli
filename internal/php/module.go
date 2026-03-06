package php

import "go.uber.org/fx"

// Module provides *Plugin to the fx dependency graph.
var Module = fx.Module("php",
	fx.Provide(NewPlugin),
)
