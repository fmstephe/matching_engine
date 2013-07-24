package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
)

type dispatchChan interface {
	SetDispatch(chan *msg.Message)
}

type appChan interface {
	SetAppMsgs(chan *msg.Message)
}

type responseChan interface {
	SetResponses(chan *msg.Message)
}

type runner interface {
	Run()
}

type listener interface {
	runner
	dispatchChan
}

type responder interface {
	runner
	dispatchChan
	responseChan
}

type app interface {
	runner
	dispatchChan
	appChan
}

func Coordinate(l listener, r responder, a app, log bool) {
	d := connect(l, r, a, log)
	run(l, r, a, d)
}

func connect(l listener, r responder, a app, log bool) *dispatcher {
	dispatch := make(chan *msg.Message, 100)
	appMsgs := make(chan *msg.Message, 100)
	responses := make(chan *msg.Message, 100)
	d := &dispatcher{dispatch: dispatch, appMsgs: appMsgs, responses: responses, log: log}
	l.SetDispatch(dispatch)
	r.SetResponses(responses)
	r.SetDispatch(dispatch)
	a.SetAppMsgs(appMsgs)
	a.SetDispatch(dispatch)
	return d
}

func run(l listener, r responder, m app, d *dispatcher) {
	go l.Run()
	go r.Run()
	go m.Run()
	go d.Run()
}

type dispatcher struct {
	dispatch  chan *msg.Message
	appMsgs   chan *msg.Message
	responses chan *msg.Message
	log       bool
}

func (d *dispatcher) Run() {
	defer d.shutdown()
	for {
		m := <-d.dispatch
		if d.log {
			println(fmt.Sprintf("Dispatcher - %v", m))
		}
		switch {
		case !m.Valid():
			d.resubmitErr(m)
		case m.Status != msg.NORMAL:
			d.responses <- m
		case m.Direction == msg.OUT, m.Route == msg.ACK:
			d.responses <- m
			if m.Route == msg.SHUTDOWN {
				return
			}
		case m.Direction == msg.IN: // Includes APP and SHUTDOWN messages
			d.appMsgs <- m
		default:
			panic(fmt.Sprintf("Dispatcher - Unkown object: %v", m))
		}
	}
}

func (d *dispatcher) resubmitErr(m *msg.Message) {
	em := &msg.Message{}
	*em = *m
	em.Status = msg.INVALID_MSG_ERROR
	em.Direction = msg.OUT
	d.dispatch <- em
}

func (d *dispatcher) shutdown() {
}
