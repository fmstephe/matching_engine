package netwk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fmstephe/matching_engine/guid"
	"github.com/fmstephe/matching_engine/msg"
	"net"
)

type Listener struct {
	conn      *net.UDPConn
	guidstore *guid.Store
	submit    chan *msg.Message
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

func (l *Listener) SetSubmit(submit chan *msg.Message) {
	l.submit = submit
}

func (l *Listener) Run() {
	defer l.conn.Close()
	for {
		s := make([]byte, msg.SizeofMessage)
		n, _, err := l.conn.ReadFromUDP(s)
		if err != nil {
			println("Listener - UDP Read: ", err.Error())
			continue
		}
		if n != msg.SizeofMessage {
			println(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d submit %v", msg.SizeofMessage, n, s))
			continue
		}
		o := &msg.Message{}
		buf := bytes.NewBuffer(s)
		err = binary.Read(buf, binary.BigEndian, o)
		if err != nil {
			println("Listener - to []byte: ", err.Error())
			continue
		}
		r := &msg.Message{}
		r.WriteServerAck(o) // TODO this is inapproppriate - the matcher itself should ack an order
		l.submit <- r
		if l.guidstore.Push(guid.MkGuid(o.TraderId, o.TradeId)) {
			l.submit <- o
		}
		if o.Route == msg.SHUTDOWN {
			return
		}
	}
}
