package client

import (
	"github.com/fmstephe/matching_engine/msg"
	"runtime"
	"testing"
)

func setupTrader(traderId uint32, t *testing.T) *trader {
	intoSvr := make(chan *msg.Message)
	outOfSvr := make(chan *msg.Message)
	tdr, _ := newTrader(traderId, intoSvr, outOfSvr)
	validate(t, tdr)
	expectBalance(t, tdr.balance, initialBalance, initialBalance)
	return tdr
}

func TestMessageProcessBuyCancelCancelled(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit buys
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 75, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
	// Cancel
	tdr.process(&msg.Message{Kind: msg.CANCEL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, tdr.balance, 75, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
	// Confirm CANCELLED
	tdr.process(&msg.Message{Kind: msg.CANCELLED, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, tdr.balance, 100, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
}

func TestMessageProcessBuyFullSimple(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit buy
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 75, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
	// Full Match
	tdr.process(&msg.Message{Kind: msg.FULL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, tdr.balance, 75, 75)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 15, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
}

func TestMessageProcessBuyFullDiffSimple(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit buy
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 75, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
	// Full Match - at lower than bid price
	tdr.process(&msg.Message{Kind: msg.FULL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 4, Amount: 5})
	expectBalance(t, tdr.balance, 80, 80)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 15, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
}

func TestMessageProcessBuyPartialSimple(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit some buys
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 75, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
	// Partial Match
	tdr.process(&msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 2})
	expectBalance(t, tdr.balance, 75, 90)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 12, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
	tdr.process(&msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 3})
	expectBalance(t, tdr.balance, 75, 75)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 15, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
}

func TestMessageProcessBuyPartialDiffSimple(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit some buys
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 75, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
	// Partial Matches at lower than bid price
	tdr.process(&msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 4, Amount: 2})
	expectBalance(t, tdr.balance, 77, 92)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 12, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
	tdr.process(&msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 3, Amount: 3})
	expectBalance(t, tdr.balance, 83, 83)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 15, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
}

func TestMessageProcessSellCancelCancelled(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 100, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{1: 5})
	validate(t, tdr)
	// Cancel sell
	tdr.process(&msg.Message{Kind: msg.CANCEL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, tdr.balance, 100, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{1: 5})
	validate(t, tdr)
	// Confirm CANCELLED
	tdr.process(&msg.Message{Kind: msg.CANCELLED, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, tdr.balance, 100, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
}

func TestMessageProcessSellFullSimple(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 100, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{1: 5})
	validate(t, tdr)
	// Match sell FULL
	tdr.process(&msg.Message{Kind: msg.FULL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, tdr.balance, 125, 125)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
}

func TestMessageProcessSellFullDiff(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 100, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{1: 5})
	validate(t, tdr)
	// Match sell FUll
	tdr.process(&msg.Message{Kind: msg.FULL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 7, Amount: 5})
	expectBalance(t, tdr.balance, 135, 135)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	validate(t, tdr)
}

func TestMessageProcessSellPartialSimple(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 100, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{1: 5})
	validate(t, tdr)
	// Match sell PARTIAL
	tdr.process(&msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 3})
	expectBalance(t, tdr.balance, 115, 115)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{1: 2})
	validate(t, tdr)
}

func TestMessageProcessSellPartialDiff(t *testing.T) {
	traderId := uint32(1)
	tdr := setupTrader(traderId, t)
	expectBalance(t, tdr.balance, 100, 100) // This test expects that initialBalance is 100
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	tdr.process(m1)
	expectBalance(t, tdr.balance, 100, 100)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{1: 5})
	validate(t, tdr)
	// Match sell PARTIAL
	tdr.process(&msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 7, Amount: 3})
	expectBalance(t, tdr.balance, 121, 121)
	expectInMap(t, tdr.stocks.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, tdr.stocks.toSell, map[uint64]uint64{1: 2})
	validate(t, tdr)
}

func validate(t *testing.T, tdr *trader) {
	// 1: current - available = sum(outstanding buys)
	totalBuys := 0
	for _, m := range tdr.outstanding {
		if m.Kind == msg.BUY {
			totalBuys += int(m.Price) * int(m.Amount)
		}
	}
	diff := (tdr.balance.current - tdr.balance.available)
	if totalBuys != int(diff) {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("Total buys outstanding: %d, current - available: %d\n%s:%d", totalBuys, diff, fname, lnum)
	}
	// 2: stocks to sell = oustanding sells
	for stockId, amount := range tdr.stocks.toSell {
		totalSells := 0
		for _, m := range tdr.outstanding {
			if m.Kind == msg.SELL {
				totalSells += int(m.Amount)
			}
		}
		if totalSells != int(amount) {
			_, fname, lnum, _ := runtime.Caller(1)
			t.Errorf("%d to sell: %d, outstanding sells for %d: %d\n%s:%d", stockId, amount, stockId, totalSells, fname, lnum)
		}
	}
}
