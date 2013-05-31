package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
)

type submitChan interface {
	SetSubmit(chan *msg.Message)
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
	submitChan
}

type responder interface {
	runner
	responseChan
}

type matcher interface {
	runner
	submitChan
	orderChan
}

func Coordinate(l listener, r responder, m matcher, log bool) {
	submit := make(chan *msg.Message, 100)
	orders := make(chan *msg.Message, 100)
	responses := make(chan *msg.Message, 100)
	d := &dispatcher{submit: submit, orders: orders, responses: responses, log: log}
	l.SetSubmit(submit)
	r.SetResponses(responses)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	go l.Run()
	go r.Run()
	go m.Run()
	go d.Run()
}

type dispatcher struct {
	submit    chan *msg.Message
	orders    chan *msg.Message
	responses chan *msg.Message
	log       bool
}

func (d *dispatcher) Run() {
	for {
		m := <-d.submit
		if d.log {
			println(fmt.Sprintf("Dispatcher - %v", m))
		}
		switch {
		case m.Route == msg.ORDER:
			d.orders <- m
		case m.Route == msg.RESPONSE, m.Route == msg.SERVER_ACK:
			d.responses <- m
		case m.Route == msg.COMMAND:
			d.responses <- m
			if m.Kind == msg.SHUTDOWN {
				return
			}
		default:
			panic(fmt.Sprintf("Dispatcher - Unkown object: %v", m))
		}
	}
}
