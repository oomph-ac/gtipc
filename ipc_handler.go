package gtipc

import (
	"net"
	"time"

	"github.com/sandertv/gophertunnel/minecraft"
)

type IpcHandler interface {
	minecraft.Network

	// BlockAddress blocks an IP address from accessing the server
	BlockAddress(addr net.IP, duration time.Duration)

	// UnblockAddress allows a blocked IP address to access te server
	UnblockAddress(addr net.IP)

	handleCustomPacket(b []byte, serverKey string)

	GetConn(key string) (*Conn, bool)
}
