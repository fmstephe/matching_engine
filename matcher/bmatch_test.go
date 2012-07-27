package matcher

import (
	"github.com/fmstephe/matching_engine/trade"
	"testing"
	"math/rand"
)

const (
	orderNum = 1000 * 1000
)

var benchRand = rand.New(rand.NewSource(1))
var buys []*trade.Order
var sells []*trade.Order

func prepare(b *testing.B) {
	b.StopTimer()
	if buys == nil {
		buys = mkBuys(orderNum, 1000, 1500)
	}
	if sells == nil {
		sells  = mkSells(orderNum, 1000, 1500)
	}
	b.StartTimer()
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
		responseFunc := func(response *trade.Response) {
			// Do Nothing
		}
		costData := trade.CostData{Price: price, Amount: 1}
		tradeData := trade.TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = trade.NewOrder(costData, tradeData, responseFunc, buySell)
	}
	return orders
}

func BenchmarkAddBuy(b *testing.B) {
	prepare(b)
	m := New(stockId)
	for _, buy := range buys {
		m.AddBuy(buy)
	}
}

func BenchmarkAddSell(b *testing.B) {
	prepare(b)
	m := New(stockId)
	for _, buy := range buys {
		m.AddBuy(buy)
	}
}

func BenchmarkMatch(b *testing.B) {
	prepare(b)
	m := New(stockId)
	for i := 0; i < orderNum; i++ {
		m.AddBuy(buys[i])
		m.AddSell(sells[i])
	}
}
