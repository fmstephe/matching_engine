package main

import (
	"flag"
	"github.com/fmstephe/fstrconv"
	"github.com/fmstephe/matching_engine/itch"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/trade"
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
	filePath = flag.String("f", "", "Relative path to an ITCH file providing test data")
	profile  = flag.String("profile", "", "Write out a profile of this application, 'cpu' and 'mem' supported")
	perfRand = rand.New(rand.NewSource(1))
)

func main() {
	flag.Parse()
	orders := getData()
	orderCount := fstrconv.Itoa64Comma(int64(len(orders)))
	println(orderCount, "Orders Built")
	buffer := matcher.NewResponseBuffer(2)
	m := matcher.NewMatcher(buffer)
	startProfile()
	defer endProfile()
	start := time.Now().UnixNano()
	for i := range orders {
		m.Submit(orders[i])
	}
	println("Buffer Writes: ", buffer.Writes())
	total := time.Now().UnixNano() - start
	println("Nanos\t", fstrconv.Itoa64Comma(total))
	println("Micros\t", fstrconv.Itoa64Comma(total/1000))
	println("Millis\t", fstrconv.Itoa64Comma(total/(1000*1000)))
	println("Seconds\t", fstrconv.Itoa64Comma(total/(1000*1000*1000)))
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

func getData() []*trade.Order {
	if *filePath == "" {
		return mkRandomData()
	}
	return getItchData()
}

func getItchData() []*trade.Order {
	ir := itch.NewItchReader(*filePath)
	orders, err := ir.ReadAll()
	if err != nil {
		panic(err.Error())
	}
	return orders
}

func mkRandomData() []*trade.Order {
	orderNum := 5 * 1000 * 1000
	sells := mkSells(orderNum, 1000, 1500)
	buys := mkBuys(orderNum, 1000, 1500)
	orders := make([]*trade.Order, orderNum*2, 0)
	for i := 0; i < orderNum; i++ {
		orders = append(orders, sells[i])
		orders = append(orders, buys[i])
	}
	return orders
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

func mkOrders(n int, low, high int32, tradeType trade.OrderKind) []*trade.Order {
	prices := valRangeFlat(n, low, high)
	orders := make([]*trade.Order, n)
	for i, price := range prices {
		costData := trade.CostData{Price: price, Amount: 1}
		tradeData := trade.TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = trade.NewOrder(costData, tradeData, tradeType)
	}
	return orders
}
