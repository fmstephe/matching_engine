package client

import (
	"runtime"
	"strconv"
	"testing"
)

func TestCanSell(t *testing.T) {
	stocks := newStockManager(map[uint32]uint32{1: 1, 2: 2, 3: 3})
	canSell(t, stocks, 1, 1)
	canSell(t, stocks, 2, 1)
	canSell(t, stocks, 2, 2)
	canSell(t, stocks, 3, 1)
	canSell(t, stocks, 3, 2)
	canSell(t, stocks, 3, 3)
}

func TestCannotSellWhatYouDoNotHave(t *testing.T) {
	stocks := newStockManager(map[uint32]uint32{1: 1, 2: 2, 3: 3})
	cannotSell(t, stocks, 4, 1)
	cannotSell(t, stocks, 5, 1)
	cannotSell(t, stocks, 6, 1)
	cannotSell(t, stocks, 7, 1)
	cannotSell(t, stocks, 8, 1)
	cannotSell(t, stocks, 9, 1)
}

func TestCannotSellMoreThanHeld(t *testing.T) {
	stocks := newStockManager(map[uint32]uint32{1: 1, 2: 2, 3: 3})
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
	init := map[uint32]uint32{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Initial state
	canSell(t, stocks, 1, 1)
	expectInHeld(t, stocks, init)
	expectInToSell(t, stocks, map[uint32]uint32{})
	// Submitted sell
	stocks.submitSell(uint32(1), uint32(1))
	expectInHeld(t, stocks, map[uint32]uint32{1: 0, 2: 7})
	expectInToSell(t, stocks, map[uint32]uint32{1: 1})
	// Completed sell
	stocks.cancelSell(uint32(1), uint32(1))
	expectInHeld(t, stocks, init)
	expectInToSell(t, stocks, map[uint32]uint32{})
}

func TestSubmitSellCancelPartialSell(t *testing.T) {
	init := map[uint32]uint32{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Initial state
	canSell(t, stocks, 2, 5)
	expectInHeld(t, stocks, init)
	expectInToSell(t, stocks, map[uint32]uint32{})
	// Submitted sell
	stocks.submitSell(uint32(2), uint32(5))
	expectInHeld(t, stocks, map[uint32]uint32{1: 1, 2: 2})
	expectInToSell(t, stocks, map[uint32]uint32{2: 5})
	// Completed sell
	stocks.cancelSell(uint32(2), uint32(5))
	expectInHeld(t, stocks, init)
	expectInToSell(t, stocks, map[uint32]uint32{})
}

func TestSubmitSellComplete(t *testing.T) {
	init := map[uint32]uint32{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Initial state
	canSell(t, stocks, 1, 1)
	expectInHeld(t, stocks, init)
	expectInToSell(t, stocks, map[uint32]uint32{})
	// Submitted sell
	stocks.submitSell(uint32(1), uint32(1))
	expectInHeld(t, stocks, map[uint32]uint32{1: 0, 2: 7})
	expectInToSell(t, stocks, map[uint32]uint32{1: 1})
	// Completed sell
	stocks.completeSell(uint32(1), uint32(1))
	expectInHeld(t, stocks, map[uint32]uint32{2: 7})
	expectInToSell(t, stocks, map[uint32]uint32{})
}

func TestSubmitSellCompletePartialSell(t *testing.T) {
	init := map[uint32]uint32{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Initial state
	canSell(t, stocks, 2, 5)
	expectInHeld(t, stocks, init)
	expectInToSell(t, stocks, map[uint32]uint32{})
	// Submitted sell
	stocks.submitSell(uint32(2), uint32(5))
	expectInHeld(t, stocks, map[uint32]uint32{1: 1, 2: 2})
	expectInToSell(t, stocks, map[uint32]uint32{2: 5})
	// Completed sell
	stocks.completeSell(uint32(2), uint32(5))
	expectInHeld(t, stocks, map[uint32]uint32{1: 1, 2: 2})
	expectInToSell(t, stocks, map[uint32]uint32{})
}

func TestCompleteBuyNewStock(t *testing.T) {
	init := map[uint32]uint32{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Complete buy
	stocks.completeBuy(3, 4)
	expectInHeld(t, stocks, map[uint32]uint32{1: 1, 2: 7, 3: 4})
	expectInToSell(t, stocks, map[uint32]uint32{})
}

func TestCompleteBuyOwnedStock(t *testing.T) {
	init := map[uint32]uint32{1: 1, 2: 7}
	stocks := newStockManager(init)
	// Complete buy
	stocks.completeBuy(1, 4)
	expectInHeld(t, stocks, map[uint32]uint32{1: 5, 2: 7})
	expectInToSell(t, stocks, map[uint32]uint32{})
}

func canSell(t *testing.T, stocks stockManager, stock, amount int) {
	expectCanSell(t, stocks, stock, amount, true)
}

func cannotSell(t *testing.T, stocks stockManager, stock, amount int) {
	expectCanSell(t, stocks, stock, amount, false)
}

func expectCanSell(t *testing.T, stocks stockManager, stock, amount int, canSell bool) {
	mod := ""
	if !canSell {
		mod = "not "
	}
	if canSell != stocks.canSell(uint32(stock), uint32(amount)) {
		_, fname, lnum, _ := runtime.Caller(2)
		t.Errorf("Expected to %sbe able to sell %d of %d. Held %v, To Sell %v\n%s:%d", mod, amount, stock, stocks.StocksHeld, stocks.StocksToSell, fname, lnum)
	}
}

func expectInHeld(t *testing.T, stocks stockManager, expected map[uint32]uint32) {
	expectInMap(t, expected, stocks.StocksHeld, "Held")
}

func expectInToSell(t *testing.T, stocks stockManager, expected map[uint32]uint32) {
	expectInMap(t, expected, stocks.StocksToSell, "To sell")
}

func expectInMap(t *testing.T, expected map[uint32]uint32, actual map[string]uint32, mod string) {
	for eStock, eAmount := range expected {
		aStock := strconv.Itoa(int(eStock))
		aAmount := actual[aStock]
		if aAmount != eAmount {
			_, fname, lnum, _ := runtime.Caller(2)
			t.Errorf("%s expected (stock %d: amount %d) but found (stock %s: amount %d)\n%s:%d", mod, eStock, eAmount, aStock, aAmount, fname, lnum)
		}
	}
	for aStock, aAmount := range actual {
		eStock, err := strconv.Atoi(aStock)
		if err != nil {
			_, fname, lnum, _ := runtime.Caller(2)
			t.Errorf("%s stock %s illegal\n%s:%d", mod, aStock, fname, lnum)
		}
		_, ok := expected[uint32(eStock)]
		if !ok {
			_, fname, lnum, _ := runtime.Caller(2)
			t.Errorf("%s (stock %d: amount %d) was not expected\n%s:%d", mod, eStock, aAmount, fname, lnum)
		}
	}
}
