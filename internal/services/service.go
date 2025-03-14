package services

import "go.uber.org/fx"

var Module = fx.Options(
	FxStateManagerService(),
	FxStreamConfig(),
	FxBlockStreamService(),
	FxBatchService(),
	FxStreamCursorService(),
	FxBatchProcessorService(),
	FxBatchSenderService(),
	FxRunnerService(),
)
