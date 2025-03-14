package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib/serializers"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	"time"
)

type IBatchSerializerService interface {
	GetReadChannel(
		ctx context.Context,
		inputChannel types.BatchReadChannel,
		errorChannel types.ErrorChannel,
	) types.SerializedBatchReadChannel
}

type batchSerializerServiceParams struct {
	fx.In

	StreamConfig IStreamConfig
	Logger       infrastructure.ILogger
}

type batchSerializerService struct {
	serializer   serializers.ISerializer
	streamConfig IStreamConfig
	logger       infrastructure.ILogger
	encoding     string
}

func FxBatchSerializerService() fx.Option {
	return fx_utils.AsProvider(newBatchSerializer, new(IBatchSerializerService))
}

func newBatchSerializer(lc fx.Lifecycle, params batchSerializerServiceParams) IBatchSerializerService {
	bs := &batchSerializerService{
		streamConfig: params.StreamConfig,
		logger:       params.Logger,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			bs.serializer = params.StreamConfig.Serializer()

			_, outputCompressor := params.StreamConfig.Compressors()

			bs.encoding = outputCompressor.EncodingType()

			return nil
		},
	})

	return bs
}

func (s *batchSerializerService) GetReadChannel(
	ctx context.Context,
	inputChannel types.BatchReadChannel,
	errorChannel types.ErrorChannel,
) types.SerializedBatchReadChannel {
	outputChannel := make(types.SerializedBatchChannel, s.streamConfig.ChannelSize())

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

					s.logger.Info(fmt.Sprintf("Batch %s collected in %s", batch.String(), batch.TimeElapsed()))

					serializedBatch, err := s.serialize(batch)

					if err != nil {
						errorChannel <- err
						return
					}

					outputChannel <- serializedBatch
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputChannel
}

func (s *batchSerializerService) serialize(batch *types.Batch) (*types.SerializedBatch, error) {
	start := time.Now()
	serialized, err := s.serializer.Serialize(batch, s.encoding)

	if err != nil {
		return nil, err
	}

	s.logger.Info(fmt.Sprintf("Batch %s serialized in %s", batch.String(), time.Since(start)))

	return serialized, nil
}
