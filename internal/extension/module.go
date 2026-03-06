package extension

import "go.uber.org/fx"

// Module is intentionally minimal — Installer is constructed per PHP version at runtime.
var Module = fx.Module("extension")
