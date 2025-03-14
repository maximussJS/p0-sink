package serializers

import (
	"fmt"
	"p0-sink/internal/enums"
	"p0-sink/internal/lib/payload_builders"
	"p0-sink/internal/types"
	pb "p0-sink/proto"
)

type NoopSerializer struct {
	compression    enums.ECompression
	reorgAction    pb.ReorgAction
	payloadBuilder payload_builders.IPayloadBuilder
}

func NewNoopSerializer(
	compression enums.ECompression,
	reorgAction pb.ReorgAction,
	payloadBuilder payload_builders.IPayloadBuilder,
) ISerializer {
	return &NoopSerializer{
		compression:    compression,
		reorgAction:    reorgAction,
		payloadBuilder: payloadBuilder,
	}
}

func (s *NoopSerializer) Serialize(blocks []*types.DownloadedBlock, direction *pb.Direction) ([]byte, int, error) {
	size := 0
	data := make([][]byte, 0)

	for _, block := range blocks {
		size += len(block.Data)
		data = append(data, block.Data)
	}

	serialized, err := s.payloadBuilder.Build(data, direction)

	if err != nil {
		return nil, 0, fmt.Errorf("cannot serialize the blocks: %w", err)
	}

	return serialized, size, nil
}
