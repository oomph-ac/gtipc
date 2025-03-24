package rak2user

import "github.com/gameparrot/gtipc/ipcprotocol/io"

const (
	DisconnectReasonClientDisconnect byte = iota
	DisconnectReasonServerDisconnect
	DisconnectReasonPeerTimeout
	DisconnectReasonClientReconnect
	DisconnectReasonServerShutdown
	DisconnectReasonSplitPacketTooLarge
	DisconnectReasonSplitPacketTooManyConcurrent
	DisconnectReasonSplitPacketInvalidPartIndex
	DisconnectReasonSplitPacketInconsistentHeader
)

type CloseSession struct {
	SessionID int32
	Reason    byte
}

func (c *CloseSession) ID() uint8 {
	return IdCloseSession
}

func (c *CloseSession) Marshal(i io.IO, len int) {
	i.BEInt32(&c.SessionID)
	i.Uint8(&c.Reason)
}
