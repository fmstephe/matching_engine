package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/pqueue/limitheap"
	"github.com/fmstephe/matching_engine/trade"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
	"strings"
	"strconv"
)

const (
	stockId = uint32(1)
)

var (
	profile  = flag.String("profile", "", "Write out a profile of this application, 'cpu' and 'mem' supported")
	perfRand = rand.New(rand.NewSource(1))
)

func Foo() {
	flag.Parse()
	orderNum := 20 * 1000 * 1000
	sells := mkSells(orderNum, 1000, 1500)
	buys := mkBuys(orderNum, 1000, 1500)
	buysQ := limitheap.New(trade.BUY, 2000, 10 * 1000 * 1000, orderNum)
	sellsQ := limitheap.New(trade.SELL, 2000, 10 * 1000 * 1000, orderNum)
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
	println("Nanos\t", total)
	println("Mircos\t", total/1000)
	println("Millis\t", total/(1000*1000))
	println("Seconds\t", total/(1000*1000*1000))
}

func main() {
	printLineCount("20120709_SPY.odat")
	readOrders("20120709_SPY.odat")
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

func printLineCount(fName string) {
	f, _ := os.Open(fName)
	r := bufio.NewReader(f)
	i := 0
	for {
		if _, err := r.ReadString('\n'); err != nil {
			if err == io.EOF {
				break
			}
			panic(err.Error())
		}
		i++
	}
	println(i)
}

func readOrders(fName string) []*trade.Order{
	f, _ := os.Open(fName)
	r := bufio.NewReader(f)
	orders := make([]*trade.Order, 10)
	i := 0
	// Read column headers
	if _, err := r.ReadString('\n'); err != nil {
		panic(err.Error())
	}
	for {
		var line string
		var err error
		if line, err = r.ReadString('\n'); err != nil {
			panic(err.Error())
		}
		orders = append(orders, mkOrder(line))
		i++
		if i > 1000 {
			break
		}
	}
	return orders
}

func mkOrder(line string) *trade.Order {
	ss := strings.Split(line, " ")
	var useful []string
	for _, w := range ss {
		if w != "" && w != "\n" {
			useful = append(useful, w)
		}
	}
	cd, td := mkData(useful)
	switch useful[3] {
		case "B" : return trade.NewBuy(cd, td)
		case "S" : return trade.NewSell(cd, td)
		case "D" : return trade.NewDelete(td)
		default : panic(fmt.Sprintf("Unrecognised Trade Type %s", useful[3]))
	}
	panic("Unreachable")
}

func mkData(useful []string) (cd trade.CostData, td trade.TradeData) {
	//	print("ID: ", useful[2], " Type: ", useful[3], " Price: ",  useful[4], " Amount: ", useful[5])
	//	println()
	var price, amount, traderId, tradeId, stockId int
	var err error
	if price, err = strconv.Atoi(useful[4]); err != nil {
		panic(err.Error())
	}
	if amount, err = strconv.Atoi(useful[5]); err != nil {
		panic(err.Error())
	}
	if traderId, err = strconv.Atoi(useful[2]); err != nil {
		panic(err.Error())
	}
	if tradeId, err = strconv.Atoi(useful[2]); err != nil {
		panic(err.Error())
	}
	stockId = 1
	cd = trade.CostData{Price: int32(price), Amount: uint32(amount)}
	td = trade.TradeData{TraderId: uint32(traderId), TradeId: uint32(tradeId), StockId: uint32(stockId)}
	return
}
