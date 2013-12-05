package client

import (
	"encoding/json"
	"github.com/fmstephe/matching_engine/msg"
)

const (
	connectedComment    = "Connected to trader"
	ordersClosedComment = "Disconnected because orders channel was closed"
	replacedComment     = "Disconnected because trader received a new connection"
	shutdownComment     = "Disconnected because trader is shutting down"
)

type trader struct {
	traderId    uint32
	curTradeId  uint32
	balance     balanceManager
	stocks      stockManager
	outstanding []msg.Message
	// Communication with external system, e.g. a websocket connection
	orders    chan *msg.Message
	responses chan []byte // serialised &response{} structs
	// Communication with internal trader server
	intoSvr   chan *msg.Message
	outOfSvr  chan *msg.Message
	connecter chan connect
}

func newTrader(traderId uint32, intoSvr, outOfSvr chan *msg.Message) (*trader, traderComm) {
	curTradeId := uint32(1)
	balance := newBalanceManager()
	stocks := newStockManager()
	outstanding := make([]msg.Message, 0)
	connecter := make(chan connect)
	t := &trader{traderId: traderId, curTradeId: curTradeId, balance: balance, outstanding: outstanding, stocks: stocks, intoSvr: intoSvr, outOfSvr: outOfSvr, connecter: connecter}
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
			accepted := t.process(m)
			t.sendState(m, accepted, "")
			if accepted {
				t.intoSvr <- m
			}
		case m := <-t.outOfSvr:
			accepted := t.process(m)
			t.sendState(m, accepted, "")
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
	t.sendState(&msg.Message{}, true, connectedComment)
}

func (t *trader) disconnect(comment string) {
	if t.responses != nil {
		t.sendState(&msg.Message{}, true, comment)
		close(t.responses)
	}
	t.responses = nil
	t.orders = nil
}

func (t *trader) sendState(m *msg.Message, accepted bool, comment string) {
	if t.responses != nil {
		rm := receivedMessage{Message: *m, Accepted: accepted}
		s := traderState{Balance: t.balance, Stocks: t.stocks, Outstanding: t.outstanding}
		r := response{State: s, Received: rm, Comment: comment}
		b, err := json.Marshal(r)
		if err != nil {
			println("Marshalling Error: ", err.Error())
		}
		t.responses <- b
	}
}

func (t *trader) process(m *msg.Message) bool {
	m.TraderId = t.traderId
	switch m.Kind {
	case msg.CANCEL:
		return t.submitCancel(m)
	case msg.BUY:
		t.curTradeId++
		m.TradeId = t.curTradeId
		return t.submitBuy(m)
	case msg.SELL:
		t.curTradeId++
		m.TradeId = t.curTradeId
		return t.submitSell(m)
	case msg.CANCELLED:
		return t.cancelOutstanding(m)
	case msg.FULL, msg.PARTIAL:
		return t.matchOutstanding(m)
	}
	return false
}

func (t *trader) submitCancel(m *msg.Message) bool {
	t.outstanding = append(t.outstanding, *m)
	return true
}

func (t *trader) submitBuy(m *msg.Message) bool {
	if !t.balance.canBuy(m.Price, m.Amount) {
		return false
	}
	t.balance.submitBuy(m.Price, m.Amount)
	t.outstanding = append(t.outstanding, *m)
	return true
}

func (t *trader) submitSell(m *msg.Message) bool {
	if !t.stocks.canSell(m.StockId, m.Amount) {
		return false
	}
	t.stocks.submitSell(m.StockId, m.Amount)
	t.outstanding = append(t.outstanding, *m)
	return true
}

func (t *trader) cancelOutstanding(m *msg.Message) bool {
	accepted := false
	newOutstanding := make([]msg.Message, 0, len(t.outstanding))
	for _, om := range t.outstanding {
		if om.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, om)
		} else {
			accepted = true
			switch om.Kind {
			case msg.BUY:
				t.balance.cancelBuy(m.Price, m.Amount)
			case msg.SELL:
				t.stocks.cancelSell(m.StockId, m.Amount)
			}
		}
	}
	t.outstanding = newOutstanding
	return accepted
}

func (t *trader) matchOutstanding(m *msg.Message) bool {
	accepted := false
	newOutstanding := make([]msg.Message, 0, len(t.outstanding))
	for i, om := range t.outstanding {
		if om.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, om)
		} else {
			accepted = true
			if m.Kind == msg.PARTIAL {
				newOutstanding = append(newOutstanding, om)
				newOutstanding[i].Amount -= m.Amount
			}
			if om.Kind == msg.SELL {
				t.balance.completeSell(m.Price, m.Amount)
				t.stocks.completeSell(m.StockId, m.Amount)
			}
			if om.Kind == msg.BUY {
				t.balance.completeBuy(om.Price, m.Price, m.Amount)
				t.stocks.completeBuy(m.StockId, m.Amount)
			}
		}
	}
	t.outstanding = newOutstanding
	return accepted
}
