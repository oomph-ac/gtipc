package io

import (
	"io"
)

type ReaderWithOffset struct {
	orig interface {
		io.Reader
		io.ByteReader
	}
	offset uint64
}

func (r *ReaderWithOffset) Read(p []byte) (n int, err error) {
	n, err = r.orig.Read(p)
	r.offset += uint64(n)
	return n, err
}

func (f *ReaderWithOffset) ReadByte() (byte, error) {
	b, err := f.orig.ReadByte()
	if err == nil {
		f.offset++
	}
	return b, err
}
