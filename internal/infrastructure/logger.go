package infrastructure

import (
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	fx_utils "p0-sink/internal/utils/fx"
)

type ILogger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
}

type logger struct {
	*zap.Logger
}

type loggerParams struct {
	fx.In

	Config IConfig
}

func FxLogger() fx.Option {
	return fx_utils.AsProvider(newLogger, new(ILogger))
}

func newLogger(params loggerParams) ILogger {
	config := zap.NewDevelopmentConfig()

	config.Level.SetLevel(params.Config.GetLoggerLevel())

	zapLogger, err := config.Build()

	if err != nil {
		panic(fmt.Errorf("failed to create logger: %s", err))
	}

	return &logger{
		zapLogger,
	}
}
