package compressors

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

type gzipCompressor struct {
}

func NewGzipCompressor() ICompressor {
	return &gzipCompressor{}
}

func (c *gzipCompressor) Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}

	if err = gz.Flush(); err != nil {
		return nil, err
	}

	if err = gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (c *gzipCompressor) Decompress(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %w", err)
	}

	return decompressedData, nil
}

func (c *gzipCompressor) EncodingType() string {
	return "gzip"
}

func (c *gzipCompressor) FileExtension() string {
	return ".gz"
}
