package trade

import (
	"testing"
)

func validateRBT(t *testing.T, rbt *tree) {
	blackBalance(t, rbt.root)
	testReds(t, rbt.root)
}

func blackBalance(t *testing.T, n *node) int {
	if n == nil {
		return 0
	}
	lb := blackBalance(t, n.left)
	rb := blackBalance(t, n.right)
	if lb != rb {
		t.Errorf("Unbalanced tree found. Left: , %d Right: %d", lb, rb)
	}
	return lb
}

func testReds(t *testing.T, n *node) {
	if n == nil {
		return
	}
	if n.isRed() && (n.left.isRed() || n.right.isRed()) {
		t.Errorf("Red violation found.")
	}
	testReds(t, n.left)
	testReds(t, n.right)
}

func TestRBTInserts(t *testing.T) {
	testRBTInserts(t, 1, 1, 1)
	testRBTInserts(t, 100, 1, 1)
	testRBTInserts(t, 100, 10, 20)
	testRBTInserts(t, 100, 1000, 20000)
}

func testRBTInserts(t *testing.T, pushCount int, lowPrice, highPrice int64) {
	rbt := &tree{}
	for i := 0; i < pushCount; i++ {
		o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), SELL)
		rbt.push(&o.priceNode)
		validateRBT(t, rbt)
	}
}
