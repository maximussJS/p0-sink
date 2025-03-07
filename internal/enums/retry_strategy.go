package enums

type ERetryStrategy string

const (
	ERetryStrategyFixed       ERetryStrategy = "fixed"
	ERetryStrategyLinear      ERetryStrategy = "linear"
	ERetryStrategyExponential ERetryStrategy = "exponential"
)
