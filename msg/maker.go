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

func (mm *MessageMaker) Between(lower, upper uint64) uint64 {
	if lower > upper {
		panic(fmt.Sprintf("lower must be less than upper. lower was %d, upper was %d", lower, upper))
	}
	if int64(lower) < 0 || int64(upper) < 0 {
		panic(fmt.Sprintf("lower and higher must be <= 2^63. lower was %d, upper was %d", lower, upper))
	}
	if lower == upper {
		return lower
	}
	low, up := int64(lower), int64(upper)
	return uint64(mm.r.Int63n(up-low) + low)
}

func (mm *MessageMaker) MkPricedOrder(price uint64, kind MsgKind) *Message {
	m := &Message{}
	mm.writePricedOrder(price, kind, m)
	return m
}

func (mm *MessageMaker) writePricedOrder(price uint64, kind MsgKind, m *Message) {
	mm.traderId++
	*m = Message{Price: price, Amount: 1, TraderId: mm.traderId, TradeId: 1, StockId: 1}
	m.Kind = kind
}

func (mm *MessageMaker) ValRangePyramid(n int, low, high uint64) []uint64 {
	seq := (high - low) / 4
	vals := make([]uint64, n)
	for i := 0; i < n; i++ {
		val := mm.Between(0, seq) + mm.Between(0, seq) + mm.Between(0, seq) + mm.Between(0, seq)
		vals[i] = uint64(val) + low
	}
	return vals
}

func (mm *MessageMaker) ValRangeFlat(n int, low, high uint64) []uint64 {
	vals := make([]uint64, n)
	for i := 0; i < n; i++ {
		vals[i] = mm.Between(low, high)
	}
	return vals
}

func (mm *MessageMaker) MkBuys(prices []uint64, stockId uint64) []Message {
	return mm.MkOrders(prices, stockId, BUY)
}

func (mm *MessageMaker) MkSells(prices []uint64, stockId uint64) []Message {
	return mm.MkOrders(prices, stockId, SELL)
}

func (mm *MessageMaker) MkOrders(prices []uint64, stockId uint64, kind MsgKind) []Message {
	msgs := make([]Message, len(prices))
	for i, price := range prices {
		mm.traderId++
		msgs[i] = Message{Price: price, Amount: 1, TraderId: mm.traderId, TradeId: uint32(i + 1), StockId: stockId}
		msgs[i].Kind = kind
	}
	return msgs
}

func (mm *MessageMaker) RndTradeSet(size, depth int, low, high uint64) ([]Message, error) {
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
