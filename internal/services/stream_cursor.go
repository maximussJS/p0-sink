package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/enums"
	"p0-sink/internal/errors"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
	pb "p0-sink/proto"
	"time"
)

type IStreamCursorService interface {
	GetBlockRequest() (*pb.BlocksRequest, error)
	ReachedEnd() bool
}

type streamCursorServiceParams struct {
	fx.In

	Logger       infrastructure.ILogger
	StreamConfig IStreamConfig
	StateManager IStateManagerService
}

type streamCursorService struct {
	lastCommitAt time.Time
	state        types.State
	logger       infrastructure.ILogger
	streamConfig IStreamConfig
	stateManager IStateManagerService
}

func FxStreamCursorService() fx.Option {
	return fx_utils.AsProvider(newStreamCursorService, new(IStreamCursorService))
}

func newStreamCursorService(lc fx.Lifecycle, params streamCursorServiceParams) IStreamCursorService {
	sc := &streamCursorService{
		lastCommitAt: time.Now(),
		logger:       params.Logger,
		streamConfig: params.StreamConfig,
		stateManager: params.StateManager,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			state := params.StreamConfig.State()
			sc.state = state
			return nil
		},
	})

	return sc
}

func (s *streamCursorService) GetBlockRequest() (*pb.BlocksRequest, error) {
	fromBlock, toBlock, lag := s.streamConfig.BlocksRange()

	if s.ReachedEnd() {
		return nil, errors.NewStreamTerminationError("stream reached end")
	}

	if !s.isRunning() && !s.isStarting() {
		return nil, errors.NewStreamTerminationError(fmt.Sprintf("stream is not running or starting, status: %s", s.state.Status))
	}

	reorgAction := s.streamConfig.ReorgAction()

	req := &pb.BlocksRequest{
		Lag:           &lag,
		From:          &pb.BlocksRequest_FromBlockNumber{FromBlockNumber: fromBlock},
		ToBlockNumber: toBlock,
		Dataset:       s.streamConfig.Dataset(),
		OnReorg:       &reorgAction,
		Format:        string(enums.DatasetFormatJSON),
		Compression:   string(enums.ECompressionGzip),
	}

	if s.state.LastCursor != "" {
		req.From = &pb.BlocksRequest_FromCursor{FromCursor: s.state.LastCursor}
	}

	return req, nil
}

func (s *streamCursorService) ReachedEnd() bool {
	if s.streamConfig.ToBlock() == -1 {
		return false
	}

	if s.state.LastBlockNumber == nil {
		return false
	}

	return s.streamConfig.ToBlock() <= *s.state.LastBlockNumber
}

func (s *streamCursorService) isRunning() bool {
	return s.state.Status == string(enums.StatusRunning)
}

func (s *streamCursorService) isStarting() bool {
	return s.state.Status == string(enums.StatusStarting)
}
