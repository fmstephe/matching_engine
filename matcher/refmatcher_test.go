package matcher

import (
	"github.com/fmstephe/matching_engine/trade"
	"github.com/fmstephe/matching_engine/tree"
)

type refmatcher struct {
	matchTrees *tree.RefMatchTrees
	submit chan interface{}
	orders chan *trade.Order
}

func newRefmatcher(lowPrice, highPrice int64, submit chan interface{}, orders chan *trade.Order) *refmatcher {
	matchTrees := tree.NewRefMatchTrees(lowPrice, highPrice)
	return &refmatcher{matchTrees: matchTrees, submit: submit, orders: orders}
}

func (m *refmatcher) Run() {
	for {
		od := <-m.orders
		o := &tree.OrderNode{}
		o.CopyFrom(od)
		if o.Kind() == trade.CANCEL {
			co := m.matchTrees.Cancel(o)
			if co != nil {
				completeCancel(m.submit, trade.CANCELLED, co)
			}
			if co == nil {
				completeCancel(m.submit, trade.NOT_CANCELLED, o)
			}
		} else {
			m.push(o)
			m.match()
		}
	}
}

func (m *refmatcher) push(o *tree.OrderNode) {
	if o.Kind() == trade.BUY {
		m.matchTrees.PushBuy(o)
		return
	}
	if o.Kind() == trade.SELL {
		m.matchTrees.PushSell(o)
		return
	}
	panic("Unsupported trade kind pushed")
}

func (m *refmatcher) match() {
	for {
		s := m.matchTrees.PeekSell()
		b := m.matchTrees.PeekBuy()
		if s == nil || b == nil {
			return
		}
		if s.Price() > b.Price() {
			return
		}
		if s.Amount() == b.Amount() {
			// pop both
			m.matchTrees.PopSell()
			m.matchTrees.PopBuy()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			completeTrade(m.submit, trade.FULL, trade.FULL, b, s, price, amount)
		}
		if s.Amount() > b.Amount() {
			// pop buy
			m.matchTrees.PopBuy()
			amount := b.Amount()
			price := price(b.Price(), s.Price())
			s.ReduceAmount(b.Amount())
			completeTrade(m.submit, trade.FULL, trade.PARTIAL, b, s, price, amount)
		}
		if b.Amount() > s.Amount() {
			// pop sell
			m.matchTrees.PopSell()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			b.ReduceAmount(s.Amount())
			completeTrade(m.submit, trade.PARTIAL, trade.FULL, b, s, price, amount)
		}
	}
}
