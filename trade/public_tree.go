package trade

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

func (m *MatchTrees) Pop(o *Order) *Order {
	po := m.orders.pop(o.Guid()).getOrder()
	if po != nil {
		m.size--
	}
	return po
}

type PriceTree struct {
	t tree
}

func (p *PriceTree) Push(o *Order) {
	p.t.push(&o.priceNode)
}

func (p *PriceTree) Pop(o *Order) *Order {
	return p.t.pop(o.Guid()).getOrder()
}

func (p *PriceTree) peekMax() *Order {
	return p.t.peekMax().getOrder()
}

func (p *PriceTree) PopMax() *Order {
	return p.t.popMax().getOrder()
}

func (p *PriceTree) PeekMin() *Order {
	return p.t.peekMin().getOrder()
}

func (p *PriceTree) PopMin() *Order {
	return p.t.popMin().getOrder()
}
