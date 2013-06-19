package matcher

import (
	"github.com/fmstephe/matching_engine/matcher/pqueue"
	"github.com/fmstephe/matching_engine/msg"
)

type refmatcher struct {
	matchQueues *pqueue.RefMatchQueues
	dispatch    chan *msg.Message
	orders      chan *msg.Message
}

func newRefmatcher(lowPrice, highPrice int64, dispatch chan *msg.Message, orders chan *msg.Message) *refmatcher {
	matchQueues := pqueue.NewRefMatchQueues(lowPrice, highPrice)
	return &refmatcher{matchQueues: matchQueues, dispatch: dispatch, orders: orders}
}

func (m *refmatcher) Run() {
	for {
		od := <-m.orders
		o := &pqueue.OrderNode{}
		o.CopyFrom(od)
		if o.Kind() == msg.CANCEL {
			co := m.matchQueues.Cancel(o)
			if co != nil {
				completeCancelled(m.dispatch, co)
			}
			if co == nil {
				completeNotCancelled(m.dispatch, o)
			}
		} else {
			m.push(o)
			m.match()
		}
	}
}

func (m *refmatcher) push(o *pqueue.OrderNode) {
	if o.Kind() == msg.BUY {
		m.matchQueues.PushBuy(o)
		return
	}
	if o.Kind() == msg.SELL {
		m.matchQueues.PushSell(o)
		return
	}
	panic("Unsupported trade kind pushed")
}

func (m *refmatcher) match() {
	for {
		s := m.matchQueues.PeekSell()
		b := m.matchQueues.PeekBuy()
		if s == nil || b == nil {
			return
		}
		if s.Price() > b.Price() {
			return
		}
		if s.Amount() == b.Amount() {
			// pop both
			m.matchQueues.PopSell()
			m.matchQueues.PopBuy()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			completeTrade(m.dispatch, msg.FULL, msg.FULL, b, s, price, amount)
		}
		if s.Amount() > b.Amount() {
			// pop buy
			m.matchQueues.PopBuy()
			amount := b.Amount()
			price := price(b.Price(), s.Price())
			s.ReduceAmount(b.Amount())
			completeTrade(m.dispatch, msg.FULL, msg.PARTIAL, b, s, price, amount)
		}
		if b.Amount() > s.Amount() {
			// pop sell
			m.matchQueues.PopSell()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			b.ReduceAmount(s.Amount())
			completeTrade(m.dispatch, msg.PARTIAL, msg.FULL, b, s, price, amount)
		}
	}
}
