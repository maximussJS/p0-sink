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
		errorChannel types.ErrorChannel,
	) types.DoneChannel
}

type batchSenderServiceParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
	StreamCursor IStreamCursorService
	Metrics      IMetricsService
}

type batchSenderService struct {
	logger        infrastructure.ILogger
	retryAttempts int
	retryDelay    time.Duration
	retryStrategy enums.ERetryStrategy
	destination   destinations.IDestination
	streamCursor  IStreamCursorService
	metrics       IMetricsService
}

func FxBatchSenderService() fx.Option {
	return fx_utils.AsProvider(newBatchSenderService, new(IBatchSenderService))
}

func newBatchSenderService(lc fx.Lifecycle, params batchSenderServiceParams) IBatchSenderService {
	bs := &batchSenderService{
		logger:       params.Logger,
		streamCursor: params.StreamCursor,
		metrics:      params.Metrics,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			bs.destination = params.StreamConfig.Destination()
			bs.retryAttempts = params.StreamConfig.RetryAttempts()
			bs.retryDelay = params.StreamConfig.RetryDelay()
			bs.retryStrategy = params.StreamConfig.RetryStrategy()

			bs.logger.Info(fmt.Sprintf("using %s ", bs.destination.String()))
			return nil
		},
	})

	return bs
}

func (s *batchSenderService) ProcessChannel(
	ctx context.Context,
	inputChannel types.ProcessedBatchReadonlyChannel,
	errorChannel types.ErrorChannel,
) types.DoneChannel {
	doneChannel := make(types.DoneChannel)

	go func() {
		for {
			select {
			case batch, ok := <-inputChannel:
				if !ok {
					return
				}

				done, err := s.process(ctx, batch)

				if err != nil {
					errorChannel <- err
					return
				}

				if done {
					doneChannel <- struct{}{}
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return doneChannel
}

func (s *batchSenderService) process(ctx context.Context, batch *types.ProcessedBatch) (bool, error) {
	err := s.sendWithRetry(ctx, batch)
	if err != nil {
		return false, fmt.Errorf("error sending %s %v", batch.String(), err)
	}

	err = s.commit(ctx, *batch)
	if err != nil {
		return false, fmt.Errorf("error committing %s %v", batch.String(), err)
	}

	s.metrics.IncTotalBatchesSent()
	s.metrics.IncTotalBytesSent(batch.BilledBytes)
	s.metrics.IncTotalBlocksSent(batch.NumBlocks())

	if s.streamCursor.ReachedEnd() {
		return true, nil
	}

	return false, nil
}

func (s *batchSenderService) commit(ctx context.Context, batch types.ProcessedBatch) error {
	return s.metrics.MeasureCommitLatency(func() error {
		start := time.Now()
		err := s.streamCursor.Commit(ctx, batch)

		if err != nil {
			return fmt.Errorf("error committing %s %v", batch.String(), err)
		}

		elapsed := time.Since(start)

		s.logger.Info(fmt.Sprintf("Commited %s. Took %s", batch.String(), elapsed))

		return nil
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
			time.Sleep(retryDelay.CalculateDelay(attempt, s.retryDelay))

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
