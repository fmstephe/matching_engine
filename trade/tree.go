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
	return m.sellTree.popMin().getOrder()
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
				in.parent = n
				in.pp = &n.left
				n.left = in
				return
			} else {
				n = n.left
			}
		case in.val > n.val:
			if n.right == nil {
				n.parent = n
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
	black bool
	// Tree fields
	val    int64
	left   *node
	right  *node
	parent *node
	pp     **node
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
		println("left")
		return false
	case n.right != nil:
		println("right")
		return false
	case n.pp != nil:
		println("pp")
		return false
	case n.next != n:
		println("next")
		return false
	case n.prev != n:
		println("prev")
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

func (n *node) giveParent(nn *node) {
	nn.parent = n.parent
	nn.pp = n.pp
	*nn.pp = nn
	n.parent = nil
	n.pp = nil
}

func (n *node) giveChildren(nn *node) {
	nn.left = n.left
	nn.right = n.right
	if nn.left != nil {
		nn.left.parent = nn
		nn.left.pp = &nn.left
	}
	if nn.right != nil {
		nn.right.parent = nn
		nn.right.pp = &nn.right
	}
	n.left = nil
	n.right = nil
}

func (n *node) givePosition(nn *node) {
	n.giveParent(nn)
	n.giveChildren(nn)
	// Guarantee: Each of n.parent/pp/left/right are now nil
}

func (n *node) detach() {
	switch {
	case n.right == nil && n.left == nil:
		*n.pp = nil
		n.pp = nil
		n.parent = nil
	case n.right == nil:
		n.giveParent(n.left)
		n.left = nil
	case n.left == nil:
		n.giveParent(n.right)
		n.right = nil
	default:
		nn := n.left.detachMax()
		n.givePosition(nn)
	}
	// Guarantee: Each of n.parent/pp/left/right are now nil
}

func (n *node) pop() {
	switch {
	case !n.isHead():
		n.prev.next = n.next
		n.next.prev = n.prev
		n.parent = nil
		n.pp = nil
		n.left = nil
		n.right = nil
	case n.next != n:
		n.prev.next = n.next
		n.next.prev = n.prev
		nn := n.prev
		n.givePosition(nn)
	default:
		n.detach()
	}
	n.next = n
	n.prev = n
	// Guarantee: Each of n.parent/pp/left/right are now nil
	// Guarantee: Both n.left/right point to n
}

func (n *node) detachMax() *node {
	m := n
	for {
		if m.right == nil {
			break
		}
		m = m.right
	}
	m.detach()
	return m
}

func (n *node) toRight(to *node) {
	to.right = n
	if n != nil {
		*n.pp = nil
		n.parent = to
		n.pp = &to.right
	}
}

func (n *node) toLeft(to *node) {
	to.left = n
	if n != nil {
		*n.pp = nil
		n.parent = to
		n.pp = &to.left
	}
}

func (n *node) rotateLeft() {
	r := n.right
	n.giveParent(r)
	r.left.toRight(n)
	n.toLeft(r)
	r.black = n.black
	n.black = false
}

func (n *node) rotateRight() {
	l := n.left
	n.giveParent(l)
	l.right.toLeft(n)
	n.toRight(l)
	l.black = n.black
	n.black = false
}

func (n *node) flip() {
	n.black = !n.black
	n.left.black = !n.left.black
	n.right.black = !n.right.black
}

func (n *node) moveRedLeft() {
	n.flip()
	if n.right.left.isRed() {
		n.right.rotateRight()
		n.rotateLeft()
		n.parent.flip()
	}
}

func (n *node) moveRedRight() {
	n.flip()
	if n.left.left.isRed() {
		n.rotateRight()
		n.parent.flip()
	}
}

func (n *node) fixup() {
	p := n
	if p.isRed() {
		p.rotateLeft()
		p = p.parent
	}
	if p.left.isRed() && p.left.left.isRed() {
		p.rotateRight()
		p = p.parent
	}
	if p.left.isRed() && p.right.isRed() {
		p.flip()
	}
}

func (n *node) isRed() bool {
	if n != nil {
		return !n.black
	}
	return false
}
