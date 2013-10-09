package msg

import (
	"errors"
	"fmt"
	"math/rand"
)

type MessageMaker struct {
	traderId uint32
	r        *rand.Rand
}

func NewMessageMaker(initTraderId uint32) *MessageMaker {
	r := rand.New(rand.NewSource(1))
	return &MessageMaker{traderId: initTraderId, r: r}
}

func (mm *MessageMaker) Seed(seed int64) {
	mm.r.Seed(seed)
}

func (mm *MessageMaker) Between(lower, upper int64) int64 {
	if lower == upper {
		return lower
	}
	d := upper - lower
	return mm.r.Int63n(d) + lower
}

func (mm *MessageMaker) MkPricedOrder(price int64, kind MsgKind) *Message {
	m := &Message{}
	mm.writePricedOrder(price, kind, m)
	return m
}

func (mm *MessageMaker) writePricedOrder(price int64, kind MsgKind, m *Message) {
	mm.traderId++
	*m = Message{Price: price, Amount: 1, TraderId: mm.traderId, TradeId: 1, StockId: 1}
	m.Kind = kind
}

func (mm *MessageMaker) ValRangePyramid(n int, low, high int64) []int64 {
	seq := (high - low) / 4
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		val := mm.Between(0, seq) + mm.Between(0, seq) + mm.Between(0, seq) + mm.Between(0, seq)
		vals[i] = int64(val) + low
	}
	return vals
}

func (mm *MessageMaker) ValRangeFlat(n int, low, high int64) []int64 {
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = mm.Between(low, high)
	}
	return vals
}

func (mm *MessageMaker) MkBuys(prices []int64, stockId uint32) []Message {
	return mm.MkOrders(prices, stockId, BUY)
}

func (mm *MessageMaker) MkSells(prices []int64, stockId uint32) []Message {
	return mm.MkOrders(prices, stockId, SELL)
}

func (mm *MessageMaker) MkOrders(prices []int64, stockId uint32, kind MsgKind) []Message {
	msgs := make([]Message, len(prices))
	for i, price := range prices {
		mm.traderId++
		msgs[i] = Message{Price: price, Amount: 1, TraderId: mm.traderId, TradeId: uint32(i + 1), StockId: stockId}
		msgs[i].Kind = kind
	}
	return msgs
}

func (mm *MessageMaker) RndTradeSet(size, depth int, low, high int64) ([]Message, error) {
	if depth > size {
		return nil, errors.New(fmt.Sprintf("Size (%d) must be greater than or equal to (%d)", size, depth))
	}
	orders := make([]Message, size*4)
	buys := make([]*Message, 0, size)
	sells := make([]*Message, 0, size)
	idx := 0
	for i := 0; i < size+depth; i++ {
		if i < size {
			b := &orders[idx]
			idx++
			mm.writePricedOrder(mm.Between(low, high), BUY, b)
			buys = append(buys, b)
			if b.Price == 0 {
				b.Price = 1 // Buys can't have price of 0
			}
			s := &orders[idx]
			idx++
			mm.writePricedOrder(mm.Between(low, high), SELL, s)
			sells = append(sells, s)
		}
		if i >= depth {
			b := buys[i-depth]
			cb := &orders[idx]
			idx++
			cb.WriteCancelFor(b)
			s := sells[i-depth]
			cs := &orders[idx]
			idx++
			cs.WriteCancelFor(s)
		}
	}
	return orders, nil
}
