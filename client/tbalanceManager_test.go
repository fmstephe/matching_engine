package client

import (
	"math/rand"
	"runtime"
	"testing"
)

const randRuns = 20

func TestNewBalanceManager(t *testing.T) {
	balance := uint64(100)
	bm := newBalanceManager(balance)
	if bm.Current != balance {
		t.Errorf("Expecting %d current balance, found %d", balance, bm.Current)
	}
	if bm.Available != balance {
		t.Errorf("Expecting %d current balance, found %d", balance, bm.Available)
	}
}

func TestCanBuySimple(t *testing.T) {
	balance := 100
	bm := newBalanceManager(uint64(balance))
	// Can Buy
	canBuy(t, bm, 1, 1)
	canBuy(t, bm, 1, balance)
	canBuy(t, bm, balance, 1)
	canBuy(t, bm, balance/2, 2)
	canBuy(t, bm, 2, balance/2)
	canBuy(t, bm, balance/3, 3)
	canBuy(t, bm, 3, balance/3)
	canBuy(t, bm, balance/4, 4)
	canBuy(t, bm, 4, balance/4)
	canBuy(t, bm, balance/5, 5)
	canBuy(t, bm, 5, balance/5)
	// Cannot Buy
	cannotBuy(t, bm, balance+1, 1)
	cannotBuy(t, bm, balance, 2)
	cannotBuy(t, bm, 2+1, balance/2)
	cannotBuy(t, bm, 2, (balance/2)+1)
	cannotBuy(t, bm, balance/2, 2+1)
	cannotBuy(t, bm, (balance/2)+1, 2)
	cannotBuy(t, bm, 3+1, balance/3)
	cannotBuy(t, bm, 3, (balance/3)+1)
	cannotBuy(t, bm, balance/3, 3+1)
	cannotBuy(t, bm, (balance/3)+1, 3)
	cannotBuy(t, bm, 4+1, balance/4)
	cannotBuy(t, bm, 4, (balance/4)+1)
	cannotBuy(t, bm, balance/4, 4+1)
	cannotBuy(t, bm, (balance/4)+1, 4)
	cannotBuy(t, bm, 5+1, balance/5)
	cannotBuy(t, bm, 5, (balance/5)+1)
	cannotBuy(t, bm, balance/5, 5+1)
	cannotBuy(t, bm, (balance/5)+1, 5)
}

func TestCanBuy(t *testing.T) {
	balance := 100
	bm := newBalanceManager(uint64(balance))
	// If the (stocks * price) <= bm.Balance then we can buy
	for stocks := 1; stocks <= balance; stocks++ {
		for price := 1; price <= balance/stocks; price++ {
			canBuy(t, bm, price, stocks)
		}
	}
	// If the (stocks * price) > bm.Balance then we can't buy
	for stocks := 1; stocks <= balance; stocks++ {
		initPrice := (balance / stocks) + 1
		for price := initPrice; price <= initPrice*5; price++ {
			cannotBuy(t, bm, price, stocks)
		}
	}
}

func TestCanBuyAfterSubmitBuySimple(t *testing.T) {
	balance := 100
	price := uint64(5)
	amount := uint32(5)
	bm := newBalanceManager(uint64(balance))
	cannotBuy(t, bm, 101, 1)
	canBuy(t, bm, 100, 1)
	canBuy(t, bm, 5, 5)
	bm.submitBuy(price, amount)
	// 75 available
	expectBalance(t, bm, 75, balance)
	cannotBuy(t, bm, 76, 1)
	canBuy(t, bm, 75, 1)
	canBuy(t, bm, 5, 5)
	bm.submitBuy(price, amount)
	// 50 available
	expectBalance(t, bm, 50, balance)
	cannotBuy(t, bm, 51, 1)
	canBuy(t, bm, 50, 1)
	canBuy(t, bm, 5, 5)
	bm.submitBuy(price, amount)
	// 25 available
	expectBalance(t, bm, 25, balance)
	cannotBuy(t, bm, 26, 1)
	canBuy(t, bm, 25, 1)
	canBuy(t, bm, 5, 5)
	bm.submitBuy(price, amount)
	// 0 available
	expectBalance(t, bm, 0, balance)
	cannotBuy(t, bm, 1, 1)
	bm.cancelBuy(price, amount)
	// 25 available
	expectBalance(t, bm, 25, balance)
	cannotBuy(t, bm, 26, 1)
	canBuy(t, bm, 25, 1)
	canBuy(t, bm, 5, 5)
	bm.cancelBuy(price, amount)
	// 50 available
	expectBalance(t, bm, 50, balance)
	cannotBuy(t, bm, 51, 1)
	canBuy(t, bm, 50, 1)
	canBuy(t, bm, 5, 5)
	bm.cancelBuy(price, amount)
	// 75 available
	expectBalance(t, bm, 75, balance)
	cannotBuy(t, bm, 76, 1)
	canBuy(t, bm, 75, 1)
	canBuy(t, bm, 5, 5)
	bm.cancelBuy(price, amount)
	// 100 available
	expectBalance(t, bm, 100, balance)
	cannotBuy(t, bm, 101, 1)
	canBuy(t, bm, 100, 1)
	canBuy(t, bm, 5, 5)
}

func TestCanBuyAfterSubmitBuyRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < randRuns; i++ {
		testCanBuyAfterSubmitBuyRandom(t, r)
	}
}

func testCanBuyAfterSubmitBuyRandom(t *testing.T, r *rand.Rand) {
	balance := 100
	bm := newBalanceManager(uint64(balance))
	total := 0
	for i := 0; i < 20; i++ {
		price, amount := makePriceAmount(balance/20, 5, r)
		if (total + price*amount) <= balance {
			total += price * amount
			canBuy(t, bm, price, amount)
			bm.submitBuy(uint64(price), uint32(amount))
			expectBalance(t, bm, balance-total, balance)
		} else {
			cannotBuy(t, bm, price, amount)
			expectBalance(t, bm, balance-total, balance)
		}
	}
}

func TestBuyAndCancelRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < randRuns; i++ {
		testBuyAndCancelRandom(t, r)
	}
}

func testBuyAndCancelRandom(t *testing.T, r *rand.Rand) {
	balance := 100
	bm := newBalanceManager(uint64(balance))
	for j := 0; j < 20; j++ {
		prices := make([]int, 0)
		amounts := make([]int, 0)
		total := 0
		for i := 0; i < 20; i++ {
			price, amount := makePriceAmount(balance/20, 5, r)
			if (total + price*amount) <= balance {
				prices = append(prices, price)
				amounts = append(amounts, amount)
				total += price * amount
				canBuy(t, bm, price, amount)
				bm.submitBuy(uint64(price), uint32(amount))
				expectBalance(t, bm, balance-total, balance)
			} else {
				cannotBuy(t, bm, price, amount)
				expectBalance(t, bm, balance-total, balance)
			}
		}
		for i := range prices {
			price := prices[i]
			amount := amounts[i]
			total -= price * amount
			bm.cancelBuy(uint64(price), uint32(amount))
			expectBalance(t, bm, balance-total, balance)
		}
	}
}

func TestBuyAndCompleteBuySimple(t *testing.T) {
	balance := 100
	price := uint64(5)
	amount := uint32(5)
	bm := newBalanceManager(uint64(balance))
	bm.submitBuy(price, amount)
	expectBalance(t, bm, 75, 100)
	bm.completeBuy(price, price, amount)
	expectBalance(t, bm, 75, 75)
	// 75
	bm.submitBuy(price, amount)
	expectBalance(t, bm, 50, 75)
	bm.completeBuy(price, price, amount)
	expectBalance(t, bm, 50, 50)
	// 50
	bm.submitBuy(price, amount)
	expectBalance(t, bm, 25, 50)
	bm.completeBuy(price, price, amount)
	expectBalance(t, bm, 25, 25)
	// 25
	bm.submitBuy(price, amount)
	expectBalance(t, bm, 0, 25)
	bm.completeBuy(price, price, amount)
	expectBalance(t, bm, 0, 0)
	// 0
}

func TestBuyAndCompleteBuyDiff(t *testing.T) {
	balance := 100
	bm := newBalanceManager(uint64(balance))
	// Buy 10 * 5
	bm.submitBuy(uint64(10), uint32(5))
	expectBalance(t, bm, 50, 100)
	// Actual 9 * 5
	bm.completeBuy(uint64(10), uint64(9), uint32(5))
	expectBalance(t, bm, 55, 55)
	// Buy 5 * 1
	bm.submitBuy(uint64(5), uint32(1))
	expectBalance(t, bm, 50, 55)
	// Actual 4 * 1
	bm.completeBuy(uint64(5), uint64(4), uint32(1))
	expectBalance(t, bm, 51, 51)
	// Buy 5 * 10
	bm.submitBuy(uint64(5), uint32(10))
	expectBalance(t, bm, 1, 51)
	// Actual 3 * 10
	bm.completeBuy(uint64(5), uint64(3), uint32(10))
	expectBalance(t, bm, 21, 21)
}

func TestCompleteSellSimple(t *testing.T) {
	balance := 100
	price := uint64(5)
	amount := uint32(5)
	bm := newBalanceManager(uint64(balance))
	bm.completeSell(price, amount)
	expectBalance(t, bm, 125, 125)
	// 125
	bm.completeSell(price, amount)
	expectBalance(t, bm, 150, 150)
	// 150
	bm.completeSell(price, amount)
	expectBalance(t, bm, 175, 175)
	// 175
	bm.completeSell(price, amount)
	expectBalance(t, bm, 200, 200)
	// 200
}

func testCompleteSellRandom(t *testing.T, r *rand.Rand) {
	balance := 100
	bm := newBalanceManager(uint64(balance))
	total := 0
	for i := 0; i < 20; i++ {
		price, amount := makePriceAmount(balance/20, 5, r)
		total += price * amount
		bm.completeSell(uint64(price), uint32(amount))
		expectBalance(t, bm, balance+total, balance+total)
	}
}

func makePriceAmount(priceCap, amountCap int, r *rand.Rand) (price, amount int) {
	return int(r.Int63n(int64(priceCap-1))) + 1, int(r.Int63n(int64(amountCap-1)) + 1)
}

func canBuy(t *testing.T, bm balanceManager, price, amount int) {
	expectCanBuy(t, bm, price, amount, true)
}

func cannotBuy(t *testing.T, bm balanceManager, price, amount int) {
	expectCanBuy(t, bm, price, amount, false)
}

func expectCanBuy(t *testing.T, bm balanceManager, price, amount int, canBuy bool) {
	mod := ""
	if !canBuy {
		mod = "not "
	}
	if canBuy != bm.canBuy(uint64(price), uint32(amount)) {
		_, fname, lnum, _ := runtime.Caller(2)
		t.Errorf("Expected to %sbe able to buy %d stock(s) at %d. Current: %d, Available: %d\n%s:%d", mod, amount, price, bm.Current, bm.Available, fname, lnum)
	}
}

func expectBalance(t *testing.T, bm balanceManager, available, current int) {
	if uint64(available) != bm.Available {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("Expected available %d, found %d\n%s:%d", available, bm.Available, fname, lnum)
	}
	if uint64(current) != bm.Current {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("Expected current %d, found %d\n%s:%d", current, bm.Current, fname, lnum)
	}
}
