package netwk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fmstephe/matching_engine/guid"
	"github.com/fmstephe/matching_engine/trade"
	"net"
)

type Listener struct {
	conn      *net.UDPConn
	guidstore *guid.Store
	submit    chan interface{}
}

func NewListener(port string) (*Listener, error) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &Listener{conn: conn, guidstore: guid.NewStore()}, nil
}

func (l *Listener) SetSubmit(submit chan interface{}) {
	l.submit = submit
}

func (l *Listener) Run() {
	defer l.conn.Close()
	for {
		s := make([]byte, trade.SizeofOrder)
		n, _, err := l.conn.ReadFromUDP(s)
		if err != nil {
			println("Listener - UDP Read: ", err.Error())
			continue
		}
		if n != trade.SizeofOrder {
			println(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d submit %v", trade.SizeofOrder, n, s))
			continue
		}
		o := &trade.Order{}
		buf := bytes.NewBuffer(s)
		err = binary.Read(buf, binary.BigEndian, o)
		if err != nil {
			println("Listener - to []byte: ", err.Error())
			continue
		}
		r := &trade.Response{}
		r.WriteAck(o)
		l.submit <- r
		if l.guidstore.Push(guid.MkGuid(o.TraderId, o.TradeId)) {
			l.submit <- o
		}
		if o.Kind == trade.SHUTDOWN {
			return
		}
	}
}
