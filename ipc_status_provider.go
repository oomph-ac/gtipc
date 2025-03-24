package gtipc

import (
	_ "unsafe"

	"github.com/sandertv/gophertunnel/minecraft"
)

type IpcStatusProvider struct {
	key     string
	handler IpcHandler
}

func NewIpcStatusProvider(key string, handler IpcHandler) *IpcStatusProvider {
	return &IpcStatusProvider{key: key, handler: handler}
}

func (i *IpcStatusProvider) ServerStatus(int, int) minecraft.ServerStatus {
	conn, ok := i.handler.GetConn(i.key)
	if !ok {
		return minecraft.ServerStatus{}
	}
	return parsePongData(conn.pongData)
}

//go:linkname parsePongData github.com/sandertv/gophertunnel/minecraft.parsePongData
func parsePongData(pong []byte) minecraft.ServerStatus
