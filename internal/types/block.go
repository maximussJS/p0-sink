package types

import (
	"fmt"
	"net/url"
	pb "p0-sink/proto"
	"strconv"
	"strings"
)

type Block struct {
	*pb.BlockWrapper
	BytesStart uint64
	BytesEnd   uint64
}

func NewBlock(block *pb.BlockWrapper) (*Block, error) {
	urlStr := block.Url

	if strings.HasPrefix(urlStr, "data:") {
		return nil, fmt.Errorf("data URL is not supported, block: %d", block.BlockNumber)
	}

	if !strings.Contains(urlStr, "bytesStart") && !strings.Contains(urlStr, "bytesEnd") {
		return nil, fmt.Errorf("URL does not contain bytesStart or bytesEnd, block: %d", block.BlockNumber)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v, block: %d", err, block.BlockNumber)
	}

	queryParams := parsedURL.Query()
	bytesStart := queryParams.Get("bytesStart")
	bytesEnd := queryParams.Get("bytesEnd")

	if bytesStart == "" || bytesEnd == "" {
		return nil, fmt.Errorf("bytesStart or bytesEnd is empty, block: %d", block.BlockNumber)
	}

	bytesStartInt, err := strconv.ParseUint(bytesStart, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bytesStart: %v, block: %d", err, block.BlockNumber)
	}

	bytesEndInt, err := strconv.ParseUint(bytesEnd, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bytesEnd: %v, block: %d", err, block.BlockNumber)
	}

	return &Block{
		BlockWrapper: block,
		BytesStart:   bytesStartInt,
		BytesEnd:     bytesEndInt,
	}, nil
}

func (b *Block) BatchNumber() string {
	batchNumber := b.BlockNumber / 100
	return fmt.Sprintf("%010d", batchNumber)
}
