package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"io"
	"net"
	"os"
	"time"
)

func Reliable(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, originId uint32, name string, log bool) {
	listenerToApp := make(chan *msg.Message)
	appToResponder := make(chan *msg.Message)
	listenerToResponder := make(chan *RMessage)
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
	toResponder chan *RMessage
	name        string
	originId    uint32
	log         bool
	ticker      *Ticker
}

func newReliableListener(reader io.ReadCloser, toApp chan *msg.Message, toResponder chan *RMessage, name string, originId uint32, log bool) *reliableListener {
	l := &reliableListener{}
	l.reader = reader
	l.toApp = toApp
	l.toResponder = toResponder
	l.name = name
	l.originId = originId
	l.log = log
	l.ticker = NewTicker()
	return l
}

func (l *reliableListener) Run() {
	defer l.shutdown()
	for {
		rm := l.deserialise()
		shutdown := rm.message.Kind == msg.SHUTDOWN
		l.forward(rm)
		if shutdown {
			return
		}
	}
}

func (l *reliableListener) deserialise() *RMessage {
	b := make([]byte, SizeofRMessage)
	rm := &RMessage{}
	n, err := l.reader.Read(b)
	rm.WriteFrom(b[:n])
	if err != nil {
		rm.status = READ_ERROR
		l.logErr("Listener - UDP Read: " + err.Error())
	} else if n != SizeofRMessage {
		rm.status = SMALL_READ_ERROR
		l.logErr(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", SizeofRMessage, n, b))
	}
	return rm
}

func (l *reliableListener) logErr(errStr string) {
	if l.log {
		println(errStr)
	}
}

func (l *reliableListener) forward(rm *RMessage) {
	if l.log {
		println(l.name, "Listener:", rm.String())
	}
	if rm.route != ACK {
		a := &RMessage{}
		a.WriteAckFor(rm)
		l.toResponder <- a
	}
	if rm.route == ACK {
		l.toResponder <- rm
	} else if l.ticker.Tick(rm) {
		m := msg.Message{}
		m = rm.message
		l.toApp <- &m
	}
}

func (l *reliableListener) shutdown() {
	l.reader.Close()
}

const RESEND_MILLIS = time.Duration(10) * time.Millisecond

type reliableResponder struct {
	writer       io.WriteCloser
	fromApp      <-chan *msg.Message
	fromListener <-chan *RMessage
	name         string
	originId     uint32
	msgId        uint32
	log          bool
	unacked      *rmsgSet
}

func newReliableResponder(writer io.WriteCloser, fromApp <-chan *msg.Message, fromListener <-chan *RMessage, name string, originId uint32, log bool) *reliableResponder {
	r := &reliableResponder{}
	r.writer = writer
	r.fromApp = fromApp
	r.fromListener = fromListener
	r.name = name
	r.originId = originId
	r.msgId = 1
	r.log = log
	r.unacked = newSet()
	return r
}

func (r *reliableResponder) Run() {
	defer r.shutdown()
	t := time.NewTimer(RESEND_MILLIS)
	for {
		select {
		case m := <-r.fromApp:
			if r.log {
				println(r.name + " Responder (A): " + m.String())
			}
			if m.Kind == msg.SHUTDOWN {
				return
			}
			rm := &RMessage{route: APP, message: *m}
			r.writeResponse(rm)
		case rm := <-r.fromListener:
			if r.log {
				println(r.name + " Responder (L): " + rm.String()) // TODO code duplication
			}
			if rm.route == ACK {
				if rm.direction == IN {
					r.handleInAck(rm)
				} else {
					r.writeResponse(rm)
				}
			} else {
				panic(fmt.Sprintf("Illegal message received from listener: %v", rm))
			}
		case <-t.C:
			r.resend()
			t = time.NewTimer(RESEND_MILLIS)
		}
	}
}

func (r *reliableResponder) handleInAck(rm *RMessage) {
	r.unacked.remove(rm)
}

func (r *reliableResponder) writeResponse(rm *RMessage) {
	r.decorateResponse(rm)
	r.addToUnacked(rm)
	r.write(rm)
}

func (r *reliableResponder) decorateResponse(rm *RMessage) {
	rm.direction = IN
	rm.originId = r.originId
	rm.msgId = r.msgId
	r.msgId++
}

func (r *reliableResponder) addToUnacked(rm *RMessage) {
	if rm.route == APP {
		r.unacked.add(rm)
	}
}

func (r *reliableResponder) resend() {
	r.unacked.do(func(rm *RMessage) {
		r.write(rm)
	})
}

func (r *reliableResponder) write(rm *RMessage) {
	b := make([]byte, SizeofRMessage)
	rm.WriteTo(b)
	n, err := r.writer.Write(b)
	if err != nil {
		r.handleError(rm, err, WRITE_ERROR)
	}
	if n != SizeofRMessage {
		r.handleError(rm, err, SMALL_WRITE_ERROR)
	}
}

func (r *reliableResponder) handleError(rm *RMessage, err error, s MsgStatus) {
	em := &RMessage{}
	*em = *rm
	em.status = s
	println(rm.String(), err.Error())
	if e, ok := err.(net.Error); ok && !e.Temporary() {
		os.Exit(1)
	}
}

func (r *reliableResponder) shutdown() {
	r.writer.Close()
}
