package rak2user

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type Encapsulated struct {
	SessionID   int32
	UserPayload []byte
}

func (e *Encapsulated) ID() uint8 {
	return IdEncapsulated
}

func (e *Encapsulated) Marshal(i io.IO, length int) {
	i.BEInt32(&e.SessionID)

	pkLen := len(e.UserPayload)
	if i.Offset() != 0 {
		pkLen = length - int(i.Offset())
	}
	io.FuncSliceOfLen(i, uint32(pkLen), &e.UserPayload, i.Uint8)
}
