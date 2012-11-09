package trade

import ()

type Tree struct {
	root *Node
}

func NewTree() *Tree {
	return &Tree{}
}

func (b *Tree) Push(n *Node) {
	if b.root != nil {
		b.root.push(n)
	} else {
		b.root = n
		n.pp = &b.root
	}
}

func (b *Tree) PeekMin() *Node {
	return b.root.peekMin()
}

func (b *Tree) PopMin() *Node {
	if b.root != nil {
		n := b.root.popMin()
		n.other.pop() // Clear complementary tree
		return n
	}
	return nil
}

func (b *Tree) PeekMax() *Node {
	return b.root.peekMax()
}

func (b *Tree) PopMax() *Node {
	if b.root != nil {
		n := b.root.popMax()
		n.other.pop() // Clear complementary tree
		return n
	}
	return nil
}

func (b *Tree) Pop(val int64) *Node {
	n := b.root.popVal(val)
	if n != nil {
		n.other.pop()
	}
	return n
}

func (b *Tree) Has(val int64) bool {
	return b.root.has(val)
}

type Node struct {
	// Tree fields
	val   int64
	pp    **Node
	left  *Node
	right *Node
	// Limit queue fields
	next *Node
	prev *Node
	// Order
	O *Order
	// This is the other node attaching O to another tree
	other *Node
}

func initNode(o *Order, val int64, n, other *Node) {
	*n = Node{val: val, O: o, other: other}
	n.next = n
	n.prev = n
}

func (n *Node) isHead() bool {
	return n.pp != nil
}

func (n *Node) push(in *Node) {
	switch {
	case in.val == n.val:
		n.addLast(in)
	case in.val < n.val:
		if n.left == nil {
			in.pp = &n.left
			n.left = in
		} else {
			n.left.push(in)
		}
	case in.val > n.val:
		if n.right == nil {
			in.pp = &n.right
			n.right = in
		} else {
			n.right.push(in)
		}
	}
}

func (n *Node) peekMax() *Node {
	if n == nil {
		return nil
	}
	if n.right == nil {
		return n
	}
	return n.right.peekMax()
}

func (n *Node) popMax() *Node {
	m := n.peekMax()
	m.pop()
	return m
}

func (n *Node) peekMin() *Node {
	if n == nil {
		return nil
	}
	if n.left == nil {
		return n
	}
	return n.left.peekMin()
}

func (n *Node) popMin() *Node {
	m := n.peekMin()
	m.pop()
	return m
}

func (n *Node) popVal(val int64) *Node {
	switch {
	case n == nil:
		return nil
	case val == n.val:
		n.pop()
		return n
	case val < n.val:
		return n.left.popVal(val)
	case val > n.val:
		return n.right.popVal(val)
	}
	panic("unreachable")
}

func (n *Node) has(val int64) bool {
	if n == nil {
		return false
	}
	if n.val == val {
		return true
	}
	if n.val > val {
		return n.left.has(val)
	}
	return n.right.has(val)
}

func (n *Node) addLast(in *Node) {
	last := n.next
	last.prev = in
	in.next = last
	in.prev = n
	n.next = in
}

func (n *Node) pop() {
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
}

func (n *Node) swapWith(nn *Node) {
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

func (n *Node) detachAll() {
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

func (n *Node) detachMax() *Node {
	m := n.peekMax()
	m.detachAll()
	return m
}
