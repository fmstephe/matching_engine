package trade

import ()

type bst struct {
	size int
	root *node
}

func NewBST() *bst {
	return &bst{}
}

func (b *bst) Insert(n *node) {
	if b.root != nil {
		b.root.insert(n)
	} else {
		b.root = n
		n.pp = &b.root
	}
}

func (b *bst) PopMin() *node {
	if b.root != nil {
		b.size--
		return b.root.popMin()
	}
	return nil
}

func (b *bst) PopMax() *node {
	if b.root != nil {
		b.size--
		return b.root.popMax()
	}
	return nil
}

func (b *bst) Pop(val int64) *node {
	n := b.root.popVal(val)
	if n != nil {
		b.size--
	}
	return n
}

func (b *bst) Size() int {
	return b.size
}

type node struct {
	pp    **node
	left  *node
	right *node
	next  *node
	prev  *node
	size  int
	val   int64
	o     *Order
}

func initNode(val int64, o *Order, n *node) {
	*n = node{val: val, o: o}
	n.next = n
	n.prev = n
}

func (n *node) peekMax() *node {
	if n == nil {
		return nil
	}
	if n.right == nil {
		return n
	}
	return n.right.peekMax()
}

func (n *node) popMax() *node {
	if n == nil {
		return nil
	}
	m := n.peekMax()
	return m.pop()
}

func (n *node) peekMin() *node {
	if n == nil {
		return nil
	}
	if n.left == nil {
		return n
	}
	return n.left.peekMin()
}

func (n *node) popMin() *node {
	if n == nil {
		return nil
	}
	m := n.peekMin()
	return m.pop()
}

func (n *node) popVal(val int64) *node {
	switch {
	case n == nil:
		return nil
	case val == n.val:
		return n.pop()
	case val < n.val:
		return n.left.popVal(val)
	case val > n.val:
		return n.right.popVal(val)
	}
	panic("unreachable")
}

func (n *node) insert(in *node) {
	switch {
	case in.val == n.val:
		n.push(in)
	case in.val < n.val:
		if n.left == nil {
			in.pp = &n
			n.left = in
		} else {
			n.left.insert(in)
		}
	case in.val > n.val:
		if n.right == nil {
			in.pp = &n
			n.right = in
		} else {
			n.right.insert(in)
		}
	}
}

func (n *node) push(in *node) {
	last := n.next
	last.prev = in
	in.next = last
	in.prev = n
	n.next = n
	n.size++
}

func (n *node) pop() *node {
	n.prev.next = n.next
	n.next.prev = n.prev
	nn := n.next
	nn.size = n.size - 1
	n.next = nil
	n.prev = nil
	n.size = 1
	if n != nn {
		swap(n, nn)
	} else {
		detatch(n)
	}
	return n
}

func swap(n *node, nn *node) {
	nn.pp = n.pp
	nn.left = n.left
	nn.right = n.right
	*nn.pp = nn
	nn.left.pp = &nn
	nn.right.pp = &nn
}

func detatch(n *node) {
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
		nn := n.left.detatchMax()
		swap(n, nn)
	}
	n.pp = nil
	n.left = nil
	n.right = nil
}

func (n *node) detatchMax() *node {
	m := n.peekMax()
	detatch(m)
	return m
}
