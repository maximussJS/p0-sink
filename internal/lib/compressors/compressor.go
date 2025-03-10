package compressors

type ICompressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	EncodingType() string
	FileExtension() string
}
