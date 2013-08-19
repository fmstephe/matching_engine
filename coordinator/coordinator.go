package coordinator

import (
	"github.com/fmstephe/matching_engine/msg"
	"io"
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

func Coordinate(reader io.ReadCloser, writer io.WriteCloser, a app, name string, log bool) {
	l := newListener(reader)
	r := newResponder(writer)
	d := connect(l, r, a, name, log)
	run(l, r, a, d)
}

func connect(l listener, r responder, a app, name string, log bool) *dispatcher {
	dispatch := make(chan *msg.Message, 10)
	appMsgs := make(chan *msg.Message, 10)
	responses := make(chan *msg.Message, 10)
	d := &dispatcher{dispatch: dispatch, appMsgs: appMsgs, responses: responses, name: name, log: log}
	l.SetDispatch(dispatch)
	r.SetResponses(responses)
	r.SetDispatch(dispatch)
	a.SetAppMsgs(appMsgs)
	a.SetDispatch(dispatch)
	return d
}

func run(l listener, r responder, a app, d *dispatcher) {
	go l.Run()
	go r.Run()
	go a.Run()
	go d.Run()
}
