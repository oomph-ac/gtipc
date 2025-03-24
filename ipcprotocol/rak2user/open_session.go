package rak2user

import (
	"net"

	"github.com/gameparrot/gtipc/ipcprotocol/io"
)

type OpenSession struct {
	SessionID int32
	Addr      net.IP
	Port      uint16
	ClientID  int64
}

func (o *OpenSession) ID() uint8 {
	return IdOpenSession
}

func (o *OpenSession) Marshal(i io.IO, len int) {
	i.BEInt32(&o.SessionID)
	addr := []byte(o.Addr)
	io.FuncSliceUint8Length(i, &addr, i.Uint8)
	i.BEUint16(&o.Port)
	i.BEInt64(&o.ClientID)
}
