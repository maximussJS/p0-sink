package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	pb "p0-sink/proto"
)

type IBatchService interface {
	Channel(
		ctx context.Context,
		inputChannel types.BlockReadonlyChannel,
		errorChannel types.ErrorChannel,
	) types.BatchReadonlyChannel
}

type blockServiceParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
}

type batchService struct {
	logger              infrastructure.ILogger
	streamConfig        IStreamConfig
	blockOrderValidator *lib.BlockOrderValidator
}

func FxBatchService() fx.Option {
	return fx_utils.AsProvider(newBlockService, new(IBatchService))
}

func newBlockService(lc fx.Lifecycle, params blockServiceParams) IBatchService {
	s := &batchService{
		logger:       params.Logger,
		streamConfig: params.StreamConfig,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			reorgAction := params.StreamConfig.ReorgAction()

			s.blockOrderValidator = lib.NewBlockOrderValidator(reorgAction)

			s.logger.Info(fmt.Sprintf("block validator reorg action: %s", reorgAction.String()))
			return nil
		},
	})

	return s
}

func (s *batchService) Channel(
	ctx context.Context,
	inputChannel types.BlockReadonlyChannel,
	errorChannel types.ErrorChannel,
) types.BatchReadonlyChannel {
	outputChannel := make(types.BatchChannel, 1)

	var block *pb.BlockWrapper
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

					if err := s.blockOrderValidator.Validate(block); err != nil {
						errorChannel <- err
						return
					}

					if block.IsHead && s.streamConfig.BatchSize() > 1 && s.streamConfig.ElasticBatchEnabled() {
						s.streamConfig.UpdateBatchSize(1)
						s.logger.Info("stream has reached realtime, scaling down batch size to 1")
					}

					if s.shouldReleaseBatchAfterBlock(block, batch) {
						outputChannel <- batch
						batch = types.NewBatch()
					}

					if s.shouldPushToBatch() {
						err := batch.PushBlock(block)
						if err != nil {
							errorChannel <- err
							return
						}
					}

					shouldReleaseBatchBeforeBlock, err := s.shouldReleaseBatchBeforeBlock(block, batch)

					if err != nil {
						errorChannel <- err
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

func (s *batchService) shouldReleaseBatchBeforeBlock(block *pb.BlockWrapper, batch *types.Batch) (bool, error) {
	batchDirection, err := batch.GetDirection()

	if err != nil {
		return false, err
	}

	return batch.HasBlocks() && batchDirection.String() != block.Direction.String(), nil
}

func (s *batchService) shouldPushToBatch() bool {
	return true
}

func (s *batchService) shouldReleaseBatchAfterBlock(_ *pb.BlockWrapper, batch *types.Batch) bool {
	return batch.Len() >= s.streamConfig.BatchSize()
}
