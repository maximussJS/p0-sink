package lib

import (
	"fmt"
	"p0-sink/internal/enums"
	"p0-sink/internal/errors"
	map_utils "p0-sink/internal/utils/object"
	"time"
)

type SinkConfig struct {
	RetryAttempts        int
	RetryDelay           time.Duration
	RetryStrategy        enums.ERetryStrategy
	RollbackBeforeResend bool
	ResendOnReorg        bool
	Compression          enums.ECompression
}

func NewSinkConfig(data map[string]interface{}) (*SinkConfig, error) {
	var retryAttempts int
	if num, ok := map_utils.GetNumberFromMap(data, "retryAttempts"); ok {
		if num < 0 || num > 99 {
			return nil, errors.NewInvalidSinkConfigError(fmt.Sprintf("invalid retryAttempts: %v", num))
		}
		retryAttempts = int(num)
	} else {
		retryAttempts = 3
	}

	var retryDelay time.Duration
	if num, ok := map_utils.GetNumberFromMap(data, "retryDelay"); ok {
		if num < 0 || num > 60 {
			return nil, errors.NewInvalidSinkConfigError(fmt.Sprintf("invalid retryDelay: %v", num))
		}
		retryDelay = time.Duration(int(num)) * time.Second
	} else {
		retryDelay = 1 * time.Second
	}

	var retryStrategy enums.ERetryStrategy
	if s, ok := map_utils.GetStringFromMap(data, "retryStrategy"); ok {
		switch s {
		case string(enums.ERetryStrategyFixed):
			retryStrategy = enums.ERetryStrategyFixed
		case string(enums.ERetryStrategyLinear):
			retryStrategy = enums.ERetryStrategyLinear
		case string(enums.ERetryStrategyExponential):
			retryStrategy = enums.ERetryStrategyExponential
		default:
			return nil, errors.NewInvalidSinkConfigError(fmt.Sprintf("invalid retryStrategy: %v", s))
		}
	} else {
		retryStrategy = enums.ERetryStrategyFixed
	}

	var rollbackBeforeResend bool
	if b, ok := map_utils.GetBoolFromMap(data, "rollbackBeforeResend"); ok {
		rollbackBeforeResend = b
	} else {
		rollbackBeforeResend = false
	}

	var compression enums.ECompression
	if s, ok := map_utils.GetStringFromMap(data, "compression"); ok {
		switch s {
		case string(enums.ECompressionNone):
			compression = enums.ECompressionNone
		case string(enums.ECompressionGzip):
			compression = enums.ECompressionGzip
		case string(enums.ECompressionSnappy):
			compression = enums.ECompressionSnappy
		case string(enums.ECompressionLz4):
			compression = enums.ECompressionLz4
		default:
			return nil, errors.NewInvalidSinkConfigError(fmt.Sprintf("invalid compression: %v", s))
		}
	} else {
		compression = enums.ECompressionNone
	}

	var resendOnReorg bool
	if b, ok := map_utils.GetBoolFromMap(data, "resendOnReorg"); ok {
		resendOnReorg = b
	} else {
		resendOnReorg = false
	}

	return &SinkConfig{
		RetryAttempts:        retryAttempts,
		RetryDelay:           retryDelay,
		RetryStrategy:        retryStrategy,
		RollbackBeforeResend: rollbackBeforeResend,
		ResendOnReorg:        resendOnReorg,
		Compression:          compression,
	}, nil
}
