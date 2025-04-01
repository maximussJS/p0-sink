package types

import (
	"fmt"
)

type DownloadedBlock struct {
	Block *Block
	Data  []byte
}

func NewDownloadedBlock(block *Block, data []byte) *DownloadedBlock {
	return &DownloadedBlock{
		Block: block,
		Data:  data,
	}
}

func CopyWithNewData(block *DownloadedBlock, data []byte) *DownloadedBlock {
	return &DownloadedBlock{
		Block: block.Block,
		Data:  data,
	}
}

func (b *DownloadedBlock) String() string {
	return fmt.Sprintf("DownloadedBlock{Block: %v, Data: %v}", b.Block, b.Data)
}

func (b *DownloadedBlock) IsEmpty() bool {
	return b.Block == nil || b.Data == nil
}

func (b *DownloadedBlock) BlockNumber() uint64 {
	return b.Block.BlockNumber
}
