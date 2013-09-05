package coordinator

import (
	"github.com/fmstephe/matching_engine/msg"
	"io"
)

type msgRunner interface {
	Config(log bool, name string, msgs chan *msg.Message)
	Run()
}

type msgHelper struct {
	log  bool
	name string
	msgs chan *msg.Message
}

func (h *msgHelper) Config(log bool, name string, msgs chan *msg.Message) {
	h.log = log
	h.name = name
	h.msgs = msgs
}

type AppMsgRunner interface {
	Config(name string, in, out chan *msg.Message)
	Run()
}

type AppMsgHelper struct {
	Name string
	In   <-chan *msg.Message
	Out  chan<- *msg.Message
}

func (a *AppMsgHelper) Config(name string, in, out chan *msg.Message) {
	a.Name = name
	a.In = in
	a.Out = out
}

func (a *AppMsgHelper) Process(m *msg.Message) (AppMsg *msg.Message, shutdown bool) {
	a.Out <- m
	switch {
	case m.Route == msg.APP && m.Status == msg.NORMAL:
		return m, false
	case m.Route == msg.SHUTDOWN:
		return nil, true
	default:
		return nil, false
	}
}

func Coordinate(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, name string, log bool) {
	listener := newListener(reader)
	responder := newResponder(writer)
	connect(listener, responder, app, name, log)
	run(listener, responder, app)
}

func connect(listener, responder msgRunner, app AppMsgRunner, name string, log bool) {
	fromListener := make(chan *msg.Message)
	fromApp := make(chan *msg.Message)
	listener.Config(log, name, fromListener)
	responder.Config(log, name, fromApp)
	app.Config(name, fromListener, fromApp)
}

func run(listener msgRunner, responder msgRunner, app AppMsgRunner) {
	go listener.Run()
	go responder.Run()
	go app.Run()
}
