package serializers

import (
	"p0-sink/internal/enums"
	"p0-sink/internal/types"
	pb "p0-sink/proto"
)

type NoopSerializer struct{}

func NewNoopSerializer(compression enums.ECompression, reorgAction pb.ReorgAction) ISerializer {
	return &NoopSerializer{}
}

func (s *NoopSerializer) Serialize(batch *types.Batch) (*types.SerializedBatch, error) {
	size, err := batch.GetData()

	if err != nil {
		return nil, err
	}
}
