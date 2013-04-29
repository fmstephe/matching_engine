package netwk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"net"
)

type Listener struct {
	conn      *net.UDPConn
	orderChan chan *trade.OrderData
}

func NewListener(port string, orderChan chan *trade.OrderData) (*Listener, error) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &Listener{conn: conn, orderChan: orderChan}, nil
}

func (l *Listener) Listen() {
	for {
		s := make([]byte, trade.SizeofOrderData)
		n, _, err := l.conn.ReadFromUDP(s)
		if err != nil {
			println("Listener - UDP Read: ", err.Error())
			continue
		}
		if n != trade.SizeofOrderData {
			println(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", trade.SizeofOrderData, n, s))
			continue
		}
		od := &trade.OrderData{}
		buf := bytes.NewBuffer(s)
		err = binary.Read(buf, binary.BigEndian, od)
		if err != nil {
			println("Listener - to []byte: ", err.Error())
			continue
		}
		println(fmt.Sprintf("Listener - order data: %v", od)) // Temporary Logging
		l.orderChan <- od
	}
}
