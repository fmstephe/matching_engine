package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/msg/msgutil"
	"io"
	"net"
	"os"
	"time"
)

func Reliable(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, originId uint32, name string, log bool) {
	listener := newListener(reader)
	responder := newResponder(writer)
	connect(listener, responder, app, originId, name, log)
	run(listener, responder, app)
}

func connect(listener, responder msgRunner, app AppMsgRunner, originId uint32, name string, log bool) {
	fromListener := make(chan *msg.Message)
	fromApp := make(chan *msg.Message)
	listener.Config(originId, log, name, fromListener)
	responder.Config(originId, log, name, fromApp)
	app.Config(name, fromListener, fromApp, reliableMsgProcessor)
}

func run(listener msgRunner, responder msgRunner, app AppMsgRunner) {
	go listener.Run()
	go responder.Run()
	go app.Run()
}

func reliableFilter(m *msg.Message, out chan<- *msg.Message) (AppMsg *msg.Message, shutdown bool) {
	out <- m
	switch {
	case m.Route == msg.APP && m.Status == msg.NORMAL:
		return m, false
	case m.Route == msg.SHUTDOWN:
		return nil, true
	default:
		return nil, false
	}
}

type reliableListener struct {
	reader io.ReadCloser
	ticker *msgutil.Ticker
	msgHelper
}

func newListener(reader io.ReadCloser) *reliableListener {
	return &reliableListener{reader: reader, ticker: msgutil.NewTicker()}
}

func (l *reliableListener) Run() {
	defer l.shutdown()
	for {
		m := l.deserialise()
		shutdown := l.forward(m)
		if shutdown {
			return
		}
	}
}

func (l *reliableListener) deserialise() *msg.Message {
	b := make([]byte, msg.SizeofMessage)
	m := &msg.Message{}
	n, err := l.reader.Read(b)
	m.WriteFrom(b[:n])
	if err != nil {
		m.Status = msg.READ_ERROR
		l.logErr("Listener - UDP Read: " + err.Error())
	} else if n != msg.SizeofMessage {
		m.Status = msg.SMALL_READ_ERROR
		l.logErr(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.SizeofMessage, n, b))
	}
	return m
}

func (l *reliableListener) logErr(errStr string) {
	if l.log {
		println(errStr)
	}
}

func (l *reliableListener) forward(o *msg.Message) (shutdown bool) {
	if o.Route != msg.ACK && o.Route != msg.SHUTDOWN {
		a := &msg.Message{}
		a.WriteAckFor(o)
		l.msgs <- a
	}
	if o.Route == msg.SHUTDOWN || o.Route == msg.ACK || l.ticker.Tick(o) {
		l.msgs <- o
	}
	return o.Route == msg.SHUTDOWN
}

func (l *reliableListener) shutdown() {
	l.reader.Close()
}

const RESEND_MILLIS = time.Duration(10) * time.Millisecond

type reliableResponder struct {
	unacked *msgutil.Set
	writer  io.WriteCloser
	msgHelper
	msgId uint32
}

func newResponder(writer io.WriteCloser) *reliableResponder {
	return &reliableResponder{unacked: msgutil.NewSet(), writer: writer, msgId: 1}
}

func (r *reliableResponder) Run() {
	defer r.shutdown()
	t := time.NewTimer(RESEND_MILLIS)
	for {
		select {
		case resp := <-r.msgs:
			if r.log {
				println(r.name + ": " + resp.String())
			}
			switch {
			case resp.Direction == msg.IN && resp.Route == msg.ACK:
				r.handleInAck(resp)
			case resp.Direction == msg.OUT && (resp.Status != msg.NORMAL || resp.Route == msg.APP || resp.Route == msg.ACK):
				r.writeResponse(resp)
			case resp.Direction == msg.IN && resp.Route == msg.APP && resp.Status == msg.NORMAL:
				continue
			case resp.Route == msg.SHUTDOWN:
				return
			default:
				panic(fmt.Sprintf("Unhandleable response %v", resp))
			}
		case <-t.C:
			r.resend()
			t = time.NewTimer(RESEND_MILLIS)
		}
	}
}

func (r *reliableResponder) handleInAck(ca *msg.Message) {
	r.unacked.Remove(ca)
}

func (r *reliableResponder) writeResponse(resp *msg.Message) {
	resp.Direction = msg.IN
	resp.OriginId = r.originId
	resp.MsgId = r.msgId
	r.msgId++
	r.addToUnacked(resp)
	r.write(resp)
}

func (r *reliableResponder) addToUnacked(resp *msg.Message) {
	if resp.Route == msg.APP {
		r.unacked.Add(resp)
	}
}

func (r *reliableResponder) resend() {
	r.unacked.Do(func(m *msg.Message) {
		r.write(m)
	})
}

func (r *reliableResponder) write(resp *msg.Message) {
	b := make([]byte, msg.SizeofMessage)
	resp.WriteTo(b)
	n, err := r.writer.Write(b)
	if err != nil {
		r.handleError(resp, err, msg.WRITE_ERROR)
	}
	if n != msg.SizeofMessage {
		r.handleError(resp, err, msg.SMALL_WRITE_ERROR)
	}
}

func (r *reliableResponder) handleError(resp *msg.Message, err error, s msg.MsgStatus) {
	em := &msg.Message{}
	*em = *resp
	em.Status = s
	println(resp.String(), err.Error())
	if e, ok := err.(net.Error); ok && !e.Temporary() {
		os.Exit(1)
	}
}

func (r *reliableResponder) shutdown() {
	r.writer.Close()
}
