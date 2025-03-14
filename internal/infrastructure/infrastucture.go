package infrastructure

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	FxConfig(),
	FxLogger(),
	//fx.WithLogger(func() fxevent.Logger {
	//	return fxevent.NopLogger
	//}),
)
