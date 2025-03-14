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

func (s *NoopSerializer) Serialize(batch *types.Batch, encoding string) (*types.SerializedBatch, error) {
	blocks, err := batch.GetData()

	if err != nil {
		return nil, fmt.Errorf("noop seralizer: cannot get data from batch: %w", err)
	}

	size := 0

	for _, block := range blocks {
		size += len(block)
	}

	var direction *pb.Direction

	if s.reorgAction == pb.ReorgAction_ROLLBACK_AND_RESEND {
		direction, err = batch.GetDirection()

		if err != nil {
			return nil, fmt.Errorf("noop seralizer: cannot get direction from batch: %w", err)
		}
	} else {
		direction = nil
	}

	payload, err := s.payloadBuilder.Build(blocks, direction)

	if err != nil {
		return nil, fmt.Errorf("noop seralizer: cannot build payload: %w", err)
	}

	return types.NewSerializedBatch(batch, payload, encoding, uint64(size))
}
