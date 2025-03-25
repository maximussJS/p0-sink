package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/constants"
	custom_errors "p0-sink/internal/errors"
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

	Logger         infrastructure.ILogger
	StreamCursor   IStreamCursorService
	BlockStream    IBlockStreamService
	Batch          IBatchService
	BatchProcessor IBatchProcessorService
	BatchSender    IBatchSenderService
}

type runnerService struct {
	shutdowner        fx.Shutdowner
	pipelineCtx       context.Context
	pipelineCtxCancel context.CancelFunc
	logger            infrastructure.ILogger
	streamCursor      IStreamCursorService
	blockStream       IBlockStreamService
	batch             IBatchService
	batchProcessor    IBatchProcessorService
	batchSender       IBatchSenderService
}

func FxRunnerService() fx.Option {
	return fx_utils.AsProvider(newRunnerService, new(IRunnerService))
}

func newRunnerService(lc fx.Lifecycle, shutdowner fx.Shutdowner, params runnerServiceParams) IRunnerService {
	r := &runnerService{
		logger:         params.Logger,
		streamCursor:   params.StreamCursor,
		blockStream:    params.BlockStream,
		batch:          params.Batch,
		batchProcessor: params.BatchProcessor,
		batchSender:    params.BatchSender,
		shutdowner:     shutdowner,
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

			if attempt == constants.PIPELINE_RETRY_MAX_ATTEMPTS {
				s.logger.Error("pipeline failed after max attempts")
				return
			}

			terminationErr, ok := err.(*custom_errors.StreamTerminationError)
			if ok {
				s.logger.Error(fmt.Sprintf("stream termination error: %v", terminationErr))
				return
			}
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

	blocksChannel := s.blockStream.Channel(ctx, blockRequest, errorChannel)

	batchChannel := s.batch.Channel(ctx, blocksChannel, errorChannel)

	processedBatchChannel := s.batchProcessor.Channel(ctx, batchChannel, errorChannel)

	doneChannel := s.batchSender.ProcessChannel(ctx, processedBatchChannel, errorChannel)

	<-doneChannel

	s.logger.Info("stream has been processed successfully. shutting down in 5 seconds")

	time.Sleep(5 * time.Second)

	err = s.shutdowner.Shutdown()

	if err != nil {
		s.logger.Error(fmt.Sprintf("error shutting down: %v", err))
	}
}
