package gtipc

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/gameparrot/gtipc/internal"
	"github.com/gameparrot/gtipc/ipcprotocol/rak2user"
)

type clientConn struct {
	addr *ipcAddr

	ctx        context.Context
	cancelFunc context.CancelFunc

	lastPacketTime time.Time
	conn           *Conn
	sessionId      int32

	userPackets *internal.ElasticChan[[]byte]

	once sync.Once
}

func newClientConn(conn *Conn, sessionId int32) *clientConn {
	c, cancel := context.WithCancel(context.Background())
	return &clientConn{conn: conn, sessionId: sessionId, ctx: c, cancelFunc: cancel, userPackets: internal.Chan[[]byte](4, 4096), addr: &ipcAddr{Key: conn.key, SessionId: sessionId}}
}

func (c *clientConn) handlePacketFromServer(payload []byte, needsAck bool, ack int32) {
	c.userPackets.Send(payload)
	if needsAck {
		c.conn.WritePacket(&rak2user.AckNotification{SessionID: c.sessionId, ACK: ack})
	}
}

func (c *clientConn) Write(b []byte) (int, error) {
	err := c.conn.WritePacket(&rak2user.Encapsulated{SessionID: c.sessionId, UserPayload: b})
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *clientConn) ReadPacket() ([]byte, error) {
	pk, ok := c.userPackets.Recv(c.ctx)
	if !ok {
		return nil, net.ErrClosed
	}
	return pk, nil
}

func (c *clientConn) Read(b []byte) (int, error) {
	pk, err := c.ReadPacket()
	if err != nil {
		return 0, err
	}
	return copy(b, pk), nil
}

func (c *clientConn) Close() error {
	c.internalClose()
	c.conn.removeSession(c.sessionId)
	return c.conn.WritePacket(&rak2user.CloseSession{SessionID: c.sessionId})
}

func (c *clientConn) LocalAddr() net.Addr {
	return c.addr
}

func (c *clientConn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *clientConn) internalClose() {
	c.once.Do(func() {
		c.cancelFunc()
	})
}

func (c *clientConn) SetDeadline(t time.Time) error {
	return errors.New("not supported")
}

func (c *clientConn) SetReadDeadline(t time.Time) error {
	return errors.New("not supported")
}

func (c *clientConn) SetWriteDeadline(t time.Time) error {
	return errors.New("not supported")
}

type ipcAddr struct {
	SessionId int32
	Key       string
}

func (i *ipcAddr) String() string {
	return i.Key + ":" + strconv.Itoa(int(i.SessionId))
}

func (i *ipcAddr) Network() string {
	return "ipc"
}
