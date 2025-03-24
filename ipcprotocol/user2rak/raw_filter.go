package user2rak

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type RawFilter struct {
	Filter string
}

func (r *RawFilter) ID() uint8 {
	return IdRawFilter
}

func (r *RawFilter) Marshal(i io.IO, len int) {
	io.StringOfLen(&r.Filter, uint32(len), i)
}
