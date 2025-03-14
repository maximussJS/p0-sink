package services

import (
	"context"
	"fmt"
	"github.com/alitto/pond/v2"
	"go.uber.org/fx"
	"net/url"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib"
	"p0-sink/internal/lib/compressors"
	"p0-sink/internal/lib/serializers"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	pb "p0-sink/proto"
	"sort"
	"strings"
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

	StreamConfig IStreamConfig
	Config       infrastructure.IConfig
	Logger       infrastructure.ILogger
}

type batchProcessorService struct {
	inputCompressor  compressors.ICompressor
	outputCompressor compressors.ICompressor
	serializer       serializers.ISerializer
	streamConfig     IStreamConfig
	logger           infrastructure.ILogger
	encoding         string
	pool             pond.Pool
	httpClient       *lib.HttpClient
}

func FxBatchProcessorService() fx.Option {
	return fx_utils.AsProvider(newBatchProcessor, new(IBatchProcessorService))
}

func newBatchProcessor(lc fx.Lifecycle, params batchProcessorServiceParams) IBatchProcessorService {
	s := &batchProcessorService{
		streamConfig: params.StreamConfig,
		logger:       params.Logger,
		httpClient:   lib.NewEmptyHttpClient(),
		pool:         pond.NewPool(params.Config.GetDownloadConcurrency()),
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			s.serializer = params.StreamConfig.Serializer()

			inputCompressor, outputCompressor := params.StreamConfig.Compressors()

			s.inputCompressor = inputCompressor
			s.outputCompressor = outputCompressor

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

	downloadedBlocks, err := s.getDownloadedBlocks(ctx, batch)

	if err != nil {
		return nil, err
	}

	serialized, size, err := s.serialize(batch, downloadedBlocks)

	if err != nil {
		return nil, err
	}

	return types.NewProcessedBatch(batch, serialized, s.encoding, uint64(size))
}

func (s *batchProcessorService) getDownloadedBlocks(ctx context.Context, batch *types.Batch) ([]*types.DownloadedBlock, error) {
	start := time.Now()

	blocks := batch.GetBlocks()

	blockLen := len(blocks)

	downloadedBlocks := make([]*types.DownloadedBlock, 0, blockLen)

	pool := pond.NewPool(blockLen + 1)

	defer pool.StopAndWait()

	group := pool.NewGroupContext(ctx)

	for _, block := range blocks {
		group.SubmitErr(func() error {
			downloadedBlock, err := s.downloadBlock(ctx, block)

			if err != nil {
				return err
			}

			downloadedBlocks = append(downloadedBlocks, downloadedBlock)

			return nil
		})
	}

	err := group.Wait()

	if err != nil {
		return nil, fmt.Errorf("error downloading blocks for %s: %w", batch.String(), err)
	}

	s.logger.Info(fmt.Sprintf("%s downloaded in %s", batch.String(), time.Since(start)))

	sort.Slice(downloadedBlocks[:], func(i, j int) bool {
		return downloadedBlocks[i].BlockNumber < downloadedBlocks[j].BlockNumber
	})

	donwloadedNumbers := make([]uint64, 0, len(downloadedBlocks))

	for _, block := range downloadedBlocks {
		donwloadedNumbers = append(donwloadedNumbers, block.BlockNumber)
	}

	s.logger.Info(fmt.Sprintf("%s downloaded blocks: %v", batch.String(), donwloadedNumbers))

	return downloadedBlocks, nil
}

func (s *batchProcessorService) serialize(batch *types.Batch, blocks []*types.DownloadedBlock) ([]byte, int, error) {
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
}

func (s *batchProcessorService) downloadBlock(ctx context.Context, block *pb.BlockWrapper) (*types.DownloadedBlock, error) {
	urlStr, headers := s.prepareBlockRequest(block)

	resp, err := s.httpClient.Get(ctx, urlStr, headers)
	if err != nil {
		return nil, err
	}

	compressedData, err := s.inputToOutputCompress(resp)

	if err != nil {
		return nil, err
	}

	return &types.DownloadedBlock{
		BlockWrapper: block,
		Data:         compressedData,
	}, nil
}

func (s *batchProcessorService) inputToOutputCompress(data []byte) ([]byte, error) {
	if s.inputCompressor == nil || s.outputCompressor == nil {
		return data, nil
	}

	if s.inputCompressor.EncodingType() == s.outputCompressor.EncodingType() {
		return data, nil
	}

	decompressedData, err := s.inputCompressor.Decompress(data)

	if err != nil {
		return nil, err
	}

	return s.outputCompressor.Compress(decompressedData)
}

func (s *batchProcessorService) prepareBlockRequest(block *pb.BlockWrapper) (string, map[string]string) {
	urlStr := block.Url

	headers := make(map[string]string)

	if strings.HasPrefix(urlStr, "data:") {
		return urlStr, headers
	}

	if !strings.Contains(urlStr, "bytesStart") && !strings.Contains(urlStr, "bytesEnd") {
		return urlStr, headers
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr, headers
	}

	queryParams := parsedURL.Query()
	bytesStart := queryParams.Get("bytesStart")
	bytesEnd := queryParams.Get("bytesEnd")

	headers["Range"] = "bytes=" + bytesStart + "-" + bytesEnd

	queryParams.Del("bytesStart")
	queryParams.Del("bytesEnd")
	parsedURL.RawQuery = queryParams.Encode()

	return parsedURL.String(), headers
}
