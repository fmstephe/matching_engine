package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
)

type M struct {
	matchTrees trade.MatchTrees // No constructor required
	slab       *trade.Slab
	output     *ResponseBuffer
}

func NewMatcher(slabSize int, output *ResponseBuffer) *M {
	slab := trade.NewSlab(slabSize)
	return &M{slab: slab, output: output}
}

func (m *M) Submit(in *trade.Order) {
	o := m.slab.Malloc()
	in.CopyInto(o)
	switch o.Kind {
	case trade.BUY:
		m.addBuy(o)
	case trade.SELL:
		m.addSell(o)
	case trade.DELETE:
		m.remove(o)
	default:
		panic(fmt.Sprintf("OrderKind %#v not supported", o.Kind))
	}
}

func (m *M) addBuy(b *trade.Order) {
	if b.Price() == trade.MARKET_PRICE {
		panic("It is illegal to submit a buy at market price")
	}
	if !m.fillableBuy(b) {
		m.matchTrees.PushBuy(b)
	}
}

func (m *M) addSell(s *trade.Order) {
	if !m.fillableSell(s) {
		m.matchTrees.PushSell(s)
	}
}

func (m *M) remove(o *trade.Order) {
	ro := m.matchTrees.Pop(o)
	if ro != nil {
		m.slab.Free(ro)
	}
	m.slab.Free(o)
}

func (m *M) fillableBuy(b *trade.Order) bool {
	for {
		s := m.matchTrees.PeekSell()
		if s == nil {
			return false
		}
		if b.Price() >= s.Price() {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := price(b.Price(), s.Price())
				m.slab.Free(m.matchTrees.PopSell())
				b.Amount -= amount
				m.completeTrade(b, s, price, amount)
				continue
			}
			if s.Amount > b.Amount {
				amount := b.Amount
				price := price(b.Price(), s.Price())
				s.Amount -= amount
				m.completeTrade(b, s, price, amount)
				m.slab.Free(b)
				return true // The buy has been used up
			}
			if s.Amount == b.Amount {
				amount := b.Amount
				price := price(b.Price(), s.Price())
				m.completeTrade(b, s, price, amount)
				m.slab.Free(m.matchTrees.PopSell())
				m.slab.Free(b)
				return true // The buy has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func (m *M) fillableSell(s *trade.Order) bool {
	for {
		b := m.matchTrees.PeekBuy()
		if b == nil {
			return false
		}
		if b.Price() >= s.Price() {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := price(b.Price(), s.Price())
				b.Amount -= amount
				m.completeTrade(b, s, price, amount)
				m.slab.Free(s)
				return true // The sell has been used up
			}
			if s.Amount > b.Amount {
				amount := b.Amount
				price := price(b.Price(), s.Price())
				s.Amount -= amount
				m.completeTrade(b, s, price, amount)
				m.slab.Free(m.matchTrees.PopBuy())
				continue
			}
			if s.Amount == b.Amount {
				amount := b.Amount
				price := price(b.Price(), s.Price())
				m.completeTrade(b, s, price, amount)
				m.slab.Free(m.matchTrees.PopBuy())
				m.slab.Free(s)
				return true // The sell has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func price(bPrice, sPrice int64) int64 {
	if sPrice == trade.MARKET_PRICE {
		return bPrice
	}
	d := bPrice - sPrice
	return sPrice + (d >> 1)
}

func (m *M) completeTrade(b, s *trade.Order, price int64, amount uint32) {
	// TODO write the response type into these responses
	// Write the buy response
	rb := m.output.getForWrite()
	rb.Price = -price
	rb.Amount = amount
	rb.TradeId = b.TradeId()
	rb.CounterParty = s.TraderId()
	// Write the sell response
	rs := m.output.getForWrite()
	rs.Price = price
	rs.Amount = amount
	rs.TradeId = s.TradeId()
	rs.CounterParty = b.TraderId()
}
