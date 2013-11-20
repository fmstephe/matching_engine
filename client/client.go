package client

import (
	"encoding/json"
	"github.com/fmstephe/matching_engine/msg"
)

type client struct {
	traderId    uint32
	curTradeId  uint32
	balance     balanceManager
	stocks      stockManager
	outstanding []msg.Message
	// Communication with external system, e.g. a websocket connection
	orders    chan *msg.Message
	responses chan []byte // serialised &response{} structs
	// Communication with internal client server
	intoSvr   chan *msg.Message
	outOfSvr  chan *msg.Message
	connecter chan connect
}

func newClient(traderId uint32, intoSvr, outOfSvr chan *msg.Message) (*client, clientComm) {
	curTradeId := uint32(1)
	balance := newBalanceManager()
	stocks := newStockManager()
	outstanding := make([]msg.Message, 0)
	connecter := make(chan connect)
	c := &client{traderId: traderId, curTradeId: curTradeId, balance: balance, outstanding: outstanding, stocks: stocks, intoSvr: intoSvr, outOfSvr: outOfSvr, connecter: connecter}
	cc := clientComm{outOfSvr: outOfSvr, connecter: connecter}
	return c, cc
}

func (c *client) run() {
	defer close(c.responses)
	for {
		select {
		case con := <-c.connecter:
			c.connectTo(con)
		case m := <-c.orders:
			accepted := c.process(m)
			c.sendState(m, accepted)
		case m := <-c.outOfSvr:
			accepted := c.process(m)
			c.sendState(m, accepted)
		}
	}
}

func (c *client) connectTo(con connect) {
	if c.responses != nil {
		// TODO we need to send the old connection an explanatory message before we close
		close(c.responses)
	}
	c.orders = con.orders
	c.responses = con.responses
}

func (c *client) sendState(m *msg.Message, accepted bool) {
	if c.responses != nil {
		rm := receivedMessage{Message: *m, Accepted: accepted}
		s := clientState{Balance: c.balance, Stocks: c.stocks, Outstanding: c.outstanding}
		r := response{State: s, Received: rm}
		b, err := json.Marshal(r)
		if err != nil {
			println("Marshalling Error: ", err.Error())
		}
		c.responses <- b
	}
}

func (c *client) process(m *msg.Message) bool {
	m.TraderId = c.traderId
	switch m.Kind {
	case msg.CANCEL:
		return c.submitCancel(m)
	case msg.BUY:
		c.curTradeId++
		m.TradeId = c.curTradeId
		return c.submitBuy(m)
	case msg.SELL:
		c.curTradeId++
		m.TradeId = c.curTradeId
		return c.submitSell(m)
	case msg.CANCELLED:
		return c.cancelOutstanding(m)
	case msg.FULL, msg.PARTIAL:
		return c.matchOutstanding(m)
	}
	return false
}

func (c *client) submitCancel(m *msg.Message) bool {
	c.outstanding = append(c.outstanding, *m)
	c.intoSvr <- m
	return true
}

func (c *client) submitBuy(m *msg.Message) bool {
	if !c.balance.canBuy(m.Price, m.Amount) {
		return false
	}
	c.balance.submitBuy(m.Price, m.Amount)
	c.outstanding = append(c.outstanding, *m)
	c.intoSvr <- m
	return true
}

func (c *client) submitSell(m *msg.Message) bool {
	if !c.stocks.canSell(m.StockId, m.Amount) {
		return false
	}
	c.stocks.submitSell(m.StockId, m.Amount)
	c.outstanding = append(c.outstanding, *m)
	c.intoSvr <- m
	return true
}

func (c *client) cancelOutstanding(m *msg.Message) bool {
	accepted := false
	newOutstanding := make([]msg.Message, 0, len(c.outstanding))
	for _, om := range c.outstanding {
		if om.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, om)
		} else {
			accepted = true
			switch om.Kind {
			case msg.BUY:
				c.balance.cancelBuy(m.Price, m.Amount)
			case msg.SELL:
				c.stocks.cancelSell(m.StockId, m.Amount)
			}
		}
	}
	c.outstanding = newOutstanding
	return accepted
}

func (c *client) matchOutstanding(m *msg.Message) bool {
	accepted := false
	newOutstanding := make([]msg.Message, 0, len(c.outstanding))
	for i, om := range c.outstanding {
		if om.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, om)
		} else {
			accepted = true
			if m.Kind == msg.PARTIAL {
				newOutstanding = append(newOutstanding, om)
				newOutstanding[i].Amount -= m.Amount
			}
			if om.Kind == msg.SELL {
				c.balance.completeSell(m.Price, m.Amount)
				c.stocks.completeSell(m.StockId, m.Amount)
			}
			if om.Kind == msg.BUY {
				c.balance.completeBuy(m.Price, m.Amount)
				c.stocks.completeBuy(m.StockId, m.Amount)
			}
		}
	}
	c.outstanding = newOutstanding
	return accepted
}
