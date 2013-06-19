package pqueue

import ()

type MatchQueues struct {
	buyTree  rbtree
	sellTree rbtree
	orders   rbtree
	size     int
}

func (m *MatchQueues) Size() int {
	return m.size
}

func (m *MatchQueues) PushBuy(b *OrderNode) {
	m.size++
	m.buyTree.push(&b.priceNode)
	m.orders.push(&b.guidNode)
}

func (m *MatchQueues) PushSell(s *OrderNode) {
	m.size++
	m.sellTree.push(&s.priceNode)
	m.orders.push(&s.guidNode)
}

func (m *MatchQueues) PeekBuy() *OrderNode {
	return m.buyTree.peekMax().getOrderNode()
}

func (m *MatchQueues) PeekSell() *OrderNode {
	return m.sellTree.peekMin().getOrderNode()
}

func (m *MatchQueues) PopBuy() *OrderNode {
	m.size--
	return m.buyTree.popMax().getOrderNode()
}

func (m *MatchQueues) PopSell() *OrderNode {
	m.size--
	return m.sellTree.popMin().getOrderNode()
}

func (m *MatchQueues) Cancel(o *OrderNode) *OrderNode {
	po := m.orders.cancel(o.Guid()).getOrderNode()
	if po != nil {
		m.size--
	}
	return po
}
