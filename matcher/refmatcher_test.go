package matcher

import (
	"github.com/fmstephe/matching_engine/trade"
	"github.com/fmstephe/matching_engine/tree"
)

type refmatcher struct {
	buys   *prioq
	sells  *prioq
	submit chan interface{}
	orders chan *trade.Order
}

func newRefmatcher(lowPrice, highPrice int64, submit chan interface{}, orders chan *trade.Order) *refmatcher {
	buys := newPrioq(lowPrice, highPrice)
	sells := newPrioq(lowPrice, highPrice)
	return &refmatcher{buys: buys, sells: sells, submit: submit, orders: orders}
}

func (m *refmatcher) Run() {
	for {
		od := <-m.orders
		o := &tree.OrderNode{}
		o.CopyFrom(od)
		if o.Kind() == trade.CANCEL {
			co := m.pop(o)
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

func (m *refmatcher) match() {
	for {
		s := m.peekSell()
		b := m.peekBuy()
		if s == nil || b == nil {
			return
		}
		if s.Price() > b.Price() {
			return
		}
		if s.Amount() == b.Amount() {
			// pop both
			m.popSell()
			m.popBuy()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			completeTrade(m.submit, trade.FULL, trade.FULL, b, s, price, amount)
		}
		if s.Amount() > b.Amount() {
			// pop buy
			m.popBuy()
			amount := b.Amount()
			price := price(b.Price(), s.Price())
			s.ReduceAmount(b.Amount())
			completeTrade(m.submit, trade.FULL, trade.PARTIAL, b, s, price, amount)
		}
		if b.Amount() > s.Amount() {
			// pop sell
			m.popSell()
			amount := s.Amount()
			price := price(b.Price(), s.Price())
			b.ReduceAmount(s.Amount())
			completeTrade(m.submit, trade.PARTIAL, trade.FULL, b, s, price, amount)
		}
	}
}

func (m *refmatcher) Size() int {
	return -1
}

func (m *refmatcher) push(o *tree.OrderNode) {
	if o.Kind() == trade.BUY {
		m.buys.push(o)
		return
	}
	if o.Kind() == trade.SELL {
		m.sells.push(o)
		return
	}
	panic("Unsupported trade kind pushed")
}

func (m *refmatcher) peekBuy() *tree.OrderNode {
	return m.buys.peekMax()
}

func (m *refmatcher) peekSell() *tree.OrderNode {
	return m.sells.peekMin()
}

func (m *refmatcher) popBuy() *tree.OrderNode {
	return m.buys.popMax()
}

func (m *refmatcher) popSell() *tree.OrderNode {
	return m.sells.popMin()
}

func (m *refmatcher) pop(o *tree.OrderNode) *tree.OrderNode {
	guid := o.Guid()
	ro := m.buys.remove(guid)
	if ro == nil {
		return m.sells.remove(guid)
	}
	return ro
}

// An easy to build priority queue
type prioq struct {
	prios               [][]*tree.OrderNode
	lowPrice, highPrice int64
}

func newPrioq(lowPrice, highPrice int64) *prioq {
	prios := make([][]*tree.OrderNode, highPrice-lowPrice+1)
	return &prioq{prios: prios, lowPrice: lowPrice, highPrice: highPrice}
}

func (q *prioq) push(o *tree.OrderNode) {
	idx := o.Price() - q.lowPrice
	prio := q.prios[idx]
	prio = append(prio, o)
	q.prios[idx] = prio
}

func (q *prioq) peekMax() *tree.OrderNode {
	if len(q.prios) == 0 {
		return nil
	}
	for i := len(q.prios) - 1; i >= 0; i-- {
		switch {
		case len(q.prios[i]) > 0:
			return q.prios[i][0]
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) popMax() *tree.OrderNode {
	if len(q.prios) == 0 {
		return nil
	}
	for i := len(q.prios) - 1; i >= 0; i-- {
		switch {
		case len(q.prios[i]) > 0:
			return q.pop(i)
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) peekMin() *tree.OrderNode {
	if len(q.prios) == 0 {
		return nil
	}
	for i := 0; i < len(q.prios); i++ {
		switch {
		case len(q.prios[i]) > 0:
			return q.prios[i][0]
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) popMin() *tree.OrderNode {
	if len(q.prios) == 0 {
		return nil
	}
	for i := 0; i < len(q.prios); i++ {
		switch {
		case len(q.prios[i]) > 0:
			return q.pop(i)
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) pop(price int) *tree.OrderNode {
	prio := q.prios[price]
	o := prio[0]
	prio = prio[1:]
	q.prios[price] = prio
	return o
}

func (q *prioq) remove(guid int64) *tree.OrderNode {
	for i := range q.prios {
		priceQ := q.prios[i]
		for j := range priceQ {
			o := priceQ[j]
			if o.Guid() == guid {
				priceQ = append(priceQ[0:j], priceQ[j+1:]...)
				q.prios[i] = priceQ
				return o
			}
		}
	}
	return nil
}
