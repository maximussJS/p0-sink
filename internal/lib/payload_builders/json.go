package payload_builders

import (
	"fmt"
	"p0-sink/internal/enums"
	pb "p0-sink/proto"
)

type JsonPayloadBuilder struct {
	startForward     []byte
	startBackward    []byte
	startNoDirection []byte
	separator        []byte
	end              []byte
}

func NewJsonPayloadBuilder() IPayloadBuilder {
	builder := &JsonPayloadBuilder{}

	builder.startBackward = builder.toBytes(fmt.Sprintf(`{"direction":"%s","data":[`, enums.DirectionStringBackward))
	builder.startForward = builder.toBytes(fmt.Sprintf(`{"direction":"%s","data":[`, enums.DirectionStringForward))
	builder.startNoDirection = builder.toBytes(`{"data":[`)
	builder.separator = builder.toBytes(`,`)
	builder.end = builder.toBytes(`]}`)

	return builder
}

func (j *JsonPayloadBuilder) Build(blocks [][]byte, direction *pb.Direction) ([]byte, error) {
	if direction == nil {
		return j.combineBytes(j.startNoDirection, blocks), nil
	}

	switch direction.String() {
	case pb.Direction_FORWARD.String():
		return j.combineBytes(j.startForward, blocks), nil
	case pb.Direction_BACKWARD.String():
		return j.combineBytes(j.startBackward, blocks), nil
	default:
		return nil, fmt.Errorf("cannot build the payload: unknown direction %s", direction)
	}
}

func (j *JsonPayloadBuilder) combineBytes(start []byte, bytes [][]byte) []byte {
	var sequence = make([]byte, 0)

	for i := 0; i < len(bytes); i++ {
		if i > 0 {
			sequence = append(sequence, j.separator...)
		} else {
			sequence = append(sequence, start...)
		}

		sequence = append(sequence, bytes[i]...)
	}

	sequence = append(sequence, j.end...)

	return sequence
}

func (j *JsonPayloadBuilder) GetPayloadStartLength(direction *pb.Direction) int {
	if direction == nil {
		return len(j.startNoDirection)
	}

	if direction.String() == pb.Direction_FORWARD.String() {
		return len(j.startForward)
	} else {
		return len(j.startBackward)
	}
}

func (j *JsonPayloadBuilder) GetPayloadEndLength() int {
	return len(j.end)
}

func (j *JsonPayloadBuilder) GetSeparatorLength() int {
	return len(j.separator)
}

func (j *JsonPayloadBuilder) CalculateOffsets(blocks [][]byte, direction *pb.Direction) []int {
	offsets := make([]int, len(blocks))
	if len(blocks) == 0 {
		return offsets
	}

	offsets[0] = j.GetPayloadStartLength(direction)

	for i := 1; i < len(blocks); i++ {
		offsets[i] = offsets[i-1] + len(blocks[i-1]) + j.GetSeparatorLength()
	}

	return offsets
}

func (j *JsonPayloadBuilder) toBytes(data string) []byte {
	return []byte(data)
}
