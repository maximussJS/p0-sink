package infrastructure

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
	fx_utils "p0-sink/internal/utils/fx"
	"time"
)

type IConfig interface {
	GetStreamId() string

	GetStateManagerUrl() string
	GetStateManagerToken() string
	GetStateManagerTimeout() time.Duration

	GetMetricsUrl() string
	GetMetricsEnabled() bool
	GetMetricsInterval() int
	GetMetricsJobName() string
	GetMetricsInstanceName() string
	GetMetricsDefaultEnabled() bool

	GetDownloadConcurrency() int

	GetLoggerLevel() zapcore.Level
}

type config struct {
	StreamId string `mapstructure:"P0_SINK_STREAM_ID" validate:"required"`

	StateManagerUrl     string `mapstructure:"P0_SINK_STATE_MANAGER_URL" validate:"required"`
	StateManagerToken   string `mapstructure:"P0_SINK_STATE_MANAGER_TOKEN" validate:"required"`
	StateManagerTimeout int    `mapstructure:"P0_SINK_STATE_MANAGER_TIMEOUT"`

	MetricsUrl            string `mapstructure:"P0_SINK_METRICS_URL" validate:"required"`
	MetricsEnabled        bool   `mapstructure:"P0_SINK_METRICS_ENABLED"`
	MetricsInterval       int    `mapstructure:"P0_SINK_METRICS_INTERVAL"`
	MetricsJobName        string `mapstructure:"P0_SINK_METRICS_JOB_NAME"`
	MetricsInstanceName   string `mapstructure:"P0_SINK_METRICS_INSTANCE_NAME"`
	MetricsDefaultEnabled bool   `mapstructure:"P0_SINK_METRICS_ENABLE_DEFAULT_METRICS"`

	DownloadConcurrency int `mapstructure:"P0_SINK_DOWNLOADER_CONCURRENCY"`

	LoggerLevel string `mapstructure:"LOGGER_LEVEL" validate:"required"`
}

func FxConfig() fx.Option {
	return fx_utils.AsProvider(newConfig, new(IConfig))
}

func newConfig() IConfig {
	_ = godotenv.Load()

	viper.SetConfigFile(".env")
	//viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("read configuration error: %s", err))
	}

	cfg := &config{
		StateManagerTimeout: 2000,
		DownloadConcurrency: 16,
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		panic(fmt.Errorf("unmarshal configuration error: %s", err))
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Errorf("validate configuration error: %s", err))
	}

	return cfg
}

func (c *config) GetStreamId() string {
	return c.StreamId
}

func (c *config) GetStateManagerUrl() string {
	return c.StateManagerUrl
}

func (c *config) GetStateManagerToken() string {
	return c.StateManagerToken
}

func (c *config) GetStateManagerTimeout() time.Duration {
	return time.Duration(c.StateManagerTimeout) * time.Second
}

func (c *config) GetMetricsUrl() string {
	return c.MetricsUrl
}

func (c *config) GetMetricsEnabled() bool {
	return c.MetricsEnabled
}

func (c *config) GetMetricsInterval() int {
	return c.MetricsInterval
}

func (c *config) GetMetricsJobName() string {
	return c.MetricsJobName
}

func (c *config) GetMetricsInstanceName() string {
	return c.MetricsInstanceName
}

func (c *config) GetMetricsDefaultEnabled() bool {
	return c.MetricsDefaultEnabled
}

func (c *config) GetDownloadConcurrency() int {
	return c.DownloadConcurrency
}

func (c *config) GetLoggerLevel() zapcore.Level {
	switch c.LoggerLevel {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.PanicLevel
	}
}
