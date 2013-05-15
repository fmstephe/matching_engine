package matcher

import (
	"fmt"
	"github.com/fmstephe/matching_engine/trade"
	"github.com/fmstephe/matching_engine/tree"
)

type M struct {
	matchTrees tree.MatchTrees // No constructor required
	slab       *tree.Slab
	submit     chan interface{}
	orders     chan *trade.Order
}

func NewMatcher(slabSize int) *M {
	slab := tree.NewSlab(slabSize)
	return &M{slab: slab}
}

func (m *M) SetSubmit(submit chan interface{}) {
	m.submit = submit
}

func (m *M) SetOrderNodes(orders chan *trade.Order) {
	m.orders = orders
}

func (m *M) Run() {
	for {
		od := <-m.orders
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
			// This should probably just be an message to m.submit
			panic(fmt.Sprintf("OrderNodeKind %s not supported", o.Kind().String()))
		}
	}
}

func (m *M) addBuy(b *tree.OrderNode) {
	if b.Price() == trade.MARKET_PRICE {
		// This should probably just be a message to m.submit
		panic("It is illegal to submit a buy at market price")
	}
	if !m.fillableBuy(b) {
		m.matchTrees.PushBuy(b)
	}
}

func (m *M) addSell(s *tree.OrderNode) {
	if !m.fillableSell(s) {
		m.matchTrees.PushSell(s)
	}
}

func (m *M) cancel(o *tree.OrderNode) {
	ro := m.matchTrees.Cancel(o)
	if ro != nil {
		completeCancel(m.submit, trade.CANCELLED, ro)
		m.slab.Free(ro)
	} else {
		completeCancel(m.submit, trade.NOT_CANCELLED, o)
	}
	m.slab.Free(o)
}

func (m *M) fillableBuy(b *tree.OrderNode) bool {
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
				completeTrade(m.submit, trade.PARTIAL, trade.FULL, b, s, price, amount)
				continue
			}
			if s.Amount() > b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				s.ReduceAmount(amount)
				completeTrade(m.submit, trade.FULL, trade.PARTIAL, b, s, price, amount)
				m.slab.Free(b)
				return true // The buy has been used up
			}
			if s.Amount() == b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				completeTrade(m.submit, trade.FULL, trade.FULL, b, s, price, amount)
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

func (m *M) fillableSell(s *tree.OrderNode) bool {
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
				completeTrade(m.submit, trade.PARTIAL, trade.FULL, b, s, price, amount)
				m.slab.Free(s)
				return true // The sell has been used up
			}
			if s.Amount() > b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				s.ReduceAmount(amount)
				completeTrade(m.submit, trade.PARTIAL, trade.FULL, b, s, price, amount)
				m.slab.Free(m.matchTrees.PopBuy())
				continue
			}
			if s.Amount() == b.Amount() {
				amount := b.Amount()
				price := price(b.Price(), s.Price())
				completeTrade(m.submit, trade.FULL, trade.FULL, b, s, price, amount)
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

func completeTrade(submit chan interface{}, brk, srk trade.ResponseKind, b, s *tree.OrderNode, price int64, amount uint32) {
	br := &trade.Response{}
	sr := &trade.Response{}
	br.WriteTrade(brk, -price, amount, b.TraderId(), b.TradeId(), s.TraderId())
	sr.WriteTrade(srk, price, amount, s.TraderId(), s.TradeId(), b.TraderId())
	submit <- br
	submit <- sr
}

func completeCancel(submit chan interface{}, rk trade.ResponseKind, d *tree.OrderNode) {
	r := &trade.Response{}
	r.WriteCancel(rk, d.TraderId(), d.TradeId())
	submit <- r
}
