package main

import (
	"encoding/json"
	"github.com/fmstephe/matching_engine/client"
	"github.com/fmstephe/matching_engine/msg"
)

type webMessage struct {
	Kind    string `json:"kind"`
	Price   int64  `json:"price"`
	Amount  uint32 `json:"amount"`
	StockId uint32 `json:"stockId"`
	TradeId uint32 `json:"tradeId"`
}

type receivedMessage struct {
	FromClient bool       `json:"fromClient"`
	Accepted   bool       `json:accepted`
	Message    webMessage `json:"message"`
}

type response struct {
	AvailBal    int64           `json:"availBal"`
	CurBal      int64           `json:"curBal"`
	Received    receivedMessage `json:"received"`
	Outstanding []webMessage    `json:"outstanding"`
}

type user struct {
	curTradeId  uint32
	availBal    int64
	curBal      int64
	clientComm  *client.Comm
	outstanding []webMessage
}

func newUser(clientComm *client.Comm) *user {
	curTradeId := uint32(1)
	bal := int64(100)
	outstanding := make([]webMessage, 0)
	return &user{curTradeId: curTradeId, availBal: bal, curBal: bal, clientComm: clientComm, outstanding: outstanding}
}

func (u *user) run(msgs chan webMessage, responses chan []byte) {
	defer close(responses)
	outOfClient := u.clientComm.Out()
	for {
		var rm receivedMessage
		select {
		case wm := <-msgs:
			rm = u.processOrder(wm)
		case m := <-outOfClient:
			rm = u.processMsg(m)
		}
		r := &response{Received: rm, AvailBal: u.availBal, CurBal: u.curBal, Outstanding: u.outstanding}
		bytes, err := json.Marshal(r)
		if err != nil {
			println("Marshalling Error", err.Error())
		} else {
			responses <- bytes
		}
	}
}

func (u *user) processOrder(wm webMessage) receivedMessage {
	if wm.Kind == "CANCEL" {
		u.clientComm.Cancel(wm.Price, wm.TradeId, wm.Amount, wm.StockId)
		u.outstanding = append(u.outstanding, wm)
		return receivedMessage{Message: wm, FromClient: true, Accepted: true}
	}
	wm.TradeId = u.curTradeId
	u.curTradeId++
	totalCost := wm.Price * int64(wm.Amount)
	if wm.Kind == "BUY" && totalCost > u.availBal {
		return receivedMessage{Message: wm, FromClient: true, Accepted: false}
	}
	if wm.Kind == "BUY" {
		if err := u.clientComm.Buy(wm.Price, wm.TradeId, wm.Amount, wm.StockId); err != nil {
			return receivedMessage{Message: wm, FromClient: true, Accepted: false}
		}
		u.availBal -= totalCost
		u.outstanding = append(u.outstanding, wm)
		return receivedMessage{Message: wm, FromClient: true, Accepted: true}
	}
	if wm.Kind == "SELL" {
		if err := u.clientComm.Sell(wm.Price, wm.TradeId, wm.Amount, wm.StockId); err != nil {
			return receivedMessage{Message: wm, FromClient: true, Accepted: false}
		}
		u.outstanding = append(u.outstanding, wm)
		return receivedMessage{Message: wm, FromClient: true, Accepted: true}
	}
	return receivedMessage{Message: wm, FromClient: true, Accepted: false}
}

func (u *user) processMsg(m *msg.Message) receivedMessage {
	switch m.Kind {
	case msg.CANCELLED:
		u.cancelOutstanding(m)
	case msg.FULL, msg.PARTIAL:
		u.matchOutstanding(m)
	}
	wm := webMessage{Kind: m.Kind.String(), Price: m.Price, Amount: m.Amount, StockId: m.StockId, TradeId: m.TradeId}
	return receivedMessage{Message: wm, FromClient: false, Accepted: true}
}

func (u *user) cancelOutstanding(c *msg.Message) {
	newOutstanding := make([]webMessage, 0, len(u.outstanding))
	for _, wm := range u.outstanding {
		if wm.TradeId != c.TradeId {
			newOutstanding = append(newOutstanding, wm)
		} else {
			if wm.Kind == msg.BUY.String() {
				// If buy then we need to increase the available balance
				totalCost := c.Price * int64(c.Amount)
				u.availBal += totalCost
			}
		}
	}
	u.outstanding = newOutstanding
}

func (u *user) matchOutstanding(m *msg.Message) {
	totalCost := m.Price * int64(m.Amount)
	u.curBal += totalCost
	// If sell (totalCost>0) then available balance is updated
	// If buy (totalCost<0) then available balance has already been updated
	if totalCost > 0 {
		u.availBal += totalCost
	}
	newOutstanding := make([]webMessage, 0, len(u.outstanding))
	for i, wm := range u.outstanding {
		if wm.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, wm)
		} else if m.Kind == msg.PARTIAL {
			newOutstanding = append(newOutstanding, wm)
			newOutstanding[i].Amount -= m.Amount
		}
	}
	u.outstanding = newOutstanding
}
