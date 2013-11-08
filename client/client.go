package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
)

type client struct {
	traderId    uint32
	curTradeId  uint32
	balance     balanceManager
	stocks      stockManager
	outstanding []ClientMessage
	// Communication with external system, e.g. a websocket connection
	orders    chan ClientMessage
	responses chan []byte // serialised &response{} structs
	// Communication with internal client server
	intoSvr  chan interface{}
	outOfSvr chan *msg.Message
}

func newClient(traderId uint32, intoSvr chan interface{}, outOfSvr chan *msg.Message) *client {
	orders := make(chan ClientMessage)
	responses := make(chan []byte)
	curTradeId := uint32(1)
	balance := newBalanceManager()
	stocks := newStockManager()
	outstanding := make([]ClientMessage, 0)
	return &client{traderId: traderId, orders: orders, responses: responses, curTradeId: curTradeId, balance: balance, outstanding: outstanding, stocks: stocks, intoSvr: intoSvr, outOfSvr: outOfSvr}
}

func (c *client) run() {
	defer close(c.responses)
	orders := c.orders
	responses := c.responses
	outOfSvr := c.outOfSvr
	for {
		var rm receivedMessage
		select {
		case cm := <-orders:
			rm = c.processClientMessage(cm)
		case m := <-outOfSvr:
			rm = c.processServerMessage(m)
		}
		r := &response{Received: rm, Balance: c.balance, Stocks: c.stocks, Outstanding: c.outstanding}
		bytes, err := json.Marshal(r)
		if err != nil {
			println("Marshalling Error", err.Error())
		} else {
			responses <- bytes
		}
	}
}

func (c *client) Chans() (orders chan ClientMessage, responses chan []byte) {
	return c.orders, c.responses
}

func (c *client) processClientMessage(cm ClientMessage) receivedMessage {
	switch cm.Kind {
	case msg.CANCEL.String():
		return c.submitCancel(cm)
	case msg.BUY.String():
		return c.submitBuy(cm)
	case msg.SELL.String():
		return c.submitSell(cm)
	default:
		return receivedMessage{Message: cm, FromClient: true, Accepted: false}
	}
}

func (c *client) submitCancel(cm ClientMessage) receivedMessage {
	if err := c.submit(msg.CANCEL, cm.Price, cm.TradeId, cm.Amount, cm.StockId); err != nil {
		println("Rejected message: ", err.Error())
		return receivedMessage{Message: cm, FromClient: true, Accepted: false}
	}
	c.outstanding = append(c.outstanding, cm)
	return receivedMessage{Message: cm, FromClient: true, Accepted: true}
}

func (c *client) submitBuy(cm ClientMessage) receivedMessage {
	cm.TradeId = c.curTradeId
	c.curTradeId++
	if !c.balance.canBuy(cm.Price, cm.Amount) {
		return receivedMessage{Message: cm, FromClient: true, Accepted: false}
	}
	if err := c.submit(msg.BUY, cm.Price, cm.TradeId, cm.Amount, cm.StockId); err != nil {
		println("Rejected message", err.Error())
		return receivedMessage{Message: cm, FromClient: true, Accepted: false}
	}
	c.balance.submitBuy(cm.Price, cm.Amount)
	c.outstanding = append(c.outstanding, cm)
	return receivedMessage{Message: cm, FromClient: true, Accepted: true}
}

func (c *client) submitSell(cm ClientMessage) receivedMessage {
	cm.TradeId = c.curTradeId
	c.curTradeId++
	if !c.stocks.canSell(cm.StockId, cm.Amount) {
		return receivedMessage{Message: cm, FromClient: true, Accepted: false}
	}
	if err := c.submit(msg.SELL, cm.Price, cm.TradeId, cm.Amount, cm.StockId); err != nil {
		println("Rejected message", err.Error())
		return receivedMessage{Message: cm, FromClient: true, Accepted: false}
	}
	c.stocks.submitSell(cm.StockId, cm.Amount)
	c.outstanding = append(c.outstanding, cm)
	return receivedMessage{Message: cm, FromClient: true, Accepted: true}
}

func (c *client) submit(kind msg.MsgKind, price uint64, tradeId, amount, stockId uint32) error {
	m := &msg.Message{Kind: kind, Price: price, Amount: amount, TraderId: c.traderId, TradeId: tradeId, StockId: stockId}
	if !m.Valid() {
		return errors.New(fmt.Sprintf("Invalid Message %v", m))
	}
	c.intoSvr <- m
	return nil
}

func (c *client) processServerMessage(m *msg.Message) receivedMessage {
	switch m.Kind {
	case msg.CANCELLED:
		c.cancelOutstanding(m)
	case msg.FULL, msg.PARTIAL:
		c.matchOutstanding(m)
	}
	cm := ClientMessage{Kind: m.Kind.String(), Price: m.Price, Amount: m.Amount, StockId: m.StockId, TradeId: m.TradeId}
	return receivedMessage{Message: cm, FromClient: false, Accepted: true}
}

func (c *client) cancelOutstanding(m *msg.Message) {
	newOutstanding := make([]ClientMessage, 0, len(c.outstanding))
	for _, cm := range c.outstanding {
		if cm.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, cm)
		} else {
			switch cm.Kind {
			case msg.BUY.String():
				c.balance.cancelBuy(m.Price, m.Amount)
			case msg.SELL.String():
				c.stocks.cancelSell(m.StockId, m.Amount)
			}
		}
	}
	c.outstanding = newOutstanding
}

func (c *client) matchOutstanding(m *msg.Message) {
	newOutstanding := make([]ClientMessage, 0, len(c.outstanding))
	for i, cm := range c.outstanding {
		if cm.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, cm)
		} else {
			if m.Kind == msg.PARTIAL {
				newOutstanding = append(newOutstanding, cm)
				newOutstanding[i].Amount -= m.Amount
			}
			if cm.Kind == msg.SELL.String() {
				c.balance.completeSell(m.Price, m.Amount)
				c.stocks.completeSell(m.StockId, m.Amount)
			}
			if cm.Kind == msg.BUY.String() {
				c.balance.completeBuy(m.Price, m.Amount)
				c.stocks.completeBuy(m.StockId, m.Amount)
			}
		}
	}
	c.outstanding = newOutstanding
}
