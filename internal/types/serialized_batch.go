package types

import (
	"fmt"
	pb "p0-sink/proto"
)

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

func NewSerializedBatch(batch *Batch, serializedData []byte, encoding, metadata string, billedBytes uint64) (*SerializedBatch, error) {
	cursor, err := batch.GetCursor()

	if err != nil {
		return nil, fmt.Errorf("failed to get cursor from batch: %v", err)
	}

	direction, err := batch.GetDirection()

	if err != nil {
		return nil, fmt.Errorf("failed to get direction from batch: %v", err)
	}

	blockNumbers, err := batch.GetBlockNumbers()

	if err != nil {
		return nil, fmt.Errorf("failed to get block numbers from batch: %v", err)
	}

	return &SerializedBatch{
		Data:              serializedData,
		Encoding:          encoding,
		Cursor:            cursor,
		Direction:         direction,
		BlockNumbers:      blockNumbers,
		HasRealtimeBlocks: batch.HasHeadBlock(),
		BilledBytes:       billedBytes,
		BlockMetadata:     metadata,
	}, nil
}

func (s *SerializedBatch) LastBlockNumber() (uint64, error) {
	if len(s.BlockNumbers) == 0 {
		return 0, fmt.Errorf("cannot get last block number from empty batch")
	}

	return s.BlockNumbers[len(s.BlockNumbers)-1], nil
}

func (s *SerializedBatch) FirstBlockNumber() (uint64, error) {
	if len(s.BlockNumbers) == 0 {
		return 0, fmt.Errorf("cannot get first block number from empty batch")
	}

	return s.BlockNumbers[0], nil
}

func (s *SerializedBatch) NumBlocks() int {
	return len(s.BlockNumbers)
}

func (s *SerializedBatch) String() string {
	return fmt.Sprintf("SerializedBatch [%d-%d]", s.BlockNumbers[0], s.BlockNumbers[len(s.BlockNumbers)-1])
}
