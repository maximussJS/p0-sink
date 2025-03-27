package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/enums"
	"p0-sink/internal/errors"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib"
	"p0-sink/internal/lib/destinations"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	"time"
)

type IBatchSenderService interface {
	ProcessChannel(
		ctx context.Context,
		inputChannel types.ProcessedBatchReadonlyChannel,
		doneChannel types.DoneChannel,
		errorChannel types.ErrorChannel,
	)
}

type batchSenderServiceParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
	StreamCursor IStreamCursorService
	StateManager IStateManagerService
	Metrics      IMetricsService
}

type batchSenderService struct {
	logger         infrastructure.ILogger
	previousStatus enums.EStatus
	retryAttempts  int
	retryDelay     time.Duration
	retryStrategy  enums.ERetryStrategy
	destination    destinations.IDestination
	streamCursor   IStreamCursorService
	metrics        IMetricsService
	stateManager   IStateManagerService
}

func FxBatchSenderService() fx.Option {
	return fx_utils.AsProvider(newBatchSenderService, new(IBatchSenderService))
}

func newBatchSenderService(lc fx.Lifecycle, params batchSenderServiceParams) IBatchSenderService {
	bs := &batchSenderService{
		logger:       params.Logger,
		streamCursor: params.StreamCursor,
		stateManager: params.StateManager,
		metrics:      params.Metrics,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			bs.destination = params.StreamConfig.Destination()
			bs.retryAttempts = params.StreamConfig.RetryAttempts()
			bs.retryDelay = params.StreamConfig.RetryDelay()
			bs.retryStrategy = params.StreamConfig.RetryStrategy()
			bs.previousStatus = params.StreamConfig.Status()

			bs.logger.Info(fmt.Sprintf("using %s ", bs.destination.String()))
			return nil
		},
	})

	return bs
}

func (s *batchSenderService) ProcessChannel(
	ctx context.Context,
	inputChannel types.ProcessedBatchReadonlyChannel,
	doneChannel types.DoneChannel,
	errorChannel types.ErrorChannel,
) {
	go func() {
		for {
			select {
			case batch, ok := <-inputChannel:
				if !ok {
					return
				}

				s.process(ctx, batch, doneChannel, errorChannel)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *batchSenderService) process(
	ctx context.Context,
	batch *types.ProcessedBatch,
	doneChannel types.DoneChannel,
	errorChannel types.ErrorChannel,
) {
	err := s.sendWithRetry(ctx, batch)
	if err != nil {
		errorChannel <- fmt.Errorf("error sending batch %s %v", batch.String(), err)
		return
	}

	status, err := s.commit(ctx, *batch)
	if err != nil {
		errorChannel <- fmt.Errorf("error committing %s %v", batch.String(), err)
		return
	}

	s.metrics.IncTotalBatchesSent()
	s.metrics.IncTotalBytesSent(batch.BilledBytes)
	s.metrics.IncTotalBlocksSent(batch.NumBlocks())

	if status == enums.StatusStarting {
		err := s.stateManager.MarkAsRunning(ctx)

		if err != nil {
			errorChannel <- fmt.Errorf("error marking as running %v", err)
			return
		}

		s.previousStatus = enums.StatusRunning

		return
	}

	if s.previousStatus == enums.StatusRunning && status == enums.StatusPaused {
		s.previousStatus = enums.StatusPaused
		errorChannel <- errors.NewStreamTerminationError("stream was paused")
		return
	}

	if s.previousStatus == enums.StatusRunning && status == enums.StatusFinished {
		s.previousStatus = enums.StatusFinished
		errorChannel <- errors.NewStreamTerminationError("stream was finished")
		return
	}

	if status != enums.StatusRunning {
		errorChannel <- errors.NewStreamTerminationError(fmt.Sprintf("stream has incorrect status %s", status))
		return
	}

	if s.streamCursor.ReachedEnd() {
		doneChannel <- struct{}{}
		return
	}
}

func (s *batchSenderService) commit(ctx context.Context, batch types.ProcessedBatch) (enums.EStatus, error) {
	return s.metrics.MeasureCommitLatency(func() (enums.EStatus, error) {
		start := time.Now()
		status, err := s.streamCursor.Commit(ctx, batch)

		if err != nil {
			return "", fmt.Errorf("error committing %s %v", batch.String(), err)
		}

		elapsed := time.Since(start)

		s.logger.Info(fmt.Sprintf("Commited %s. Took %s", batch.String(), elapsed))

		return status, nil
	})
}

func (s *batchSenderService) send(ctx context.Context, batch *types.ProcessedBatch) error {
	return s.metrics.MeasureSendLatency(func() error {
		err := s.sendWithRetry(ctx, batch)

		if err != nil {
			return err
		}

		s.metrics.SetLastBlockSentAt()

		return nil
	})
}

func (s *batchSenderService) sendWithRetry(
	ctx context.Context,
	batch *types.ProcessedBatch,
) error {
	retryDelay := lib.NewRetryDelay(s.retryStrategy)

	for attempt := 0; attempt < s.retryAttempts; attempt++ {
		start := time.Now()

		if attempt > 0 {
			sleepTime := retryDelay.CalculateDelay(attempt, s.retryDelay)
			s.logger.Warn(fmt.Sprintf("sleeping for %v before %d data send retry attempt", sleepTime, attempt))
			time.Sleep(sleepTime)

			s.metrics.IncRetriesCount()
			s.logger.Warn(fmt.Sprintf("retrying data send. attempt %d of %d", attempt, s.retryAttempts))
		}

		err := s.destination.Send(ctx, batch)

		if err != nil {
			if s.isNotRetryableError(err) {
				return errors.NewStreamTerminationError(fmt.Sprintf("error sending data: %v", err))
			}

			s.metrics.IncErrorsCount()
			s.logger.Error(fmt.Sprintf("error sending %s: %v", batch.String(), err))

			continue
		}

		elapsed := time.Since(start)

		s.logger.Info(fmt.Sprintf("Sent %s. Took %s", batch.String(), elapsed))

		return nil
	}

	s.logger.Error(fmt.Sprintf("max %d send attempts reached", s.retryAttempts))
	return errors.NewStreamTerminationError(fmt.Sprintf("max %d send attempts reached", s.retryAttempts))
}

func (s *batchSenderService) isNotRetryableError(_ error) bool {
	return false
}
