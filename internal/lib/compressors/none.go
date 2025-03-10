package compressors

type noneCompressor struct{}

func NewNoneCompressor() ICompressor {
	return &noneCompressor{}
}

func (c *noneCompressor) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (c *noneCompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}

func (c *noneCompressor) EncodingType() string {
	return ""
}

func (c *noneCompressor) FileExtension() string {
	return ""
}
