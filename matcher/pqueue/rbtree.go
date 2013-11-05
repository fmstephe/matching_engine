package pqueue

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

type rbtree struct {
	root *node
}

func (b *rbtree) String() string {
	return b.root.String()
}

func (b *rbtree) push(in *node) {
	if b.root == nil {
		b.root = in
		in.pp = &b.root
		return
	}
	b.root.push(in)
}

func (b *rbtree) peekMin() *node {
	n := b.root
	if n == nil {
		return nil
	}
	for n.left != nil {
		n = n.left
	}
	return n
}

func (b *rbtree) popMin() *node {
	if b.root != nil {
		n := b.peekMin()
		n.pop()
		n.other.pop() // Clear complementary rbtree
		return n
	}
	return nil
}

func (b *rbtree) peekMax() *node {
	n := b.root
	if n == nil {
		return nil
	}
	for n.right != nil {
		n = n.right
	}
	return n
}

func (b *rbtree) popMax() *node {
	if b.root != nil {
		n := b.peekMax()
		n.pop()
		n.other.pop() // Clear complementary rbtree
		return n
	}
	return nil
}

func (b *rbtree) cancel(val uint64) *node {
	n := b.get(val)
	if n == nil {
		return nil
	}
	n.pop()
	n.other.pop()
	return n
}

func (b *rbtree) Has(val uint64) bool {
	return b.get(val) != nil
}

func (b *rbtree) get(val uint64) *node {
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
	val    uint64
	left   *node
	right  *node
	parent *node
	pp     **node
	// Limit queue fields
	next *node
	prev *node
	// OrderNode
	order *OrderNode
	// This is the other node tying order to another rbtree
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

func initNode(o *OrderNode, val uint64, n, other *node) {
	*n = node{val: val, order: o, other: other}
	n.next = n
	n.prev = n
	n.black = false
}

func (n *node) getOrderNode() *OrderNode {
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

func (n *node) push(in *node) {
	for {
		switch {
		case in.val == n.val:
			n.addLast(in)
			return
		case in.val < n.val:
			if n.left == nil {
				in.toLeftOf(n)
				repairInsert(n)
				return
			} else {
				n = n.left
			}
		case in.val > n.val:
			if n.right == nil {
				in.toRightOf(n)
				repairInsert(n)
				return
			} else {
				n = n.right
			}
		}
	}
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
	repairDetach(p, n, s, nn)
}

func repairDetach(p, n, s, nn *node) {
	// Guarantee: Each of n.parent/pp/left/right are now nil
	if n.isRed() {
		return
	}
	if nn.isRed() {
		// Since n was black we can happily make its red replacement black
		nn.black = true
		return
	}
	repairToRoot(p, s)
}

func repairToRoot(p, s *node) {
	for p != nil {
		if s == nil {
			return
		}
		if s.isRed() { // Perform a rotation to make sibling black
			if p.left == s {
				p.rotateRight()
				s = p.left
			} else {
				p.rotateLeft()
				s = p.right
			}
		}
		pRed := p.isRed()
		slRed := s.left.isRed()
		srRed := s.right.isRed()
		if !slRed && !srRed {
			if pRed { // Sibling's children are black and parent is red
				p.black = true
				s.black = false
				return
			} else { // Sibling's children and parent are black, makes a black violation
				s.black = false
			}
		} else { // One of sibling's children is red
			if p.left == s {
				if slRed {
					p = p.rotateRight()
				} else {
					s.rotateLeft()
					p = p.rotateRight()
				}
			} else {
				if srRed {
					p = p.rotateLeft()
				} else {
					s.rotateRight()
					p = p.rotateLeft()
				}
			}
			p.black = !pRed
			p.left.black = true
			p.right.black = true
			return
		}
		s = p.getSibling()
		p = p.parent
	}
}

func repairInsert(n *node) {
	for n != nil {
		if n.left.isRed() && n.right.isRed() {
			n.flip()
		}
		if n.left.isRed() {
			if n.left.left.isRed() {
				n = n.rotateRight()
			}
			if n.left.right.isRed() {
				n.left.rotateLeft()
				n = n.rotateRight()
			}
		}
		if n.right.isRed() {
			if n.right.right.isRed() {
				n = n.rotateLeft()
			}
			if n.right.left.isRed() {
				n.right.rotateRight()
				n = n.rotateLeft()
			}
		}
		n = n.parent
	}
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

func validateRBT(rbt *rbtree) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	blackBalance(rbt.root, 0)
	testReds(rbt.root, 0)
	return nil
}

func blackBalance(n *node, depth int) int {
	if n == nil {
		return 0
	}
	lb := blackBalance(n.left, depth+1)
	rb := blackBalance(n.right, depth+1)
	if lb != rb {
		panic(errors.New(fmt.Sprintf("Unbalanced rbtree found at depth %d. Left: , %d Right: %d", depth, lb, rb)))
	}
	b := lb
	if !n.isRed() {
		b++
	}
	return b
}

func testReds(n *node, depth int) {
	if n == nil {
		return
	}
	if n.isRed() && (n.left.isRed() || n.right.isRed()) && depth != 0 {
		panic(errors.New(fmt.Sprintf("Red violation found at depth %d", depth)))
	}
	testReds(n.left, depth+1)
	testReds(n.right, depth+1)
}
