package gtipc

import "log/slog"

type IpcOptions struct {
	// Handler function for custom packets
	CustomPacketHandler func(b []byte, serverKey string)
	// Upstream handler for additional features intended for proxy usage
	Upstream *UpstreamHandler
	// Logger
	Log *slog.Logger
}
