package pgproto3

import (
	"io"

	"github.com/jackc/chunkreader/v2"
)

// ChunkReader is an interface to decouple github.com/jackc/chunkreader from this package.

// ChunkReader 是一个接口，用于将 github.com/jackc/chunkreader 与此包解耦。
type ChunkReader interface {
	// Next returns buf filled with the next n bytes. If an error (including a partial read) occurs,
	// buf must be nil. Next must preserve any partially read data. Next must not reuse buf.

	// Next 返回 buf 填充下一个 n 字节。
	// 如果发生错误（包括读取部分），buf 必须为 nil。
	// Next 必须保留任何部分读取的数据。Next 不得重用 buf。
	Next(n int) (buf []byte, err error)
}

// NewChunkReader creates and returns a new default ChunkReader.
func NewChunkReader(r io.Reader) ChunkReader {
	return chunkreader.New(r)
}
