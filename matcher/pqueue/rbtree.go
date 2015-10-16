package pqueue

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/fmstephe/flib/fmath"
	"strconv"
)

const (
	leftChild     = 0
	rightChild    = 1
	nilParentDir  = 255
	rootParentDir = 254
)

type rbtree struct {
	fakeRoot *node
}

func (b *rbtree) String() string {
	return b.getRoot().String()
}

func (b *rbtree) getRoot() *node {
	if b.fakeRoot == nil {
		b.fakeRoot = &node{parentDir: rootParentDir}
	}
	return b.fakeRoot.children[leftChild]
}

func (b *rbtree) push(nn *node) {
	root := b.getRoot()
	if root != nil {
		root := b.fakeRoot.children[leftChild]
		insert(root, nn)
		b.getRoot().black = true
	} else {
		b.fakeRoot.children[leftChild] = nn
		nn.parent = b.fakeRoot
		nn.parentDir = leftChild
		nn.black = true
	}
}

func (b *rbtree) peekMin() *node {
	if b.fakeRoot == nil {
		return nil
	}
	n := b.fakeRoot.children[leftChild]
	for n.children[leftChild] != nil {
		n = n.children[leftChild]
	}
	return n
}

func (b *rbtree) popMin() *node {
	if b.fakeRoot != nil {
		n := b.peekMin()
		n.pop()
		n.other.pop() // Clear complementary rbtree
		return n
	}
	return nil
}

func (b *rbtree) peekMax() *node {
	if b.fakeRoot == nil {
		return nil
	}
	n := b.fakeRoot.children[leftChild]
	for n.children[rightChild] != nil {
		n = n.children[rightChild]
	}
	return n
}

func (b *rbtree) popMax() *node {
	if b.fakeRoot != nil {
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
	n := b.fakeRoot
	for {
		if n == nil {
			return nil
		}
		if val == n.val {
			return n
		}
		child := fmath.UILT(n.val, val)
		n = n.children[child]
	}
}

type node struct {
	black bool
	// Tree fields
	val       uint64
	children  [2]*node
	parent    *node
	parentDir uint8 // TODO this is bigger than it needs to be uint8?
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
	valStr := strconv.FormatUint(n.val, 10)
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
	n.parentDir = nilParentDir
}

func (n *node) getOrderNode() *OrderNode {
	if n != nil {
		return n.order
	}
	return nil
}

func (n *node) isRed() bool {
	// TODO return n == nil || !n.black
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
	case n.parentDir != nilParentDir:
		return false
	case n.next != n:
		return false
	case n.prev != n:
		return false
	}
	return true
}

func (n *node) isHead() bool {
	return n.parentDir != nilParentDir
}

func (n *node) rotSingle(dir uint8) *node {
	oDir := dir ^ 1
	nr := n.children[oDir]
	n.setChild(nr.children[dir], oDir)
	n.giveParent(nr)
	nr.setChild(n, dir)
	n.black = false
	nr.black = true
	return nr
}

func (n *node) rotDouble(dir uint8) *node {
	oDir := dir ^ 1
	n.children[oDir].rotSingle(oDir)
	return n.rotSingle(dir)
}

func insert(root, nn *node) {
	n := root
	// TODO val := nn.val
	for {
		if n.val == nn.val {
			n.appendNode(nn)
			return
		}
		dir := fmath.UILT(n.val, nn.val)
		child := n.children[dir]
		if child == nil {
			n.setChild(nn, dir)
			break
		}
		n = child
	}
	repairInsert(nn)
}

func repairInsert(nn *node) {
	n := nn       // node
	p := n.parent // parent
	g := p.parent // grandparent
	for g != nil && !p.black {
		dir := n.parentDir
		oDir := dir ^ 1
		if n.isRed() {
			u := g.getOtherChild(p.parentDir) // uncle
			if u.isRed() {
				g.black = false
				p.black = true
				u.black = true
				n = g // rebalancing skips to grandparent
				if g.parent == nil {
					return
				}
			} else {
				if n.parentDir == p.parentDir {
					g.rotSingle(oDir)
				} else {
					g.rotDouble(dir)
				}
				return
			}
		}
		p = n.parent
		g = p.parent
	}
}

func (n *node) getOtherChild(dir uint8) *node {
	return n.children[dir^1]
}

func (n *node) getSibling() *node {
	p := n.parent
	if p == nil {
		return nil
	}
	return p.children[n.parentDir^1]
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
	// Guarantee: Each of n.parent/parentDir/left/right are now nil
}

func (n *node) giveParent(nn *node) {
	nn.parent = n.parent
	nn.parentDir = n.parentDir
	if n.parent != nil {
		nn.parent.children[nn.parentDir] = nn
	}
	n.parent = nil
	n.parentDir = nilParentDir
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
		n.parent.children[n.parentDir] = nil
		n.parent = nil
		n.parentDir = nilParentDir
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
	s := p.getOtherChild(n.parentDir)
	repairDetach(p, n, s, nn)
}

func repairDetach(p, n, s, nn *node) {
	// Guarantee: Each of n.parent/parentDir/left/right are now nil
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
				p.rotSingle(rightChild)
				s = p.children[leftChild]
			} else {
				p.rotSingle(leftChild)
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
					p = p.rotSingle(rightChild)
				} else {
					s.rotSingle(leftChild)
					p = p.rotSingle(rightChild)
				}
			} else {
				if srRed {
					p = p.rotSingle(leftChild)
				} else {
					s.rotSingle(rightChild)
					p = p.rotSingle(leftChild)
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

func (n *node) pop() {
	switch {
	case !n.isHead():
		n.prev.next = n.next
		n.next.prev = n.prev
		/*
			n.parent = nil
			n.parentDir = -1
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
	// Guarantee: Each of n.parent/parentDir/left/right are now nil
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

func (n *node) setChild(child *node, childDir uint8) {
	n.children[childDir] = child
	if child != nil {
		child.parent = n
		child.parentDir = childDir
	}
}

func (n *node) flip() {
	n.black = !n.black
	n.children[leftChild].black = !n.children[leftChild].black
	n.children[rightChild].black = !n.children[rightChild].black
}

func validateRBT(rbt *rbtree) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	blackBalance(rbt.getRoot(), 0)
	testReds(rbt.getRoot(), 0)
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
