package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/enums"
	"p0-sink/internal/lib"
	"p0-sink/internal/lib/compressors"
	"p0-sink/internal/lib/serializers"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	pb "p0-sink/proto"
)

type IStreamConfig interface {
	Id() string
	GrpcUrl() string
	BlocksRange() (fromBlock, toBlock int64, lag int32)
	ToBlock() int64
	FromBlock() int64
	State() types.State
	Dataset() string
	Compression() enums.ECompression
	BatchSize() int
	DestinationConfig() interface{}
	ElasticBatchEnabled() bool
	Destination() enums.EDestinationType
	ReorgAction() pb.ReorgAction
	Network() string
	UpdateBatchSize(int)
	Compressors() (input compressors.ICompressor, output compressors.ICompressor)
	Serializer() serializers.ISerializer
}

type streamConfigParams struct {
	fx.In

	StateManager IStateManagerService
}

type streamConfig struct {
	stream     *types.Stream
	sinkConfig *lib.SinkConfig
}

func FxStreamConfig() fx.Option {
	return fx_utils.AsProvider(newStreamConfig, new(IStreamConfig))
}

func newStreamConfig(lc fx.Lifecycle, params streamConfigParams) IStreamConfig {
	streamCfg := &streamConfig{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			stream, err := params.StateManager.GetStream(ctx)

			if err != nil {
				return fmt.Errorf("stream config get stream error: %w", err)
			}

			sinkConfig, err := lib.NewSinkConfig(stream.DestinationEntity.Config)

			if err != nil {
				return fmt.Errorf("stream config new sink config error: %w", err)
			}

			streamCfg.stream = stream
			streamCfg.sinkConfig = sinkConfig

			return nil
		},
	})

	return streamCfg
}

func (s *streamConfig) Id() string {
	return s.stream.Id
}

func (s *streamConfig) GrpcUrl() string {
	return "ssl://eth-bs.internal.troiszero.net"
	return s.stream.Network.BlockStreamGrpcUrl
}

func (s *streamConfig) BlocksRange() (fromBlock, toBlock int64, lag int32) {
	return s.stream.FromBlock, s.stream.ToBlock, s.stream.LagFromRealtime
}

func (s *streamConfig) ToBlock() int64 {
	return s.stream.ToBlock
}

func (s *streamConfig) FromBlock() int64 {
	return s.stream.FromBlock
}

func (s *streamConfig) State() types.State {
	return s.stream.State
}

func (s *streamConfig) Dataset() string {
	return s.stream.Dataset.Value
}

func (s *streamConfig) BatchSize() int {
	return s.stream.MaxBatchSize
}

func (s *streamConfig) ElasticBatchEnabled() bool {
	return true
}

func (s *streamConfig) Destination() enums.EDestinationType {
	switch s.stream.DestinationType {
	case "webhook":
		return enums.EDestinationTypeWebhook
	case "noop":
		return enums.EDestinationTypeNoop
	default:
		panic(fmt.Sprintf("unknown destination type: %s", s.stream.DestinationType))
	}
}

func (s *streamConfig) Compression() enums.ECompression {
	return s.sinkConfig.Compression
}

func (s *streamConfig) Compressors() (input compressors.ICompressor, output compressors.ICompressor) {
	input = compressors.NewGzipCompressor()

	switch s.Compression() {
	case enums.ECompressionGzip:
		return input, compressors.NewGzipCompressor()
	case enums.ECompressionNone:
		return input, compressors.NewNoneCompressor()
	default:
		panic(fmt.Sprintf("cannot find compressor for compression type: %s", s.Compression()))
	}
}

func (s *streamConfig) Serializer() serializers.ISerializer {
	switch s.Destination() {
	case enums.EDestinationTypeNoop:
		return serializers.NewNoopSerializer(s.Compression(), s.ReorgAction())
	case enums.EDestinationTypeWebhook:
		return serializers.NewWebhookSerializer(s.Compression(), s.ReorgAction())
	default:
		panic(fmt.Sprintf("cannot find serializer for destination type: %s", s.Destination()))
	}
}

func (s *streamConfig) DestinationConfig() interface{} {
	return s.stream.DestinationEntity.Config
}

func (s *streamConfig) Network() string {
	return s.stream.Network.ShortName
}

func (s *streamConfig) ReorgAction() pb.ReorgAction {
	if s.Destination() == enums.EDestinationTypeS3 {
		return pb.ReorgAction_ROLLBACK_AND_RESEND
	}

	if !s.sinkConfig.ResendOnReorg {
		return pb.ReorgAction_IGNORE
	}

	if s.sinkConfig.RollbackBeforeResend {
		return pb.ReorgAction_ROLLBACK_AND_RESEND
	}

	return pb.ReorgAction_RESEND
}

func (s *streamConfig) UpdateBatchSize(size int) {
	s.stream.MaxBatchSize = size
}
