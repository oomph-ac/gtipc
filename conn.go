package gtipc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"time"

	goio "io"

	"github.com/gameparrot/gtipc/ipcprotocol"
	"github.com/gameparrot/gtipc/ipcprotocol/io"
	"github.com/gameparrot/gtipc/ipcprotocol/rak2user"
	"github.com/gameparrot/gtipc/ipcprotocol/user2rak"
)

type Conn struct {
	unixConn net.Conn
	handler  IpcHandler

	user2RakPool map[uint8]func() ipcprotocol.Packet
	rak2UserPool map[uint8]func() ipcprotocol.Packet

	sessionId int32

	sessionsMut sync.Mutex
	sessions    map[int32]*clientConn

	key      string
	isClient bool
	pongData []byte

	writeMu sync.Mutex

	reader *packetReader

	log *slog.Logger
}

// NewConn returns a new conn. Call ReadLoop to start reading packets.
func NewConn(log *slog.Logger, conn net.Conn, key string, handler IpcHandler) *Conn {
	_, isClient := handler.(*IpcClient)
	if !isClient {
		log.Info("Server connected", "key", key)
	}
	c := &Conn{unixConn: conn, log: log, user2RakPool: user2rak.NewUser2RakPool(), rak2UserPool: rak2user.NewRak2UserPool(), sessions: make(map[int32]*clientConn), key: key, reader: newPacketReader(), isClient: isClient, handler: handler}
	return c
}

// Read loop
func (c *Conn) ReadLoop() {
	for {
		pks, err := c.ReadPacket()
		if err != nil {
			if errors.Is(err, goio.EOF) {
				if !c.isClient {
					c.log.Info("Server disconnected", "key", c.key)
				}
				c.Close()
				return
			} else if errors.Is(err, net.ErrClosed) {
				c.Close()
				return
			} else {
				c.log.Error("Failed to read packet", "err", err.Error())
			}
			continue
		}
		for _, pk := range pks {
			switch pk := pk.(type) {
			case *user2rak.Encapsulated:
				if pk.SessionID == -1 {
					c.handler.handleCustomPacket(pk.UserPayload, c.key)
					continue
				}
				c.sessionsMut.Lock()
				if session, ok := c.sessions[pk.SessionID]; ok {
					session.handlePacketFromServer(pk.UserPayload, (pk.Flags&(1<<0)) != 0, pk.Ack)
				}
				c.sessionsMut.Unlock()
			case *user2rak.SetName:
				c.pongData = pk.Name
			case *user2rak.CloseSession:
				c.sessionsMut.Lock()
				if session, ok := c.sessions[pk.SessionID]; ok {
					session.internalClose()
					delete(c.sessions, pk.SessionID)
				}
				c.sessionsMut.Unlock()
			case *user2rak.BlockAddress:
				c.handler.BlockAddress(net.ParseIP(pk.Addr), time.Duration(pk.Timeout)*time.Second)
			case *user2rak.UnblockAddress:
				c.handler.UnblockAddress(net.ParseIP(pk.Addr))
			}
		}
	}
}

func (c *Conn) removeSession(sessionId int32) {
	c.sessionsMut.Lock()
	delete(c.sessions, sessionId)
	c.sessionsMut.Unlock()
}

// OpenSession opens a new session on the PM server
func (c *Conn) OpenSession(clientAddr string) (net.Conn, error) {
	c.sessionsMut.Lock()

	c.sessionId++
	sid := c.sessionId

	var port int
	var addr []byte
	if clientAddr == "" {
		addr = make([]byte, 4)
		binary.BigEndian.PutUint32(addr, uint32(sid))
	} else {
		ip, portStr, err := net.SplitHostPort(clientAddr)
		if err != nil {
			addr = make([]byte, 4)
			binary.BigEndian.PutUint32(addr, uint32(sid))
		} else {
			ip, ok := netip.AddrFromSlice(net.ParseIP(ip))
			if ok {
				if ip.Is4In6() {
					ip = ip.Unmap()
				}
				addr = ip.AsSlice()
			}
			port, _ = strconv.Atoi(portStr)
		}
	}

	err := c.WritePacket(&rak2user.OpenSession{
		SessionID: sid,
		Addr:      addr,
		Port:      uint16(port),
		ClientID:  rand.Int63(),
	})
	if err != nil {
		c.sessionsMut.Unlock()
		return nil, err
	}

	clientConn := newClientConn(c, sid)
	c.sessions[sid] = clientConn
	c.sessionsMut.Unlock()

	return clientConn, nil
}

// Close closes the unix conn and all raknet conns
func (c *Conn) Close() {
	if c.unixConn != nil {
		c.unixConn.Close()
	}
	for _, s := range c.sessions {
		s.Close()
	}
}

// WritePacket writes an IPC packet to the conn
func (c *Conn) WritePacket(pk ipcprotocol.Packet) error {
	c.writeMu.Lock()
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(pk.ID())
	writer := io.NewWriter(buf)
	pk.Marshal(writer, 0)
	var b = make([]byte, 4)
	bytes := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(bytes)))
	_, err := c.unixConn.Write(append(b, buf.Bytes()...))
	c.writeMu.Unlock()
	return err
}

// WriteCustomPacket writes a custom packet (Encapsulated with session id set to -1)
func (c *Conn) WriteCustomPacket(b []byte) {
	c.WritePacket(&rak2user.Encapsulated{
		SessionID:   -1,
		UserPayload: b,
	})
}

// ReadPacket reads an IPC packet from the conn
func (c *Conn) ReadPacket() (packets []ipcprotocol.Packet, err error) {
	pks, err := c.reader.takePackets(c.unixConn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(error); ok {
				err = rErr
			}
		}
	}()
	for _, pk := range pks {
		packetFunc, ok := c.user2RakPool[pk[0]]
		if !ok {
			return packets, fmt.Errorf("invalid packet id %d", pk[0])
		}
		reader := bytes.NewReader(pk[1:])
		pkReader := io.NewReader(reader)
		packet := packetFunc()
		packet.Marshal(pkReader, len(pk)-1)
		packets = append(packets, packet)
	}
	return packets, nil
}
