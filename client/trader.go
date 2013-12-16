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

// Temporary constant while we are creating new traders when a connection is established
const initialBalance = 100

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
	curTradeId := uint32(0)
	balance := newBalanceManager(initialBalance)
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
			t.sendResponse(m, accepted, "")
			if accepted {
				t.intoSvr <- m
			}
		case m := <-t.outOfSvr:
			accepted := t.process(m)
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
		r := t.makeResponse(m, accepted, comment)
		b, err := json.Marshal(r)
		if err != nil {
			// TODO what should we send to responses if we had a marshalling error?
			println("Marshalling Error: ", err.Error())
		}
		t.responses <- b
	}
}

func (t *trader) makeResponse(m *msg.Message, accepted bool, comment string) *response {
	rm := receivedMessage{Message: *m, Accepted: accepted}
	s := traderState{Balance: t.balance, Stocks: t.stocks, Outstanding: t.outstanding}
	return &response{State: s, Received: rm, Comment: comment}
}

// NB: After this method returns BUYs and SELLs are guaranteed to have the correct TradeId
// BUYs, SELLs and CANCELs are guaranteed to have the correct TraderId
// CANCELLEDs and FULLs are assumed to have the correct values and are unchanged
func (t *trader) process(m *msg.Message) bool {
	switch m.Kind {
	case msg.CANCEL:
		m.TraderId = t.traderId
		return t.submitCancel(m)
	case msg.BUY:
		m.TraderId = t.traderId
		t.curTradeId++
		m.TradeId = t.curTradeId
		return t.submitBuy(m)
	case msg.SELL:
		m.TraderId = t.traderId
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
