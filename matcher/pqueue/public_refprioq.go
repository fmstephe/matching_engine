package pqueue

import ()

type RefMatchQueues struct {
	buys  *pqueue
	sells *pqueue
	size  int
}

func NewRefMatchQueues(lowPrice, highPrice uint64) *RefMatchQueues {
	buys := mkPrioq(lowPrice, highPrice)
	sells := mkPrioq(lowPrice, highPrice)
	return &RefMatchQueues{buys: buys, sells: sells}
}

func (m *RefMatchQueues) Size() int {
	return m.size
}

func (m *RefMatchQueues) PushBuy(b *OrderNode) {
	m.size++
	m.buys.push(b)
}

func (m *RefMatchQueues) PushSell(s *OrderNode) {
	m.size++
	m.sells.push(s)
}

func (m *RefMatchQueues) PeekBuy() *OrderNode {
	return m.buys.peekMax()
}

func (m *RefMatchQueues) PeekSell() *OrderNode {
	return m.sells.peekMin()
}

func (m *RefMatchQueues) PopBuy() *OrderNode {
	m.size--
	return m.buys.popMax()
}

func (m *RefMatchQueues) PopSell() *OrderNode {
	m.size--
	return m.sells.popMin()
}

func (m *RefMatchQueues) Cancel(o *OrderNode) *OrderNode {
	c := m.buys.cancel(o.Guid())
	if c != nil {
		m.size--
		return c
	}
	c = m.sells.cancel(o.Guid())
	if c != nil {
		m.size--
		return c
	}
	return nil
}
