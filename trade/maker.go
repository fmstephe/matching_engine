package trade

import (
	"errors"
	"fmt"
	"math/rand"
)

var (
	stockId = uint32(1)
)

type OrderNodeMaker struct {
	traderId uint32
	r        *rand.Rand
}

func NewOrderMaker() *OrderNodeMaker {
	r := rand.New(rand.NewSource(1))
	return &OrderNodeMaker{traderId: 0, r: r}
}

func (o *OrderNodeMaker) Seed(seed int64) {
	o.r.Seed(seed)
}

func (o *OrderNodeMaker) Between(lower, upper int64) int64 {
	if lower == upper {
		return lower
	}
	d := upper - lower
	return o.r.Int63n(d) + lower
}

func (o *OrderNodeMaker) MkPricedOrder(price int64, kind OrderKind) *Order {
	od := &Order{}
	o.writePricedOrder(price, kind, od)
	return od
}

func (o *OrderNodeMaker) writePricedOrder(price int64, kind OrderKind, od *Order) {
	costData := CostData{Price: price, Amount: 1}
	tradeData := TradeData{TraderId: o.traderId, TradeId: 1, StockId: 1}
	o.traderId++
	od.Write(costData, tradeData, kind)
}

func (o *OrderNodeMaker) ValRangePyramid(n int, low, high int64) []int64 {
	seq := (high - low) / 4
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		val := o.Between(0, seq) + o.Between(0, seq) + o.Between(0, seq) + o.Between(0, seq)
		vals[i] = int64(val) + low
	}
	return vals
}

func (o *OrderNodeMaker) ValRangeFlat(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = o.Between(low, high)
	}
	return vals
}

func (o *OrderNodeMaker) MkBuys(prices []int64) []Order {
	return o.MkOrders(prices, BUY)
}

func (o *OrderNodeMaker) MkSells(prices []int64) []Order {
	return o.MkOrders(prices, SELL)
}

func (o *OrderNodeMaker) MkOrders(prices []int64, kind OrderKind) []Order {
	orders := make([]Order, len(prices))
	for i, price := range prices {
		costData := CostData{Price: price, Amount: 1}
		tradeData := TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i].Write(costData, tradeData, kind)
	}
	return orders
}

func (o *OrderNodeMaker) RndTradeSet(size, depth int, low, high int64) ([]Order, error) {
	if depth > size {
		return nil, errors.New(fmt.Sprintf("Size (%d) must be greater than or equal to (%d)", size, depth))
	}
	orders := make([]Order, size*4)
	buys := make([]*Order, 0, size)
	sells := make([]*Order, 0, size)
	idx := 0
	for i := 0; i < size+depth; i++ {
		if i < size {
			b := &orders[idx]
			idx++
			o.writePricedOrder(o.Between(low, high), BUY, b)
			buys = append(buys, b)
			if b.Price == 0 {
				b.Price = 1 // Buys can't have price of 0
			}
			s := &orders[idx]
			idx++
			o.writePricedOrder(o.Between(low, high), SELL, s)
			sells = append(sells, s)
		}
		if i >= depth {
			b := buys[i-depth]
			cb := &orders[idx]
			idx++
			cb.WriteCancelFromOrder(b)
			s := sells[i-depth]
			cs := &orders[idx]
			idx++
			cs.WriteCancelFromOrder(s)
		}
	}
	return orders, nil
}
