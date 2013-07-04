package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
)

type dispatchChan interface {
	SetDispatch(chan *msg.Message)
}

type orderChan interface {
	SetOrders(chan *msg.Message)
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

type matcher interface {
	runner
	dispatchChan
	orderChan
}

func Coordinate(l listener, r responder, m matcher, log bool) {
	d := connect(l, r, m, log)
	run(l, r, m, d)
}

func connect(l listener, r responder, m matcher, log bool) *dispatcher {
	dispatch := make(chan *msg.Message, 100)
	orders := make(chan *msg.Message, 100)
	responses := make(chan *msg.Message, 100)
	d := &dispatcher{dispatch: dispatch, orders: orders, responses: responses, log: log}
	l.SetDispatch(dispatch)
	r.SetResponses(responses)
	r.SetDispatch(dispatch)
	m.SetOrders(orders)
	m.SetDispatch(dispatch)
	return d
}

func run(l listener, r responder, m matcher, d *dispatcher) {
	go l.Run()
	go r.Run()
	go m.Run()
	go d.Run()
}

type dispatcher struct {
	dispatch  chan *msg.Message
	orders    chan *msg.Message
	responses chan *msg.Message
	log       bool
}

func (d *dispatcher) Run() {
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
		case m.Route == msg.ORDER:
			d.orders <- m
		case m.Route == msg.MATCHER_RESPONSE, m.Route == msg.SERVER_ACK, m.Route == msg.CLIENT_ACK:
			d.responses <- m
		case m.Route == msg.COMMAND:
			d.orders <- m
			d.responses <- m
			if m.Kind == msg.SHUTDOWN {
				return
			}
		default:
			panic(fmt.Sprintf("Dispatcher - Unkown object: %v", m))
		}
	}
}

func (d *dispatcher) resubmitErr(m *msg.Message) {
	em := &msg.Message{}
	*em = *m
	em.WriteStatus(msg.INVALID_MSG_ERROR)
	d.dispatch <- em
}
