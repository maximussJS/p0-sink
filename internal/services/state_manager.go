package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/fx"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/lib"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
)

type IStateManagerService interface {
	GetStream(ctx context.Context) (*types.Stream, error)
	GetStatus(ctx context.Context) (*types.StatusResponse, error)
	UpdateCursor(ctx context.Context, payload *types.UpdateCursorPayload) (*types.StatusResponse, error)
	MarkAsRunning(ctx context.Context) error
}

type stateManagerServiceParams struct {
	fx.In

	Logger infrastructure.ILogger
	Config infrastructure.IConfig
}

type stateManagerService struct {
	streamId   string
	url        string
	token      string
	httpClient *lib.HttpClient
	logger     infrastructure.ILogger
}

func FxStateManagerService() fx.Option {
	return fx_utils.AsProvider(newStateManagerService, new(IStateManagerService))
}

func newStateManagerService(params stateManagerServiceParams) IStateManagerService {
	headers := map[string]string{
		"Authorization": params.Config.GetStateManagerToken(),
	}

	return &stateManagerService{
		streamId: params.Config.GetStreamId(),
		url:      params.Config.GetStateManagerUrl(),
		token:    params.Config.GetStateManagerToken(),
		logger:   params.Logger,
		httpClient: lib.NewHttpClient(
			params.Config.GetStateManagerUrl(),
			headers,
			params.Config.GetStateManagerTimeout(),
		),
	}
}

func (s *stateManagerService) GetStream(ctx context.Context) (*types.Stream, error) {
	path := fmt.Sprintf("/stream/%s", s.streamId)

	responseBody, err := s.httpClient.Get(ctx, path, lib.EmptyHttpHeaders)

	if err != nil {
		return nil, fmt.Errorf("state manager get stream error: %w", err)
	}

	var stream types.Stream
	if err := json.Unmarshal(responseBody, &stream); err != nil {
		return nil, fmt.Errorf("state manager unmarshal stream error: %w", err)
	}

	return &stream, nil
}

func (s *stateManagerService) GetStatus(ctx context.Context) (*types.StatusResponse, error) {
	path := fmt.Sprintf("/stream/%s/status", s.streamId)

	responseBody, err := s.httpClient.Get(ctx, path, lib.EmptyHttpHeaders)
	if err != nil {
		return nil, fmt.Errorf("state manager get status error: %w", err)
	}

	var statusResp types.StatusResponse
	if err := json.Unmarshal(responseBody, &statusResp); err != nil {
		return nil, fmt.Errorf("state manager unmarshal status error: %w", err)
	}

	return &statusResp, nil
}

func (s *stateManagerService) UpdateCursor(ctx context.Context, payload *types.UpdateCursorPayload) (*types.StatusResponse, error) {
	path := fmt.Sprintf("/stream/%s/state/cursor", s.streamId)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("state manager update cursor: error marshaling payload: %w", err)
	}

	responseBody, err := s.httpClient.Patch(ctx, path, lib.EmptyHttpHeaders, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("state manager update cursor error: %w", err)
	}

	var statusResp types.StatusResponse
	if err := json.Unmarshal(responseBody, &statusResp); err != nil {
		return nil, fmt.Errorf("state manager update cursor unmarshal error: %w", err)
	}

	return &statusResp, nil
}

func (s *stateManagerService) MarkAsRunning(ctx context.Context) error {
	path := fmt.Sprintf("/stream/%s/state/status/running", s.streamId)
	_, err := s.httpClient.Patch(ctx, path, lib.EmptyHttpHeaders, nil)

	if err != nil {
		return fmt.Errorf("state manager mark as running error: %w", err)
	}

	return nil
}
