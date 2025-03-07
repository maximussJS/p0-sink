package enums

type ECompression string

const (
	ECompressionNone   ECompression = "none"
	ECompressionGzip   ECompression = "gzip"
	ECompressionSnappy ECompression = "snappy"
	ECompressionLz4    ECompression = "lz4"
)
