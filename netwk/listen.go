package netwk

import (
	"bytes"
	"encoding/binary"
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
		println("Listener - UDP Read: ", err.Error())
		return nil
	}
	if n != msg.SizeofMessage {
		println(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.SizeofMessage, n, b))
		return nil
	}
	return b
}

func (l *Listener) deserialise(b []byte) *msg.Message {
	o := &msg.Message{}
	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.BigEndian, o)
	if err != nil {
		println("Listener - to []byte: ", err.Error())
		return nil
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
