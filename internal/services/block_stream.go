package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"io"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	grpc_utils "p0-sink/internal/utils/grpc"
	pb "p0-sink/proto"
	"time"
)

type IBlockStreamService interface {
	Channel(ctx context.Context, req *pb.BlocksRequest, errorChannel types.ErrorChannel) types.BlockReadonlyChannel
}

type blockStreamParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
}

type blockStream struct {
	streamConfig   IStreamConfig
	grpcConnection *grpc.ClientConn
	logger         infrastructure.ILogger
}

func FxBlockStreamService() fx.Option {
	return fx_utils.AsProvider(newBlockStream, new(IBlockStreamService))
}

func newBlockStream(lc fx.Lifecycle, params blockStreamParams) IBlockStreamService {
	bs := &blockStream{
		streamConfig: params.StreamConfig,
		logger:       params.Logger,
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			err := bs.closeConnection()

			if err != nil {
				return err
			}

			bs.logger.Info("BlockStreamService stopped")

			return nil
		},
	})

	return bs
}

func (s *blockStream) Channel(
	ctx context.Context,
	req *pb.BlocksRequest,
	errorChannel types.ErrorChannel,
) types.BlockReadonlyChannel {
	if err := s.createConnection(); err != nil {
		errorChannel <- err
		return nil
	}

	client := pb.NewBlockStreamClient(s.grpcConnection)

	stream, err := client.Blocks(ctx, req)
	if err != nil {
		errorChannel <- fmt.Errorf("stream blocks error %v", err)
		return nil
	}

	blockCh := make(types.BlockChannel, s.streamConfig.BatchSize())

	go func() {
		defer close(blockCh)
		for {
			block, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				errorChannel <- fmt.Errorf("error receiving block: %w", err)
				return
			}

			select {
			case <-ctx.Done():
				return
			default:
				blockCh <- block
			}
		}
	}()

	return blockCh
}

func (s *blockStream) createConnection() error {
	err := s.closeConnection()
	if err != nil {
		return err
	}

	grpcUrl, credentials := grpc_utils.ResolveGrpcCredentials(s.streamConfig.GrpcUrl())

	client, err := grpc.NewClient(
		grpcUrl,
		credentials,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)),
		grpc.WithTimeout(5*time.Second),
	)

	if err != nil {
		return fmt.Errorf("failed to create grpc connection: %v", err)
	}

	s.grpcConnection = client

	return nil
}

func (s *blockStream) closeConnection() error {
	if s.grpcConnection != nil {
		err := s.grpcConnection.Close()

		if err != nil {
			return fmt.Errorf("failed to close grpc connection: %v", err)
		}
	}

	return nil
}
