package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/fx"
	"io"
	"net/http"
	"p0-sink/internal/infrastructure"
	"p0-sink/internal/types"
	fx_utils "p0-sink/internal/utils/fx"
)

type IStateManagerService interface {
	GetStream(ctx context.Context) (*types.Stream, error)
	GetStatus(ctx context.Context) (*types.StatusResponse, error)
	UpdateCursor(ctx context.Context, payload types.UpdateCursorPayload) (*types.StatusResponse, error)
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
	httpClient *http.Client
	logger     infrastructure.ILogger
}

func FxStateManagerService() fx.Option {
	return fx_utils.AsProvider(newStateManagerService, new(IStateManagerService))
}

func newStateManagerService(params stateManagerServiceParams) IStateManagerService {
	return &stateManagerService{
		streamId: params.Config.GetStreamId(),
		url:      params.Config.GetStateManagerUrl(),
		token:    params.Config.GetStateManagerToken(),
		logger:   params.Logger,
		httpClient: &http.Client{
			Timeout: params.Config.GetStateManagerTimeout(),
		},
	}
}

func (s *stateManagerService) GetStream(ctx context.Context) (*types.Stream, error) {
	path := fmt.Sprintf("/stream/%s", s.streamId)

	responseBody, err := s.makeGetRequest(ctx, path)

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

	responseBody, err := s.makeGetRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("state manager get status error: %w", err)
	}

	var statusResp types.StatusResponse
	if err := json.Unmarshal(responseBody, &statusResp); err != nil {
		return nil, fmt.Errorf("state manager unmarshal status error: %w", err)
	}

	return &statusResp, nil
}

func (s *stateManagerService) UpdateCursor(ctx context.Context, payload types.UpdateCursorPayload) (*types.StatusResponse, error) {
	path := fmt.Sprintf("/stream/%s/state/cursor", s.streamId)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("state manager update cursor: error marshaling payload: %w", err)
	}

	responseBody, err := s.makePatchRequest(ctx, path, bytes.NewReader(payloadBytes))
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
	_, err := s.makePatchRequest(ctx, path, nil)

	if err != nil {
		return fmt.Errorf("state manager mark as running error: %w", err)
	}

	return nil
}

func (s *stateManagerService) makeGetRequest(ctx context.Context, path string) ([]byte, error) {
	req, err := s.createRequest(ctx, http.MethodGet, path, nil)

	if err != nil {
		return nil, fmt.Errorf("state manager get request error: %w", err)
	}

	responseBody, statusCode, err := s.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("state manager get request error: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("state manager get request error: %s", responseBody)
	}

	return responseBody, nil
}

func (s *stateManagerService) makePostRequest(ctx context.Context, path string, body io.Reader) ([]byte, error) {
	req, err := s.createRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("state manager post request error: %w", err)
	}

	responseBody, statusCode, err := s.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("state manager post request error: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("state manager post request error: %s", responseBody)
	}

	return responseBody, nil
}

func (s *stateManagerService) makePatchRequest(ctx context.Context, path string, body io.Reader) ([]byte, error) {
	req, err := s.createRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, fmt.Errorf("state manager patch request error: %w", err)
	}

	responseBody, statusCode, err := s.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("state manager patch request error: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("state manager patch request error: %s", responseBody)
	}

	return responseBody, nil
}

func (s *stateManagerService) createRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	requestPath := fmt.Sprintf("%s%s", s.url, path)

	req, err := http.NewRequestWithContext(ctx, method, requestPath, body)

	if err != nil {
		return nil, fmt.Errorf("state manager create request error: %w", err)
	}

	req.Header.Set("Authorization", s.token)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (s *stateManagerService) doRequest(req *http.Request) ([]byte, int, error) {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("state manager send request error: %w", err)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("state manager read response error: %w", err)
	}

	return responseBody, resp.StatusCode, nil
}
