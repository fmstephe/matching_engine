package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
)

type submitChan interface {
	SetSubmit(chan interface{})
}

type orderChan interface {
	SetOrderNodes(chan *trade.Order)
}

type responseChan interface {
	SetResponses(chan *trade.Response)
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
	submit := make(chan interface{}, 100)
	orders := make(chan *trade.Order, 100)
	responses := make(chan *trade.Response, 100)
	d := &dispatcher{submit: submit, orders: orders, responses: responses, log: log}
	l.SetSubmit(submit)
	r.SetResponses(responses)
	m.SetSubmit(submit)
	m.SetOrderNodes(orders)
	go l.Run()
	go r.Run()
	go m.Run()
	go d.Run()
}

type dispatcher struct {
	submit    chan interface{}
	orders    chan *trade.Order
	responses chan *trade.Response
	log       bool
}

func (d *dispatcher) Run() {
	for {
		v := <-d.submit
		switch v := v.(type) {
		case *trade.Order:
			if d.log {
				println(fmt.Sprintf("Order - %v", v))
			}
			d.orders <- v
		case *trade.Response:
			if d.log {
				println(fmt.Sprintf("Response - %v", v))
			}
			d.responses <- v
			if v.Kind == trade.SHUTDOWN {
				return
			}
		default:
			panic(fmt.Sprintf("Unkown object received: %v", v))
		}
	}
}
