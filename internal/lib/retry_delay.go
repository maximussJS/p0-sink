package lib

import (
	"fmt"
	"math"
	"p0-sink/internal/enums"
	"time"
)

type RetryDelay struct {
	strategy enums.ERetryStrategy
}

func NewRetryDelay(strategy enums.ERetryStrategy) *RetryDelay {
	return &RetryDelay{
		strategy: strategy,
	}
}

func (rd *RetryDelay) CalculateDelay(attempt int, delay time.Duration) time.Duration {
	switch rd.strategy {
	case enums.ERetryStrategyFixed:
		return rd.fixed(attempt, delay)
	case enums.ERetryStrategyLinear:
		return rd.linear(attempt, delay)
	case enums.ERetryStrategyExponential:
		return rd.exponential(attempt, delay)
	default:
		panic(fmt.Sprintf("cannot calculate delay for unknown strategy: %v", rd.strategy))
	}
}

func (rd *RetryDelay) fixed(_ int, delay time.Duration) time.Duration {
	return delay
}

func (rd *RetryDelay) linear(attempt int, delay time.Duration) time.Duration {
	if attempt == 0 || attempt == 1 {
		return delay
	}

	return delay * time.Duration(attempt)
}

func (rd *RetryDelay) exponential(attempt int, delay time.Duration) time.Duration {
	if attempt == 0 || attempt == 1 {
		return delay
	}

	return delay * time.Duration(math.Pow(2, float64(attempt-1)))
}
