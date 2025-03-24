package user2rak

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type Raw struct {
	Addr    string
	Port    uint16
	Payload []byte
}

func (r *Raw) ID() uint8 {
	return IdRaw
}

func (r *Raw) Marshal(i io.IO, length int) {
	io.StringUint8Length(&r.Addr, i)
	i.BEUint16(&r.Port)

	pkLen := len(r.Payload)
	if i.Offset() != 0 {
		pkLen = length - int(i.Offset())
	}
	io.FuncSliceOfLen(i, uint32(pkLen), &r.Payload, i.Uint8)
}
