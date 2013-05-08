package dispatch

import (
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
)

type dispatcher struct {
	submit    chan interface{}
	orders    chan *trade.OrderData
	responses chan *trade.Response
	log       bool
}

func New(submit chan interface{}, orders chan *trade.OrderData, responses chan *trade.Response) *dispatcher {
	return &dispatcher{submit: submit, orders: orders, responses: responses, log: true}
}

func (d *dispatcher) SetLogging(log bool) {
	d.log = log
}

func (d *dispatcher) Run() {
	for {
		v := <-d.submit
		if d.log {
			println(fmt.Sprintf("%v", v))
		}
		switch v := v.(type) {
		case *trade.OrderData:
			d.orders <- v
		case *trade.Response:
			d.responses <- v
		default:
			panic(fmt.Sprintf("Unkown object received: %v", v))
		}
	}
}
