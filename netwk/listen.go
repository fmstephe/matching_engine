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
	dispatch  chan *msg.Message
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

func (l *Listener) SetDispatch(dispatch chan *msg.Message) {
	l.dispatch = dispatch
}

func (l *Listener) Run() {
	defer l.conn.Close()
	for {
		b := make([]byte, msg.SizeofMessage)
		n, _, err := l.conn.ReadFromUDP(b)
		if err != nil {
			println("Listener - UDP Read: ", err.Error())
			continue
		}
		if n != msg.SizeofMessage {
			println(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.SizeofMessage, n, b))
			continue
		}
		o := &msg.Message{}
		buf := bytes.NewBuffer(b)
		err = binary.Read(buf, binary.BigEndian, o)
		if err != nil {
			println("Listener - to []byte: ", err.Error())
			continue
		}
		if o.Route != msg.CLIENT_ACK {
			a := &msg.Message{}
			a.WriteServerAckFor(o)
			l.dispatch <- a
		}
		if l.guidstore.Push(guid.MkGuid(o.TraderId, o.TradeId)) {
			l.dispatch <- o
		}
		if o.Route == msg.CLIENT_ACK {
			l.dispatch <- o
		}
		if o.Route == msg.COMMAND && o.Kind == msg.SHUTDOWN {
			return
		}
	}
}
