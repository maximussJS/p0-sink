package services

import "go.uber.org/fx"

var Module = fx.Options(
	FxStateManagerService(),
	FxStreamConfig(),
	FxMetricsService(),
	FxBlockStreamService(),
	FxBatchService(),
	FxBlockDownloaderService(),
	FxStreamCursorService(),
	FxBatchProcessorService(),
	FxBatchSenderService(),
	FxRunnerService(),
)
