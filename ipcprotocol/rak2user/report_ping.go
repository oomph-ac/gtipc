package rak2user

import "github.com/gameparrot/gtipc/ipcprotocol/io"

type ReportPing struct {
	SessionID int32
	Ping      int32
}

func (r *ReportPing) ID() uint8 {
	return IdReportPing
}

func (r *ReportPing) Marshal(i io.IO, len int) {
	i.BEInt32(&r.SessionID)
	i.BEInt32(&r.Ping)
}
