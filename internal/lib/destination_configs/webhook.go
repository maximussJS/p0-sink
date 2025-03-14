package destination_configs

import (
	"fmt"
	net_url "net/url"
	"p0-sink/internal/enums"
	"p0-sink/internal/errors"
	map_utils "p0-sink/internal/utils/object"
	validate_utils "p0-sink/internal/utils/validate"
	"time"
)

type WebhookDestinationConfig struct {
	Url                  string
	Timeout              time.Duration
	Headers              map[string]string
	Compression          enums.ECompression
	RollbackBeforeResend bool
}

func NewWebhookDestinationConfig(data map[string]interface{}) (*WebhookDestinationConfig, error) {
	var url string
	if s, ok := map_utils.GetStringFromMap(data, "url"); ok {
		if s == "" {
			return nil, errors.NewInvalidDestinationConfigError("url is required")
		}

		_, err := net_url.Parse(s)

		if err != nil {
			return nil, errors.NewInvalidDestinationConfigError("invalid url")
		}

		url = s
	} else {
		return nil, errors.NewInvalidDestinationConfigError("url is required")
	}

	var timeout time.Duration

	if num, ok := map_utils.GetNumberFromMap(data, "timeout"); ok {
		if num < 0 || num > 60 {
			return nil, errors.NewInvalidDestinationConfigError("invalid timeout")
		}
		timeout = time.Duration(int(num)) * time.Second
	} else {
		timeout = 30 * time.Second
	}

	var headers = make(map[string]string)

	if h, ok := map_utils.GetStringMapFromMap(data, "headers"); ok {
		result := validate_utils.ValidateHeaders(h)

		if result.Valid {
			headers = h
		} else {
			return nil, errors.NewInvalidDestinationConfigError(fmt.Sprintf("invalid headers: %v", result.Message))
		}
	}

	var rollbackBeforeResend bool
	if b, ok := map_utils.GetBoolFromMap(data, "rollbackBeforeResend"); ok {
		rollbackBeforeResend = b
	} else {
		return nil, errors.NewInvalidDestinationConfigError("rollbackBeforeResend is required")
	}

	var compression enums.ECompression
	if s, ok := map_utils.GetStringFromMap(data, "compression"); ok {
		switch s {
		case string(enums.ECompressionNone):
			compression = enums.ECompressionNone
		case string(enums.ECompressionGzip):
			compression = enums.ECompressionGzip
		default:
			return nil, errors.NewInvalidDestinationConfigError(fmt.Sprintf("invalid webhook 	compression: %v", s))
		}
	} else {
		return nil, errors.NewInvalidDestinationConfigError("compression is required")
	}

	return &WebhookDestinationConfig{
		Url:                  url,
		Timeout:              timeout,
		Headers:              headers,
		Compression:          compression,
		RollbackBeforeResend: rollbackBeforeResend,
	}, nil
}
