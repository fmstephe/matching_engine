package client

import (
	"github.com/fmstephe/matching_engine/msg"
	"strconv"
)

type balanceManager struct {
	traderId    uint32
	curTradeId  uint32
	outstanding []msg.Message
	current     uint64
	available   uint64
	held        map[uint64]uint64
	toSell      map[uint64]uint64
}

func newBalanceManager(traderId uint32, balance uint64, initialStocks map[uint64]uint64) *balanceManager {
	bm := &balanceManager{current: balance, available: balance}
	bm.traderId = traderId
	bm.curTradeId = 0
	bm.held = make(map[uint64]uint64)
	bm.toSell = make(map[uint64]uint64)
	for k, v := range initialStocks {
		bm.held[k] = v
	}
	return bm
}

// NB: After this method returns BUYs and SELLs are guaranteed to have the correct TradeId
// BUYs, SELLs and CANCELs are guaranteed to have the correct TraderId
// CANCELLEDs and FULLs are assumed to have the correct values and are unchanged
func (bm *balanceManager) process(m *msg.Message) bool {
	switch m.Kind {
	case msg.CANCEL:
		m.TraderId = bm.traderId
		return bm.processCancel(m)
	case msg.BUY:
		m.TraderId = bm.traderId
		bm.curTradeId++
		m.TradeId = bm.curTradeId
		return bm.processBuy(m)
	case msg.SELL:
		m.TraderId = bm.traderId
		bm.curTradeId++
		m.TradeId = bm.curTradeId
		return bm.processSell(m)
	case msg.CANCELLED:
		return bm.cancelOutstanding(m)
	case msg.FULL, msg.PARTIAL:
		return bm.matchOutstanding(m)
	}
	return false
}

func (bm *balanceManager) processCancel(m *msg.Message) bool {
	bm.outstanding = append(bm.outstanding, *m)
	return true
}

func (bm *balanceManager) processBuy(m *msg.Message) bool {
	if bm.available < (m.Price * m.Amount) {
		return false
	}
	bm.available -= m.Price * m.Amount
	bm.outstanding = append(bm.outstanding, *m)
	return true
}

func (bm *balanceManager) processSell(m *msg.Message) bool {
	if bm.held[m.StockId] < m.Amount {
		return false
	}
	bm.addHeld(m.StockId, -m.Amount)
	bm.addToSell(m.StockId, m.Amount)
	bm.outstanding = append(bm.outstanding, *m)
	// Don't clean up, we want the zeroed held stocks to remain
	return true
}

func (bm *balanceManager) cancelOutstanding(m *msg.Message) bool {
	accepted := false
	newOutstanding := make([]msg.Message, 0, len(bm.outstanding))
	for _, om := range bm.outstanding {
		if om.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, om)
		} else {
			accepted = true
			switch om.Kind {
			case msg.BUY:
				bm.cancelBuy(m.Price, m.Amount)
			case msg.SELL:
				bm.cancelSell(m.StockId, m.Amount)
			}
		}
	}
	bm.outstanding = newOutstanding
	return accepted
}

func (bm *balanceManager) matchOutstanding(m *msg.Message) bool {
	accepted := false
	newOutstanding := make([]msg.Message, 0, len(bm.outstanding))
	for i, om := range bm.outstanding {
		if om.TradeId != m.TradeId {
			newOutstanding = append(newOutstanding, om)
		} else {
			accepted = true
			if m.Kind == msg.PARTIAL {
				newOutstanding = append(newOutstanding, om)
				newOutstanding[i].Amount -= m.Amount
			}
			if om.Kind == msg.SELL {
				bm.completeSell(m.StockId, m.Price, m.Amount)
			}
			if om.Kind == msg.BUY {
				bm.completeBuy(m.StockId, om.Price, m.Price, m.Amount)
			}
		}
	}
	bm.outstanding = newOutstanding
	return accepted
}

func (bm *balanceManager) cleanup(stockId uint64) {
	if bm.toSell[stockId] == 0 {
		delete(bm.toSell, stockId)
	}
	if bm.held[stockId] == 0 {
		delete(bm.held, stockId)
	}
}

func (bm *balanceManager) addHeld(stockId, amount uint64) {
	held := bm.held[stockId]
	bm.held[stockId] = held + amount
}

func (bm *balanceManager) addToSell(stockId, amount uint64) {
	toSell := bm.toSell[stockId]
	bm.toSell[stockId] = toSell + amount
}

func (bm *balanceManager) cancelBuy(price, amount uint64) {
	bm.available += price * amount
}

func (bm *balanceManager) cancelSell(stockId, amount uint64) {
	bm.addHeld(stockId, amount)
	bm.addToSell(stockId, -amount)
	bm.cleanup(stockId)
}

func (bm *balanceManager) completeBuy(stockId, bidPrice, actualPrice, amount uint64) {
	bidTotal := bidPrice * amount
	actualTotal := actualPrice * amount
	bm.available += bidTotal
	bm.available -= actualTotal
	bm.current -= actualTotal
	bm.addHeld(stockId, amount)
}

func (bm *balanceManager) completeSell(stockId, price, amount uint64) {
	total := price * amount
	bm.current += total
	bm.available += total
	bm.addToSell(stockId, -amount)
	bm.cleanup(stockId)
}

func (bm *balanceManager) makeResponse(m *msg.Message, accepted bool, comment string) *Response {
	rm := receivedMessage{Message: *m, Accepted: accepted}
	current := bm.current
	available := bm.available
	held := mapToResponse(bm.held)
	toSell := mapToResponse(bm.toSell)
	os := make([]msg.Message, len(bm.outstanding))
	copy(os, bm.outstanding)
	s := traderState{CurrentBalance: current, AvailableBalance: available, StocksHeld: held, StocksToSell: toSell, Outstanding: os}
	return &Response{State: s, Received: rm, Comment: comment}
}

func mapToResponse(in map[uint64]uint64) map[string]uint64 {
	out := make(map[string]uint64)
	for k, v := range in {
		ks := strconv.FormatUint(k, 10)
		out[ks] = v
	}
	return out
}
