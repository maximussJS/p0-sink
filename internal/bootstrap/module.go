package bootstrap

import (
	"go.uber.org/fx"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/services"
)

var SinkModule = fx.Options(
	infrastructure.Module,
	services.Module,
)
