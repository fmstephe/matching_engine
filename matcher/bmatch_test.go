package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
	"math/rand"
)

const (
	orderNum = 1000 * 10
)

var benchRand = rand.New(rand.NewSource(1))
var buys []*trade.Order
var sells []*trade.Order

func init() {
	buys = mkBuys(orderNum, 1000, 1500)
	println("Finished Buys")
	sells  = mkSells(orderNum, 1000, 1500)
	println("Finished Sells")
}

func valRange(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = rand.Int63n(high - low) + low
	}
	return vals
}

func mkBuys(n int, low, high int64) []*trade.Order {
	return mkOrders(n, low, high, trade.BUY)
}

func mkSells(n int, low, high int64) []*trade.Order {
	return mkOrders(n, low, high, trade.SELL)
}

func mkOrders(n int, low, high int64, buySell trade.TradeType) []*trade.Order {
	prices := valRange(n, low, high)
	orders := make([]*trade.Order, n)
	for i, price := range prices {
		rc := make(chan *trade.Response, 256)
		orders[i] = trade.NewOrder(int64(i), 1, price, stockId, fmt.Sprintf("benchTrader%d",i), rc, buySell)
		go func() {
			<-rc
		}()
	}
	return orders
}

func BenchmarkAddBuy(b *testing.B) {
	m := New(stockId)
	for _, buy := range buys {
		m.AddBuy(buy)
	}
}

func BenchmarkAddSell(b *testing.B) {
	m := New(stockId)
	for _, buy := range buys {
		m.AddBuy(buy)
	}
}

func BenchmarkMatch(b *testing.B) {
	m := New(stockId)
	for i := 0; i < orderNum; i++ {
		m.AddBuy(buys[i])
		m.AddSell(sells[i])
	}
}
