package infrastructure

import "go.uber.org/fx"

var Module = fx.Options(
	FxConfig(),
	FxLogger(),
)
