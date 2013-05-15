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

func (m *MatchTrees) PushBuy(b *Order) {
	m.size++
	m.buyTree.push(&b.priceNode)
	m.orders.push(&b.guidNode)
}

func (m *MatchTrees) PushSell(s *Order) {
	m.size++
	m.sellTree.push(&s.priceNode)
	m.orders.push(&s.guidNode)
}

func (m *MatchTrees) PeekBuy() *Order {
	return m.buyTree.peekMax().getOrder()
}

func (m *MatchTrees) PeekSell() *Order {
	return m.sellTree.peekMin().getOrder()
}

func (m *MatchTrees) PopBuy() *Order {
	m.size--
	return m.buyTree.popMax().getOrder()
}

func (m *MatchTrees) PopSell() *Order {
	m.size--
	return m.sellTree.popMin().getOrder()
}

func (m *MatchTrees) Cancel(o *Order) *Order {
	po := m.orders.cancel(o.Guid()).getOrder()
	if po != nil {
		m.size--
	}
	return po
}
