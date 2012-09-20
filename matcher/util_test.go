package matcher

import (
	"math/rand"
)

type orderMaker struct {
	traderId uint32
	r *rand.Rand
}

func newOrderMaker() *orderMaker {
	r := rand.New(rand.NewSource(1))
	return &orderMaker{traderId: 0, r: r}
}

func myRand(lim int64, r *rand.Rand) int64 {
	return r.Int63n(lim)
}

func (o *orderMaker) mkPricedBuy(price int64) *Order {
	return o.mkPricedOrder(price, BUY)
}

func (o *orderMaker) mkPricedSell(price int64) *Order {
	return o.mkPricedOrder(price, SELL)
}

func (o *orderMaker) mkPricedOrder(price int64, buySell TradeType) *Order {
	costData := CostData{Price: price, Amount: 1}
	tradeData := TradeData{TraderId: o.traderId, TradeId: 1, StockId: 1}
	o.traderId++
	return NewOrder(costData, tradeData, buySell)
}

func (o *orderMaker) valRangePyramid(n int, low, high int64) []int64 {
	seq := (high - low) / 4
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		val := myRand(seq, o.r) + myRand(seq, o.r) + myRand(seq, o.r) + myRand(seq, o.r)
		vals[i] = int64(val) + low
	}
	return vals
}

func (o *orderMaker) valRangeFlat(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = myRand(high-low, o.r) + low
	}
	return vals
}

func (o *orderMaker) mkBuys(prices []int64) []*Order {
	return o.mkOrders(prices, BUY)
}

func (o *orderMaker) mkSells(prices []int64) []*Order {
	return o.mkOrders(prices, SELL)
}

func (o *orderMaker) mkOrders(prices []int64, buySell TradeType) []*Order {
	orders := make([]*Order, len(prices))
	for i, price := range prices {
		costData := CostData{Price: price, Amount: 1}
		tradeData := TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = NewOrder(costData, tradeData, buySell)
	}
	return orders
}
