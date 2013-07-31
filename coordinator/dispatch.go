package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
)

type dispatcher struct {
	dispatch  chan *msg.Message
	appMsgs   chan *msg.Message
	responses chan *msg.Message
	name      string
	log       bool
}

func (d *dispatcher) Run() {
	defer d.shutdown()
	for {
		m := <-d.dispatch
		if d.log {
			println(fmt.Sprintf("%s - %v", d.name, m))
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
