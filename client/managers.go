package client

import ()

// TODO should rethink the approach taken to make this code safer, right now the managers are just adders and subtracters. No safety checking is being done
type balanceManager struct {
	current   uint64
	available uint64
}

func newBalanceManager(balance uint64) balanceManager {
	return balanceManager{current: balance, available: balance}
}

func (bm *balanceManager) total(price, amount uint64) uint64 {
	return price * uint64(amount)
}

func (bm *balanceManager) canBuy(price, amount uint64) bool {
	return bm.available >= bm.total(price, amount)
}

// TODO if this wraps below 0 we really need to cope with that error?
func (bm *balanceManager) submitBuy(price, amount uint64) {
	bm.available -= bm.total(price, amount)
}

func (bm *balanceManager) cancelBuy(price, amount uint64) {
	bm.available += bm.total(price, amount)
}

func (bm *balanceManager) completeBuy(bidPrice, actualPrice, amount uint64) {
	bidTotal := bm.total(bidPrice, amount)
	actualTotal := bm.total(actualPrice, amount)
	bm.available += bidTotal
	bm.available -= actualTotal
	bm.current -= actualTotal
}

func (bm *balanceManager) completeSell(price, amount uint64) {
	total := bm.total(price, amount)
	bm.current += total
	bm.available += total
}

type stockManager struct {
	held   map[uint64]uint64
	toSell map[uint64]uint64
}

func newStockManager(initialStocks map[uint64]uint64) stockManager {
	sm := stockManager{}
	sm.held = make(map[uint64]uint64)
	sm.toSell = make(map[uint64]uint64)
	for k, v := range initialStocks {
		sm.held[k] = v
	}
	return sm
}

func (sm *stockManager) cleanup(stockId uint64) {
	if sm.toSell[stockId] == 0 {
		delete(sm.toSell, stockId)
	}
	if sm.held[stockId] == 0 {
		delete(sm.held, stockId)
	}
}

func (sm *stockManager) getHeld(stockId uint64) uint64 {
	return sm.held[stockId]
}

func (sm *stockManager) addHeld(stockId, amount uint64) {
	held := sm.held[stockId]
	sm.held[stockId] = held + amount
}

func (sm *stockManager) getToSell(stockId uint64) uint64 {
	return sm.toSell[stockId]
}

func (sm *stockManager) addToSell(stockId, amount uint64) {
	toSell := sm.toSell[stockId]
	sm.toSell[stockId] = toSell + amount
}

func (sm *stockManager) canSell(stockId, amount uint64) bool {
	return sm.getHeld(stockId) >= amount
}

func (sm *stockManager) submitSell(stockId, amount uint64) {
	sm.addHeld(stockId, -amount)
	sm.addToSell(stockId, amount)
	// Don't clean up, we want the zeroed held stocks to remain
}

func (sm *stockManager) cancelSell(stockId, amount uint64) {
	sm.addHeld(stockId, amount)
	sm.addToSell(stockId, -amount)
	sm.cleanup(stockId)
}

func (sm *stockManager) completeSell(stockId, amount uint64) {
	sm.addToSell(stockId, -amount)
	sm.cleanup(stockId)
}

func (sm *stockManager) completeBuy(stockId, amount uint64) {
	sm.addHeld(stockId, amount)
}
