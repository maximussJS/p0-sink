package services

import (
	"context"
	"go.uber.org/fx"
	fx_utils "p0-sink/internal/utils/fx"
)

type IRunnerService interface {
	Run(ctx context.Context) error
}

type runnerServiceParams struct {
	fx.In

	StreamCursor IStreamCursorService
	BlockStream  IBlockStreamService
}

type runnerService struct {
	streamCursor IStreamCursorService
	blockStream  IBlockStreamService
}

func FxRunnerService() fx.Option {
	return fx_utils.AsProvider(newRunnerService, new(IRunnerService))
}

func newRunnerService(lc fx.Lifecycle, params runnerServiceParams) IRunnerService {
	r := &runnerService{
		streamCursor: params.StreamCursor,
		blockStream:  params.BlockStream,
	}

	lc.Append(fx.Hook{
		OnStart: r.Run,
	})

	return r
}

func (s *runnerService) Run(_ context.Context) error {
	blockRequest, err := s.streamCursor.GetBlockRequest()
	if err != nil {
		return err
	}

	ctx := context.Background()

	go func() {
		err = s.blockStream.Stream(ctx, blockRequest)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}
