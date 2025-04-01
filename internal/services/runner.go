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
	start          time.Time
	shutdowner     fx.Shutdowner
	logger         infrastructure.ILogger
	streamCursor   IStreamCursorService
	blockStream    IBlockStreamService
	batch          IBatchService
	batchProcessor IBatchProcessorService
	batchSender    IBatchSenderService
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
	go s.runPipeline()

	return nil
}

func (s *runnerService) stop(_ context.Context) error {
	s.logger.Info("stopping runner service")

	return nil
}

func (s *runnerService) runPipeline() {
	s.start = time.Now()
	errorChannel := make(types.ErrorChannel, 1)
	doneChannel := make(types.DoneChannel, 1)

	pipelineCtx, pipelineCtxCancel := context.WithCancel(context.Background())
	defer pipelineCtxCancel()

	var err error

	for attempt := 1; attempt <= constants.PIPELINE_RETRY_MAX_ATTEMPTS; attempt++ {
		if attempt == 1 {
			s.logger.Info("starting pipeline")
		} else {
			s.logger.Info(fmt.Sprintf("retrying pipeline, attempt %d", attempt))
		}

		childCtx, childCtxCancel := context.WithCancel(pipelineCtx)

		childCtx = context_utils.SetAttempt(childCtx, attempt)

		s.startStream(childCtx, childCtxCancel, doneChannel, errorChannel)

		select {
		case <-doneChannel:
			s.logger.Info("stream has been processed successfully.")
			childCtxCancel()
			s.shutdown()
			return
		case <-pipelineCtx.Done():
			childCtxCancel()
			s.logger.Info("context cancelled. stopping pipeline.")
			s.shutdown()
			return
		case err = <-errorChannel:
			s.logger.Error(fmt.Sprintf("error running pipeline: %v", err))

			if attempt == constants.PIPELINE_RETRY_MAX_ATTEMPTS {
				s.logger.Error("pipeline failed after max attempts")
				childCtxCancel()
				s.shutdown()
				return
			}

			if s.isNotRetryableError(err) {
				s.logger.Error(fmt.Sprintf("stopping pipeline because of non-retryable error: %v", err))
				childCtxCancel()
				s.shutdown()
				return
			}

			s.logger.Info(fmt.Sprintf("retrying pipeline in %v", constants.PIPELINE_RETRY_DELAY))
			time.Sleep(constants.PIPELINE_RETRY_DELAY)
		}

		childCtxCancel()
	}
}

func (s *runnerService) startStream(ctx context.Context, cancelFunc context.CancelFunc, doneChannel types.DoneChannel, errorChannel types.ErrorChannel) {
	blockRequest, err := s.streamCursor.GetBlockRequest()

	if err != nil {
		errorChannel <- err
		cancelFunc()
		return
	}

	s.logger.Debug(fmt.Sprintf("block request: %v", blockRequest))

	blocksChannel := s.blockStream.Channel(ctx, blockRequest, errorChannel)

	batchChannel := s.batch.Channel(ctx, blocksChannel, errorChannel)

	processedBatchChannel := s.batchProcessor.Channel(ctx, batchChannel, errorChannel)

	s.batchSender.ProcessChannel(ctx, processedBatchChannel, doneChannel, errorChannel)
}

func (s *runnerService) isNotRetryableError(err error) bool {
	terminationErr, ok := err.(*custom_errors.StreamTerminationError)
	if ok {
		s.logger.Error(fmt.Sprintf("stream termination error: %v", terminationErr))
		return true
	}

	return false
}

func (s *runnerService) shutdown() {
	s.logger.Info(fmt.Sprintf("pipeline took %v", time.Since(s.start)))
	s.logger.Info("shutting down in 5 seconds...")
	time.Sleep(5 * time.Second)

	err := s.shutdowner.Shutdown()

	if err != nil {
		s.logger.Error(fmt.Sprintf("error shutting down: %v", err))
	}
}
