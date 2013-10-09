package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"io"
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
		shutdown := m.Kind == msg.SHUTDOWN
		l.toApp <- m
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
		l.logErr("Listener - UDP Read: " + err.Error())
	} else if n != msg.SizeofMessage {
		l.logErr(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.SizeofMessage, n, b))
	}
	return m
}

func (l *inMemoryListener) logErr(errStr string) {
	if l.log {
		println(errStr)
	}
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
		m := <-r.fromApp
		if r.log {
			println(r.name + ": " + m.String())
		}
		shutdown := m.Kind == msg.SHUTDOWN
		r.write(m)
		if shutdown {
			return
		}
	}
}

func (r *inMemoryResponder) write(m *msg.Message) {
	b := make([]byte, msg.SizeofMessage)
	m.WriteTo(b)
	n, err := r.writer.Write(b)
	if err != nil {
		r.handleError(err.Error())
	}
	if n != msg.SizeofMessage {
		r.handleError(fmt.Sprintf("Write Error: Wrong sized message. Found %d, expecting %d", n, msg.SizeofMessage))
	}
}

func (r *inMemoryResponder) handleError(errMsg string) {
	if r.log {
		println("Write Error: ", errMsg)
	}
}

func (r *inMemoryResponder) shutdown() {
	r.writer.Close()
}
