package trade

import ()

type Tree struct {
	root *Node
}

func NewTree() *Tree {
	return &Tree{}
}

func (b *Tree) Push(in *Node) {
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

func (b *Tree) PeekMin() *Node {
	n := b.root
	if n == nil {
		return nil
	}
	for {
		if n.left == nil {
			return n
		}
		n = n.left
	}
	panic("Unreachable")
}

func (b *Tree) PopMin() *Node {
	if b.root != nil {
		n := b.PeekMin()
		n.pop()
		n.other.pop() // Clear complementary tree
		return n
	}
	return nil
}

func (b *Tree) PeekMax() *Node {
	n := b.root
	if n == nil {
		return nil
	}
	for {
		if n.right == nil {
			return n
		}
		n = n.right
	}
	panic("Unreachable")
}

func (b *Tree) PopMax() *Node {
	if b.root != nil {
		n := b.PeekMax()
		n.pop()
		n.other.pop() // Clear complementary tree
		return n
	}
	return nil
}

func (b *Tree) Pop(val int64) *Node {
	n := b.get(val)
	if n == nil {
		return nil
	}
	n.pop()
	n.other.pop()
	return n
}

func (b *Tree) Has(val int64) bool {
	return b.get(val) != nil
}

func (b *Tree) get(val int64) *Node {
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

type Node struct {
	// Tree fields
	val   int64
	left  *Node
	right *Node
	pp    **Node
	// Limit queue fields
	next *Node
	prev *Node
	// This is the other node attaching O to another tree
	other *Node
	// Order
	O *Order
}

func initNode(o *Order, val int64, n, other *Node) {
	*n = Node{val: val, O: o, other: other}
	n.next = n
	n.prev = n
}

func (n *Node) isFree() bool {
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

func (n *Node) isHead() bool {
	return n.pp != nil
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
	n.next = n
	n.prev = n
	n.pp = nil
	n.left = nil
	n.right = nil
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
