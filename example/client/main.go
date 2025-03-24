package main

import (
	"log/slog"

	"github.com/gameparrot/gtipc"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func main() {
	ipc := gtipc.NewIpcClient(nil)
	minecraft.RegisterNetwork("ipc", func(l *slog.Logger) minecraft.Network {
		return ipc
	})
	dial, err := minecraft.Dialer{
		IdentityData:        login.IdentityData{XUID: "4357534"},
		KeepXBLIdentityData: true,
	}.Dial("ipc", "/tmp/gtipc.sock")
	if err != nil {
		panic(err)
	}
	dial.WritePacket(&packet.Text{TextType: packet.TextTypeChat, Message: "Hello, world!"})
	dial.Close()
}
