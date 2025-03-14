package infrastructure

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var Module = fx.Options(
	FxConfig(),
	FxLogger(),
	fx.WithLogger(func() fxevent.Logger {
		return fxevent.NopLogger
	}),
)
