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
	defer l.reader.Close()
	for {
		b := l.read()
		if b == nil {
			continue
		}
		o := l.deserialise(b)
		shutdown := l.forward(o)
		if shutdown {
			return
		}
	}
}

func (l *Listener) read() []byte {
	b := make([]byte, msg.SizeofMessage)
	n, err := l.reader.Read(b)
	if err != nil {
		em := &msg.Message{}
		em.WriteStatus(msg.READ_ERROR)
		l.dispatch <- em
		println("Listener - UDP Read: ", err.Error())
		return nil
	}
	return b[:n]
}

func (l *Listener) deserialise(b []byte) *msg.Message {
	o := &msg.Message{}
	o.WriteFrom(b)
	if len(b) != msg.SizeofMessage {
		o.WriteStatus(msg.SMALL_READ_ERROR)
		println(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.SizeofMessage, len(b), b))
	}
	return o
}

func (l *Listener) forward(o *msg.Message) (shutdown bool) {
	if o.Route != msg.CLIENT_ACK {
		a := &msg.Message{}
		a.WriteServerAckFor(o)
		l.dispatch <- a
	}
	if l.guidstore.Push(guid.MkGuid(o.TraderId, o.TradeId)) {
		l.dispatch <- o
	} else if o.Route == msg.CLIENT_ACK {
		l.dispatch <- o
	}
	return o.Route == msg.COMMAND && o.Kind == msg.SHUTDOWN
}
