package bootstrap

import (
	"go.uber.org/fx"
	"p0-sink/internal/services"
)

func CreateApp() fx.Option {
	return fx.Options(
		SinkModule,
		fx.Invoke(func(service services.IRunnerService) {}),
	)
}
