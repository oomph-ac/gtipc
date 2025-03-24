package user2rak

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type BlockAddress struct {
	Addr    string
	Timeout uint32
}

func (b *BlockAddress) ID() uint8 {
	return IdBlockAddress
}

func (b *BlockAddress) Marshal(i io.IO, len int) {
	io.StringUint8Length(&b.Addr, i)
	i.BEUint32(&b.Timeout)
}
