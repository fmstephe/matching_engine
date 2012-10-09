package matcher

import (
	"encoding/json"
	"github.com/fmstephe/matching_engine/pqueue/limitheap"
	"github.com/fmstephe/matching_engine/trade"
	"io/ioutil"
	"testing"
)

const (
	orderNum        = 500 * 1000
	buySellOffset   = 0
	dataDir         = "data/"
	buysWideFile    = dataDir + "buyswide"
	sellsWideFile   = dataDir + "sellswide"
	buysMediumFile  = dataDir + "buymedium"
	sellsMediumFile = dataDir + "sellsmedium"
	buysNarrowFile  = dataDir + "buysnarrow"
	sellsNarrowFile = dataDir + "sellsnarrow"
)

var (
	isInitialised   = false
	benchOrderMaker = trade.NewOrderMaker()
	buysWide        []*trade.Order
	buysMedium      []*trade.Order
	buysNarrow      []*trade.Order
	sellsWide       []*trade.Order
	sellsMedium     []*trade.Order
	sellsNarrow     []*trade.Order
	output          *ResponseBuffer
)

func writeOut() {
	println("Beginning Test Data Write")
	//
	buysWideSlice, _ := json.Marshal(buysWide)
	sellsWideSlice, _ := json.Marshal(sellsWide)
	buysMediumSlice, _ := json.Marshal(buysMedium)
	sellsMediumSlice, _ := json.Marshal(sellsMedium)
	buysNarrowSlice, _ := json.Marshal(buysNarrow)
	sellsNarrowSlice, _ := json.Marshal(sellsNarrow)
	//
	ioutil.WriteFile(buysWideFile, buysWideSlice, 0777)
	ioutil.WriteFile(sellsWideFile, sellsWideSlice, 0777)
	ioutil.WriteFile(buysMediumFile, buysMediumSlice, 0777)
	ioutil.WriteFile(sellsMediumFile, sellsMediumSlice, 0777)
	ioutil.WriteFile(buysNarrowFile, buysNarrowSlice, 0777)
	ioutil.WriteFile(sellsNarrowFile, sellsNarrowSlice, 0777)
	println("Test Data Write Complete")
}

func readIn() {
	println("Beginning Test Data Read")
	dataDir := "data/"
	buysWideFile := dataDir + "buyswide"
	sellsWideFile := dataDir + "sellswide"
	buysMediumFile := dataDir + "buymedium"
	sellsMediumFile := dataDir + "sellsmedium"
	buysNarrowFile := dataDir + "buysnarrow"
	sellsNarrowFile := dataDir + "sellsnarrow"
	//
	buysWideSlice, _ := ioutil.ReadFile(buysWideFile)
	sellsWideSlice, _ := ioutil.ReadFile(sellsWideFile)
	buysMediumSlice, _ := ioutil.ReadFile(buysMediumFile)
	sellsMediumSlice, _ := ioutil.ReadFile(sellsMediumFile)
	buysNarrowSlice, _ := ioutil.ReadFile(buysNarrowFile)
	sellsNarrowSlice, _ := ioutil.ReadFile(sellsNarrowFile)
	//
	json.Unmarshal(buysWideSlice, &buysWide)
	json.Unmarshal(sellsWideSlice, &sellsWide)
	json.Unmarshal(buysMediumSlice, &buysMedium)
	json.Unmarshal(sellsMediumSlice, &sellsMedium)
	json.Unmarshal(buysNarrowSlice, &buysNarrow)
	json.Unmarshal(sellsNarrowSlice, &sellsNarrow)
	//
	println("Test Data Read Complete")
}

func create() {
	// Wide range 
	if buysWide == nil {
		buysWide = benchOrderMaker.MkBuys(benchOrderMaker.ValRangePyramid(orderNum, 1000-buySellOffset, (100*1000)-buySellOffset))
	}
	if sellsWide == nil {
		sellsWide = benchOrderMaker.MkSells(benchOrderMaker.ValRangePyramid(orderNum, 1000+buySellOffset, (100*1000)+buySellOffset))
	}
	// Medium Range
	if buysMedium == nil {
		buysMedium = benchOrderMaker.MkBuys(benchOrderMaker.ValRangePyramid(orderNum, 1000-buySellOffset, 5000-buySellOffset))
	}
	if sellsMedium == nil {
		sellsMedium = benchOrderMaker.MkSells(benchOrderMaker.ValRangePyramid(orderNum, 1000+buySellOffset, 5000+buySellOffset))
	}
	// Narrow Range
	if buysNarrow == nil {
		buysNarrow = benchOrderMaker.MkBuys(benchOrderMaker.ValRangePyramid(orderNum, 1000-buySellOffset, 1500-buySellOffset))
	}
	if sellsNarrow == nil {
		sellsNarrow = benchOrderMaker.MkSells(benchOrderMaker.ValRangePyramid(orderNum, 1000+buySellOffset, 1500+buySellOffset))
	}
}

func prepare(b *testing.B) {
	b.StopTimer()
	if !isInitialised {
		create()
		//	writeOut()
		//	readIn()
		output = NewResponseBuffer(4)
		isInitialised = true
	}
	output.clear()
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

func benchmarkAddBuy(b *testing.B, buys []*trade.Order) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buyQ := limitheap.New(trade.BUY, 2000, 10*1000*1000, orderNum)
		sellQ := limitheap.New(trade.SELL, 2000, 10*1000*1000, orderNum)
		m := NewMatcher(buyQ, sellQ, output)
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

func benchmarkAddSell(b *testing.B, sells []*trade.Order) {
	for i := 0; i < b.N; i++ {
		prepare(b)
		b.StopTimer()
		buyQ := limitheap.New(trade.BUY, 2000, 10*1000*1000, orderNum)
		sellQ := limitheap.New(trade.SELL, 2000, 10*1000*1000, orderNum)
		m := NewMatcher(buyQ, sellQ, output)
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
	benchmarkMatch(b, buysMedium, sellsMedium, 989)
}

func BenchmarkMatchNarrow(b *testing.B) {
	prepare(b)
	benchmarkMatch(b, buysNarrow, sellsNarrow, 994)
}

func benchmarkMatch(b *testing.B, buys, sells []*trade.Order, expMatches int) {
	for i := 0; i < b.N; i++ {
		prepare(b)
		buyQ := limitheap.New(trade.BUY, 2000, 10*1000*1000, orderNum)
		sellQ := limitheap.New(trade.SELL, 2000, 10*1000*1000, orderNum)
		m := NewMatcher(buyQ, sellQ, output)
		for j := 0; j < orderNum; j++ {
			m.AddBuy(buys[j])
			m.AddSell(sells[j])
		}
		if output.write/2 != expMatches {
			println("Expecting", expMatches, "found", output.write/2, "instead")
		}
	}
}
