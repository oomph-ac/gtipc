package rak2user

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type AckNotification struct {
	SessionID int32
	ACK       int32
}

func (a *AckNotification) ID() uint8 {
	return IdAckNotification
}

func (a *AckNotification) Marshal(i io.IO, len int) {
	i.BEInt32(&a.SessionID)
	i.BEInt32(&a.ACK)
}
