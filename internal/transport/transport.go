package transport

import (
	"io"
)

// QilinIO wraps ReadWriteCloseDoner to provide a type that can be used in Qilin's transport layer.
type QilinIO struct {
	_     struct{}
	Inner io.ReadWriteCloser
}

func (q *QilinIO) Read(p []byte) (n int, err error) {
	return q.Inner.Read(p)
}

func (q *QilinIO) Write(p []byte) (n int, err error) {
	return q.Inner.Write(p)
}

func (q *QilinIO) Close() error {
	return q.Inner.Close()
}

func NewQilinIO(inner io.ReadWriteCloser) *QilinIO {
	return &QilinIO{Inner: inner}
}
