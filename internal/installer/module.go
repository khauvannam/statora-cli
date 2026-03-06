package installer

import "go.uber.org/fx"

// Module provides Pipeline constructor to the fx dependency graph.
// Callers (php, composer, ext) build their own Pipeline using New().
var Module = fx.Module("installer")
