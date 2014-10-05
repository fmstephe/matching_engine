package client

import (
	"github.com/fmstephe/matching_engine/msg"
	"runtime"
	"testing"
)

func TestMessageProcessBuyCancelCancelled(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit buys
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 75, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
	// Cancel
	canProcess(t, bm, &msg.Message{Kind: msg.CANCEL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, bm, 75, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
	// Confirm CANCELLED
	canProcess(t, bm, &msg.Message{Kind: msg.CANCELLED, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, bm, 100, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
}

func TestMessageProcessBuyFullSimple(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit buy
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 75, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
	// Full Match
	canProcess(t, bm, &msg.Message{Kind: msg.FULL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, bm, 75, 75)
	expectInMap(t, bm.held, map[uint64]uint64{1: 15, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
}

func TestMessageProcessBuyFullDiffSimple(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit buy
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 75, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
	// Full Match - at lower than bid price
	canProcess(t, bm, &msg.Message{Kind: msg.FULL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 4, Amount: 5})
	expectBalance(t, bm, 80, 80)
	expectInMap(t, bm.held, map[uint64]uint64{1: 15, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
}

func TestMessageProcessBuyPartialSimple(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit some buys
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 75, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
	// Partial Match
	canProcess(t, bm, &msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 2})
	expectBalance(t, bm, 75, 90)
	expectInMap(t, bm.held, map[uint64]uint64{1: 12, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
	canProcess(t, bm, &msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 3})
	expectBalance(t, bm, 75, 75)
	expectInMap(t, bm.held, map[uint64]uint64{1: 15, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
}

func TestMessageProcessBuyPartialDiffSimple(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit some buys
	m1 := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 75, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
	// Partial Matches at lower than bid price
	canProcess(t, bm, &msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 4, Amount: 2})
	expectBalance(t, bm, 77, 92)
	expectInMap(t, bm.held, map[uint64]uint64{1: 12, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
	canProcess(t, bm, &msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 3, Amount: 3})
	expectBalance(t, bm, 83, 83)
	expectInMap(t, bm.held, map[uint64]uint64{1: 15, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
}

func TestMessageProcessSellCancelCancelled(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 100, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{1: 5})
	validate(t, bm)
	// Cancel sell
	canProcess(t, bm, &msg.Message{Kind: msg.CANCEL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, bm, 100, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{1: 5})
	validate(t, bm)
	// Confirm CANCELLED
	canProcess(t, bm, &msg.Message{Kind: msg.CANCELLED, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, bm, 100, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
}

func TestMessageProcessSellFullSimple(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 100, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{1: 5})
	validate(t, bm)
	// Match sell FULL
	canProcess(t, bm, &msg.Message{Kind: msg.FULL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 5})
	expectBalance(t, bm, 125, 125)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
}

func TestMessageProcessSellFullDiff(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 100, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{1: 5})
	validate(t, bm)
	// Match sell FUll
	canProcess(t, bm, &msg.Message{Kind: msg.FULL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 7, Amount: 5})
	expectBalance(t, bm, 135, 135)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	validate(t, bm)
}

func TestMessageProcessSellPartialSimple(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 100, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{1: 5})
	validate(t, bm)
	// Match sell PARTIAL
	canProcess(t, bm, &msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 5, Amount: 3})
	expectBalance(t, bm, 115, 115)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{1: 2})
	validate(t, bm)
}

func TestMessageProcessSellPartialDiff(t *testing.T) {
	traderId := uint32(1)
	bm := newBalanceManager(traderId, 100, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectBalance(t, bm, 100, 100) // This test expects that 100 is 100
	expectInMap(t, bm.held, map[uint64]uint64{1: 10, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{})
	// Submit sell
	m1 := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 5, Amount: 5}
	canProcess(t, bm, m1)
	expectBalance(t, bm, 100, 100)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{1: 5})
	validate(t, bm)
	// Match sell PARTIAL
	canProcess(t, bm, &msg.Message{Kind: msg.PARTIAL, TraderId: traderId, TradeId: m1.TradeId, StockId: 1, Price: 7, Amount: 3})
	expectBalance(t, bm, 121, 121)
	expectInMap(t, bm.held, map[uint64]uint64{1: 5, 2: 10, 3: 10})
	expectInMap(t, bm.toSell, map[uint64]uint64{1: 2})
	validate(t, bm)
}

func TestCanBuyCannotBuy(t *testing.T) {
	traderId := uint32(1)
	balance := uint64(100)
	bm := newBalanceManager(traderId, balance, map[uint64]uint64{})
	// If the (stocks * price) <= bm.Balance then we can buy
	for amount := uint64(1); amount <= balance; amount++ {
		for price := uint64(1); price <= balance/amount; price++ {
			m := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: price, Amount: amount}
			canProcess(t, bm, m)
			c := &msg.Message{Kind: msg.CANCEL, TraderId: traderId, TradeId: m.TradeId, StockId: 1, Price: price, Amount: amount}
			canProcess(t, bm, c)
			cd := &msg.Message{Kind: msg.CANCELLED, TraderId: traderId, TradeId: m.TradeId, StockId: 1, Price: price, Amount: amount}
			canProcess(t, bm, cd)
		}
	}
	// If the (stocks * price) > bm.Balance then we can't buy
	for amount := uint64(1); amount <= balance; amount++ {
		initPrice := (balance / amount) + 1
		for price := initPrice; price <= initPrice*5; price++ {
			m := &msg.Message{Kind: msg.BUY, TraderId: traderId, TradeId: 0, StockId: 1, Price: price, Amount: amount}
			cannotProcess(t, bm, m)
		}
	}
}

func TestCanSellCannotSellAmount(t *testing.T) {
	traderId := uint32(1)
	balance := uint64(100)
	amount := uint64(100)
	bm := newBalanceManager(traderId, balance, map[uint64]uint64{1: amount})
	// If the stock amount <= stock held then we can sell
	for i := uint64(1); i < amount; i++ {
		m := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 1, Amount: i}
		canProcess(t, bm, m)
		c := &msg.Message{Kind: msg.CANCEL, TraderId: traderId, TradeId: m.TradeId, StockId: 1, Price: 1, Amount: i}
		canProcess(t, bm, c)
		cd := &msg.Message{Kind: msg.CANCELLED, TraderId: traderId, TradeId: m.TradeId, StockId: 1, Price: 1, Amount: i}
		canProcess(t, bm, cd)
	}
	// If the stock amount >= stock held then we can't sell
	for i := amount + 1; i < amount*3; i++ {
		m := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: 1, Price: 1, Amount: i}
		cannotProcess(t, bm, m)
	}
}

func TestCanSellCannotSellStockId(t *testing.T) {
	traderId := uint32(1)
	balance := uint64(100)
	for stockId := uint64(2); stockId < 100; stockId++ {
		bm := newBalanceManager(traderId, balance, map[uint64]uint64{stockId: 1})
		mTooLow := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: stockId - 1, Price: 1, Amount: 1}
		cannotProcess(t, bm, mTooLow)
		mTooHigh := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: stockId + 1, Price: 1, Amount: 1}
		cannotProcess(t, bm, mTooHigh)
		mJustRight := &msg.Message{Kind: msg.SELL, TraderId: traderId, TradeId: 0, StockId: stockId, Price: 1, Amount: 1}
		canProcess(t, bm, mJustRight)
	}
}

func validate(t *testing.T, bm *balanceManager) {
	// 1: current - available = sum(outstanding buys)
	totalBuys := 0
	for _, m := range bm.outstanding {
		if m.Kind == msg.BUY {
			totalBuys += int(m.Price) * int(m.Amount)
		}
	}
	diff := (bm.current - bm.available)
	if totalBuys != int(diff) {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("Total buys outstanding: %d, current - available: %d\n%s:%d", totalBuys, diff, fname, lnum)
	}
	// 2: balance to sell = oustanding sells
	for stockId, amount := range bm.toSell {
		totalSells := 0
		for _, m := range bm.outstanding {
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

func canProcess(t *testing.T, bm *balanceManager, m *msg.Message) {
	expectCanProcess(t, bm, m, true)
}

func cannotProcess(t *testing.T, bm *balanceManager, m *msg.Message) {
	expectCanProcess(t, bm, m, false)
}

func expectCanProcess(t *testing.T, bm *balanceManager, m *msg.Message, can bool) {
	mod := ""
	if !can {
		mod = "not "
	}
	if can != bm.process(m) {
		_, fname, lnum, _ := runtime.Caller(2)
		t.Errorf("Expected to %sbe able to process %v - available: %d, current %d.\n%s:%d", mod, m, bm.available, bm.current, fname, lnum)
	}
}

func expectBalance(t *testing.T, bm *balanceManager, available, current uint64) {
	if available != bm.available {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("Expected available %d, found %d\n%s:%d", available, bm.available, fname, lnum)
	}
	if current != bm.current {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("Expected current %d, found %d\n%s:%d", current, bm.current, fname, lnum)
	}
}

func expectInMap(t *testing.T, expected, actual map[uint64]uint64) {
	for stock, eAmount := range expected {
		aAmount := actual[stock]
		if aAmount != eAmount {
			_, fname, lnum, _ := runtime.Caller(1)
			t.Errorf("Expected (stock %d: amount %d) but found (stock %d: amount %d)\n%s:%d", stock, eAmount, stock, aAmount, fname, lnum)
		}
	}
	for stock, aAmount := range actual {
		_, ok := expected[stock]
		if !ok {
			_, fname, lnum, _ := runtime.Caller(1)
			t.Errorf("(stock %d: amount %d) was not expected\n%s:%d", stock, aAmount, fname, lnum)
		}
	}
}
