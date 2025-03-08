package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	pb "p0-sink/proto"
)

type IBlockOrderService interface {
	GetReadChannel(
		ctx context.Context,
		inputChannel types.BlockWrapperReadChannel,
		errorChannel types.ErrorChannel,
	) types.BlockWrapperReadChannel
}

type blockOrderServiceParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
}

type blockOrderService struct {
	logger              infrastructure.ILogger
	streamConfig        IStreamConfig
	blockOrderValidator *lib.BlockOrderValidator
}

func FxBlockOrderService() fx.Option {
	return fx_utils.AsProvider(newBlockOrderService, new(IBlockOrderService))
}

func newBlockOrderService(lc fx.Lifecycle, params blockOrderServiceParams) IBlockOrderService {
	s := &blockOrderService{
		logger:       params.Logger,
		streamConfig: params.StreamConfig,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			reorgAction := params.StreamConfig.ReorgAction()

			s.blockOrderValidator = lib.NewBlockOrderValidator(reorgAction)

			s.logger.Info(fmt.Sprintf("block validator reorg action: %s", reorgAction.String()))
			return nil
		},
	})

	return s
}

func (s *blockOrderService) GetReadChannel(
	ctx context.Context,
	inputChannel types.BlockWrapperReadChannel,
	errorChannel types.ErrorChannel,
) types.BlockWrapperReadChannel {
	outputChannel := make(types.BlockWrapperChannel, s.streamConfig.BatchSize())

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

					if err := s.blockOrderValidator.Validate(block); err != nil {
						errorChannel <- err
						return
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
