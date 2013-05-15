package tree

import ()

type MatchTrees struct {
	buyTree  tree
	sellTree tree
	orders   tree
	size     int
}

func (m *MatchTrees) Size() int {
	return m.size
}

func (m *MatchTrees) PushBuy(b *OrderNode) {
	m.size++
	m.buyTree.push(&b.priceNode)
	m.orders.push(&b.guidNode)
}

func (m *MatchTrees) PushSell(s *OrderNode) {
	m.size++
	m.sellTree.push(&s.priceNode)
	m.orders.push(&s.guidNode)
}

func (m *MatchTrees) PeekBuy() *OrderNode {
	return m.buyTree.peekMax().getOrderNode()
}

func (m *MatchTrees) PeekSell() *OrderNode {
	return m.sellTree.peekMin().getOrderNode()
}

func (m *MatchTrees) PopBuy() *OrderNode {
	m.size--
	return m.buyTree.popMax().getOrderNode()
}

func (m *MatchTrees) PopSell() *OrderNode {
	m.size--
	return m.sellTree.popMin().getOrderNode()
}

func (m *MatchTrees) Cancel(o *OrderNode) *OrderNode {
	po := m.orders.cancel(o.Guid()).getOrderNode()
	if po != nil {
		m.size--
	}
	return po
}
