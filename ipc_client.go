package gtipc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/sandertv/gophertunnel/minecraft"
)

// IpcClient implements a client that connects to an ipc server in PM. This is not recommended for proxy usage.
type IpcClient struct {
	opts    *IpcOptions
	conns   map[string]*Conn
	connsMu sync.RWMutex
}

// NewIpcClient returns a new IPC client
func NewIpcClient(opts *IpcOptions) *IpcClient {
	if opts == nil {
		opts = &IpcOptions{}
	}
	if opts.Log == nil {
		opts.Log = slog.Default()
	}
	return &IpcClient{opts: opts, conns: make(map[string]*Conn)}
}

func (c *IpcClient) openConn(path string) (*Conn, error) {
	unixConn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	conn := NewConn(c.opts.Log, unixConn, path, c)
	go func() {
		conn.ReadLoop()
		c.connsMu.Lock()
		delete(c.conns, path)
		c.connsMu.Unlock()
	}()
	c.connsMu.Lock()
	c.conns[path] = conn
	c.connsMu.Unlock()
	return conn, nil
}

func (c *IpcClient) GetOrCreateConn(path string) (*Conn, error) {
	path = filepath.Clean(path)
	c.connsMu.RLock()
	if conn, ok := c.conns[path]; ok {
		c.connsMu.RUnlock()
		return conn, nil
	}
	c.connsMu.RUnlock()
	return c.openConn(path)
}

func (c *IpcClient) GetConn(key string) (*Conn, bool) {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()
	conn, ok := c.conns[key]
	return conn, ok
}

func (c *IpcClient) DialContext(ctx context.Context, address string) (net.Conn, error) {
	conn, err := c.GetOrCreateConn(address)
	if err != nil {
		return nil, err
	}
	return conn.OpenSession("")
}

func (c *IpcClient) PingContext(ctx context.Context, address string) ([]byte, error) {
	conn, err := c.GetOrCreateConn(address)
	if err != nil {
		return nil, err
	}
	return conn.pongData, nil
}

func (c *IpcClient) Listen(address string) (minecraft.NetworkListener, error) {
	return nil, errors.New("not supported")
}

// BlockAddress blocks an IP address from accessing the server
func (c *IpcClient) BlockAddress(addr net.IP, duration time.Duration) {
	if len(addr) != net.IPv4len && len(addr) != net.IPv6len {
		return
	}
	if c.opts.Upstream != nil {
		c.opts.Upstream.blockAddress(addr, duration)
	}
}

// UnblockAddress allows a blocked IP address to access te server
func (c *IpcClient) UnblockAddress(addr net.IP) {
	if len(addr) != net.IPv4len && len(addr) != net.IPv6len {
		return
	}
	if c.opts.Upstream != nil {
		c.opts.Upstream.unblockAddress(addr)
	}
}

func (c *IpcClient) handleCustomPacket(b []byte, serverKey string) {
	if c.opts.CustomPacketHandler != nil {
		c.opts.CustomPacketHandler(b, serverKey)
	}
}
