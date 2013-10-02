package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"io"
	"net"
	"os"
)

func InMemory(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, unused uint32, name string, log bool) {
	toApp := make(chan *msg.Message)
	toResponder := make(chan *msg.Message)
	listener := newInMemoryListener(reader, toApp, name, log)
	responder := newInMemoryResponder(writer, toResponder, name, log)
	app.Config(name, toApp, toResponder)
	go listener.Run()
	go responder.Run()
	go app.Run()
}

type inMemoryListener struct {
	reader io.ReadCloser
	toApp  chan *msg.Message
	name   string
	log    bool
}

func newInMemoryListener(reader io.ReadCloser, toApp chan *msg.Message, name string, log bool) *inMemoryListener {
	l := &inMemoryListener{}
	l.reader = reader
	l.toApp = toApp
	l.name = name
	l.log = log
	return l
}

func (l *inMemoryListener) Run() {
	defer l.shutdown()
	for {
		m := l.deserialise()
		shutdown := l.forward(m)
		if shutdown {
			return
		}
	}
}

func (l *inMemoryListener) deserialise() *msg.Message {
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

func (l *inMemoryListener) logErr(errStr string) {
	if l.log {
		println(errStr)
	}
}

func (l *inMemoryListener) forward(o *msg.Message) (shutdown bool) {
	l.toApp <- o
	return o.Route == msg.SHUTDOWN
}

func (l *inMemoryListener) shutdown() {
	l.reader.Close()
}

type inMemoryResponder struct {
	writer  io.WriteCloser
	fromApp <-chan *msg.Message
	name    string
	log     bool
}

func newInMemoryResponder(writer io.WriteCloser, fromApp chan *msg.Message, name string, log bool) *inMemoryResponder {
	r := &inMemoryResponder{}
	r.writer = writer
	r.fromApp = fromApp
	r.name = name
	r.log = log
	return r
}

func (r *inMemoryResponder) Run() {
	defer r.shutdown()
	for {
		resp := <-r.fromApp
		if r.log {
			println(r.name + ": " + resp.String())
		}
		if resp.Route == msg.SHUTDOWN {
			return
		} else {
			r.write(resp)
		}
	}
}

func (r *inMemoryResponder) write(resp *msg.Message) {
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

func (r *inMemoryResponder) handleError(resp *msg.Message, err error, s msg.MsgStatus) {
	em := &msg.Message{}
	*em = *resp
	em.Status = s
	println(resp.String(), err.Error())
	if e, ok := err.(net.Error); ok && !e.Temporary() {
		os.Exit(1)
	}
}

func (r *inMemoryResponder) shutdown() {
	r.writer.Close()
}
