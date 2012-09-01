package main

import (
	"math/rand"
	"log"
	"os"
	"runtime/pprof"
	"github.com/fmstephe/matching_engine/matcher"
	"time"
)

const (
	stockId = uint32(1)
)

var (
	responseFunc = func(response *matcher.Response) {
		// Do Nothing
	}
)

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	orderNum := 20 * 1000 * 1000
	sells := mkSells(orderNum, 1000, 1500)
	buys := mkBuys(orderNum, 1000, 1500)
	buffer := matcher.NewResponseBuffer(2)
	m := matcher.NewMatcher(stockId, buffer)
	pprof.StartCPUProfile(f)
	start := time.Now().UnixNano()
	for i := 0; i < orderNum; i++ {
		m.AddBuy(buys[i])
		m.AddSell(sells[i])
	}
	total := time.Now().UnixNano() - start
	println(total)
	defer pprof.StopCPUProfile()
}

func valRangeFlat(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = rand.Int63n(high-low) + low
	}
	return vals
}

func valRangePyramid(n int, low, high int64) []int64 {
	seq := (high - low) / 4
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		val := rand.Int63n(seq) + rand.Int63n(seq) + rand.Int63n(seq) + rand.Int63n(seq)
		vals[i] = val + low
	}
	return vals
}

func mkBuys(n int, low, high int64) []*matcher.Order {
	return mkOrders(n, low, high, matcher.BUY)
}

func mkSells(n int, low, high int64) []*matcher.Order {
	return mkOrders(n, low, high, matcher.SELL)
}

func mkOrders(n int, low, high int64, buySell matcher.TradeType) []*matcher.Order {
	prices := valRangeFlat(n, low, high)
	orders := make([]*matcher.Order, n)
	for i, price := range prices {
		costData := matcher.CostData{Price: price, Amount: 1}
		tradeData := matcher.TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = matcher.NewOrder(costData, tradeData, responseFunc, buySell)
	}
	return orders
}
