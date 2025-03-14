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
	ProcessChannel(ctx context.Context, inputChannel types.SerializedBatchReadChannel, errorChannel types.ErrorChannel)
}

type batchSenderServiceParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
}

type batchSenderService struct {
	logger        infrastructure.ILogger
	retryAttempts int
	retryDelay    time.Duration
	retryStrategy enums.ERetryStrategy
	destination   destinations.IDestination
}

func FxBatchSenderService() fx.Option {
	return fx_utils.AsProvider(newBatchSenderService, new(IBatchSenderService))
}

func newBatchSenderService(lc fx.Lifecycle, params batchSenderServiceParams) IBatchSenderService {
	bs := &batchSenderService{
		logger: params.Logger,
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
	inputChannel types.SerializedBatchReadChannel,
	errorChannel types.ErrorChannel,
) {
	for {
		select {
		case batch, ok := <-inputChannel:
			if !ok {
				return
			}

			err := s.SendWithRetry(ctx, batch)
			if err != nil {
				errorChannel <- err
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *batchSenderService) SendWithRetry(
	ctx context.Context,
	batch *types.SerializedBatch,
) error {
	retryDelay := lib.NewRetryDelay(s.retryStrategy)

	for attempt := 0; attempt < s.retryAttempts; attempt++ {
		start := time.Now()

		if attempt > 0 {
			time.Sleep(retryDelay.CalculateDelay(attempt, s.retryDelay))

			s.logger.Warn(fmt.Sprintf("retrying data send. attempt %d of %d", attempt, s.retryAttempts))
		}

		err := s.destination.Send(ctx, batch)

		if err != nil {
			s.logger.Error(fmt.Sprintf("error sending data: %v", err))

			if s.isNotRetryableError(err) {
				return errors.NewStreamTerminationError(fmt.Sprintf("error sending data: %v", err))
			}

			continue
		}

		elapsed := time.Since(start)

		s.logger.Info(fmt.Sprintf("Sent %s. Commiting batch took %s", batch.String(), elapsed))

		return nil
	}

	s.logger.Error(fmt.Sprintf("max %d send attempts reached", s.retryAttempts))
	return errors.NewStreamTerminationError(fmt.Sprintf("max %d send attempts reached", s.retryAttempts))
}

func (s *batchSenderService) isNotRetryableError(err error) bool {
	return false
}
