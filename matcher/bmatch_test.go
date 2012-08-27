package matcher

import (
	"math/rand"
	"testing"
)

const (
	orderNum = 500 * 1000
)

var (
	benchRand = rand.New(rand.NewSource(1))
	buysWide []*Order
	buysMedium []*Order
	buysNarrow []*Order
	sellsWide []*Order
	sellsMedium []*Order
	sellsNarrow []*Order
	output *ResponseBuffer
)

func prepare(b *testing.B) {
	b.StopTimer()
	if buysWide == nil {
		buysWide = mkBuys(orderNum, 1000, 100*1000)
	}
	if buysMedium == nil {
		buysMedium = mkBuys(orderNum, 1000, 5000)
	}
	if buysNarrow == nil {
		buysNarrow = mkBuys(orderNum, 1000, 1500)
	}
	if sellsWide == nil {
		sellsWide = mkSells(orderNum, 1000, 100*1000)
	}
	if sellsMedium == nil {
		sellsMedium = mkSells(orderNum, 1000, 5000)
	}
	if sellsNarrow == nil {
		sellsNarrow = mkSells(orderNum, 1000, 1500)
	}
	if output == nil {
		output = NewResponseBuffer(orderNum*32)
	} else {
		output.clear()
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

func BenchmarkAddBuyWide(b *testing.B) {
	prepare(b)
	benchmarkAddBuy(b, buysWide)
}

func BenchmarkAddBuyMedium(b *testing.B) {
	prepare(b)
	benchmarkAddBuy(b, buysMedium)
}

func BenchmarkAddBuyNarrow(b *testing.B) {
	prepare(b)
	benchmarkAddBuy(b, buysNarrow)
}

func benchmarkAddBuy(b *testing.B, buys []*Order) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := NewMatcher(stockId, output)
		b.StartTimer()
		for _, buy := range buys {
			m.AddBuy(buy)
		}
	}
}

func BenchmarkAddSellWide(b *testing.B) {
	prepare(b)
	benchmarkAddSell(b, sellsWide)
}

func BenchmarkAddSellMedium(b *testing.B) {
	prepare(b)
	benchmarkAddSell(b, sellsMedium)
}

func BenchmarkAddSellNarrow(b *testing.B) {
	prepare(b)
	benchmarkAddSell(b, sellsNarrow)
}

func benchmarkAddSell(b *testing.B, sells []*Order) {
	for i := 0; i < b.N; i++ {
		prepare(b)
		b.StopTimer()
		m := NewMatcher(stockId, output)
		b.StartTimer()
		for _, sell := range sells {
			m.AddSell(sell)
		}
	}
}

func BenchmarkMatchWide(b *testing.B) {
	prepare(b)
	benchmarkMatch(b, buysWide, sellsWide)
}

func BenchmarkMatchMedium(b *testing.B) {
	prepare(b)
	benchmarkMatch(b, buysMedium, sellsMedium)
}

func BenchmarkMatchNarrow(b *testing.B) {
	prepare(b)
	benchmarkMatch(b, buysNarrow, sellsNarrow)
}

func benchmarkMatch(b *testing.B, buys, sells []*Order) {
	for i := 0; i < b.N; i++ {
		prepare(b)
		m := NewMatcher(stockId, output)
		for j := 0; j < orderNum; j++ {
			m.AddBuy(buys[j])
			m.AddSell(sells[j])
		}
	}
}
