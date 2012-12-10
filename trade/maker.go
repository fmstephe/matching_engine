package trade

import (
	"math/rand"
)

var (
	stockId = uint32(1)
)

type OrderMaker struct {
	traderId uint32
	r        *rand.Rand
}

func NewOrderMaker() *OrderMaker {
	r := rand.New(rand.NewSource(1))
	return &OrderMaker{traderId: 0, r: r}
}

func (o *OrderMaker) Between(lower, upper int64) int64 {
	if lower == upper {
		return lower
	}
	r := upper - lower
	return o.r.Int63n(r) + lower
}

func (o *OrderMaker) MkPricedBuyData(price int64) *OrderData {
	if price == 0 {
		price = 1 // 'market' buys are not allowed
	}
	return o.MkPricedOrderData(price, BUY)
}

func (o *OrderMaker) MkPricedBuy(price int64) *Order {
	return NewOrderFromData(o.MkPricedBuyData(price))
}

func (o *OrderMaker) MkPricedSellData(price int64) *OrderData {
	return o.MkPricedOrderData(price, SELL)
}

func (o *OrderMaker) MkPricedSell(price int64) *Order {
	return o.MkPricedSell(price)
}

func (o *OrderMaker) MkPricedOrder(price int64, kind OrderKind) *Order {
	return NewOrderFromData(o.MkPricedOrderData(price, kind))
}

func (o *OrderMaker) MkPricedOrderData(price int64, kind OrderKind) *OrderData {
	costData := CostData{Price: price, Amount: 1}
	tradeData := TradeData{TraderId: o.traderId, TradeId: 1, StockId: 1}
	o.traderId++
	return NewOrderData(costData, tradeData, kind)
}

func (o *OrderMaker) ValRangePyramid(n int, low, high int64) []int64 {
	seq := (high - low) / 4
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		val := o.Between(0, seq) + o.Between(0, seq) + o.Between(0, seq) + o.Between(0, seq)
		vals[i] = int64(val) + low
	}
	return vals
}

func (o *OrderMaker) ValRangeFlat(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = o.Between(low, high)
	}
	return vals
}

func (o *OrderMaker) MkBuys(prices []int64) []*OrderData {
	return o.MkOrderDatas(prices, BUY)
}

func (o *OrderMaker) MkSells(prices []int64) []*OrderData {
	return o.MkOrderDatas(prices, SELL)
}

func (o *OrderMaker) MkOrderDatas(prices []int64, kind OrderKind) []*OrderData {
	orders := make([]*OrderData, len(prices))
	for i, price := range prices {
		costData := CostData{Price: price, Amount: 1}
		tradeData := TradeData{TraderId: uint32(i), TradeId: uint32(i), StockId: stockId}
		orders[i] = NewOrderData(costData, tradeData, kind)
	}
	return orders
}
