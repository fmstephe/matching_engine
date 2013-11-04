package main

import (
	"encoding/json"
	"github.com/fmstephe/matching_engine/client"
	"github.com/fmstephe/matching_engine/msg"
	"strconv"
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
	Accepted   bool       `json:"accepted"`
	Message    webMessage `json:"message"`
}

type response struct {
	Balance     balanceManager  `json:"balance"`
	Stocks      stockManager    `json:"stocks"`
	Received    receivedMessage `json:"received"`
	Outstanding []webMessage    `json:"outstanding"`
}

type balanceManager struct {
	Current   int64 `json:"current"`
	Available int64 `json:"available"`
}

func newBalanceManager() balanceManager {
	bal := int64(100)
	return balanceManager{Current: bal, Available: bal}
}

func (bm *balanceManager) total(price int64, amount uint32) int64 {
	return price * int64(amount)
}

func (bm *balanceManager) canBuy(price int64, amount uint32) bool {
	return bm.Available >= bm.total(price, amount)
}

func (bm *balanceManager) submitBuy(price int64, amount uint32) {
	bm.Available -= bm.total(price, amount)
}

func (bm *balanceManager) cancelBuy(price int64, amount uint32) {
	bm.Available += bm.total(price, amount)
}

func (bm *balanceManager) completeBuy(price int64, amount uint32) {
	// TODO bm.Current -= bm.total(price, amount)
	bm.Current += bm.total(price, amount)
}

func (bm *balanceManager) completeSell(price int64, amount uint32) {
	total := bm.total(price, amount)
	bm.Current += total
	bm.Available += total
}

// TODO the naming of this hasn't worked out very well
type stockManager struct {
	StocksHeld   map[string]uint32 `json:"stocksHeld"`
	StocksToSell map[string]uint32 `json:"stocksToSell"`
}

func newStockManager() stockManager {
	sm := stockManager{}
	sm.StocksHeld = map[string]uint32{"1": 10, "2": 10, "3": 10}
	sm.StocksToSell = make(map[string]uint32)
	return sm
}

func (sm *stockManager) getKey(stockId uint32) string {
	return strconv.Itoa(int(stockId))
}

func (sm *stockManager) cleanup(stockKey string) {
	if sm.StocksToSell[stockKey] == 0 {
		delete(sm.StocksToSell, stockKey)
	}
	if sm.StocksHeld[stockKey] == 0 {
		delete(sm.StocksHeld, stockKey)
	}
}

func (sm *stockManager) held(stockKey string) uint32 {
	return sm.StocksHeld[stockKey]
}

func (sm *stockManager) addHeld(stockKey string, amount uint32) {
	held := sm.StocksHeld[stockKey]
	sm.StocksHeld[stockKey] = held + amount
}

func (sm *stockManager) toSell(stockKey string) uint32 {
	return sm.StocksToSell[stockKey]
}

func (sm *stockManager) addToSell(stockKey string, amount uint32) {
	toSell := sm.StocksToSell[stockKey]
	sm.StocksToSell[stockKey] = toSell + amount
}

func (sm *stockManager) canSell(stockId, amount uint32) bool {
	stockKey := sm.getKey(stockId)
	return sm.held(stockKey) >= amount
}

func (sm *stockManager) submitSell(stockId, amount uint32) {
	stockKey := sm.getKey(stockId)
	sm.addHeld(stockKey, -amount)
	sm.addToSell(stockKey, amount)
	// Don't clean up, we want the zeroed held stocks to remain
}

func (sm *stockManager) cancelSell(stockId, amount uint32) {
	stockKey := sm.getKey(stockId)
	sm.addHeld(stockKey, amount)
	sm.addToSell(stockKey, -amount)
	sm.cleanup(stockKey)
}

func (sm *stockManager) completeSell(stockId, amount uint32) {
	stockKey := sm.getKey(stockId)
	sm.addToSell(stockKey, -amount)
	sm.cleanup(stockKey)
}

func (sm *stockManager) completeBuy(stockId, amount uint32) {
	stockKey := sm.getKey(stockId)
	sm.addHeld(stockKey, amount)
}

type user struct {
	curTradeId  uint32
	balance     balanceManager
	stocks      stockManager
	outstanding []webMessage
	clientComm  *client.Comm
}

func newUser(clientComm *client.Comm) *user {
	curTradeId := uint32(1)
	outstanding := make([]webMessage, 0)
	balance := newBalanceManager()
	stocks := newStockManager()
	return &user{curTradeId: curTradeId, balance: balance, outstanding: outstanding, stocks: stocks, clientComm: clientComm}
}

func (u *user) run(msgs chan webMessage, responses chan []byte) {
	defer close(responses)
	outOfClient := u.clientComm.Out()
	for {
		var rm receivedMessage
		select {
		case wm := <-msgs:
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

func (u *user) processWebMessage(wm webMessage) receivedMessage {
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

func (u *user) submitCancel(c webMessage) receivedMessage {
	if err := u.clientComm.Cancel(c.Price, c.TradeId, c.Amount, c.StockId); err != nil {
		println("Rejected message: ", err.Error())
		return receivedMessage{Message: c, FromClient: true, Accepted: false}
	}
	u.outstanding = append(u.outstanding, c)
	return receivedMessage{Message: c, FromClient: true, Accepted: true}
}

func (u *user) submitBuy(b webMessage) receivedMessage {
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

func (u *user) submitSell(s webMessage) receivedMessage {
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
	wm := webMessage{Kind: m.Kind.String(), Price: m.Price, Amount: m.Amount, StockId: m.StockId, TradeId: m.TradeId}
	return receivedMessage{Message: wm, FromClient: false, Accepted: true}
}

func (u *user) cancelOutstanding(c *msg.Message) {
	newOutstanding := make([]webMessage, 0, len(u.outstanding))
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
	newOutstanding := make([]webMessage, 0, len(u.outstanding))
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
