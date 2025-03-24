package destinations

import (
	"bytes"
	"context"
	"fmt"
	"p0-sink/internal/lib"
	"p0-sink/internal/lib/destination_configs"
	"p0-sink/internal/types"
	"strconv"
)

type WebhookDestination struct {
	streamId      string
	httpUserAgent string
	config        *destination_configs.WebhookDestinationConfig
	httpClient    *lib.HttpClient
}

func NewWebhookDestination(streamId string, destinationConfig map[string]interface{}) (IDestination, error) {
	config, err := destination_configs.NewWebhookDestinationConfig(destinationConfig)

	if err != nil {
		return nil, err
	}

	return &WebhookDestination{
		streamId:   streamId,
		config:     config,
		httpClient: lib.NewHttpClient(config.Url, config.Headers, config.Timeout),
	}, nil
}

func (d *WebhookDestination) String() string {
	return fmt.Sprintf("webhook destination: url %s, timeout %s", d.config.Url, d.config.Timeout)
}

func (d *WebhookDestination) Send(ctx context.Context, batch *types.ProcessedBatch) error {

	_, err := d.httpClient.Post(ctx, "", d.getBatchHeaders(batch), bytes.NewReader(batch.Data))

	if err != nil {
		return fmt.Errorf("failed to send batch to webhook destination: %v", err)
	}

	return nil
}

func (d *WebhookDestination) getBatchHeaders(batch *types.ProcessedBatch) map[string]string {
	headers := make(map[string]string)

	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json, text/plain, */*"
	headers["User-Agent"] = d.httpUserAgent

	headers["X-Block-Stream-ID"] = d.streamId
	headers["X-Block-Stream-From-Block"] = strconv.FormatUint(batch.FirstBlockNumber(), 10)
	headers["X-Block-Stream-To-Block"] = strconv.FormatUint(batch.LastBlockNumber(), 10)

	if d.config.RollbackBeforeResend {
		headers["X-Block-Stream-Direction"] = string(batch.Direction)
	}

	if batch.Encoding != "" {
		headers["Content-Encoding"] = batch.Encoding
		headers["Vary"] = "Accept-Encoding"
	}

	return headers
}
