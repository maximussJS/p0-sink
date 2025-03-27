package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"net/url"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	"sync"
)

type IBlockDownloaderService interface {
	GetDownloadedBlocksData(ctx context.Context, batch *types.Batch) ([]*types.DownloadedBlock, error)
}

type blockDownloaderServiceParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
	Metrics      IMetricsService
}

type blockDownloaderService struct {
	metrics          IMetricsService
	logger           infrastructure.ILogger
	blockDataMap     map[string][]byte
	blockDataMapLock *sync.RWMutex
	httpClient       *lib.HttpClient
}

func FxBlockDownloaderService() fx.Option {
	return fx_utils.AsProvider(newBlockDownloaderService, new(IBlockDownloaderService))
}

func newBlockDownloaderService(params blockDownloaderServiceParams) IBlockDownloaderService {
	s := &blockDownloaderService{
		logger:           params.Logger,
		blockDataMap:     make(map[string][]byte),
		blockDataMapLock: &sync.RWMutex{},
		metrics:          params.Metrics,
		httpClient:       lib.NewHttpClientWithDisabledCompression(),
	}

	return s
}

func (s *blockDownloaderService) GetDownloadedBlocksData(ctx context.Context, batch *types.Batch) ([]*types.DownloadedBlock, error) {
	blocks := batch.GetBlocks()
	data := make([]*types.DownloadedBlock, 0)

	for _, block := range blocks {
		blockData, err := s.getBlockData(ctx, block)
		if err != nil {
			return nil, fmt.Errorf("failed to get block data: %w", err)
		}

		data = append(data, types.NewDownloadedBlock(block, blockData))
	}

	return data, nil
}

func (s *blockDownloaderService) getBlockData(ctx context.Context, block *types.Block) ([]byte, error) {
	s.blockDataMapLock.Lock()

	defer s.blockDataMapLock.Unlock()

	batchNumber := block.BatchNumber()

	if data, ok := s.blockDataMap[batchNumber]; ok {
		return data[block.BytesStart : block.BytesEnd+1], nil
	}

	urlStr, err := s.prepareUrl(block.Url)

	if err != nil {
		return nil, err
	}

	s.logger.Debug(fmt.Sprintf("Downloading batch %s", batchNumber))

	resp, err := s.metrics.MeasureDownloadLatency(func() ([]byte, error) {
		resp, err := s.httpClient.Get(ctx, urlStr, lib.EmptyHttpHeaders)

		if err != nil {
			return nil, fmt.Errorf("failed to download block: %w", err)
		}

		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	s.blockDataMap[batchNumber] = resp

	return resp[block.BytesStart : block.BytesEnd+1], nil
}

func (s *blockDownloaderService) prepareUrl(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	queryParams := parsedURL.Query()

	queryParams.Del("bytesStart")
	queryParams.Del("bytesEnd")
	parsedURL.RawQuery = queryParams.Encode()

	return parsedURL.String(), nil
}
