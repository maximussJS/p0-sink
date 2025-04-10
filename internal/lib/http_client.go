package lib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

var EmptyHttpHeaders = make(map[string]string)

type HttpClient struct {
	client  *http.Client
	url     string
	headers map[string]string
}

func NewHttpClient(url string, headers map[string]string, timeout time.Duration) *HttpClient {
	return &HttpClient{
		url:     url,
		headers: headers,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func NewEmptyHttpClient() *HttpClient {
	return &HttpClient{
		headers: make(map[string]string),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url: "",
	}
}

func NewHttpClientWithDisabledCompression() *HttpClient {
	return &HttpClient{
		headers: make(map[string]string),
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DisableCompression: true,
			},
		},
		url: "",
	}
}

func (s *HttpClient) Get(ctx context.Context, path string, headers map[string]string) ([]byte, error) {
	req, err := s.createRequest(ctx, http.MethodGet, path, headers, nil)

	if err != nil {
		return nil, fmt.Errorf("http client get request error: %w", err)
	}

	responseBody, statusCode, err := s.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("http client get request error: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("http client get request error: %s", responseBody)
	}

	return responseBody, nil
}

func (s *HttpClient) Post(ctx context.Context, path string, headers map[string]string, body io.Reader) ([]byte, error) {
	req, err := s.createRequest(ctx, http.MethodPost, path, headers, body)
	if err != nil {
		return nil, fmt.Errorf("http client post request error: %w", err)
	}

	responseBody, statusCode, err := s.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("http client post request error: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("http client post request error: %s", responseBody)
	}

	return responseBody, nil
}

func (s *HttpClient) Patch(ctx context.Context, path string, headers map[string]string, body io.Reader) ([]byte, error) {
	req, err := s.createRequest(ctx, http.MethodPatch, path, headers, body)
	if err != nil {
		return nil, fmt.Errorf("http client patch request error: %w", err)
	}

	responseBody, statusCode, err := s.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("http client patch request error: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("http client patch request error: %s", responseBody)
	}

	return responseBody, nil
}

func (s *HttpClient) createRequest(ctx context.Context, method, path string, headers map[string]string, body io.Reader) (*http.Request, error) {
	requestPath := fmt.Sprintf("%s%s", s.url, path)

	req, err := http.NewRequestWithContext(ctx, method, requestPath, body)

	if err != nil {
		return nil, fmt.Errorf("http client create request error: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	for key, value := range s.headers {
		req.Header.Set(key, value)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (s *HttpClient) doRequest(req *http.Request) ([]byte, int, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("http client send request error: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("http client close response body error: %v\n", err)
		}
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("http client read response error: %w", err)
	}

	return responseBody, resp.StatusCode, nil
}
