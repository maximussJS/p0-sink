package payload_builders

import pb "p0-sink/proto"

type IPayloadBuilder interface {
	Build(blocks [][]byte, direction *pb.Direction) ([]byte, error)
	CalculateOffsets(blocks [][]byte, direction *pb.Direction) []int
	GetPayloadStartLength(direction *pb.Direction) int
	GetPayloadEndLength() int
	GetSeparatorLength() int
}
