// TODO this approach to remembering received messages just doesn't work
package coordinator

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/fmstephe/matching_engine/ints"
	"strconv"
)

const (
	blockSize = 1024 * 1024 // 2^20
	blockMask = blockSize - 1
)

type Ticker struct {
	r *root
}

func NewTicker() *Ticker {
	return &Ticker{}
}

func (t *Ticker) Tick(m *RMessage) (ticked bool) {
	val := ints.Combine(m.originId, m.msgId)
	if t.r == nil {
		t.r = newRoot(val)
	}
	return t.r.tick(val)
}

type root struct {
	n *node
}

func newRoot(val uint64) *root {
	n := newNode(val)
	r := &root{n: n}
	n.pp = &r.n
	return r
}

func (r *root) tick(val uint64) bool {
	return r.n.tick(val)
}

type node struct {
	black    bool
	left     *node
	right    *node
	parent   *node
	pp       **node
	min, max uint64
	block    [blockSize]bool
}

func (n *node) String() string {
	if n == nil {
		return "()"
	}
	valStr := strconv.Itoa(int(n.min))
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

func newNode(val uint64) *node {
	min := val &^ blockMask
	max := min + blockMask
	return &node{min: min, max: max}
}

func (n *node) tick(val uint64) (ticked bool) {
	for {
		switch {
		case val >= n.min && val <= n.max:
			return n.record(val)
		case val < n.min:
			if n.left == nil {
				nn := newNode(val)
				nn.record(val)
				nn.toLeftOf(n)
				rebalance(n)
				return true
			} else {
				n = n.left
			}
		case val > n.max:
			if n.right == nil {
				nn := newNode(val)
				nn.record(val)
				nn.toRightOf(n)
				rebalance(n)
				return true
			} else {
				n = n.right
			}
		}
	}
	panic("Unreachable")
}

func (n *node) isRed() bool {
	if n != nil {
		return !n.black
	}
	return false
}

func (n *node) record(val uint64) (ticked bool) {
	idx := val - n.min
	if n.block[idx] {
		return false
	}
	n.block[idx] = true
	return true
}

func rebalance(n *node) {
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

func (n *node) giveParent(nn *node) {
	nn.parent = n.parent
	nn.pp = n.pp
	*nn.pp = nn
	n.parent = nil
	n.pp = nil
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

func validateRBT(t *Ticker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	blackBalance(t.r.n, 0)
	testReds(t.r.n, 0)
	checkStructure(t.r.n)
	return nil
}

func blackBalance(n *node, depth int) int {
	if n == nil {
		return 0
	}
	lb := blackBalance(n.left, depth+1)
	rb := blackBalance(n.right, depth+1)
	if lb != rb {
		panic(errors.New(fmt.Sprintf("Unbalanced tree found at depth %d. Left: , %d Right: %d", depth, lb, rb)))
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

func checkStructure(n *node) {
	if n == nil {
		return
	}
	if *n.pp != n {
		panic(errors.New(fmt.Sprintf("Parent pointer not pointing at me!")))
	}
	if n.left != nil && n.min <= n.left.min {
		panic(errors.New(fmt.Sprintf("Left value is greater than or equal to node value. Left value: %d Node value %d", n.left.min, n.min)))
	}
	if n.right != nil && n.min >= n.right.min {
		panic(errors.New(fmt.Sprintf("Right value is less than or equal to node value. Right value: %d Node value %d", n.right.min, n.min)))
	}
	checkStructure(n.left)
	checkStructure(n.right)
}
