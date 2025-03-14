package serializers

import "p0-sink/internal/types"

type ISerializer interface {
	Serialize(batch *types.Batch, encoding string) (*types.SerializedBatch, error)
}
