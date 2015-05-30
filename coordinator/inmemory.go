package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"io"
)

func InMemory(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, unused uint32, name string, log bool) {
	fromListener, toResponder := InMemoryListenerResponder(reader, writer, name, log)
	app.Config(name, fromListener, toResponder)
	go app.Run()
}

func InMemoryListenerResponder(reader io.ReadCloser, writer io.WriteCloser, name string, log bool) (MsgReader, MsgWriter) {
	fromListener := NewChanReaderWriter(1000)
	toResponder := NewChanReaderWriter(1000)
	listener := newInMemoryListener(reader, fromListener, name, log)
	responder := newInMemoryResponder(writer, toResponder, name, log)
	go listener.Run()
	go responder.Run()
	return fromListener, toResponder
}

type inMemoryListener struct {
	reader io.ReadCloser
	toApp  MsgWriter
	name   string
	log    bool
}

func newInMemoryListener(reader io.ReadCloser, toApp MsgWriter, name string, log bool) *inMemoryListener {
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
		l.toApp.Write(*m)
		if shutdown {
			return
		}
	}
}

func (l *inMemoryListener) deserialise() *msg.Message {
	b := make([]byte, msg.ByteSize)
	m := &msg.Message{}
	n, err := l.reader.Read(b)
	if err != nil {
		panic("Listener - UDP Read: " + err.Error())
	} else if n != msg.ByteSize {
		panic(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.ByteSize, n, b))
	}
	if err := m.Unmarshal(b[:n]); err != nil {
		panic(err.Error())
	}
	return m
}

func (l *inMemoryListener) shutdown() {
	l.reader.Close()
}

type inMemoryResponder struct {
	writer  io.WriteCloser
	fromApp MsgReader
	name    string
	log     bool
}

func newInMemoryResponder(writer io.WriteCloser, fromApp MsgReader, name string, log bool) *inMemoryResponder {
	r := &inMemoryResponder{}
	r.writer = writer
	r.fromApp = fromApp
	r.name = name
	r.log = log
	return r
}

func (r *inMemoryResponder) Run() {
	defer r.shutdown()
	m := &msg.Message{}
	for {
		*m = r.fromApp.Read()
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
	b := make([]byte, msg.ByteSize)
	if err := m.Marshal(b); err != nil {
		panic(err.Error())
	}
	n, err := r.writer.Write(b)
	if err != nil {
		panic(err.Error())
	}
	if n != msg.ByteSize {
		panic(fmt.Sprintf("Write Error: Wrong sized message. Found %d, expecting %d", n, msg.ByteSize))
	}
}

func (r *inMemoryResponder) shutdown() {
	r.writer.Close()
}
