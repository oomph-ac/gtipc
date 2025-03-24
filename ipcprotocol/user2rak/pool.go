package user2rak

import "github.com/gameparrot/gtipc/ipcprotocol"

func NewUser2RakPool() map[uint8]func() ipcprotocol.Packet {
	return map[uint8]func() ipcprotocol.Packet{
		IdEncapsulated: func() ipcprotocol.Packet {
			return &Encapsulated{}
		},
		IdCloseSession: func() ipcprotocol.Packet {
			return &CloseSession{}
		},
		IdRaw: func() ipcprotocol.Packet {
			return &Raw{}
		},
		IdBlockAddress: func() ipcprotocol.Packet {
			return &BlockAddress{}
		},
		IdUnblockAddress: func() ipcprotocol.Packet {
			return &UnblockAddress{}
		},
		IdRawFilter: func() ipcprotocol.Packet {
			return &RawFilter{}
		},
		IdSetName: func() ipcprotocol.Packet {
			return &SetName{}
		},
	}
}
