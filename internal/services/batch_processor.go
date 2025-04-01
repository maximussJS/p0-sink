package services

import (
	"context"
	"fmt"
	"github.com/alitto/pond/v2"
	"go.uber.org/fx"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib"
	"p0-sink/internal/lib/compressors"
	"p0-sink/internal/lib/serializers"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	"sync"
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
	BlockFunction   IBlockFunctionService
	Config          infrastructure.IConfig
	Logger          infrastructure.ILogger
}

type batchProcessorService struct {
	inputCompressor  compressors.ICompressor
	outputCompressor compressors.ICompressor
	serializer       serializers.ISerializer
	streamConfig     IStreamConfig
	pool             pond.Pool
	metrics          IMetricsService
	blockFunction    IBlockFunctionService
	blockDownloader  IBlockDownloaderService
	logger           infrastructure.ILogger
	encoding         string
	httpClient       *lib.HttpClient
}

func FxBatchProcessorService() fx.Option {
	return fx_utils.AsProvider(newBatchProcessor, new(IBatchProcessorService))
}

func newBatchProcessor(lc fx.Lifecycle, params batchProcessorServiceParams) IBatchProcessorService {
	s := &batchProcessorService{
		streamConfig:    params.StreamConfig,
		blockDownloader: params.BlockDownloader,
		blockFunction:   params.BlockFunction,
		logger:          params.Logger,
		metrics:         params.Metrics,
		pool:            pond.NewPool(100),
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			inputCompressor, outputCompressor := params.StreamConfig.Compressors()

			s.inputCompressor = inputCompressor
			s.outputCompressor = outputCompressor

			s.encoding = outputCompressor.EncodingType()
			s.serializer = params.StreamConfig.Serializer()

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
	s.logger.Debug(fmt.Sprintf("%s collected in %s", batch.String(), batch.TimeElapsed()))

	downloadedBlocks, err := s.blockDownloader.GetDownloadedBlocksData(ctx, batch)

	if err != nil {
		return nil, err
	}

	processedBlocks, err := s.processDownloadedBlocks(downloadedBlocks)

	if err != nil {
		return nil, err
	}

	serialized, size, err := s.serialize(batch, processedBlocks)

	if err != nil {
		return nil, err
	}

	return types.NewProcessedBatch(batch, serialized, s.encoding, uint64(size))
}

func (s *batchProcessorService) processDownloadedBlocks(
	downloadedBlocks []*types.DownloadedBlock,
) ([]*types.DownloadedBlock, error) {
	var (
		wg       sync.WaitGroup
		results  = make([]*types.DownloadedBlock, len(downloadedBlocks))
		errOnce  sync.Once
		firstErr error
	)

	for i, block := range downloadedBlocks {
		i := i
		block := block
		wg.Add(1)

		s.pool.Submit(func() {
			defer wg.Done()

			var result *types.DownloadedBlock
			var err error

			if s.streamConfig.FunctionEnabled() {
				var decompressedData []byte
				decompressedData, err = s.decompress(block.Data)
				if err == nil {
					var processedData []byte
					processedData, err = s.blockFunction.ApplyFunction(decompressedData)
					if err == nil {
						var compressedData []byte
						compressedData, err = s.compress(processedData)
						if err == nil {
							result = types.CopyWithNewData(block, compressedData)
						}
					}
				}
			} else {
				var blockData []byte
				blockData, err = s.inputToOutputCompress(block.Data)
				if err == nil {
					result = types.CopyWithNewData(block, blockData)
				}
			}

			if err != nil {
				errOnce.Do(func() { firstErr = err })
				return
			}

			results[i] = result
		})
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return results, nil
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

		s.logger.Debug(fmt.Sprintf("%s serialized in %s", batch.String(), time.Since(start)))

		return serialized, size, nil
	})
}

func (s *batchProcessorService) inputToOutputCompress(data []byte) ([]byte, error) {
	if s.inputCompressor.EncodingType() == s.outputCompressor.EncodingType() {
		return data, nil
	}

	decompressedData, err := s.decompress(data)

	if err != nil {
		return nil, err
	}

	return s.compress(decompressedData)
}

func (s *batchProcessorService) decompress(data []byte) ([]byte, error) {
	return s.metrics.MeasureDecompressLatency(func() ([]byte, error) {
		data, err := s.inputCompressor.Decompress(data)

		if err != nil {
			return nil, fmt.Errorf("failed to decompress data with %s : %w", s.inputCompressor.EncodingType(), err)
		}

		return data, nil
	})
}

func (s *batchProcessorService) compress(data []byte) ([]byte, error) {
	return s.metrics.MeasureCompressLatency(func() ([]byte, error) {
		data, err := s.outputCompressor.Compress(data)

		if err != nil {
			return nil, fmt.Errorf("failed to compress data with %s : %w", s.outputCompressor.EncodingType(), err)
		}

		return data, nil
	})
}
