package trade

import ()

type MatchTrees struct {
	buyTree  tree
	sellTree tree
	orders   tree
}

func (m *MatchTrees) PushBuy(b *Order) {
	m.buyTree.push(&b.priceNode)
	m.orders.push(&b.guidNode)
}

func (m *MatchTrees) PushSell(s *Order) {
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
	return m.buyTree.popMax().getOrder()
}

func (m *MatchTrees) PopSell() *Order {
	return m.sellTree.popMax().getOrder()
}

func (m *MatchTrees) Pop(o *Order) *Order {
	return m.orders.pop(o.Guid()).getOrder()
}

type tree struct {
	root *node
}

func (b *tree) push(in *node) {
	if b.root == nil {
		b.root = in
		in.pp = &b.root
		return
	}
	n := b.root
	for {
		switch {
		case in.val == n.val:
			n.addLast(in)
			return
		case in.val < n.val:
			if n.left == nil {
				in.pp = &n.left
				n.left = in
				return
			} else {
				n = n.left
			}
		case in.val > n.val:
			if n.right == nil {
				in.pp = &n.right
				n.right = in
				return
			} else {
				n = n.right
			}
		}
	}
}

func (b *tree) peekMin() *node {
	n := b.root
	if n == nil {
		return nil
	}
	for n.left != nil {
		n = n.left
	}
	return n
}

func (b *tree) popMin() *node {
	if b.root != nil {
		n := b.peekMin()
		n.pop()
		n.other.pop() // Clear complementary tree
		return n
	}
	return nil
}

func (b *tree) peekMax() *node {
	n := b.root
	if n == nil {
		return nil
	}
	for n.right != nil {
		n = n.right
	}
	return n
}

func (b *tree) popMax() *node {
	if b.root != nil {
		n := b.peekMax()
		n.pop()
		n.other.pop() // Clear complementary tree
		return n
	}
	return nil
}

func (b *tree) pop(val int64) *node {
	n := b.get(val)
	if n == nil {
		return nil
	}
	n.pop()
	n.other.pop()
	return n
}

func (b *tree) Has(val int64) bool {
	return b.get(val) != nil
}

func (b *tree) get(val int64) *node {
	n := b.root
	for {
		if n == nil {
			return nil
		}
		if val == n.val {
			return n
		}
		if val < n.val {
			n = n.left
		} else {
			n = n.right
		}
	}
	panic("Unreachable")
}

type node struct {
	// Tree fields
	val   int64
	left  *node
	right *node
	pp    **node
	// Limit queue fields
	next *node
	prev *node
	// This is the other node attaching O to another tree
	other *node
	// Order
	order *Order
}

func initNode(o *Order, val int64, n, other *node) {
	*n = node{val: val, order: o, other: other}
	n.next = n
	n.prev = n
}

func (n *node) getOrder() *Order {
	if n != nil {
		return n.order
	}
	return nil
}

func (n *node) isFree() bool {
	switch {
	case n.left != nil:
		return false
	case n.right != nil:
		return false
	case n.pp != nil:
		return false
	case n.next != n:
		return false
	case n.prev != n:
		return false
	}
	return true
}

func (n *node) isHead() bool {
	return n.pp != nil
}

func (n *node) addLast(in *node) {
	last := n.next
	last.prev = in
	in.next = last
	in.prev = n
	n.next = in
}

func (n *node) pop() {
	switch {
	case !n.isHead():
		n.prev.next = n.next
		n.next.prev = n.prev
	case n.next != n:
		n.prev.next = n.next
		n.next.prev = n.prev
		nn := n.prev
		n.next = nil
		n.prev = nil
		n.swapWith(nn)
	default:
		n.detachAll()
	}
	n.next = n
	n.prev = n
	n.pp = nil
	n.left = nil
	n.right = nil
}

func (n *node) swapWith(nn *node) {
	nn.pp = n.pp
	*nn.pp = nn
	nn.left = n.left
	nn.right = n.right
	if nn.left != nil {
		nn.left.pp = &nn.left
	}
	if nn.right != nil {
		nn.right.pp = &nn.right
	}
}

func (n *node) detachAll() {
	switch {
	case n.right == nil && n.left == nil:
		*n.pp = nil
	case n.right == nil:
		*n.pp = n.left
		n.left.pp = n.pp
	case n.left == nil:
		*n.pp = n.right
		n.right.pp = n.pp
	default:
		nn := n.left.detachMax()
		n.swapWith(nn)
	}
	n.pp = nil
	n.left = nil
	n.right = nil
}

func (n *node) detachMax() *node {
	m := n
	for {
		if m.right == nil {
			break
		}
		m = m.right
	}
	m.detachAll()
	return m
}
