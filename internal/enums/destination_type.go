package enums

type EDestinationType string

const (
	EDestinationTypeWebhook  EDestinationType = "webhook"
	EDestinationTypeS3       EDestinationType = "s3"
	EDestinationTypeKafka    EDestinationType = "kafka"
	EDestinationTypeNoop     EDestinationType = "noop"
	EDestinationTypePostgres EDestinationType = "postgres"
)
