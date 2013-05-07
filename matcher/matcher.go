package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/cbuf"
	"github.com/fmstephe/matching_engine/trade"
)

type M struct {
	matchTrees trade.MatchTrees // No constructor required
	slab       *trade.Slab
	rb         *cbuf.Response
}

func NewMatcher(slabSize int, rb *cbuf.Response) *M {
	slab := trade.NewSlab(slabSize)
	return &M{slab: slab, rb: rb}
}

func (m *M) Submit(od *trade.OrderData) {
	o := m.slab.Malloc()
	o.CopyFrom(od)
	switch o.Kind() {
	case trade.BUY:
		m.addBuy(o)
	case trade.SELL:
		m.addSell(o)
	case trade.CANCEL:
		m.cancel(o)
	default:
		panic(fmt.Sprintf("OrderKind %s not supported", o.Kind().String()))
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

func (m *M) cancel(o *trade.Order) {
	ro := m.matchTrees.Cancel(o)
	if ro != nil {
		completeCancel(m.rb, trade.CANCELLED, ro)
		m.slab.Free(ro)
	} else {
		completeCancel(m.rb, trade.NOT_CANCELLED, o)
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
			if b.Amount() > s.Amount() {
				amount := s.Amount()
				price := price(b.Price(), s.Price())
				m.slab.Free(m.matchTrees.PopSell())
				b.ReduceAmount(amount)
				completeTrade(m.rb, trade.PARTIAL, trade.FULL, b, s, price, amount)
				continue
			}
			if s.Amount() > b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				s.ReduceAmount(amount)
				completeTrade(m.rb, trade.FULL, trade.PARTIAL, b, s, price, amount)
				m.slab.Free(b)
				return true // The buy has been used up
			}
			if s.Amount() == b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				completeTrade(m.rb, trade.FULL, trade.FULL, b, s, price, amount)
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
			if b.Amount() > s.Amount() {
				amount := s.Amount()
				price := price(b.Price(), s.Price())
				b.ReduceAmount(amount)
				completeTrade(m.rb, trade.PARTIAL, trade.FULL, b, s, price, amount)
				m.slab.Free(s)
				return true // The sell has been used up
			}
			if s.Amount() > b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				s.ReduceAmount(amount)
				completeTrade(m.rb, trade.PARTIAL, trade.FULL, b, s, price, amount)
				m.slab.Free(m.matchTrees.PopBuy())
				continue
			}
			if s.Amount() == b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				completeTrade(m.rb, trade.FULL, trade.FULL, b, s, price, amount)
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
	return sPrice + (d / 2)
}

func completeTrade(rb *cbuf.Response, brk, srk trade.ResponseKind, b, s *trade.Order, price int64, amount uint32) {
	br, berr := rb.GetForWrite()
	if berr != nil {
		panic(berr.Error())
	}
	sr, serr := rb.GetForWrite()
	if serr != nil {
		panic(serr.Error())
	}
	br.WriteTrade(brk, -price, amount, b.TraderId(), b.TradeId(), s.TraderId())
	sr.WriteTrade(srk, price, amount, s.TraderId(), s.TradeId(), b.TraderId())
}

func completeCancel(rb *cbuf.Response, rk trade.ResponseKind, d *trade.Order) {
	r, err := rb.GetForWrite()
	if err != nil {
		panic(err.Error())
	}
	r.WriteCancel(rk, d.TraderId(), d.TradeId())
}
