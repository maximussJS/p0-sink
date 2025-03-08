package services

import (
	"context"
	"go.uber.org/fx"
	"net/url"
	"p0-sink/internal/lib"
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
	) types.BlockWrapperReadChannel
}

type blockDownloaderServiceParams struct {
	fx.In
}

type blockDownloaderService struct {
	httpClient *lib.HttpClient
}

func FxBlockDownloaderService() fx.Option {
	return fx_utils.AsProvider(newBlockDownloader, new(IBlockDownloaderService))
}

func newBlockDownloader(params blockDownloaderService) IBlockDownloaderService {
	return &blockDownloaderService{
		httpClient: lib.NewEmptyHttpClient(),
	}
}

func (s *blockDownloaderService) GetReadChannel(
	ctx context.Context,
	inputChannel types.BlockWrapperReadChannel,
	_ types.ErrorChannel,
) types.BlockWrapperReadChannel {
	outputChannel := make(types.BlockWrapperChannel)

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
					outputChannel <- block
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return outputChannel
}

func PrepareBlockRequest(urlStr string) (string, map[string]string) {
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
