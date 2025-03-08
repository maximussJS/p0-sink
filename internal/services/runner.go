package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/constants"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/types"
	context_utils "p0-sink/internal/utils/context"
	fx_utils "p0-sink/internal/utils/fx"
	"time"
)

type IRunnerService interface {
}

type runnerServiceParams struct {
	fx.In

	Logger           infrastructure.ILogger
	StreamCursor     IStreamCursorService
	BlockStream      IBlockStreamService
	BlockOrder       IBlockOrderService
	BatchSizeTracker IBatchSizeTrackerService
}

type runnerService struct {
	pipelineCtx       context.Context
	pipelineCtxCancel context.CancelFunc
	logger            infrastructure.ILogger
	streamCursor      IStreamCursorService
	blockStream       IBlockStreamService
	blockOrder        IBlockOrderService
	batchSizeTracker  IBatchSizeTrackerService
}

func FxRunnerService() fx.Option {
	return fx_utils.AsProvider(newRunnerService, new(IRunnerService))
}

func newRunnerService(lc fx.Lifecycle, params runnerServiceParams) IRunnerService {
	r := &runnerService{
		logger:           params.Logger,
		streamCursor:     params.StreamCursor,
		blockStream:      params.BlockStream,
		blockOrder:       params.BlockOrder,
		batchSizeTracker: params.BatchSizeTracker,
	}

	lc.Append(fx.Hook{
		OnStart: r.run,
		OnStop:  r.stop,
	})

	return r
}

func (s *runnerService) run(_ context.Context) error {
	s.pipelineCtx, s.pipelineCtxCancel = context.WithCancel(context.Background())

	go s.runPipeline()

	return nil
}

func (s *runnerService) stop(_ context.Context) error {
	s.logger.Info("stopping runner service")

	if s.pipelineCtxCancel != nil {
		s.pipelineCtxCancel()
	}

	return nil
}

func (s *runnerService) runPipeline() {
	errorChannel := make(types.ErrorChannel, 1)
	var err error
	for attempt := 1; attempt <= constants.PIPELINE_RETRY_MAX_ATTEMPTS; attempt++ {
		if attempt == 1 {
			s.logger.Info("starting pipeline")
		} else {
			s.logger.Info(fmt.Sprintf("retrying pipeline, attempt %d", attempt))
		}

		childCtx, cancelChildCtx := context.WithCancel(s.pipelineCtx)

		childCtx = context_utils.SetAttempt(childCtx, attempt)

		go s.startPipeline(childCtx, errorChannel)

		select {
		case <-s.pipelineCtx.Done():
			cancelChildCtx()
			s.logger.Info("pipeline stopped")
			return
		case err = <-errorChannel:
			s.logger.Error(fmt.Sprintf("error running pipeline: %v", err))
			cancelChildCtx()
			time.Sleep(constants.PIPELINE_RETRY_DELAY)
		}
	}
}

func (s *runnerService) startPipeline(ctx context.Context, errorChannel types.ErrorChannel) {
	blockRequest, err := s.streamCursor.GetBlockRequest()
	if err != nil {
		errorChannel <- err
		return
	}

	s.logger.Debug(fmt.Sprintf("block request: %v", blockRequest))

	blockStreamReadChannel := s.blockStream.GetReadChannel(ctx, blockRequest, errorChannel)

	if blockStreamReadChannel == nil {
		return
	}

	blockOrderReadChannel := s.blockOrder.GetReadChannel(ctx, blockStreamReadChannel, errorChannel)

	if blockOrderReadChannel == nil {
		return
	}

	batchSizeTrackerReadChannel := s.batchSizeTracker.GetReadChannel(ctx, blockOrderReadChannel, errorChannel)

	if batchSizeTrackerReadChannel == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case block := <-batchSizeTrackerReadChannel:
			if block == nil {
				return
			}

			s.logger.Info(fmt.Sprintf("block: %v", block.BlockNumber))
		}
	}
}
