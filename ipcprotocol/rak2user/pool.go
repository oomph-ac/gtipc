package rak2user

import "github.com/gameparrot/gtipc/ipcprotocol"

func NewRak2UserPool() map[uint8]func() ipcprotocol.Packet {
	return map[uint8]func() ipcprotocol.Packet{
		IdEncapsulated: func() ipcprotocol.Packet {
			return &Encapsulated{}
		},
		IdCloseSession: func() ipcprotocol.Packet {
			return &CloseSession{}
		},
		IdAckNotification: func() ipcprotocol.Packet {
			return &AckNotification{}
		},
		IdRaw: func() ipcprotocol.Packet {
			return &Raw{}
		},
		IdReportPing: func() ipcprotocol.Packet {
			return &ReportPing{}
		},
	}
}
