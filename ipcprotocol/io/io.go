package io

import (
	"unsafe"
)

type IO interface {
	BEUint16(x *uint16)
	Uint16(x *uint16)
	Int16(x *int16)
	Uint32(x *uint32)
	Int32(x *int32)
	BEInt32(x *int32)
	BEUint32(x *uint32)
	Uint64(x *uint64)
	Int64(x *int64)
	BEInt64(x *int64)
	Float32(x *float32)
	Uint8(x *uint8)
	Bool(x *bool)
	String(x *string)
	StringVaruint32(x *string)
	Varuint32(x *uint32)

	Offset() uint64
}

// FuncSliceOfLen reads/writes the elements of a slice of type T with length l using func f.
func FuncSliceOfLen[T any, S ~*[]T](r IO, l uint32, x S, f func(*T)) {
	_, reader := r.(*Reader)
	if reader {
		*x = make([]T, l)
	}

	for i := uint32(0); i < l; i++ {
		f(&(*x)[i])
	}
}

// FuncSliceUint8Length reads/writes a slice of T using function f with a uint16 length prefix.
func FuncSliceUint8Length[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint8(len(*x))
	r.Uint8(&count)
	FuncSliceOfLen(r, uint32(count), x, f)
}

// String reads a string from the underlying buffer.
func StringUint8Length(x *string, r IO) {
	var length uint8 = uint8(len(*x))
	r.Uint8(&length)
	StringOfLen(x, uint32(length), r)
}

func StringOfLen(x *string, length uint32, r IO) {
	if length <= 0 {
		length = uint32(len(*x))
	}
	var data = []byte(*x)
	FuncSliceOfLen(r, length, &data, r.Uint8)
	*x = *(*string)(unsafe.Pointer(&data))
}

// Marshaler is a type that can be written to or read from an IO.
type Marshaler interface {
	Marshal(r IO)
}

// Slice reads/writes a slice of T with a varuint32 prefix.
func Slice[T any, S ~*[]T, A PtrMarshaler[T]](r IO, x S) {
	count := uint32(len(*x))
	r.Varuint32(&count)
	SliceOfLen[T, S, A](r, count, x)
}

// FuncSlice reads/writes a slice of T using function f with a varuint32 length prefix.
func FuncSlice[T any, S ~*[]T](r IO, x S, f func(*T)) {
	count := uint32(len(*x))
	r.Varuint32(&count)
	FuncSliceOfLen(r, count, x, f)
}

// SliceOfLen reads/writes the elements of a slice of type T with length l.
func SliceOfLen[T any, S ~*[]T, A PtrMarshaler[T]](r IO, l uint32, x S) {
	_, ok := r.(*Reader)
	if ok {
		*x = make([]T, l)
	}

	for i := uint32(0); i < l; i++ {
		A(&(*x)[i]).Marshal(r)
	}
}

// PtrMarshaler represents a type that implements Marshaler for its pointer.
type PtrMarshaler[T any] interface {
	Marshal(i IO)
	*T
}
