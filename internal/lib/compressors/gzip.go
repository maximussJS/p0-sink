package compressors

import (
	"bytes"
	"compress/gzip"
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
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return resB.Bytes(), nil
}

func (c *gzipCompressor) EncodingType() string {
	return "gzip"
}

func (c *gzipCompressor) FileExtension() string {
	return ".gz"
}
