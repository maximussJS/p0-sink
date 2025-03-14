package destinations

import (
	"context"
	"p0-sink/internal/types"
)

type IDestination interface {
	Send(ctx context.Context, batch *types.SerializedBatch) error
	String() string
}
