package services

import (
	"context"
	"go.uber.org/fx"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
)

type IBatchCollectorService interface {
	GetReadChannel(
		ctx context.Context,
		inputChannel types.DownloadedBlockReadChannel,
		errorChannel types.ErrorChannel,
	) types.BatchReadChannel
}

type batchCollectorServiceParams struct {
	fx.In

	StreamConfig IStreamConfig
}

type batchCollectorService struct {
	streamConfig IStreamConfig
}

func FxBatchCollectorService() fx.Option {
	return fx_utils.AsProvider(newBatchCollector, new(IBatchCollectorService))
}

func newBatchCollector(params batchCollectorServiceParams) IBatchCollectorService {
	return &batchCollectorService{
		streamConfig: params.StreamConfig,
	}
}

func (s *batchCollectorService) GetReadChannel(
	ctx context.Context,
	inputChannel types.DownloadedBlockReadChannel,
	errorChanel types.ErrorChannel,
) types.BatchReadChannel {
	outputChannel := make(types.BatchChannel)

	var block *types.DownloadedBlock
	var ok bool

	go func() {
		defer close(outputChannel)
		batch := types.NewBatch()

		defer func() {
			if batch.HasBlocks() {
				outputChannel <- batch
			}
		}()

		for {
			select {
			case block, ok = <-inputChannel:
				{
					if !ok {
						return // input channel closed
					}

					if s.shouldReleaseBatchAfterBlock(block, batch) {
						outputChannel <- batch
						batch = types.NewBatch()
					}

					if s.shouldPushToBatch() {
						err := batch.PushBlock(block)
						if err != nil {
							errorChanel <- err
							return
						}
					}

					shouldReleaseBatchBeforeBlock, err := s.shouldReleaseBatchBeforeBlock(block, batch)

					if err != nil {
						errorChanel <- err
						return
					}

					if shouldReleaseBatchBeforeBlock {
						outputChannel <- batch
						batch = types.NewBatch()
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputChannel
}

func (s *batchCollectorService) shouldReleaseBatchBeforeBlock(block *types.DownloadedBlock, batch *types.Batch) (bool, error) {
	batchDirection, err := batch.GetDirection()

	if err != nil {
		return false, err
	}

	return batch.HasBlocks() && batchDirection.String() != block.Direction.String(), nil
}

func (s *batchCollectorService) shouldPushToBatch() bool {
	return true
}

func (s *batchCollectorService) shouldReleaseBatchAfterBlock(_ *types.DownloadedBlock, batch *types.Batch) bool {
	return batch.Len() >= s.streamConfig.BatchSize()
}
