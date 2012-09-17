package matcher

import (
	"math/rand"
	"testing"
)

const (
	orderNum      = 500 * 1000
	buySellOffset = 0
)

var (
	benchRand   = rand.New(rand.NewSource(1))
	buysWide    []*Order
	buysMedium  []*Order
	buysNarrow  []*Order
	sellsWide   []*Order
	sellsMedium []*Order
	sellsNarrow []*Order
	output      *ResponseBuffer
)

func prepare(b *testing.B) {
	b.StopTimer()
	// Wide range 
	if buysWide == nil {
		buysWide = mkBuys(valRangePyramid(orderNum, 1000-buySellOffset, (100*1000)-buySellOffset))
	}
	if sellsWide == nil {
		sellsWide = mkSells(valRangePyramid(orderNum, 1000+buySellOffset, (100*1000)+buySellOffset))
	}
	// Medium Range
	if buysMedium == nil {
		buysMedium = mkBuys(valRangePyramid(orderNum, 1000-buySellOffset, 5000-buySellOffset))
	}
	if sellsMedium == nil {
		sellsMedium = mkSells(valRangePyramid(orderNum, 1000+buySellOffset, 5000+buySellOffset))
	}
	// Narrow Range
	if buysNarrow == nil {
		buysNarrow = mkBuys(valRangePyramid(orderNum, 1000-buySellOffset, 1500-buySellOffset))
	}
	if sellsNarrow == nil {
		sellsNarrow = mkSells(valRangePyramid(orderNum, 1000+buySellOffset, 1500+buySellOffset))
	}
	// Output buffer
	if output == nil {
		output = NewResponseBuffer(4)
	} else {
		output.clear()
	}
	b.StartTimer()
}

func valRangePyramid(n int, low, high int64) []int64 {
	seq := (high - low) / 4
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		val := benchRand.Int63n(seq) + benchRand.Int63n(seq) + benchRand.Int63n(seq) + benchRand.Int63n(seq)
		vals[i] = int64(val) + low
	}
	return vals
}

func valRangeFlat(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = benchRand.Int63n(high-low) + low
	}
	return vals
}

func mkBuys(prices []int64) []*Order {
	return mkOrders(prices, BUY)
}

func mkSells(prices []int64) []*Order {
	return mkOrders(prices, SELL)
}

func mkOrders(prices []int64, buySell TradeType) []*Order {
	orders := make([]*Order, len(prices))
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
	benchmarkMatch(b, buysWide, sellsWide, 499987)
}

func BenchmarkMatchMedium(b *testing.B) {
	prepare(b)
	benchmarkMatch(b, buysMedium, sellsMedium, 499994)
}

func BenchmarkMatchNarrow(b *testing.B) {
	prepare(b)
	benchmarkMatch(b, buysNarrow, sellsNarrow, 499990)
}

func benchmarkMatch(b *testing.B, buys, sells []*Order, expMatches int) {
	for i := 0; i < b.N; i++ {
		prepare(b)
		m := NewMatcher(stockId, output)
		for j := 0; j < orderNum; j++ {
			m.AddBuy(buys[j])
			m.AddSell(sells[j])
		}
		if (output.write/2 != expMatches) {
			println("Expecting", expMatches, "found", output.write/2, "instead")
		}
	}
}
