package modbus

import (
	"errors"
	"net"
	"os"
	"time"
)

// RTUNetClientHandler implements RTU On TCP server modbus client
type RTUNetClientHandler struct {
	rtuPackager
	netTransporter
}

type netTransporter struct {
	Port net.Conn
	Latency     time.Duration
	ByteTimeout time.Duration
	EndTimeout  time.Duration
}

// Send sends data to server and ensures response length is greater than header length.
func (mb *netTransporter) Send(aduRequest []byte) (aduResponse []byte, err error) {
	for {
		err = mb.Port.SetWriteDeadline(time.Now().Add(time.Duration(len(aduRequest)) * mb.ByteTimeout))
		if err != nil {
			return
		}
		n, err := mb.Port.Write(aduRequest)
		if err != nil {
			aduResponse = make([]byte, 0)
			return aduResponse, err
		}
		if n >= len(aduRequest) {
			break
		}
		aduRequest = aduRequest[n:]
	}
	return mb.readAdu()
}

func (mb *netTransporter) readAdu() (adu []byte, err error) {
	adu = make([]byte, 0, 256)
	buf := make([]byte, 256)
	timeout := mb.Latency + 255 * mb.ByteTimeout
	for {
		err = mb.Port.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return
		}
		n, err := mb.Port.Read(buf)
		adu = append(adu, buf[:n]...)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				err = nil
			}
			return adu, err
		}
		timeout = mb.EndTimeout
	}
}
