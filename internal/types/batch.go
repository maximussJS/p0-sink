package types

import (
	"fmt"
	pb "p0-sink/proto"
	"time"
)

type Batch struct {
	start     time.Time
	direction *pb.Direction
	cursor    string
	blocks    []*DownloadedBlock
}

func NewBatch() *Batch {
	return &Batch{
		start:  time.Now(),
		blocks: make([]*DownloadedBlock, 0),
	}
}

func (b *Batch) PushBlock(block *DownloadedBlock) error {
	if err := b.verifyDirection(block); err != nil {
		return err
	}

	b.blocks = append(b.blocks, block)
	b.cursor = block.Cursor
	return nil
}

func (b *Batch) GetCursor() (string, error) {
	if !b.HasBlocks() {
		return "", fmt.Errorf("cannot get cursor from empty batch")
	}

	return b.cursor, nil
}

func (b *Batch) GetBlockNumbers() ([]uint64, error) {
	if !b.HasBlocks() {
		return nil, fmt.Errorf("cannot get block numbers from empty batch")
	}

	blockNumbers := make([]uint64, len(b.blocks))
	for i, block := range b.blocks {
		blockNumbers[i] = block.BlockNumber
	}

	return blockNumbers, nil
}

func (b *Batch) GetData() ([][]byte, error) {
	if !b.HasBlocks() {
		return nil, fmt.Errorf("cannot get data from empty batch")
	}

	data := make([][]byte, 0)
	for _, block := range b.blocks {
		data = append(data, block.Data)
	}

	return data, nil
}

func (b *Batch) GetDirection() (*pb.Direction, error) {
	if !b.HasBlocks() {
		return nil, fmt.Errorf("cannot get cursor from empty batch")
	}

	return b.direction, nil
}

func (b *Batch) GetBlockRange() (uint64, uint64, error) {
	if !b.HasBlocks() {
		return 0, 0, fmt.Errorf("cannot get block range from empty batch")
	}

	from := b.blocks[0].BlockNumber
	to := b.blocks[len(b.blocks)-1].BlockNumber

	if to > from {
		return from, to, nil
	}

	return to, from, nil
}

func (b *Batch) HasBlocks() bool {
	return len(b.blocks) > 0
}

func (b *Batch) Len() int {
	return len(b.blocks)
}

func (b *Batch) HasHeadBlock() bool {
	for _, block := range b.blocks {
		if block.IsHead {
			return true
		}
	}

	return false
}

func (b *Batch) TimeElapsed() time.Duration {
	return time.Since(b.start)
}

func (b *Batch) String() string {
	from, to, _ := b.GetBlockRange()
	return fmt.Sprintf("Batch {Direction: %v, Range: %d-%d }", b.direction, from, to)
}

func (b *Batch) verifyDirection(block *DownloadedBlock) error {
	if b.direction == nil {
		b.direction = &block.Direction
		return nil
	}

	if b.direction.String() != block.Direction.String() {
		return fmt.Errorf("block direction %s does not match batch direction %s", block.Direction.String(), b.direction.String())
	}

	return nil
}
