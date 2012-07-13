package matcher

import (
	"testing"
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
)

const (
	stockId = "stock1"
	trader1 = "trader1"
	trader2 = "trader2"
	trader3 = "trader3"
)

func verify(t *testing.T, r *trade.Response, tradeId, amount, price int64, counterParty string) {
	if (r.TradeId != tradeId) {
		t.Errorf("Expecting %d trade-id, got %d instead", tradeId, r.TradeId)
	}
	if (r.Amount != amount) {
		t.Errorf("Expecting %d amount, got %d instead", amount, r.Amount)
	}
	if (r.Price != price) {
		t.Errorf("Expecting %d price, got %d instead", price, r.Price)
	}
	if (r.CounterParty != counterParty) {
		t.Errorf("Expecting %d counter party, got %d instead", counterParty, r.CounterParty)
	}
}

func responseChan() chan *trade.Response {
	return make(chan *trade.Response, 256)
}

func TestMidPoint(t *testing.T) {
	midpoint(t, 1, 1, 1)
	midpoint(t, 2, 1, 1)
	midpoint(t, 3, 1, 2)
	midpoint(t, 4, 1, 2)
	midpoint(t, 5, 1, 3)
	midpoint(t, 6, 1, 3)
	midpoint(t, 20, 10, 15)
	midpoint(t, 21, 10, 15)
	midpoint(t, 22, 10, 16)
	midpoint(t, 23, 10, 16)
	midpoint(t, 24, 10, 17)
	midpoint(t, 25, 10, 17)
	midpoint(t, 26, 10, 18)
	midpoint(t, 27, 10, 18)
	midpoint(t, 28, 10, 19)
	midpoint(t, 29, 10, 19)
	midpoint(t, 30, 10, 20)
}

func midpoint(t *testing.T, bPrice, sPrice, expected int64) {
	result := price(bPrice, sPrice)
	if result != expected {
		t.Errorf("price(%d,%d) does not equal %d, got %d instead.", bPrice, sPrice, expected, result)
	}
}

// Basic test matches lonely buy/sell trade pair which match exactly
func TestSimpleMatch(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 1)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	b := trade.NewBuy(1, 1, 2, stockId, trader1, trader1Chan)
	s := trade.NewSell(2, 1, 2, stockId, trader2, trader2Chan)
	m.AddBuy(b)
	m.AddSell(s)
	verify(t, <-trader1Chan, 1, 1, -2, trader2)
	verify(t, <-trader2Chan, 2, 1, 2, trader1)
}

// Test matches one buy order to two separate sells
func TestDoubleSellMatch(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 1)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	trader3Chan := responseChan()
	m.AddBuy(trade.NewBuy(1, 2, 2, stockId, trader1, trader1Chan))
	m.AddSell(trade.NewSell(2, 1, 2, stockId, trader2, trader2Chan))
	verify(t, <-trader1Chan, 1, 1, -2, trader2)
	verify(t, <-trader2Chan, 2, 1, 2, trader1)
	m.AddSell(trade.NewSell(3, 1, 2, stockId, trader3, trader3Chan))
	verify(t, <-trader1Chan, 1, 1, -2, trader3)
	verify(t, <-trader3Chan, 3, 1, 2, trader1)
}

// Test matches two buy orders to one sell
func TestDoubleBuyMatch(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 1)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	trader3Chan := responseChan()
	m.AddSell(trade.NewSell(1, 2, 2, stockId, trader1, trader1Chan))
	m.AddBuy(trade.NewBuy(2, 1, 2, stockId, trader2, trader2Chan))
	verify(t, <-trader1Chan, 1, 1, 2, trader2)
	verify(t, <-trader2Chan, 2, 1, -2, trader1)
	m.AddBuy(trade.NewBuy(3, 1, 2, stockId, trader3, trader3Chan))
	verify(t, <-trader1Chan, 1, 1, 2, trader3)
	verify(t, <-trader3Chan, 3, 1, -2, trader1)
}

// Test matches lonely buy/sell pair, with same quantity, uses the mid-price point for trade price
func TestMidPrice(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 1)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	m.AddBuy(trade.NewBuy(1, 1, 6, stockId, trader1, trader1Chan))
	m.AddSell(trade.NewSell(1, 1, 2, stockId, trader2, trader2Chan))
	verify(t, <-trader1Chan, 1, 1, -4, trader2)
	verify(t, <-trader2Chan, 1, 1, 4, trader1)
}

// Test matches lonely buy/sell pair, sell > quantity, and uses the mid-price point for trade price
func TestMidPriceBigSell(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 1)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	m.AddBuy(trade.NewBuy(1, 1, 6, stockId, trader1, trader1Chan))
	m.AddSell(trade.NewSell(1, 10, 2, stockId, trader2, trader2Chan))
	verify(t, <-trader1Chan, 1, 1, -4, trader2)
	verify(t, <-trader2Chan, 1, 1, 4, trader1)
}

// Test matches lonely buy/sell pair, buy > quantity, and uses the mid-price point for trade price
func TestMidPriceBigBuy(t *testing.T) {
	m := New(stockId)
	addLowBuys(m, 1)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	m.AddBuy(trade.NewBuy(1, 10, 6, stockId, trader1, trader1Chan))
	m.AddSell(trade.NewSell(1, 1, 2, stockId, trader2, trader2Chan))
	verify(t, <-trader1Chan, 1, 1, -4, trader2)
	verify(t, <-trader2Chan, 1, 1, 4, trader1)
}

func addLowBuys(m *M, highestPrice int64) {
	price := highestPrice
	for i := 0; i < 10; i++ {
		rc := responseChan()
		m.AddBuy(trade.NewBuy(1, 1, price, stockId, fmt.Sprintf("spamBuyTrader%d",i), rc))
		go func() {
			println(fmt.Sprintf("%v", <-rc))
			println("I should not have sucessfully made a buy")
		}()
		price--
		if price == 0 {
			price = highestPrice
		}
	}
}

func addHighSells(m *M, lowestPrice int64) {
	price := lowestPrice
	for i := 0; i < 10; i++ {
		rc := responseChan()
		m.AddSell(trade.NewSell(1, 1, price, stockId, fmt.Sprintf("spamSellTrader%d",i), rc))
		go func() {
			println(fmt.Sprintf("%v", <-rc))
			println("I should not have successfully make a sell")
		}()
		price++
		if price == lowestPrice*2 {
			price = lowestPrice
		}
	}
}
