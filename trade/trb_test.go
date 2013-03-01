package trade

import (
	"testing"
)

func validateRBT(t *testing.T, rbt *tree) {
	blackBalance(t, rbt.root, 0)
	testReds(t, rbt.root, 0)
}

func blackBalance(t *testing.T, n *node, depth int) int {
	if n == nil {
		return 0
	}
	lb := blackBalance(t, n.left, depth+1)
	rb := blackBalance(t, n.right, depth+1)
	if lb != rb {
		t.Errorf("Unbalanced tree found at depth %d. Left: , %d Right: %d", depth, lb, rb)
	}
	b := lb
	if !n.isRed() {
		b++
	}
	return b
}

func testReds(t *testing.T, n *node, depth int) {
	if n == nil {
		return
	}
	if n.isRed() && (n.left.isRed() || n.right.isRed()) && depth != 0 {
		t.Errorf("Red violation found at depth %d", depth)
	}
	if !n.left.isRed() && n.right.isRed() {
		t.Errorf("Right leaning red leaf found at depth %d", depth)
	}
	if n.left.isRed() && n.right.isRed() {
		t.Errorf("Red child pair found at depth", depth)
	}
	testReds(t, n.left, depth+1)
	testReds(t, n.right, depth+1)
}
