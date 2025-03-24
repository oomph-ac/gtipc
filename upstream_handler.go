package gtipc

import (
	"context"
	"log/slog"
	"maps"
	"net"
	"sync"
	"time"

	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type UpstreamHandler struct {
	sentBytes     int64
	receivedBytes int64

	blocks   map[[16]byte]time.Time
	blocksMu sync.Mutex

	lastBlockGcTime time.Time

	parent raknet.UpstreamPacketListener

	stop chan bool
}

// NewUpstreamHandler returns an upstream packet listener with ip block and bandwidth monitoring
func NewUpstreamHandler(parent raknet.UpstreamPacketListener) *UpstreamHandler {
	q := &UpstreamHandler{parent: parent, blocks: make(map[[16]byte]time.Time), stop: make(chan bool)}
	go q.gc()
	return q
}

func (q *UpstreamHandler) ListenPacket(network, address string) (conn net.PacketConn, err error) {
	if q.parent != nil {
		conn, err = q.parent.ListenPacket(network, address)
	} else {
		conn, err = net.ListenPacket(network, address)
	}
	if err != nil {
		return nil, err
	}
	return newHandlerConn(conn, q), nil
}

func (q *UpstreamHandler) gc() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			q.gcBlocks()
		case <-q.stop:
			return
		}
	}
}

func (q *UpstreamHandler) gcBlocks() {
	q.blocksMu.Lock()
	defer q.blocksMu.Unlock()

	now := time.Now()
	maps.DeleteFunc(q.blocks, func(ip [16]byte, t time.Time) bool {
		return now.After(t)
	})
}

func (q *UpstreamHandler) close() {
	q.stop <- true
}

func (q *UpstreamHandler) blockAddress(addr net.IP, duration time.Duration) {
	q.blocksMu.Lock()
	q.blocks[[16]byte(addr.To16())] = time.Now().Add(duration)
	q.blocksMu.Unlock()
}

func (q *UpstreamHandler) unblockAddress(addr net.IP) {
	q.blocksMu.Lock()
	delete(q.blocks, [16]byte(addr.To16()))
	q.blocksMu.Unlock()
}

type handlerConn struct {
	parent net.PacketConn

	upstream *UpstreamHandler
}

func newHandlerConn(parent net.PacketConn, upstream *UpstreamHandler) *handlerConn {
	return &handlerConn{parent: parent, upstream: upstream}
}

func (q *handlerConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = q.parent.ReadFrom(p)
	q.upstream.receivedBytes += int64(n)
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		q.upstream.blocksMu.Lock()
		defer q.upstream.blocksMu.Unlock()
		addrBytes := [16]byte(udpAddr.IP.To16())
		if unblockTime, ok := q.upstream.blocks[addrBytes]; ok {
			if time.Now().Before(unblockTime) {
				return 0, addr, nil
			} else {
				delete(q.upstream.blocks, addrBytes)
			}
		}
	}

	return
}

func (q *handlerConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	num, err := q.parent.WriteTo(p, addr)
	q.upstream.sentBytes += int64(num)
	return num, err
}

func (q *handlerConn) Close() error {
	q.upstream.close()
	return q.parent.Close()
}

func (q *handlerConn) LocalAddr() net.Addr {
	return q.parent.LocalAddr()
}

func (q *handlerConn) SetDeadline(t time.Time) error {
	return q.parent.SetDeadline(t)
}

func (q *handlerConn) SetReadDeadline(t time.Time) error {
	return q.parent.SetReadDeadline(t)
}

func (q *handlerConn) SetWriteDeadline(t time.Time) error {
	return q.parent.SetWriteDeadline(t)
}

func CreateGophertunnelUpstreamHandler(name string) *UpstreamHandler {
	upstream := NewUpstreamHandler(nil)
	minecraft.RegisterNetwork(name, func(l *slog.Logger) minecraft.Network { return RakNetUpstream{l: l, upstream: upstream} })
	return upstream
}

// RakNet is an implementation of a RakNet v10 Network with an upstream handler.
type RakNetUpstream struct {
	l        *slog.Logger
	upstream *UpstreamHandler
}

// DialContext ...
func (r RakNetUpstream) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return raknet.Dialer{ErrorLog: r.l.With("net origin", "raknet")}.DialContext(ctx, address)
}

// PingContext ...
func (r RakNetUpstream) PingContext(ctx context.Context, address string) (response []byte, err error) {
	return raknet.Dialer{ErrorLog: r.l.With("net origin", "raknet")}.PingContext(ctx, address)
}

// Listen ...
func (r RakNetUpstream) Listen(address string) (minecraft.NetworkListener, error) {
	return raknet.ListenConfig{ErrorLog: r.l.With("net origin", "raknet"), UpstreamPacketListener: r.upstream}.Listen(address)
}

func (r RakNetUpstream) Compression(net.Conn) packet.Compression { return packet.FlateCompression }
