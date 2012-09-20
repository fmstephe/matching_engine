package matcher

import (
	"testing"
)

const (
	orderNum      = 500 * 1000
	buySellOffset = 0
)

var (
	benchOrderMaker = newOrderMaker()
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
		buysWide = benchOrderMaker.mkBuys(benchOrderMaker.valRangePyramid(orderNum, 1000-buySellOffset, (100*1000)-buySellOffset))
	}
	if sellsWide == nil {
		sellsWide = benchOrderMaker.mkSells(benchOrderMaker.valRangePyramid(orderNum, 1000+buySellOffset, (100*1000)+buySellOffset))
	}
	// Medium Range
	if buysMedium == nil {
		buysMedium = benchOrderMaker.mkBuys(benchOrderMaker.valRangePyramid(orderNum, 1000-buySellOffset, 5000-buySellOffset))
	}
	if sellsMedium == nil {
		sellsMedium = benchOrderMaker.mkSells(benchOrderMaker.valRangePyramid(orderNum, 1000+buySellOffset, 5000+buySellOffset))
	}
	// Narrow Range
	if buysNarrow == nil {
		buysNarrow = benchOrderMaker.mkBuys(benchOrderMaker.valRangePyramid(orderNum, 1000-buySellOffset, 1500-buySellOffset))
	}
	if sellsNarrow == nil {
		sellsNarrow = benchOrderMaker.mkSells(benchOrderMaker.valRangePyramid(orderNum, 1000+buySellOffset, 1500+buySellOffset))
	}
	// Output buffer
	if output == nil {
		output = NewResponseBuffer(4)
	} else {
		output.clear()
	}
	b.StartTimer()
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
	benchmarkMatch(b, buysWide, sellsWide, 499988)
}

func BenchmarkMatchMedium(b *testing.B) {
	prepare(b)
	benchmarkMatch(b, buysMedium, sellsMedium, 499481)
}

func BenchmarkMatchNarrow(b *testing.B) {
	prepare(b)
	benchmarkMatch(b, buysNarrow, sellsNarrow, 499653)
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
