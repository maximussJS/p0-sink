package serializers

import "p0-sink/internal/types"

type ISerializer interface {
	Serialize(data *types.Batch) (*types.SerializedBatch, error)
}
