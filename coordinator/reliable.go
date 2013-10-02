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
	listenerToApp := make(chan *msg.Message)
	appToResponder := make(chan *msg.Message)
	listenerToResponder := make(chan *msg.Message)
	listener := newReliableListener(reader, listenerToApp, listenerToResponder, name, originId, log)
	responder := newReliableResponder(writer, appToResponder, listenerToResponder, name, originId, log)
	app.Config(name, listenerToApp, appToResponder)
	go listener.Run()
	go responder.Run()
	go app.Run()
}

type reliableListener struct {
	reader      io.ReadCloser
	toApp       chan *msg.Message
	toResponder chan *msg.Message
	name        string
	originId    uint32
	log         bool
	ticker      *msgutil.Ticker
}

func newReliableListener(reader io.ReadCloser, toApp, toResponder chan *msg.Message, name string, originId uint32, log bool) *reliableListener {
	l := &reliableListener{}
	l.reader = reader
	l.toApp = toApp
	l.toResponder = toResponder
	l.name = name
	l.originId = originId
	l.log = log
	l.ticker = msgutil.NewTicker()
	return l
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
	if l.log {
		println(l.name, "Listener:", o.String())
	}
	if o.Route != msg.ACK {
		a := &msg.Message{}
		a.WriteAckFor(o)
		l.toResponder <- a
	}
	if o.Route == msg.ACK {
		l.toResponder <- o
	} else if l.ticker.Tick(o) {
		l.toApp <- o
	}
	return o.Route == msg.SHUTDOWN
}

func (l *reliableListener) shutdown() {
	l.reader.Close()
}

const RESEND_MILLIS = time.Duration(10) * time.Millisecond

type reliableResponder struct {
	writer       io.WriteCloser
	fromApp      <-chan *msg.Message
	fromListener <-chan *msg.Message
	name         string
	originId     uint32
	msgId        uint32
	log          bool
	unacked      *msgutil.Set
}

func newReliableResponder(writer io.WriteCloser, fromApp, fromListener <-chan *msg.Message, name string, originId uint32, log bool) *reliableResponder {
	r := &reliableResponder{}
	r.writer = writer
	r.fromApp = fromApp
	r.fromListener = fromListener
	r.name = name
	r.originId = originId
	r.msgId = 1
	r.log = log
	r.unacked = msgutil.NewSet()
	return r
}

func (r *reliableResponder) Run() {
	defer r.shutdown()
	t := time.NewTimer(RESEND_MILLIS)
	for {
		select {
		case resp := <-r.fromApp:
			if r.log {
				println(r.name + " Responder (A): " + resp.String())
			}
			switch {
			case resp.Route == msg.APP:
				r.writeResponse(resp)
			case resp.Route == msg.SHUTDOWN:
				return
			default:
				panic(fmt.Sprintf("Unhandleable response %v", resp))
			}
		case resp := <-r.fromListener:
			if r.log {
				println(r.name + " Responder (L): " + resp.String()) // TODO code duplication
			}
			if resp.Route == msg.ACK {
				if resp.Direction == msg.IN {
					r.handleInAck(resp)
				} else {
					r.writeResponse(resp)
				}
			}
		case <-t.C:
			r.resend()
			t = time.NewTimer(RESEND_MILLIS)
		}
	}
}

func (r *reliableResponder) handleInAck(m *msg.Message) {
	r.unacked.Remove(m)
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
