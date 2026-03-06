package config

import "go.uber.org/fx"

// Module provides Config to the fx dependency graph.
var Module = fx.Module("config",
	fx.Provide(func(p Params) (*Config, error) {
		return New(p.Debug)
	}),
)

// Params are the inputs needed to construct Config via fx.
type Params struct {
	fx.In
	Debug bool `name:"debug"`
}
