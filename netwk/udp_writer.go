package netwk

import (
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"net"
)

type udpWriter struct{}

func (w *udpWriter) Write(data []byte, ip [4]byte, port int) error {
	addr := &net.UDPAddr{}
	addr.IP = net.IPv4(ip[0], ip[1], ip[2], ip[3])
	addr.Port = port
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	n, err := conn.Write(data)
	if err != nil {
		return err
	}
	if n != msg.SizeofMessage {
		return errors.New(fmt.Sprintf("Insufficient data written. Expecting %d, found %d", msg.SizeofMessage, n))
	}
	return nil
}
