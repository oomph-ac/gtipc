package main

import (
	"fmt"
	"log/slog"

	"github.com/gameparrot/gtipc"
	"github.com/sandertv/gophertunnel/minecraft"
)

func main() {
	ipc, err := gtipc.NewIPCServer("/tmp/gtipc.sock", &gtipc.IpcOptions{
		Upstream: gtipc.CreateGophertunnelUpstreamHandler("raknetupstream"),
	})
	if err != nil {
		panic(err)
	}
	minecraft.RegisterNetwork("ipc", func(l *slog.Logger) minecraft.Network {
		return ipc
	})

	list, err := minecraft.ListenConfig{
		StatusProvider: gtipc.NewIpcStatusProvider("default", ipc),
	}.Listen("raknetupstream", "0.0.0.0:19132")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := list.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn.(*minecraft.Conn), list)
	}
}

func handleConn(conn *minecraft.Conn, list *minecraft.Listener) {
	dial, err := minecraft.Dialer{
		KeepXBLIdentityData: true,
		IdentityData:        conn.IdentityData(),
		ClientData:          conn.ClientData(),
	}.Dial("ipc", "default;"+conn.RemoteAddr().String())
	if err != nil {
		fmt.Println(err)
		list.Disconnect(conn, err.Error())
		return
	}
	if err := dial.DoSpawn(); err != nil {
		list.Disconnect(conn, err.Error())
		return
	}
	if err := conn.StartGame(dial.GameData()); err != nil {
		list.Disconnect(conn, err.Error())
		return
	}
	go func() {
		defer conn.Close()
		defer dial.Close()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				list.Disconnect(conn, err.Error())
				fmt.Println(err)
				return
			}
			dial.WritePacket(pk)
		}
	}()
	go func() {
		defer conn.Close()
		defer dial.Close()
		for {
			pk, err := dial.ReadPacket()
			if err != nil {
				list.Disconnect(conn, err.Error())
				fmt.Println(err)
				return
			}
			conn.WritePacket(pk)
		}
	}()
}
