package destinations

import (
	"context"
	"p0-sink/internal/types"
)

type NoopDestination struct{}

func NewNoopDestination() IDestination {
	return &NoopDestination{}
}

func (d *NoopDestination) Send(_ context.Context, _ *types.SerializedBatch) error {
	return nil
}

func (d *NoopDestination) String() string {
	return "noop destination"
}
