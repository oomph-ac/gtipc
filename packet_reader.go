package gtipc

import (
	"encoding/binary"
	"io"
)

type packetReader struct {
	readBuf []byte
	pkData  []byte
}

func newPacketReader() *packetReader {
	return &packetReader{readBuf: make([]byte, 65535*4)}
}

func (p *packetReader) takePackets(r io.Reader) ([][]byte, error) {
	pks := [][]byte{}
	for {
		n, err := r.Read(p.readBuf)
		if err != nil {
			return nil, err
		}
		pk := p.readBuf[:n]
		pkData := append(p.pkData, pk...)

		totalLen := len(pkData)
		numRead := 0
		for numRead < totalLen {
			if len(pkData) < 4 {
				p.pkData = pkData
				break
			}

			length := int(binary.BigEndian.Uint32(pkData))

			numRead += length + 4

			if len(pkData) < length+4 {
				p.pkData = pkData
				break
			}

			pkData = pkData[4:]

			data := pkData[:length]
			pks = append(pks, data)
			p.pkData = []byte{}
			pkData = pkData[length:]
		}

		if len(pks) > 0 {
			return pks, nil
		}
	}
}
