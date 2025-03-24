package user2rak

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type CloseSession struct {
	SessionID int32
}

func (c *CloseSession) ID() uint8 {
	return IdCloseSession
}

func (c *CloseSession) Marshal(i io.IO, len int) {
	i.BEInt32(&c.SessionID)
}
