package types

import (
	"fmt"
	pb "p0-sink/proto"
)

type DownloadedBlock struct {
	*pb.BlockWrapper
	Data []byte
}

func (b *DownloadedBlock) String() string {
	return fmt.Sprintf("DownloadedBlock{BlockWrapper: %v, Data: %v}", b.BlockWrapper, b.Data)
}
