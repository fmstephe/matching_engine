package matcher

import (
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/matcher/pqueue"
	"github.com/fmstephe/matching_engine/msg"
)

type refmatcher struct {
	matchQueues *pqueue.RefMatchQueues
	coordinator.AppMsgHelper
}

func newRefmatcher(lowPrice, highPrice int64) *refmatcher {
	matchQueues := pqueue.NewRefMatchQueues(lowPrice, highPrice)
	return &refmatcher{matchQueues: matchQueues}
}

func (rm *refmatcher) Run() {
	for {
		m := <-rm.In
		if m.Route == msg.SHUTDOWN {
			return
		}
		if m != nil {
			o := &pqueue.OrderNode{}
			o.CopyFrom(m)
			if o.Kind() == msg.CANCEL {
				co := rm.matchQueues.Cancel(o)
				if co != nil {
					completeCancelled(rm.Out, co)
				}
				if co == nil {
					completeNotCancelled(rm.Out, o)
				}
			} else {
				rm.push(o)
				rm.match()
			}
		}
	}
}

func (rm *refmatcher) push(o *pqueue.OrderNode) {
	if o.Kind() == msg.BUY {
		rm.matchQueues.PushBuy(o)
		return
	}
	if o.Kind() == msg.SELL {
		rm.matchQueues.PushSell(o)
		return
	}
	panic("Unsupported trade kind pushed")
}

func (rm *refmatcher) match() {
	for {
		s := rm.matchQueues.PeekSell()
		b := rm.matchQueues.PeekBuy()
		if s == nil || b == nil {
			return
		}
		if s.Price() > b.Price() {
			return
		}
		if s.Amount() == b.Amount() {
			// pop both
			rm.matchQueues.PopSell()
			rm.matchQueues.PopBuy()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			completeTrade(rm.Out, msg.FULL, msg.FULL, b, s, price, amount)
		}
		if s.Amount() > b.Amount() {
			// pop buy
			rm.matchQueues.PopBuy()
			amount := b.Amount()
			price := price(b.Price(), s.Price())
			s.ReduceAmount(b.Amount())
			completeTrade(rm.Out, msg.FULL, msg.PARTIAL, b, s, price, amount)
		}
		if b.Amount() > s.Amount() {
			// pop sell
			rm.matchQueues.PopSell()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			b.ReduceAmount(s.Amount())
			completeTrade(rm.Out, msg.PARTIAL, msg.FULL, b, s, price, amount)
		}
	}
}
