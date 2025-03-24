package rak2user

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type ReportBandwidthStats struct {
	SentBytesDiff     int64
	ReceivedBytesDiff int64
}

func (r *ReportBandwidthStats) ID() uint8 {
	return IdReportBandwidthStats
}

func (r *ReportBandwidthStats) Marshal(io io.IO, len int) {
	io.BEInt64(&r.SentBytesDiff)
	io.BEInt64(&r.ReceivedBytesDiff)
}
