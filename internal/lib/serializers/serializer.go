package serializers

import (
	"p0-sink/internal/types"
	pb "p0-sink/proto"
)

type ISerializer interface {
	Serialize(blocks []*types.DownloadedBlock, direction *pb.Direction) ([]byte, int, error)
}
