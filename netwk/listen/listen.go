package listen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"net"
)

const (
	orderSize = 8 + 8 + 4 + 4 + 4
)

type OrderListener struct {
	conn      *net.UDPConn
	orderChan chan *trade.OrderData
}

func NewOrderListener(port string, orderChan chan *trade.OrderData) (*OrderListener, error) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &OrderListener{conn: conn, orderChan: orderChan}, nil
}

func (l *OrderListener) Listen() {
	s := make([]byte, 64)
	nbuf := bytes.NewBuffer(s)
	for {
		n, _, err := l.conn.ReadFromUDP(s)
		if err != nil {
			println("Error reading UDP packet", err.Error())
			continue
		}
		if n != orderSize {
			println(fmt.Sprintf("Error incorrect number of bytes. Expecting %d, found %d", orderSize, n))
		}
		od := &trade.OrderData{}
		binary.Read(nbuf, binary.LittleEndian, od)
		l.orderChan <- od
	}
}
