package client

import (
	"runtime"
	"testing"
)

func TestCanSell(t *testing.T) {
	stocks := newStockManager(map[uint64]uint64{1: 1, 2: 2, 3: 3})
	canSell(t, stocks, 1, 1)
	canSell(t, stocks, 2, 1)
	canSell(t, stocks, 2, 2)
	canSell(t, stocks, 3, 1)
	canSell(t, stocks, 3, 2)
	canSell(t, stocks, 3, 3)
}

func TestCannotSellWhatYouDoNotHave(t *testing.T) {
	stocks := newStockManager(map[uint64]uint64{1: 1, 2: 2, 3: 3})
	cannotSell(t, stocks, 4, 1)
	cannotSell(t, stocks, 5, 1)
	cannotSell(t, stocks, 6, 1)
	cannotSell(t, stocks, 7, 1)
	cannotSell(t, stocks, 8, 1)
	cannotSell(t, stocks, 9, 1)
}

func TestCannotSellMoreThanHeld(t *testing.T) {
	stocks := newStockManager(map[uint64]uint64{1: 1, 2: 2, 3: 3})
	cannotSell(t, stocks, 1, 2)
	cannotSell(t, stocks, 1, 3)
	cannotSell(t, stocks, 1, 4)
	cannotSell(t, stocks, 2, 3)
	cannotSell(t, stocks, 2, 4)
	cannotSell(t, stocks, 2, 5)
	cannotSell(t, stocks, 3, 4)
	cannotSell(t, stocks, 3, 5)
	cannotSell(t, stocks, 3, 6)
}

func TestSubmitSellCancel(t *testing.T) {
	init := map[uint64]uint64{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Initial state
	canSell(t, stocks, 1, 1)
	expectInMap(t, stocks.held, init)
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
	// Submitted sell
	stocks.submitSell(1, 1)
	expectInMap(t, stocks.held, map[uint64]uint64{1: 0, 2: 7})
	expectInMap(t, stocks.toSell, map[uint64]uint64{1: 1})
	// Completed sell
	stocks.cancelSell(1, 1)
	expectInMap(t, stocks.held, init)
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
}

func TestSubmitSellCancelPartialSell(t *testing.T) {
	init := map[uint64]uint64{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Initial state
	canSell(t, stocks, 2, 5)
	expectInMap(t, stocks.held, init)
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
	// Submitted sell
	stocks.submitSell(2, 5)
	expectInMap(t, stocks.held, map[uint64]uint64{1: 1, 2: 2})
	expectInMap(t, stocks.toSell, map[uint64]uint64{2: 5})
	// Completed sell
	stocks.cancelSell(2, 5)
	expectInMap(t, stocks.held, init)
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
}

func TestSubmitSellComplete(t *testing.T) {
	init := map[uint64]uint64{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Initial state
	canSell(t, stocks, 1, 1)
	expectInMap(t, stocks.held, init)
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
	// Submitted sell
	stocks.submitSell(1, 1)
	expectInMap(t, stocks.held, map[uint64]uint64{1: 0, 2: 7})
	expectInMap(t, stocks.toSell, map[uint64]uint64{1: 1})
	// Completed sell
	stocks.completeSell(1, 1)
	expectInMap(t, stocks.held, map[uint64]uint64{2: 7})
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
}

func TestSubmitSellCompletePartialSell(t *testing.T) {
	init := map[uint64]uint64{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Initial state
	canSell(t, stocks, 2, 5)
	expectInMap(t, stocks.held, init)
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
	// Submitted sell
	stocks.submitSell(2, 5)
	expectInMap(t, stocks.held, map[uint64]uint64{1: 1, 2: 2})
	expectInMap(t, stocks.toSell, map[uint64]uint64{2: 5})
	// Completed sell
	stocks.completeSell(2, 5)
	expectInMap(t, stocks.held, map[uint64]uint64{1: 1, 2: 2})
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
}

func TestCompleteBuyNewStock(t *testing.T) {
	init := map[uint64]uint64{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Complete buy
	stocks.completeBuy(3, 4)
	expectInMap(t, stocks.held, map[uint64]uint64{1: 1, 2: 7, 3: 4})
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
}

func TestCompleteBuyOwnedStock(t *testing.T) {
	init := map[uint64]uint64{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Complete buy
	stocks.completeBuy(1, 4)
	expectInMap(t, stocks.held, map[uint64]uint64{1: 5, 2: 7})
	expectInMap(t, stocks.toSell, map[uint64]uint64{})
}

func canSell(t *testing.T, stocks stockManager, stock, amount uint64) {
	expectCanSell(t, stocks, stock, amount, true)
}

func cannotSell(t *testing.T, stocks stockManager, stock, amount uint64) {
	expectCanSell(t, stocks, stock, amount, false)
}

func expectCanSell(t *testing.T, stocks stockManager, stock, amount uint64, canSell bool) {
	mod := ""
	if !canSell {
		mod = "not "
	}
	if canSell != stocks.canSell(stock, amount) {
		_, fname, lnum, _ := runtime.Caller(2)
		t.Errorf("Expected to %sbe able to sell %d of %d. Held %v, To Sell %v\n%s:%d", mod, amount, stock, stocks.held, stocks.toSell, fname, lnum)
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
