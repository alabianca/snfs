package fs

import (
	"hash"
	"io"
)

type StorageWriter interface {
	io.Closer
	hash.Hash
}

func NewWriter(hash hash.Hash, writer io.WriteCloser) StorageWriter {
	return newWriter(hash, writer)
}

type writer struct {
	hash   hash.Hash
	file   io.WriteCloser
	writer io.Writer
}

func newWriter(hash hash.Hash, file io.WriteCloser) *writer {
	return &writer{
		hash:   hash,
		file:   file,
		writer: io.MultiWriter(file, hash),
	}
}

func (w *writer) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func (w *writer) Close() error {
	return w.file.Close()
}

func (w *writer) Sum(b []byte) []byte {
	return w.hash.Sum(b)
}

func (w *writer) Reset() {
	w.hash.Reset()
}

func (w *writer) Size() int {
	return w.hash.Size()
}

func (w *writer) BlockSize() int {
	return w.hash.BlockSize()
}
