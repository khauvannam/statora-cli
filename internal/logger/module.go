package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides *zap.Logger to the fx dependency graph.
var Module = fx.Module("logger",
	fx.Provide(func(p Params) (*zap.Logger, error) {
		return New(p.Debug)
	}),
)

// Params are the inputs needed to construct Logger via fx.
type Params struct {
	fx.In
	Debug bool `name:"debug"`
}
