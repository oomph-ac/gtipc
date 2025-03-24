package user2rak

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type Encapsulated struct {
	SessionID    int32
	Flags        byte
	Reliability  byte
	Ack          int32
	OrderChannel byte
	UserPayload  []byte
}

func (e *Encapsulated) ID() uint8 {
	return IdEncapsulated
}

func (e *Encapsulated) Marshal(i io.IO, length int) {
	i.BEInt32(&e.SessionID)

	i.Uint8(&e.Flags)
	i.Uint8(&e.Reliability)

	if (e.Flags & (1 << 0)) != 0 {
		i.BEInt32(&e.Ack)
	}

	if e.Reliability != 0 && e.Reliability != 5 {
		i.Uint8(&e.OrderChannel)
	}

	pkLen := len(e.UserPayload)
	if i.Offset() != 0 {
		pkLen = length - int(i.Offset())
	}
	io.FuncSliceOfLen(i, uint32(pkLen), &e.UserPayload, i.Uint8)
}
