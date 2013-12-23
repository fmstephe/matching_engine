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

func (bm *balanceManager) total(price uint64, amount uint32) uint64 {
	return price * uint64(amount)
}

func (bm *balanceManager) canBuy(price uint64, amount uint32) bool {
	return bm.available >= bm.total(price, amount)
}

// TODO if this wraps below 0 we really need to cope with that error?
func (bm *balanceManager) submitBuy(price uint64, amount uint32) {
	bm.available -= bm.total(price, amount)
}

func (bm *balanceManager) cancelBuy(price uint64, amount uint32) {
	bm.available += bm.total(price, amount)
}

func (bm *balanceManager) completeBuy(bidPrice, actualPrice uint64, amount uint32) {
	bidTotal := bm.total(bidPrice, amount)
	actualTotal := bm.total(actualPrice, amount)
	bm.available += bidTotal
	bm.available -= actualTotal
	bm.current -= actualTotal
}

func (bm *balanceManager) completeSell(price uint64, amount uint32) {
	total := bm.total(price, amount)
	bm.current += total
	bm.available += total
}

type stockManager struct {
	held   map[uint32]uint32
	toSell map[uint32]uint32
}

func newStockManager(initialStocks map[uint32]uint32) stockManager {
	sm := stockManager{}
	sm.held = make(map[uint32]uint32)
	sm.toSell = make(map[uint32]uint32)
	for k, v := range initialStocks {
		sm.held[k] = v
	}
	return sm
}

func (sm *stockManager) cleanup(stockId uint32) {
	if sm.toSell[stockId] == 0 {
		delete(sm.toSell, stockId)
	}
	if sm.held[stockId] == 0 {
		delete(sm.held, stockId)
	}
}

func (sm *stockManager) getHeld(stockId uint32) uint32 {
	return sm.held[stockId]
}

func (sm *stockManager) addHeld(stockId uint32, amount uint32) {
	held := sm.held[stockId]
	sm.held[stockId] = held + amount
}

func (sm *stockManager) getToSell(stockId uint32) uint32 {
	return sm.toSell[stockId]
}

func (sm *stockManager) addToSell(stockId uint32, amount uint32) {
	toSell := sm.toSell[stockId]
	sm.toSell[stockId] = toSell + amount
}

func (sm *stockManager) canSell(stockId, amount uint32) bool {
	return sm.getHeld(stockId) >= amount
}

func (sm *stockManager) submitSell(stockId, amount uint32) {
	sm.addHeld(stockId, -amount)
	sm.addToSell(stockId, amount)
	// Don't clean up, we want the zeroed held stocks to remain
}

func (sm *stockManager) cancelSell(stockId, amount uint32) {
	sm.addHeld(stockId, amount)
	sm.addToSell(stockId, -amount)
	sm.cleanup(stockId)
}

func (sm *stockManager) completeSell(stockId, amount uint32) {
	sm.addToSell(stockId, -amount)
	sm.cleanup(stockId)
}

func (sm *stockManager) completeBuy(stockId, amount uint32) {
	sm.addHeld(stockId, amount)
}
