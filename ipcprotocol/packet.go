package ipcprotocol

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type Packet interface {
	ID() uint8
	Marshal(io io.IO, len int)
}
