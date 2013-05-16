package tree

import ()

type RefMatchTrees struct {
	buys  *prioq
	sells *prioq
	size     int
}

func NewRefMatchTrees(lowPrice, highPrice int64) *RefMatchTrees {
	buys := mkPrioq(lowPrice, highPrice)
	sells := mkPrioq(lowPrice, highPrice)
	return &RefMatchTrees{buys: buys, sells: sells}
}

func (m *RefMatchTrees) Size() int {
	return m.size
}

func (m *RefMatchTrees) PushBuy(b *OrderNode) {
	m.size++
	m.buys.push(b)
}

func (m *RefMatchTrees) PushSell(s *OrderNode) {
	m.size++
	m.sells.push(s)
}

func (m *RefMatchTrees) PeekBuy() *OrderNode {
	return m.buys.peekMax()
}

func (m *RefMatchTrees) PeekSell() *OrderNode {
	return m.sells.peekMin()
}

func (m *RefMatchTrees) PopBuy() *OrderNode {
	m.size--
	return m.buys.popMax()
}

func (m *RefMatchTrees) PopSell() *OrderNode {
	m.size--
	return m.sells.popMin()
}

func (m *RefMatchTrees) Cancel(o *OrderNode) *OrderNode {
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
