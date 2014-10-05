package client

import (
	"github.com/fmstephe/matching_engine/msg"
)

const (
	connectedComment    = "Connected to trader"
	ordersClosedComment = "Disconnected because orders channel was closed"
	replacedComment     = "Disconnected because trader received a new connection"
	shutdownComment     = "Disconnected because trader is shutting down"
)

// Temporary constant while we are creating new traders when a connection is established
const initialBalance = 100

// Temporary function while we are creating new traders when a connection is established
func initialStocks() map[uint64]uint64 {
	return map[uint64]uint64{1: 10, 2: 10, 3: 10}
}

type trader struct {
	balance *balanceManager
	// Communication with external system, e.g. a websocket connection
	orders    chan *msg.Message
	responses chan *Response
	// Communication with internal trader server
	intoSvr   chan *msg.Message
	outOfSvr  chan *msg.Message
	connecter chan connect
}

func newTrader(traderId uint32, intoSvr, outOfSvr chan *msg.Message) (*trader, traderComm) {
	balance := newBalanceManager(traderId, initialBalance, initialStocks())
	connecter := make(chan connect)
	t := &trader{balance: balance, intoSvr: intoSvr, outOfSvr: outOfSvr, connecter: connecter}
	tc := traderComm{outOfSvr: outOfSvr, connecter: connecter}
	return t, tc
}

func (t *trader) run() {
	defer t.shutdown()
	for {
		select {
		case con := <-t.connecter:
			t.connect(con)
		case m := <-t.orders:
			if m == nil { // channel has been closed
				t.disconnect(ordersClosedComment)
				continue
			}
			accepted := t.balance.process(m)
			t.sendResponse(m, accepted, "")
			if accepted {
				t.intoSvr <- m
			}
		case m := <-t.outOfSvr:
			accepted := t.balance.process(m)
			t.sendResponse(m, accepted, "")
		}
	}
}

// TODO currently trader never shuts down. How do we want to deal with this?
func (t *trader) shutdown() {
	t.disconnect(shutdownComment)
}

func (t *trader) connect(con connect) {
	t.disconnect(replacedComment)
	t.orders = con.orders
	t.responses = con.responses
	// Send a hello state message
	t.sendResponse(&msg.Message{}, true, connectedComment)
}

func (t *trader) disconnect(comment string) {
	if t.responses != nil {
		t.sendResponse(&msg.Message{}, true, comment)
		close(t.responses)
	}
	t.responses = nil
	t.orders = nil
}

func (t *trader) sendResponse(m *msg.Message, accepted bool, comment string) {
	if t.responses != nil {
		r := t.balance.makeResponse(m, accepted, comment)
		t.responses <- r
	}
}
