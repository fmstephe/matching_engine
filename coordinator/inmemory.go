package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"io"
	"net"
	"os"
)

func InMemory(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, originId uint32, name string, log bool) {
	listener := newInMemoryListener(reader)
	responder := newInMemoryResponder(writer)
	fromListener := make(chan *msg.Message)
	fromApp := make(chan *msg.Message)
	listener.Config(originId, log, name, fromListener)
	responder.Config(originId, log, name, fromApp)
	app.Config(name, fromListener, fromApp, DefaultMsgProcessor)
	go listener.Run()
	go responder.Run()
	go app.Run()
}

type inMemoryListener struct {
	reader io.ReadCloser
	msgHelper
}

func newInMemoryListener(reader io.ReadCloser) *inMemoryListener {
	return &inMemoryListener{reader: reader}
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
	l.msgs <- o
	return o.Route == msg.SHUTDOWN
}

func (l *inMemoryListener) shutdown() {
	l.reader.Close()
}

type inMemoryResponder struct {
	writer io.WriteCloser
	msgHelper
}

func newInMemoryResponder(writer io.WriteCloser) *inMemoryResponder {
	return &inMemoryResponder{writer: writer}
}

func (r *inMemoryResponder) Run() {
	defer r.shutdown()
	for {
		resp := <-r.msgs
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
