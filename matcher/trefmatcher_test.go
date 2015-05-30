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

func newRefmatcher(lowPrice, highPrice uint64) *refmatcher {
	matchQueues := pqueue.NewRefMatchQueues(lowPrice, highPrice)
	return &refmatcher{matchQueues: matchQueues}
}

func (rm *refmatcher) Run() {
	m := &msg.Message{}
	for {
		*m = rm.In.Read()
		if m.Kind == msg.SHUTDOWN {
			rm.Out.Write(*m)
			return
		}
		if m != nil {
			o := &pqueue.OrderNode{}
			o.CopyFrom(m)
			if o.Kind() == msg.CANCEL {
				co := rm.matchQueues.Cancel(o)
				if co != nil {
					rm.completeCancelled(co)
				}
				if co == nil {
					rm.completeNotCancelled(o)
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
			rm.completeTrade(msg.FULL, msg.FULL, b, s, price, amount)
		}
		if s.Amount() > b.Amount() {
			// pop buy
			rm.matchQueues.PopBuy()
			amount := b.Amount()
			price := price(b.Price(), s.Price())
			s.ReduceAmount(b.Amount())
			rm.completeTrade(msg.FULL, msg.PARTIAL, b, s, price, amount)
		}
		if b.Amount() > s.Amount() {
			// pop sell
			rm.matchQueues.PopSell()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			b.ReduceAmount(s.Amount())
			rm.completeTrade(msg.PARTIAL, msg.FULL, b, s, price, amount)
		}
	}
}

func (rm *refmatcher) completeTrade(brk, srk msg.MsgKind, b, s *pqueue.OrderNode, price, amount uint64) {
	rm.Out.Write(msg.Message{Kind: brk, Price: price, Amount: amount, TraderId: b.TraderId(), TradeId: b.TradeId(), StockId: b.StockId()})
	rm.Out.Write(msg.Message{Kind: srk, Price: price, Amount: amount, TraderId: s.TraderId(), TradeId: s.TradeId(), StockId: s.StockId()})
}

func (rm *refmatcher) completeCancelled(c *pqueue.OrderNode) {
	cm := msg.Message{}
	c.CopyTo(&cm)
	cm.Kind = msg.CANCELLED
	rm.Out.Write(cm)
}

func (rm *refmatcher) completeNotCancelled(nc *pqueue.OrderNode) {
	ncm := msg.Message{}
	nc.CopyTo(&ncm)
	ncm.Kind = msg.NOT_CANCELLED
	rm.Out.Write(ncm)
}
