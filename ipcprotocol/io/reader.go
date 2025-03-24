package io

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"unsafe"
)

var errVarIntOverflow = errors.New("varint overflows integer")

// NewReader creates a new Reader using the io.ByteReader passed as underlying source to read bytes from.
func NewReader(r interface {
	io.Reader
	io.ByteReader
}) *Reader {
	return &Reader{r: ReaderWithOffset{orig: r}}
}

type Reader struct {
	r ReaderWithOffset
}

// Offset returns the offset of the underlying buffer.
func (r *Reader) Offset() uint64 {
	return r.r.offset
}

// Uint8 reads a uint8 from the underlying buffer.
func (r *Reader) Uint8(x *uint8) {
	var err error
	*x, err = r.r.ReadByte()
	if err != nil {
		r.panic(err)
	}
}

// Bool reads a bool from the underlying buffer.
func (r *Reader) Bool(x *bool) {
	u, err := r.r.ReadByte()
	if err != nil {
		r.panic(err)
	}
	*x = *(*bool)(unsafe.Pointer(&u))
}

// String reads a string from the underlying buffer.
func (r *Reader) String(x *string) {
	var length uint32
	r.Uint32(&length)
	l := int(length)
	data := make([]byte, l)
	if _, err := r.r.Read(data); err != nil {
		r.panic(err)
	}
	*x = *(*string)(unsafe.Pointer(&data))
}

// Varuint32 reads up to 5 bytes from the underlying buffer into a uint32.
func (r *Reader) Varuint32(x *uint32) {
	var v uint32
	for i := 0; i < 35; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		v |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			*x = v
			return
		}
	}
	r.panic(errVarIntOverflow)
}

// String reads a string from the underlying buffer with a varuint32 length prefix.
func (r *Reader) StringVaruint32(x *string) {
	var length uint32
	r.Varuint32(&length)
	if length == 0 {
		return
	}
	l := int(length)
	data := make([]byte, l)
	if _, err := r.r.Read(data); err != nil {
		r.panic(err)
	}
	*x = *(*string)(unsafe.Pointer(&data))
}

// panic panics with the error passed, similarly to panicf.
func (r *Reader) panic(err error) {
	panic(err)
}

// Uint16 reads a little endian uint16 from the underlying buffer.
func (r *Reader) Uint16(x *uint16) {
	b := make([]byte, 2)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = binary.LittleEndian.Uint16(b)
}

// BEUint16 reads a big endian uint16 from the underlying buffer.
func (r *Reader) BEUint16(x *uint16) {
	b := make([]byte, 2)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = binary.BigEndian.Uint16(b)
}

// Int16 reads a little endian int16 from the underlying buffer.
func (r *Reader) Int16(x *int16) {
	b := make([]byte, 2)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = int16(binary.LittleEndian.Uint16(b))
}

// Uint32 reads a little endian uint32 from the underlying buffer.
func (r *Reader) Uint32(x *uint32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = binary.LittleEndian.Uint32(b)
}

// BEUint32 reads a big endian uint32 from the underlying buffer.
func (r *Reader) BEUint32(x *uint32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = binary.BigEndian.Uint32(b)
}

// Int32 reads a little endian int32 from the underlying buffer.
func (r *Reader) Int32(x *int32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = int32(binary.LittleEndian.Uint32(b))
}

// BEInt32 reads a big endian int32 from the underlying buffer.
func (r *Reader) BEInt32(x *int32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = int32(binary.BigEndian.Uint32(b))
}

// Uint64 reads a little endian uint64 from the underlying buffer.
func (r *Reader) Uint64(x *uint64) {
	b := make([]byte, 8)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = binary.LittleEndian.Uint64(b)
}

// Int64 reads a little endian int64 from the underlying buffer.
func (r *Reader) Int64(x *int64) {
	b := make([]byte, 8)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = int64(binary.LittleEndian.Uint64(b))
}

// BEInt64 reads a big endian int64 from the underlying buffer.
func (r *Reader) BEInt64(x *int64) {
	b := make([]byte, 8)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = int64(binary.BigEndian.Uint64(b))
}

// Float32 reads a little endian float32 from the underlying buffer.
func (r *Reader) Float32(x *float32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = math.Float32frombits(binary.LittleEndian.Uint32(b))
}
