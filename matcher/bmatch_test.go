package matcher

import (
	"math/rand"
	"testing"
)

const (
	orderNum = 1000 * 1000
)

var benchRand = rand.New(rand.NewSource(1))
var buys []*Order
var sells []*Order

func prepare(b *testing.B) {
	b.StopTimer()
	if buys == nil {
		buys = mkBuys(orderNum, 1000, 1500)
	}
	if sells == nil {
		sells = mkSells(orderNum, 1000, 1500)
	}
	b.StartTimer()
}

func valRange(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = rand.Int63n(high-low) + low
	}
	return vals
}

func mkBuys(n int, low, high int64) []*Order {
	return mkOrders(n, low, high, BUY)
}

func mkSells(n int, low, high int64) []*Order {
	return mkOrders(n, low, high, SELL)
}

func mkOrders(n int, low, high int64, buySell TradeType) []*Order {
	prices := valRange(n, low, high)
	orders := make([]*Order, n)
	for i, price := range prices {
		responseFunc := func(response *Response) {
			// Do Nothing
		}
		costData := CostData{Price: price, Amount: 1}
		tradeData := TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = NewOrder(costData, tradeData, responseFunc, buySell)
	}
	return orders
}

func BenchmarkAddBuy(b *testing.B) {
	prepare(b)
	m := NewMatcher(stockId)
	for _, buy := range buys {
		m.AddBuy(buy)
	}
}

func BenchmarkAddSell(b *testing.B) {
	prepare(b)
	m := NewMatcher(stockId)
	for _, buy := range buys {
		m.AddBuy(buy)
	}
}

func BenchmarkMatch(b *testing.B) {
	prepare(b)
	m := NewMatcher(stockId)
	for i := 0; i < orderNum; i++ {
		m.AddBuy(buys[i])
		m.AddSell(sells[i])
	}
}
