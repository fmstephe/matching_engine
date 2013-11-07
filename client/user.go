package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
)

type UserMaker struct {
	intoClient chan interface{}
}

func (um *UserMaker) Connect(traderId uint32) (orders chan WebMessage, responses chan []byte) {
	// TODO in the future this should take an existing user and establish a connection with it
	outOfClient := make(chan *msg.Message)
	cr := &traderRegMsg{traderId: traderId, outOfClient: outOfClient}
	um.intoClient <- cr // Register this comm
	c := comm{traderId: traderId, intoClient: um.intoClient, outOfClient: outOfClient}
	u := newUser(c)
	go u.run()
	return u.orders, u.responses
}

type comm struct {
	traderId    uint32
	intoClient  chan interface{}
	outOfClient chan *msg.Message
}

func (c *comm) Buy(price uint64, tradeId, amount, stockId uint32) error {
	return c.submit(msg.BUY, price, tradeId, amount, stockId)
}

func (c *comm) Sell(price uint64, tradeId, amount, stockId uint32) error {
	return c.submit(msg.SELL, price, tradeId, amount, stockId)
}

func (c *comm) Cancel(price uint64, tradeId, amount, stockId uint32) error {
	return c.submit(msg.CANCEL, price, tradeId, amount, stockId)
}

func (c *comm) submit(kind msg.MsgKind, price uint64, tradeId, amount, stockId uint32) error {
	m := &msg.Message{Kind: kind, Price: price, Amount: amount, TraderId: c.traderId, TradeId: tradeId, StockId: stockId}
	if !m.Valid() {
		return errors.New(fmt.Sprintf("Invalid Message %v", m))
	}
	c.intoClient <- m
	return nil
}

type user struct {
	orders      chan WebMessage
	responses   chan []byte
	curTradeId  uint32
	balance     balanceManager
	stocks      stockManager
	outstanding []WebMessage
	clientComm  comm
}

func newUser(clientComm comm) *user {
	orders := make(chan WebMessage)
	responses := make(chan []byte)
	curTradeId := uint32(1)
	balance := newBalanceManager()
	stocks := newStockManager()
	outstanding := make([]WebMessage, 0)
	return &user{orders: orders, responses: responses, curTradeId: curTradeId, balance: balance, outstanding: outstanding, stocks: stocks, clientComm: clientComm}
}

func (u *user) run() {
	defer close(u.responses)
	orders := u.orders
	responses := u.responses
	outOfClient := u.clientComm.outOfClient
	for {
		var rm receivedMessage
		select {
		case wm := <-orders:
			println("Recieved Message")
			rm = u.processWebMessage(wm)
		case m := <-outOfClient:
			rm = u.processMsg(m)
		}
		r := &response{Received: rm, Balance: u.balance, Stocks: u.stocks, Outstanding: u.outstanding}
		bytes, err := json.Marshal(r)
		if err != nil {
			println("Marshalling Error", err.Error())
		} else {
			responses <- bytes
		}
	}
}

func (u *user) Chans() (orders chan WebMessage, responses chan []byte) {
	return u.orders, u.responses
}

func (u *user) processWebMessage(wm WebMessage) receivedMessage {
	switch wm.Kind {
	case msg.CANCEL.String():
		return u.submitCancel(wm)
	case msg.BUY.String():
		return u.submitBuy(wm)
	case msg.SELL.String():
		return u.submitSell(wm)
	default:
		return receivedMessage{Message: wm, FromClient: true, Accepted: false}
	}
}

func (u *user) submitCancel(c WebMessage) receivedMessage {
	if err := u.clientComm.Cancel(c.Price, c.TradeId, c.Amount, c.StockId); err != nil {
		println("Rejected message: ", err.Error())
		return receivedMessage{Message: c, FromClient: true, Accepted: false}
	}
	u.outstanding = append(u.outstanding, c)
	return receivedMessage{Message: c, FromClient: true, Accepted: true}
}

func (u *user) submitBuy(b WebMessage) receivedMessage {
	b.TradeId = u.curTradeId
	u.curTradeId++
	if !u.balance.canBuy(b.Price, b.Amount) {
		return receivedMessage{Message: b, FromClient: true, Accepted: false}
	}
	if err := u.clientComm.Buy(b.Price, b.TradeId, b.Amount, b.StockId); err != nil {
		println("Rejected message", err.Error())
		return receivedMessage{Message: b, FromClient: true, Accepted: false}
	}
	u.balance.submitBuy(b.Price, b.Amount)
	u.outstanding = append(u.outstanding, b)
	return receivedMessage{Message: b, FromClient: true, Accepted: true}
}

func (u *user) submitSell(s WebMessage) receivedMessage {
	s.TradeId = u.curTradeId
	u.curTradeId++
	if !u.stocks.canSell(s.StockId, s.Amount) {
		return receivedMessage{Message: s, FromClient: true, Accepted: false}
	}
	if err := u.clientComm.Sell(s.Price, s.TradeId, s.Amount, s.StockId); err != nil {
		println("Rejected message", err.Error())
		return receivedMessage{Message: s, FromClient: true, Accepted: false}
	}
	u.stocks.submitSell(s.StockId, s.Amount)
	u.outstanding = append(u.outstanding, s)
	return receivedMessage{Message: s, FromClient: true, Accepted: true}
}

func (u *user) processMsg(m *msg.Message) receivedMessage {
	switch m.Kind {
	case msg.CANCELLED:
		u.cancelOutstanding(m)
	case msg.FULL, msg.PARTIAL:
		u.matchOutstanding(m)
	}
	wm := WebMessage{Kind: m.Kind.String(), Price: m.Price, Amount: m.Amount, StockId: m.StockId, TradeId: m.TradeId}
	return receivedMessage{Message: wm, FromClient: false, Accepted: true}
}

func (u *user) cancelOutstanding(c *msg.Message) {
	newOutstanding := make([]WebMessage, 0, len(u.outstanding))
	for _, wm := range u.outstanding {
		if wm.TradeId != c.TradeId {
			newOutstanding = append(newOutstanding, wm)
		} else {
			switch wm.Kind {
			case msg.BUY.String():
				u.balance.cancelBuy(c.Price, c.Amount)
			case msg.SELL.String():
				u.stocks.cancelSell(c.StockId, c.Amount)
			}
		}
	}
	u.outstanding = newOutstanding
}

func (u *user) matchOutstanding(m *msg.Message) {
	newOutstanding := make([]WebMessage, 0, len(u.outstanding))
	for i, wm := range u.outstanding {
		if wm.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, wm)
		} else {
			if m.Kind == msg.PARTIAL {
				newOutstanding = append(newOutstanding, wm)
				newOutstanding[i].Amount -= m.Amount
			}
			if wm.Kind == msg.SELL.String() {
				u.balance.completeSell(m.Price, m.Amount)
				u.stocks.completeSell(m.StockId, m.Amount)
			}
			if wm.Kind == msg.BUY.String() {
				u.balance.completeBuy(m.Price, m.Amount)
				u.stocks.completeBuy(m.StockId, m.Amount)
			}
		}
	}
	u.outstanding = newOutstanding
}
