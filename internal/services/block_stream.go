package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"io"
	"log"
	"p0-sink/internal/infrastructure"
	fx_utils "p0-sink/internal/utils/fx"
	grpc_utils "p0-sink/internal/utils/grpc"
	pb "p0-sink/proto"
	"time"
)

type IBlockStreamService interface {
	GetBlockChannel(ctx context.Context, req *pb.BlocksRequest) (<-chan *pb.BlockWrapper, error)
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

func newBlockStream(params blockStreamParams) IBlockStreamService {
	return &blockStream{
		streamConfig: params.StreamConfig,
		logger:       params.Logger,
	}
}

func (s *blockStream) GetBlockChannel(ctx context.Context, req *pb.BlocksRequest) (<-chan *pb.BlockWrapper, error) {
	if err := s.createConnection(); err != nil {
		return nil, err
	}
	client := pb.NewBlockStreamClient(s.grpcConnection)

	stream, err := client.Blocks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error calling Blocks: %w", err)
	}

	blockCh := make(chan *pb.BlockWrapper, s.streamConfig.BatchSize())

	go func() {
		defer close(blockCh)
		for {
			block, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				s.logger.Error(fmt.Sprintf("error receiving block: %v", err))
				return
			}
			select {
			case blockCh <- block:
			case <-ctx.Done():
				return
			}
		}
	}()

	return blockCh, nil
}

func (s *blockStream) streamBlock() error {
	return s.closeConnection()
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
			log.Printf("failed to close grpc connection: %v", err)
		}
	}

	return nil
}
