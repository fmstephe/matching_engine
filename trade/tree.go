package trade

import (
	"bytes"
	"strconv"
)

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
		b.root.black = true
		return
	}
	n := b.root
	for {
		switch {
		case in.val == n.val:
			n.addLast(in)
			goto repair
		case in.val < n.val:
			if n.left == nil {
				in.toLeftOf(n)
				goto repair
			} else {
				n = n.left
			}
		case in.val > n.val:
			if n.right == nil {
				in.toRightOf(n)
				goto repair
			} else {
				n = n.right
			}
		}
	}
repair:
	for n != nil {
		if n.right.isRed() && !n.left.isRed() {
			n = n.rotateLeft()
		}
		if n.left.isRed() && n.left.left.isRed() {
			n = n.rotateRight()
		}
		if n.left.isRed() && n.right.isRed() {
			n.flip()
		}
		n = n.parent
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
	// Order
	order *Order
	// This is the other node tying order to another tree
	other *node
}

func (n *node) String() string {
	if n == nil {
		return "()"
	}
	valStr := strconv.Itoa(int(n.val))
	colour := "R"
	if n.black {
		colour = "B"
	}
	b := bytes.NewBufferString("")
	b.WriteString("(")
	b.WriteString(valStr)
	b.WriteString(colour)
	if !(n.left == nil && n.right == nil) {
		b.WriteString(", ")
		b.WriteString(n.left.String())
		b.WriteString(", ")
		b.WriteString(n.right.String())
	}
	b.WriteString(")")
	return b.String()
}

func initNode(o *Order, val int64, n, other *node) {
	*n = node{val: val, order: o, other: other}
	n.next = n
	n.prev = n
	n.black = false
}

func (n *node) getOrder() *Order {
	if n != nil {
		return n.order
	}
	return nil
}

func (n *node) isRed() bool {
	if n != nil {
		return !n.black
	}
	return false
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

func (n *node) getSibling() *node {
	p := n.parent
	if p == nil {
		return nil
	}
	if p.left == n {
		return p.right
	}
	return p.left
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
	nn.black = n.black
	// Guarantee: Each of n.parent/pp/left/right are now nil
}

func (n *node) detach() {
	p := n.parent
	s := n.getSibling()
	var nn *node
	switch {
	case n.right == nil && n.left == nil:
		*n.pp = nil
		n.pp = nil
		n.parent = nil
	case n.right == nil:
		nn = n.left
		n.giveParent(nn)
		n.left = nil
	case n.left == nil:
		nn = n.right
		n.giveParent(nn)
		n.right = nil
	default:
		nn = n.left.detachMax()
		n.givePosition(nn)
		return
	}
	// Guarantee: Each of n.parent/pp/left/right are now nil
	if n.isRed() {
		return
	}
	if nn.isRed() {
		nn.black = true
		return
	}
	if s.isRed() {
		// Perform a rotation to make sibling black
		if p.left == s {
			p.rotateRight()
			s = p.left
		} else {
			p.rotateLeft()
			s = p.right
		}
	}
	// repair
	for p != nil {
		if s == nil {
			for p != nil {
				p = p.fixPop()
				p = p.parent
			}
			return
		}
		pRed := p.isRed()
		sRed := s.isRed()
		slRed := s.left.isRed()
		if !sRed && !slRed && pRed {
			p.black = true
			s.black = false
			for p != nil {
				p = p.fixPop()
				p = p.parent
			}
			return
		}
		if !sRed && !slRed && !pRed {
			// Introduce black violation
			s.black = false
		} else if !sRed && slRed {
			if p.left == s {
				p = p.rotateRight()
			} else {
				s.rotateRight()
				p = p.rotateLeft()
			}
			for p != nil {
				p = p.fixPop()
				p = p.parent
			}
			return
		}
		p = p.fixPop()
		s = p.getSibling()
		p = p.parent
	}
}

func (n *node) fixPop() *node {
	if n.right.isRed() && !n.left.isRed() {
		n = n.rotateLeft()
	}
	if n.left.isRed() && n.left.left.isRed() {
		n = n.rotateRight()
	}
	if n.left.isRed() && n.right.isRed() {
		n.flip()
	}
	return n
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

func (n *node) toRightOf(to *node) {
	to.right = n
	if n != nil {
		n.parent = to
		n.pp = &to.right
	}
}

func (n *node) toLeftOf(to *node) {
	to.left = n
	if n != nil {
		n.parent = to
		n.pp = &to.left
	}
}

func (n *node) rotateLeft() *node {
	r := n.right
	n.giveParent(r)
	r.left.toRightOf(n)
	n.toLeftOf(r)
	r.black = n.black
	n.black = false
	return r
}

func (n *node) rotateRight() *node {
	l := n.left
	n.giveParent(l)
	l.right.toLeftOf(n)
	n.toRightOf(l)
	l.black = n.black
	n.black = false
	return l
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
