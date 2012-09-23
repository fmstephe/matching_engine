package trade

import (
	"math/rand"
)

var (
	stockId = uint32(1)
)

type orderMaker struct {
	traderId uint32
	r        *rand.Rand
}

func NewOrderMaker() *orderMaker {
	r := rand.New(rand.NewSource(1))
	return &orderMaker{traderId: 0, r: r}
}

func (o *orderMaker) Rand32(lim int32) int32 {
	return int32(o.r.Int63n(int64(lim)))
}

func (o *orderMaker) MkPricedBuy(price int32) *Order {
	return o.MkPricedOrder(price, BUY)
}

func (o *orderMaker) MkPricedSell(price int32) *Order {
	return o.MkPricedOrder(price, SELL)
}

func (o *orderMaker) MkPricedOrder(price int32, buySell TradeType) *Order {
	costData := CostData{Price: price, Amount: 1}
	tradeData := TradeData{TraderId: o.traderId, TradeId: 1, StockId: 1}
	o.traderId++
	return NewOrder(costData, tradeData, buySell)
}

func (o *orderMaker) ValRangePyramid(n int, low, high int32) []int32 {
	seq := (high - low) / 4
	vals := make([]int32, n)
	for i := 0; i < n; i++ {
		val := o.Rand32(seq) + o.Rand32(seq) + o.Rand32(seq) + o.Rand32(seq)
		vals[i] = int32(val) + low
	}
	return vals
}

func (o *orderMaker) ValRangeFlat(n int, low, high int32) []int32 {
	vals := make([]int32, n)
	for i := 0; i < n; i++ {
		vals[i] = o.Rand32(high-low) + low
	}
	return vals
}

func (o *orderMaker) MkBuys(prices []int32) []*Order {
	return o.MkOrders(prices, BUY)
}

func (o *orderMaker) MkSells(prices []int32) []*Order {
	return o.MkOrders(prices, SELL)
}

func (o *orderMaker) MkOrders(prices []int32, buySell TradeType) []*Order {
	orders := make([]*Order, len(prices))
	for i, price := range prices {
		costData := CostData{Price: price, Amount: 1}
		tradeData := TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = NewOrder(costData, tradeData, buySell)
	}
	return orders
}
