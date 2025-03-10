package types

import pb "p0-sink/proto"

type SerializedBatch struct {
	Data              []byte
	Encoding          string
	Cursor            string
	Direction         *pb.Direction
	BlockNumbers      []uint64
	HasRealtimeBlocks bool
	BilledBytes       uint64
	BlockMetadata     string
}

func NewSerializedBatch(data any) *SerializedBatch {
	return &SerializedBatch{
		Data: data.data,
	}
}
