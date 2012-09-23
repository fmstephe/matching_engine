package matcher

import (
	"github.com/fmstephe/matching_engine/trade"
	"math/rand"
)

func rand32(lim int32) int32 {
	return int32(rand.Int63n(int64(lim)))
}

type orderMaker struct {
	traderId uint32
	r        *rand.Rand
}

func newOrderMaker() *orderMaker {
	r := rand.New(rand.NewSource(1))
	return &orderMaker{traderId: 0, r: r}
}

func myRand(lim int32, r *rand.Rand) int32 {
	return int32(r.Int63n(int64(lim)))
}

func (o *orderMaker) mkPricedBuy(price int32) *trade.Order {
	return o.mkPricedOrder(price, trade.BUY)
}

func (o *orderMaker) mkPricedSell(price int32) *trade.Order {
	return o.mkPricedOrder(price, trade.SELL)
}

func (o *orderMaker) mkPricedOrder(price int32, buySell trade.TradeType) *trade.Order {
	costData := trade.CostData{Price: price, Amount: 1}
	tradeData := trade.TradeData{TraderId: o.traderId, TradeId: 1, StockId: 1}
	o.traderId++
	return trade.NewOrder(costData, tradeData, buySell)
}

func (o *orderMaker) valRangePyramid(n int, low, high int32) []int32 {
	seq := (high - low) / 4
	vals := make([]int32, n)
	for i := 0; i < n; i++ {
		val := myRand(seq, o.r) + myRand(seq, o.r) + myRand(seq, o.r) + myRand(seq, o.r)
		vals[i] = int32(val) + low
	}
	return vals
}

func (o *orderMaker) valRangeFlat(n int, low, high int32) []int32 {
	vals := make([]int32, n)
	for i := 0; i < n; i++ {
		vals[i] = myRand(high-low, o.r) + low
	}
	return vals
}

func (o *orderMaker) mkBuys(prices []int32) []*trade.Order {
	return o.mkOrders(prices, trade.BUY)
}

func (o *orderMaker) mkSells(prices []int32) []*trade.Order {
	return o.mkOrders(prices, trade.SELL)
}

func (o *orderMaker) mkOrders(prices []int32, buySell trade.TradeType) []*trade.Order {
	orders := make([]*trade.Order, len(prices))
	for i, price := range prices {
		costData := trade.CostData{Price: price, Amount: 1}
		tradeData := trade.TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = trade.NewOrder(costData, tradeData, buySell)
	}
	return orders
}
