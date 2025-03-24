package user2rak

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type UnblockAddress struct {
	Addr string
}

func (b *UnblockAddress) ID() uint8 {
	return IdUnblockAddress
}

func (b *UnblockAddress) Marshal(i io.IO, len int) {
	io.StringUint8Length(&b.Addr, i)
}
