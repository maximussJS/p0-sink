package types

import (
	"fmt"
	pb "p0-sink/proto"
)

type DownloadedBlock struct {
	*pb.BlockWrapper
	Data []byte
}

func NewDownloadedBlock(block *pb.BlockWrapper) *DownloadedBlock {
	return &DownloadedBlock{
		BlockWrapper: block,
		Data:         nil,
	}
}

func (b *DownloadedBlock) String() string {
	return fmt.Sprintf("DownloadedBlock{BlockWrapper: %v, Data: %v}", b.BlockWrapper, b.Data)
}

func (b *DownloadedBlock) IsEmpty() bool {
	return b.BlockWrapper == nil || b.Data == nil
}
