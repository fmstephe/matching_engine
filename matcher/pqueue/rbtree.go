package pqueue

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/fmstephe/flib/fmath"
	"strconv"
)

type rbtree struct {
	root *node
}

func (b *rbtree) String() string {
	return b.root.String()
}

func (b *rbtree) push(nn *node) {
	if b.root == nil {
		b.root = nn
		nn.parentIdx = -2 // Special case for root
		return
	}
	b.root.push(in)
}

func (b *rbtree) peekMin() *node {
	n := b.root
	if n == nil {
		return nil
	}
	for n.children[leftChild] != nil {
		n = n.children[leftChild]
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
	for n.children[rightChild] != nil {
		n = n.children[rightChild]
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
		child := fmath.UGreaterThan(val, n.val)
		n = n.children[child]
	}
}

const (
	leftChild  = 0
	rightChild = 1
)

type node struct {
	black bool
	// Tree fields
	val       uint64
	children  [2]*node
	parent    *node
	parentIdx int // TODO this is bigger than it needs to be uint8?
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
	if !(n.children[leftChild] == nil && n.children[rightChild] == nil) {
		b.WriteString(", ")
		b.WriteString(n.children[leftChild].String())
		b.WriteString(", ")
		b.WriteString(n.children[rightChild].String())
	}
	b.WriteString(")")
	return b.String()
}

func initNode(o *OrderNode, val uint64, n, other *node) {
	*n = node{val: val, order: o, other: other}
	n.next = n
	n.prev = n
	n.black = false
	n.parentIdx = -1
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
	case n.children[leftChild] != nil:
		return false
	case n.children[rightChild] != nil:
		return false
	case n.parentIdx != -1:
		return false
	case n.next != n:
		return false
	case n.prev != n:
		return false
	}
	return true
}

func (n *node) isHead() bool {
	return n.parentIdx != -1
}

func (n *node) rotSingle(dir int) {
	oDir := fmath.UINot(dir)
	s := n.children[oDir]
	n.setChild(s.children[dir], oDir)
	s.setChild(n, dir)
	n.red = true
	s.red = false
}

func (n *node) rotDouble() {
	oDir := fmath.UINot(dir)
	rotSingle(n.children[oDir], oDir)
	rotSingle(n, dir)
}

func insert(root, nn *node) {
	p := root
	for {
		if p.val == nn.val {
			p.appendNode(nn)
			return
		}
		dir := fmath.UIGTE(p.val, nn.val)
		child := p.children[dir]
		if child == nil {
			p.setChild(nn, dir)
			break
		}
		p = child
	}
	n = p
	p = n.parent
	for p != nil {
		dir := n.parentIdx
		oDir := fmath.UINot(dir)
		if n.isRed() {
			s := p.children[oDir]
			if s.isRed() {
				// Asking for a nil pointer panic
				p.black = false
				n.black = true
				s.black = true
			} else {
				if n.children[dir].isRed() {
					p.rotSingle(oDir)
				} else if n.children[oDir].isRed() {
					p.rotDouble(oDir)
				}
			}
		}
		n = p
		p = n.parent
	}
}

func (n *node) getSibling() *node {
	p := n.parent
	if p == nil {
		return nil
	}
	p.children[fmath.UINot(n.parentIdx)]
}

func (n *node) appendNode(in *node) {
	last := n.next
	last.prev = in
	in.next = last
	in.prev = n
	n.next = in
}

func (n *node) givePosition(nn *node) {
	n.giveParent(nn)
	n.giveChildren(nn)
	nn.black = n.black
	// Guarantee: Each of n.parent/parentIdx/left/right are now nil
}

func (n *node) giveParent(nn *node) {
	nn.parent = n.parent
	nn.parentIdx = n.parentIdx
	nn.parent.children[nn.parentIdx] = nn
	n.parent = nil
	n.parentIdx = -1
}

func (n *node) giveChildren(nn *node) {
	nn.setChild(n.children[leftChild], leftChild)
	nn.setChild(n.children[rightChild], rightChild)
	n.children[leftChild] = nil
	n.children[rightChild] = nil
}

