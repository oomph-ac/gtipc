package io

import (
	"encoding/binary"
	"io"
	"math"
	"unsafe"
)

// NewWriter creates a new initialised Writer with an underlying io.ByteWriter to write to.
func NewWriter(w interface {
	io.Writer
	io.ByteWriter
}) *Writer {
	return &Writer{w: w}
}

type Writer struct {
	w interface {
		io.Writer
		io.ByteWriter
	}
}

// Offset...
func (w *Writer) Offset() uint64 {
	return 0
}

// Uint8 writes a uint8 to the underlying buffer.
func (w *Writer) Uint8(x *uint8) {
	_ = w.w.WriteByte(*x)
}

// Bool writes a bool as either 0 or 1 to the underlying buffer.
func (w *Writer) Bool(x *bool) {
	_ = w.w.WriteByte(*(*byte)(unsafe.Pointer(x)))
}

// String writes a string, prefixed with a uint32, to the underlying buffer.
func (w *Writer) String(x *string) {
	l := uint32(len(*x))
	w.Uint32(&l)
	_, _ = w.w.Write([]byte(*x))
}

// Varuint32 writes a uint32 as 1-5 bytes to the underlying buffer.
func (w *Writer) Varuint32(x *uint32) {
	u := *x
	for u >= 0x80 {
		_ = w.w.WriteByte(byte(u) | 0x80)
		u >>= 7
	}
	_ = w.w.WriteByte(byte(u))
}

// String writes a string, prefixed with a varuint32, to the underlying buffer.
func (w *Writer) StringVaruint32(x *string) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	_, _ = w.w.Write([]byte(*x))
}

// Uint16 writes a little endian uint16 to the underlying buffer.
func (w *Writer) Uint16(x *uint16) {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, *x)
	_, _ = w.w.Write(data)
}

// BEUint16 writes a big endian uint16 to the underlying buffer.
func (w *Writer) BEUint16(x *uint16) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, *x)
	_, _ = w.w.Write(data)
}

// Int16 writes a little endian int16 to the underlying buffer.
func (w *Writer) Int16(x *int16) {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(*x))
	_, _ = w.w.Write(data)
}

// Uint32 writes a little endian uint32 to the underlying buffer.
func (w *Writer) Uint32(x *uint32) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, *x)
	_, _ = w.w.Write(data)
}

// BEUint32 writes a big endian uint32 to the underlying buffer.
func (w *Writer) BEUint32(x *uint32) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, *x)
	_, _ = w.w.Write(data)
}

// Int32 writes a little endian int32 to the underlying buffer.
func (w *Writer) Int32(x *int32) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(*x))
	_, _ = w.w.Write(data)
}

// BEInt32 writes a big endian int32 to the underlying buffer.
func (w *Writer) BEInt32(x *int32) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(*x))
	_, _ = w.w.Write(data)
}

// Uint64 writes a little endian uint64 to the underlying buffer.
func (w *Writer) Uint64(x *uint64) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, *x)
	_, _ = w.w.Write(data)
}

// Int64 writes a little endian int64 to the underlying buffer.
func (w *Writer) Int64(x *int64) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(*x))
	_, _ = w.w.Write(data)
}

// BEInt64 writes a big endian int64 to the underlying buffer.
func (w *Writer) BEInt64(x *int64) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(*x))
	_, _ = w.w.Write(data)
}

// Float32 writes a little endian float32 to the underlying buffer.
func (w *Writer) Float32(x *float32) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, math.Float32bits(*x))
	_, _ = w.w.Write(data)
}
