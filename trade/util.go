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

func (o *orderMaker) Between(lower, upper int64) int64 {
	if lower == upper {
		return lower
	}
	r := upper - lower
	return o.r.Int63n(r) + lower
}

func (o *orderMaker) MkPricedBuy(price int64) *Order {
	return o.MkPricedOrder(price, BUY)
}

func (o *orderMaker) MkPricedSell(price int64) *Order {
	return o.MkPricedOrder(price, SELL)
}

func (o *orderMaker) MkPricedOrder(price int64, kind OrderKind) *Order {
	costData := CostData{Price: price, Amount: 1}
	tradeData := TradeData{TraderId: o.traderId, TradeId: 1, StockId: 1}
	o.traderId++
	return NewOrder(costData, tradeData, kind)
}

func (o *orderMaker) ValRangePyramid(n int, low, high int64) []int64 {
	seq := (high - low) / 4
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		val := o.Between(0, seq) + o.Between(0, seq) + o.Between(0, seq) + o.Between(0, seq)
		vals[i] = int64(val) + low
	}
	return vals
}

func (o *orderMaker) ValRangeFlat(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = o.Between(low, high)
	}
	return vals
}

func (o *orderMaker) MkBuys(prices []int64) []*Order {
	return o.MkOrders(prices, BUY)
}

func (o *orderMaker) MkSells(prices []int64) []*Order {
	return o.MkOrders(prices, SELL)
}

func (o *orderMaker) MkOrders(prices []int64, kind OrderKind) []*Order {
	orders := make([]*Order, len(prices))
	for i, price := range prices {
		costData := CostData{Price: price, Amount: 1}
		tradeData := TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = NewOrder(costData, tradeData, kind)
	}
	return orders
}