func (n *node) detach() {
	var nn *node
	switch {
	case n.children[rightChild] == nil && n.children[leftChild] == nil:
		n.parent.children[n.parentIdx] = nil
		n.parent = nil
		n.parentIdx = -1
	case n.children[rightChild] == nil:
		nn = n.children[leftChild]
		n.giveParent(nn)
		n.children[leftChild] = nil
	case n.children[leftChild] == nil:
		nn = n.children[rightChild]
		n.giveParent(nn)
		n.children[rightChild] = nil
	default:
		nn = n.children[leftChild].detachMax()
		n.givePosition(nn)
		return
	}
	p := n.parent
	s := n.getSibling()
	repairDetach(p, n, s, nn)
}

func repairDetach(p, n, s, nn *node) {
	// Guarantee: Each of n.parent/parentIdx/left/right are now nil
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
			if p.children[leftChild] == s {
				p.rotateRight()
				s = p.children[leftChild]
			} else {
				p.rotateLeft()
				s = p.children[rightChild]
			}
		}
		pRed := p.isRed()
		slRed := s.children[leftChild].isRed()
		srRed := s.children[rightChild].isRed()
		if !slRed && !srRed {
			if pRed { // Sibling's children are black and parent is red
				p.black = true
				s.black = false
				return
			} else { // Sibling's children and parent are black, makes a black violation
				s.black = false
			}
		} else { // One of sibling's children is red
			if p.children[leftChild] == s {
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
			p.children[leftChild].black = true
			p.children[rightChild].black = true
			return
		}
		s = p.getSibling()
		p = p.parent
	}
}

func repairInsert(n *node) {
	for n != nil {
		if n.children[leftChild].isRed() && n.children[rightChild].isRed() {
			n.flip()
		}
		if n.children[leftChild].isRed() {
			if n.children[leftChild].children[leftChild].isRed() {
				n = n.rotateRight()
			}
			if n.children[leftChild].children[rightChild].isRed() {
				n.children[leftChild].rotateLeft()
				n = n.rotateRight()
			}
		}
		if n.children[rightChild].isRed() {
			if n.children[rightChild].children[rightChild].isRed() {
				n = n.rotateLeft()
			}
			if n.children[rightChild].children[leftChild].isRed() {
				n.children[rightChild].rotateRight()
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
		/*
			n.parent = nil
			n.parentIdx = -1
			n.children[leftChild] = nil
			n.children[rightChild] = nil
		*/
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
	// Guarantee: Each of n.parent/parentIdx/left/right are now nil
	// Guarantee: Both n.children[leftChild]/right point to n
}

func (n *node) detachMax() *node {
	m := n
	for {
		if m.children[rightChild] == nil {
			break
		}
		m = m.children[rightChild]
	}
	m.detach()
	return m
}

func (n *node) setChild(child *node, childIdx int) {
	n.children[childIdx] = child
	if child != nil {
		child.parent = n
		child.parentIdx = childIdx
	}
}

func (n *node) rotate(dir int) *node {
	odir := fmath.UIComplement(dir)
	l := n.children[dir]
	n.giveParent(l)
	l.children[odir].makeChildOf(n, dir)
	n.makeChildOf(l, odir)
	l.black = n.black
	n.black = false
	return l
}

func (n *node) flip() {
	n.black = !n.black
	n.children[leftChild].black = !n.children[leftChild].black
	n.children[rightChild].black = !n.children[rightChild].black
}

func (n *node) moveRed(dir int) {
	odir := fmath.UIComplement(dir)
	n.flip()
	if n.children[odir].children[dir].isRed() {
		n.children[odir].rotate(odir)
		n.rotate(dir)
		n.parent.flip()
	}
}

func (n *node) moveRedLeft() {
	n.flip()
	if n.children[rightChild].children[leftChild].isRed() {
		n.children[rightChild].rotate(rightChild)
		n.rotate(leftChild)
		n.parent.flip()
	}
}

func (n *node) moveRedRight() {
	n.flip()
	if n.children[leftChild].children[leftChild].isRed() {
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
	lb := blackBalance(n.children[leftChild], depth+1)
	rb := blackBalance(n.children[rightChild], depth+1)
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
	if n.isRed() && (n.children[leftChild].isRed() || n.children[rightChild].isRed()) && depth != 0 {
		panic(errors.New(fmt.Sprintf("Red violation found at depth %d", depth)))
	}
	testReds(n.children[leftChild], depth+1)
	testReds(n.children[rightChild], depth+1)
}
