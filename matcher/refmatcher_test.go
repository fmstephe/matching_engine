package matcher

import (
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/prioq"
)

type refmatcher struct {
	matchQueues *prioq.RefMatchQueues
	submit      chan *msg.Message
	orders      chan *msg.Message
}

func newRefmatcher(lowPrice, highPrice int64, submit chan *msg.Message, orders chan *msg.Message) *refmatcher {
	matchQueues := prioq.NewRefMatchQueues(lowPrice, highPrice)
	return &refmatcher{matchQueues: matchQueues, submit: submit, orders: orders}
}

func (m *refmatcher) Run() {
	for {
		od := <-m.orders
		o := &prioq.OrderNode{}
		o.CopyFrom(od)
		if o.Kind() == msg.CANCEL {
			co := m.matchQueues.Cancel(o)
			if co != nil {
				completeCancelled(m.submit, co)
			}
			if co == nil {
				completeNotCancelled(m.submit, o)
			}
		} else {
			m.push(o)
			m.match()
		}
	}
}

func (m *refmatcher) push(o *prioq.OrderNode) {
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
			completeTrade(m.submit, msg.FULL, msg.FULL, b, s, price, amount)
		}
		if s.Amount() > b.Amount() {
			// pop buy
			m.matchQueues.PopBuy()
			amount := b.Amount()
			price := price(b.Price(), s.Price())
			s.ReduceAmount(b.Amount())
			completeTrade(m.submit, msg.FULL, msg.PARTIAL, b, s, price, amount)
		}
		if b.Amount() > s.Amount() {
			// pop sell
			m.matchQueues.PopSell()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			b.ReduceAmount(s.Amount())
			completeTrade(m.submit, msg.PARTIAL, msg.FULL, b, s, price, amount)
		}
	}
}
