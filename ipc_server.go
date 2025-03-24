package gtipc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gameparrot/gtipc/ipcprotocol/rak2user"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var (
	errNotFound = errors.New("pocketmine server not found")
)

// IpcServer implements a server that PM clients connect to. This is recommended for proxy usage.
type IpcServer struct {
	listener       net.Listener
	ipcRaknetConns map[string]*Conn
	connsMu        sync.RWMutex
	socketPath     string

	close chan bool

	opts *IpcOptions
}

// NewIpcServer returns a new IPC server
func NewIPCServer(socketPath string, opts *IpcOptions) (*IpcServer, error) {
	if opts == nil {
		opts = &IpcOptions{}
	}
	if opts.Log == nil {
		opts.Log = slog.Default()
	}

	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	c := &IpcServer{listener: listener, ipcRaknetConns: make(map[string]*Conn), socketPath: socketPath, opts: opts, close: make(chan bool)}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go func() {
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				var nameLen = make([]byte, 1)
				if _, err := conn.Read(nameLen); err != nil {
					conn.Close()
					return
				}
				var name = make([]byte, nameLen[0])
				if _, err := conn.Read(name); err != nil {
					conn.Close()
					return
				}
				conn.SetReadDeadline(time.Time{})
				ipcConn := NewConn(c.opts.Log, conn, string(name), c)
				c.connsMu.Lock()
				c.ipcRaknetConns[string(name)] = ipcConn
				c.connsMu.Unlock()
				ipcConn.ReadLoop()
				c.connsMu.Lock()
				delete(c.ipcRaknetConns, string(name))
				c.connsMu.Unlock()
			}()
		}
	}()

	if opts.Upstream != nil {
		ticker := time.NewTicker(1 * time.Second)
		go func() {
			for {
				select {
				case <-c.close:
					return
				case <-ticker.C:
					c.connsMu.RLock()
					for _, i := range c.ipcRaknetConns {
						i.WritePacket(&rak2user.ReportBandwidthStats{SentBytesDiff: c.opts.Upstream.sentBytes, ReceivedBytesDiff: c.opts.Upstream.receivedBytes})
					}
					c.opts.Upstream.sentBytes = 0
					c.opts.Upstream.receivedBytes = 0
					c.connsMu.RUnlock()
				}
			}
		}()
	}

	return c, nil
}

// BlockAddress blocks an IP address from accessing the server
func (l *IpcServer) BlockAddress(addr net.IP, duration time.Duration) {
	if len(addr) != net.IPv4len && len(addr) != net.IPv6len {
		return
	}
	if l.opts.Upstream != nil {
		l.opts.Upstream.blockAddress(addr, duration)
	}
}

// UnblockAddress allows a blocked IP address to access te server
func (l *IpcServer) UnblockAddress(addr net.IP) {
	if len(addr) != net.IPv4len && len(addr) != net.IPv6len {
		return
	}
	if l.opts.Upstream != nil {
		l.opts.Upstream.unblockAddress(addr)
	}
}

// GetConn returns a conn with the key, or false if not found
func (l *IpcServer) GetConn(key string) (*Conn, bool) {
	l.connsMu.RLock()
	defer l.connsMu.RUnlock()
	conn, ok := l.ipcRaknetConns[key]
	return conn, ok
}

// Close closes all conns and the unix socket
func (l *IpcServer) Close() {
	l.close <- true
	l.connsMu.RLock()
	defer l.connsMu.RUnlock()
	for _, i := range l.ipcRaknetConns {
		i.Close()
	}
	l.listener.Close()
	os.Remove(l.socketPath)
}

func (l *IpcServer) DialContext(ctx context.Context, address string) (net.Conn, error) {
	l.connsMu.RLock()
	defer l.connsMu.RUnlock()

	split := strings.Split(address, ";")
	key := split[0]
	conn, ok := l.ipcRaknetConns[key]
	if !ok {
		return nil, errNotFound
	}
	clientIp := ""
	if len(split) > 1 {
		clientIp = split[1]
	}
	return conn.OpenSession(clientIp)
}

func (l *IpcServer) PingContext(ctx context.Context, address string) (response []byte, err error) {
	l.connsMu.RLock()
	defer l.connsMu.RUnlock()

	conn, ok := l.ipcRaknetConns[address]
	if !ok {
		return nil, errNotFound
	}
	return conn.pongData, nil
}

func (l *IpcServer) Listen(address string) (minecraft.NetworkListener, error) {
	return nil, errors.New("not supported")
}

func (l *IpcServer) handleCustomPacket(b []byte, serverKey string) {
	if l.opts.CustomPacketHandler != nil {
		l.opts.CustomPacketHandler(b, serverKey)
	}
}

func (*IpcServer) Compression(net.Conn) packet.Compression { return packet.FlateCompression }
