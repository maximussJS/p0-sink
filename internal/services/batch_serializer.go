package services

import (
	"context"
	"go.uber.org/fx"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
)

type IBatchSerializerService interface {
	GetReadStream(
		ctx context.Context,
		inputChannel types.BatchReadChannel,
		errorChannel types.ErrorChannel,
	) types.SerializedBatchReadChannel
}

type batchSerializerServiceParams struct {
	fx.In
}

type batchSerializerService struct {
}

func FxBatchSerializerService() fx.Option {
	return fx_utils.AsProvider(newBatchSerializer, new(IBatchSerializerService))
}

func newBatchSerializer(params batchSerializerServiceParams) IBatchSerializerService {
	return &batchSerializerService{}
}

func (s *batchSerializerService) GetReadStream(
	ctx context.Context,
	inputChannel types.BatchReadChannel,
	_ types.ErrorChannel,
) types.SerializedBatchReadChannel {
	outputChannel := make(types.SerializedBatchChannel)

	var batch *types.Batch
	var ok bool

	go func() {
		defer close(outputChannel)
		for {
			select {
			case batch, ok = <-inputChannel:
				{
					if !ok {
						return // input channel closed
					}
					outputChannel <- s.serialize(batch)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputChannel
}

func (s *batchSerializerService) serialize(batch *types.Batch) *types.SerializedBatch {
	return batch
}
