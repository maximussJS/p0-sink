package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/enums"
	"p0-sink/internal/errors"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/types"
	"p0-sink/internal/utils/direction"
	fx_utils "p0-sink/internal/utils/fx"
	pb "p0-sink/proto"
	"time"
)

type IStreamCursorService interface {
	Commit(ctx context.Context, batch types.ProcessedBatch) (enums.EStatus, error)
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

func (s *streamCursorService) Commit(ctx context.Context, batch types.ProcessedBatch) (enums.EStatus, error) {
	start := time.Now()

	timeDelta := int64(time.Since(s.lastCommitAt).Seconds())

	payload := types.NewUpdateCursorPayload(batch, s.state, timeDelta)

	result, err := s.stateManager.UpdateCursor(ctx, payload)

	if err != nil {
		return "", fmt.Errorf("failed to update cursor: %w", err)
	}

	s.lastCommitAt = start
	lastBlockNumber := batch.LastBlockNumber()

	s.state = types.State{
		StreamId:        s.streamConfig.Id(),
		Status:          result.Status,
		LastCursor:      batch.Cursor,
		LastBlockNumber: &lastBlockNumber,
		BlocksSent:      s.state.BlocksSent + int64(batch.NumBlocks()),
		BytesSent:       s.state.BytesSent + int64(batch.BilledBytes),
		TimeSpent:       s.state.TimeSpent + timeDelta,
		CursorUpdatedAt: time.Now().Unix(),
		LastDirection:   int(direction.ArrowToDirection(batch.Direction)),
	}

	return result.Status, nil
}

func (s *streamCursorService) GetBlockRequest() (*pb.BlocksRequest, error) {
	fromBlock, toBlock, lag := s.streamConfig.BlocksRange()

	reorgAction := s.streamConfig.ReorgAction()

	// for testing purposes

	//return &pb.BlocksRequest{
	//	Lag:           &lag,
	//	From:          &pb.BlocksRequest_FromBlockNumber{FromBlockNumber: fromBlock},
	//	ToBlockNumber: toBlock,
	//	Dataset:       s.streamConfig.Dataset(),
	//	OnReorg:       &reorgAction,
	//	Format:        string(enums.DatasetFormatJSON),
	//	Compression:   string(enums.ECompressionGzip),
	//}, nil

	if s.ReachedEnd() {
		return nil, errors.NewStreamTerminationError("stream reached end")
	}

	if !s.isRunning() && !s.isStarting() {
		return nil, errors.NewStreamTerminationError(fmt.Sprintf("stream is not running or starting, status: %s", s.state.Status))
	}

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

	bn := s.state.LastBlockNumber

	return s.streamConfig.ToBlock() <= int64(*bn)
}

func (s *streamCursorService) isRunning() bool {
	return s.state.Status == enums.StatusRunning
}

func (s *streamCursorService) isStarting() bool {
	return s.state.Status == enums.StatusStarting
}
