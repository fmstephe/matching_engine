package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/msg/msgutil"
	"io"
)

type stdListener struct {
	reader   io.ReadCloser
	ticker   *msgutil.Ticker
	dispatch chan *msg.Message
}

func newListener(reader io.ReadCloser) listener {
	return &stdListener{reader: reader, ticker: msgutil.NewTicker()}
}

func (l *stdListener) SetDispatch(dispatch chan *msg.Message) {
	l.dispatch = dispatch
}

func (l *stdListener) Run() {
	defer l.shutdown()
	for {
		m := l.deserialise()
		shutdown := l.forward(m)
		if shutdown {
			return
		}
	}
}

func (l *stdListener) deserialise() *msg.Message {
	b := make([]byte, msg.SizeofMessage)
	m := &msg.Message{}
	n, err := l.reader.Read(b)
	m.WriteFrom(b[:n])
	if err != nil {
		m.Status = msg.READ_ERROR
		println("Listener - UDP Read: ", err.Error())
	} else if n != msg.SizeofMessage {
		m.Status = msg.SMALL_READ_ERROR
		println(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.SizeofMessage, n, b))
	}
	return m
}

func (l *stdListener) forward(o *msg.Message) (shutdown bool) {
	if o.Route != msg.ACK && o.Route != msg.SHUTDOWN {
		a := &msg.Message{}
		a.WriteAckFor(o)
		l.dispatch <- a
	}
	if o.Route == msg.SHUTDOWN || l.ticker.Tick(o) {
		l.dispatch <- o
	}
	return o.Route == msg.SHUTDOWN
}

func (l *stdListener) shutdown() {
	l.reader.Close()
}
