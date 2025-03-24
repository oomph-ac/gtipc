package user2rak

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type SetName struct {
	Name []byte
}

func (s *SetName) ID() uint8 {
	return IdSetName
}

func (s *SetName) Marshal(i io.IO, len int) {
	io.FuncSliceOfLen(i, uint32(len), &s.Name, i.Uint8)
}
