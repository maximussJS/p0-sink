package types

import (
	"fmt"
	"p0-sink/internal/enums"
	directon_utils "p0-sink/internal/utils/direction"
)

type ProcessedBatch struct {
	Data              []byte
	Encoding          string
	Cursor            string
	Direction         enums.DirectionString
	BlockNumbers      []uint64
	HasRealtimeBlocks bool
	BilledBytes       uint64
}

func NewProcessedBatch(batch *Batch, serializedData []byte, encoding string, billedBytes uint64) (*ProcessedBatch, error) {
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

	return &ProcessedBatch{
		Data:              serializedData,
		Encoding:          encoding,
		Cursor:            cursor,
		Direction:         directon_utils.DirectionToArrow(direction),
		BlockNumbers:      blockNumbers,
		HasRealtimeBlocks: batch.HasHeadBlock(),
		BilledBytes:       billedBytes,
	}, nil
}

func (s *ProcessedBatch) LastBlockNumber() uint64 {
	return s.BlockNumbers[len(s.BlockNumbers)-1]
}

func (s *ProcessedBatch) FirstBlockNumber() uint64 {
	return s.BlockNumbers[0]
}

func (s *ProcessedBatch) NumBlocks() int {
	return len(s.BlockNumbers)
}

func (s *ProcessedBatch) String() string {
	return fmt.Sprintf("Batch [%d-%d]", s.BlockNumbers[0], s.BlockNumbers[len(s.BlockNumbers)-1])
}

func (s *ProcessedBatch) LongString() string {
	return fmt.Sprintf("Batch [%v]", s.BlockNumbers)
}
