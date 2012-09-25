package main

import (
	"flag"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/trade"
	"github.com/fmstephe/matching_engine/pqueue/rtree"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

const (
	stockId = uint32(1)
)

var (
	profile  = flag.String("profile", "", "Write out a profile of this application, 'cpu' and 'mem' supported")
	perfRand = rand.New(rand.NewSource(1))
)

func main() {
	flag.Parse()
	orderNum := 20 * 1000 * 1000
	sells := mkSells(orderNum, 1000, 1500)
	buys := mkBuys(orderNum, 1000, 1500)
	buysQ := rtree.New(trade.BUY)
	sellsQ := rtree.New(trade.SELL)
	buffer := matcher.NewResponseBuffer(2)
	m := matcher.NewMatcher(buysQ, sellsQ, buffer)
	startProfile()
	defer endProfile()
	start := time.Now().UnixNano()
	for i := 0; i < orderNum; i++ {
		m.AddBuy(buys[i])
		m.AddSell(sells[i])
	}
	total := time.Now().UnixNano() - start
	println(total)
}

func startProfile() {
	if *profile == "cpu" {
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}
}

func endProfile() {
	if *profile == "cpu" {
		pprof.StopCPUProfile()
	}
	if *profile == "mem" {
		f, err := os.Create("mem.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
	}
}

func myRand(lim int32, r *rand.Rand) int32 {
	return int32(r.Int63n(int64(lim)))
}

func valRangeFlat(n int, low, high int32) []int32 {
	vals := make([]int32, n)
	for i := 0; i < n; i++ {
		vals[i] = myRand(high-low, perfRand) + low
	}
	return vals
}

func valRangePyramid(n int, low, high int32) []int32 {
	seq := (high - low) / 4
	vals := make([]int32, n)
	for i := 0; i < n; i++ {
		val := myRand(seq, perfRand) + myRand(seq, perfRand) + myRand(seq, perfRand) + myRand(seq, perfRand)
		vals[i] = val + low
	}
	return vals
}

func mkBuys(n int, low, high int32) []*trade.Order {
	return mkOrders(n, low, high, trade.BUY)
}

func mkSells(n int, low, high int32) []*trade.Order {
	return mkOrders(n, low, high, trade.SELL)
}

func mkOrders(n int, low, high int32, buySell trade.TradeType) []*trade.Order {
	prices := valRangeFlat(n, low, high)
	orders := make([]*trade.Order, n)
	for i, price := range prices {
		costData := trade.CostData{Price: price, Amount: 1}
		tradeData := trade.TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = trade.NewOrder(costData, tradeData, buySell)
	}
	return orders
}
