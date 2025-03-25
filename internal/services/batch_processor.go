package services

import (
	"context"
	"fmt"
	"github.com/alitto/pond/v2"
	"go.uber.org/fx"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib"
	"p0-sink/internal/lib/serializers"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	"time"
)

type IBatchProcessorService interface {
	Channel(
		ctx context.Context,
		inputChannel types.BatchReadonlyChannel,
		errorChannel types.ErrorChannel,
	) types.ProcessedBatchReadonlyChannel
}

type batchProcessorServiceParams struct {
	fx.In

	StreamConfig    IStreamConfig
	BlockDownloader IBlockDownloaderService
	Metrics         IMetricsService
	Config          infrastructure.IConfig
	Logger          infrastructure.ILogger
}

type batchProcessorService struct {
	serializer      serializers.ISerializer
	streamConfig    IStreamConfig
	metrics         IMetricsService
	blockDownloader IBlockDownloaderService
	logger          infrastructure.ILogger
	encoding        string
	pool            pond.Pool
	httpClient      *lib.HttpClient
}

func FxBatchProcessorService() fx.Option {
	return fx_utils.AsProvider(newBatchProcessor, new(IBatchProcessorService))
}

func newBatchProcessor(lc fx.Lifecycle, params batchProcessorServiceParams) IBatchProcessorService {
	s := &batchProcessorService{
		streamConfig:    params.StreamConfig,
		blockDownloader: params.BlockDownloader,
		logger:          params.Logger,
		metrics:         params.Metrics,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			s.serializer = params.StreamConfig.Serializer()

			_, outputCompressor := params.StreamConfig.Compressors()

			s.encoding = outputCompressor.EncodingType()

			return nil
		},
	})

	return s
}

func (s *batchProcessorService) Channel(
	ctx context.Context,
	inputChannel types.BatchReadonlyChannel,
	errorChannel types.ErrorChannel,
) types.ProcessedBatchReadonlyChannel {
	outputChannel := make(types.ProcessedBatchChannel, 1)

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

					processedBatch, err := s.process(ctx, batch)

					if err != nil {
						errorChannel <- err
						return
					}

					outputChannel <- processedBatch
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputChannel
}

func (s *batchProcessorService) process(
	ctx context.Context,
	batch *types.Batch,
) (*types.ProcessedBatch, error) {
	s.logger.Info(fmt.Sprintf("%s collected in %s", batch.String(), batch.TimeElapsed()))

	downloadedBlocks, err := s.blockDownloader.GetDownloadedBlocksData(ctx, batch)

	if err != nil {
		return nil, err
	}

	serialized, size, err := s.serialize(batch, downloadedBlocks)

	if err != nil {
		return nil, err
	}

	return types.NewProcessedBatch(batch, serialized, s.encoding, uint64(size))
}

func (s *batchProcessorService) serialize(batch *types.Batch, blocks []*types.DownloadedBlock) ([]byte, int, error) {
	return s.metrics.MeasureSerializeLatency(func() ([]byte, int, error) {
		start := time.Now()

		direction, err := batch.GetDirection()

		if err != nil {
			return nil, 0, err
		}

		serialized, size, err := s.serializer.Serialize(blocks, direction)

		if err != nil {
			return nil, 0, err
		}

		s.logger.Info(fmt.Sprintf("%s serialized in %s", batch.String(), time.Since(start)))

		return serialized, size, nil
	})
}
