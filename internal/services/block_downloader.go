package services

import (
	"context"
	"go.uber.org/fx"
	"net/url"
	"p0-sink/internal/lib"
	"p0-sink/internal/lib/compressors"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	pb "p0-sink/proto"
	"strings"
)

type IBlockDownloaderService interface {
	GetReadChannel(
		ctx context.Context,
		inputChannel types.BlockWrapperReadChannel,
		errorChannel types.ErrorChannel,
	) types.DownloadedBlockReadChannel
}

type blockDownloaderServiceParams struct {
	fx.In

	StreamConfig IStreamConfig
}

type blockDownloaderService struct {
	inputCompressor  compressors.ICompressor
	outputCompressor compressors.ICompressor
	httpClient       *lib.HttpClient
	streamConfig     IStreamConfig
}

func FxBlockDownloaderService() fx.Option {
	return fx_utils.AsProvider(newBlockDownloader, new(IBlockDownloaderService))
}

func newBlockDownloader(lc fx.Lifecycle, params blockDownloaderServiceParams) IBlockDownloaderService {
	s := &blockDownloaderService{
		streamConfig: params.StreamConfig,
		httpClient:   lib.NewEmptyHttpClient(),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			inputCompressor, outputCompressor := params.StreamConfig.Compressors()

			s.inputCompressor = inputCompressor
			s.outputCompressor = outputCompressor
			return nil
		},
	})

	return s
}

func (s *blockDownloaderService) GetReadChannel(
	ctx context.Context,
	inputChannel types.BlockWrapperReadChannel,
	errorChannel types.ErrorChannel,
) types.DownloadedBlockReadChannel {
	outputChannel := make(types.DownloadedBlockChannel)

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

					downloadedBlock, err := s.downloadBlock(ctx, block)

					if err != nil {
						errorChannel <- err
						return
					}

					outputChannel <- downloadedBlock
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputChannel
}

func (s *blockDownloaderService) downloadBlock(ctx context.Context, block *pb.BlockWrapper) (*types.DownloadedBlock, error) {
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

func (s *blockDownloaderService) inputToOutputCompress(data []byte) ([]byte, error) {
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

func (s *blockDownloaderService) prepareBlockRequest(block *pb.BlockWrapper) (string, map[string]string) {
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
