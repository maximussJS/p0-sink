package services

import (
	"context"
	"go.uber.org/fx"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	pb "p0-sink/proto"
)

type IBatchSizeTrackerService interface {
	GetReadChannel(
		ctx context.Context,
		inputChannel types.BlockWrapperReadChannel,
		errorChannel types.ErrorChannel,
	) types.BlockWrapperReadChannel
}

type batchSizeTrackerServiceParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
}

type batchSizeTrackerService struct {
	logger       infrastructure.ILogger
	streamConfig IStreamConfig
}

func FxBatchSizeTrackerService() fx.Option {
	return fx_utils.AsProvider(newBatchSizeTracker, new(IBatchSizeTrackerService))
}

func newBatchSizeTracker(params batchSizeTrackerServiceParams) IBatchSizeTrackerService {
	return &batchSizeTrackerService{
		logger:       params.Logger,
		streamConfig: params.StreamConfig,
	}
}

func (s *batchSizeTrackerService) GetReadChannel(
	ctx context.Context,
	inputChannel types.BlockWrapperReadChannel,
	_ types.ErrorChannel,
) types.BlockWrapperReadChannel {
	outputChannel := make(types.BlockWrapperChannel, s.streamConfig.ChannelSize())

	var block *pb.BlockWrapper
	var ok bool

	go func() {
		defer close(outputChannel)
		for {
			select {
			case block, ok = <-inputChannel:
				{
					if !ok {
						return // input channel closed
					}

					if block.IsHead && s.streamConfig.BatchSize() > 1 && s.streamConfig.ElasticBatchEnabled() {
						s.streamConfig.UpdateBatchSize(1)
						s.logger.Info("stream has reached realtime, scaling down batch size to 1")
					}

					outputChannel <- block
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputChannel
}
