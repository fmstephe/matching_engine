package netwk

import (
	"fmt"
	"github.com/fmstephe/matching_engine/guid"
	"github.com/fmstephe/matching_engine/msg"
	"io"
)

type Listener struct {
	reader    io.ReadCloser
	guidstore *guid.Store
	dispatch  chan *msg.Message
}

func NewListener(reader io.ReadCloser) *Listener {
	return &Listener{reader: reader, guidstore: guid.NewStore()}
}

func (l *Listener) SetDispatch(dispatch chan *msg.Message) {
	l.dispatch = dispatch
}

func (l *Listener) Run() {
	defer l.shutdown()
	for {
		m := l.deserialise()
		shutdown := l.forward(m)
		if shutdown {
			return
		}
	}
}

func (l *Listener) deserialise() *msg.Message {
	b := make([]byte, msg.SizeofMessage)
	m := &msg.Message{}
	n, err := l.reader.Read(b)
	m.WriteFrom(b[:n])
	if err != nil {
		m.WriteStatus(msg.READ_ERROR)
		println("Listener - UDP Read: ", err.Error())
	} else if n != msg.SizeofMessage {
		m.WriteStatus(msg.SMALL_READ_ERROR)
		println(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.SizeofMessage, n, b))
	}
	return m
}

func (l *Listener) forward(o *msg.Message) (shutdown bool) {
	if o.Route != msg.CLIENT_ACK {
		a := &msg.Message{}
		a.WriteServerAckFor(o)
		l.dispatch <- a
	}
	if o.Route == msg.CLIENT_ACK || o.Kind == msg.CANCEL || l.guidstore.Push(guid.MkGuid(o.TraderId, o.TradeId)) {
		l.dispatch <- o
	}
	return o.Route == msg.COMMAND && o.Kind == msg.SHUTDOWN
}

func (l *Listener) shutdown() {
	l.reader.Close()
}
