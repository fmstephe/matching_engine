package pqueue

import (
	"github.com/fmstephe/matching_engine/msg"
	"math/rand"
	"testing"
)

// A function signature allowing us to switch easily between min and max queues
type popperFun func(*testing.T, *rbtree, *rbtree, *pqueue) (*OrderNode, *OrderNode, *OrderNode)

var msgMkr = msg.NewMessageMaker(1)

func TestPush(t *testing.T) {
	// buys
	testPushAscDesc(t, 100, msg.BUY)
	// buys
	testPushSimple(t, 1, 1, 1, msg.BUY)
	testPushSimple(t, 4, 1, 1, msg.SELL)
	testPushSimple(t, 100, 10, 20, msg.BUY)
	testPushSimple(t, 100, 100, 10000, msg.SELL)
	testPushSimple(t, 1000, 100, 10000, msg.BUY)
}

func TestPushPopSimpleMin(t *testing.T) {
	// buys
	testPushPopSimple(t, 1, 1, 1, msg.BUY, maxPopper)
	testPushPopSimple(t, 4, 1, 1, msg.BUY, maxPopper)
	testPushPopSimple(t, 100, 10, 20, msg.BUY, maxPopper)
	testPushPopSimple(t, 100, 100, 10000, msg.BUY, maxPopper)
	testPushPopSimple(t, 1000, 100, 10000, msg.BUY, maxPopper)
	// sells
	testPushPopSimple(t, 1, 1, 1, msg.SELL, minPopper)
	testPushPopSimple(t, 100, 1, 1, msg.SELL, minPopper)
	testPushPopSimple(t, 100, 10, 20, msg.SELL, minPopper)
	testPushPopSimple(t, 100, 100, 10000, msg.SELL, minPopper)
	testPushPopSimple(t, 1000, 100, 10000, msg.SELL, minPopper)
}

func TestRandomPushPop(t *testing.T) {
	// buys
	testPushPopRandom(t, 1, 1, 1, msg.BUY, maxPopper)
	testPushPopRandom(t, 100, 1, 1, msg.BUY, maxPopper)
	testPushPopRandom(t, 100, 10, 20, msg.BUY, maxPopper)
	testPushPopRandom(t, 100, 100, 10000, msg.BUY, maxPopper)
	testPushPopRandom(t, 1000, 100, 10000, msg.BUY, maxPopper)
	// sells
	testPushPopRandom(t, 1, 1, 1, msg.SELL, minPopper)
	testPushPopRandom(t, 100, 1, 1, msg.SELL, minPopper)
	testPushPopRandom(t, 100, 10, 20, msg.SELL, minPopper)
	testPushPopRandom(t, 100, 100, 10000, msg.SELL, minPopper)
	testPushPopRandom(t, 1000, 100, 10000, msg.SELL, minPopper)
}

func TestAddRemoveSimple(t *testing.T) {
	// Buys
	testAddRemoveSimple(t, 1, 1, 1, msg.BUY)
	testAddRemoveSimple(t, 100, 1, 1, msg.BUY)
	testAddRemoveSimple(t, 100, 10, 20, msg.BUY)
	testAddRemoveSimple(t, 100, 100, 10000, msg.BUY)
	testAddRemoveSimple(t, 1000, 100, 10000, msg.BUY)
	// Sells
	testAddRemoveSimple(t, 1, 1, 1, msg.SELL)
	testAddRemoveSimple(t, 100, 1, 1, msg.SELL)
	testAddRemoveSimple(t, 100, 10, 20, msg.SELL)
	testAddRemoveSimple(t, 100, 100, 10000, msg.SELL)
	testAddRemoveSimple(t, 1000, 100, 10000, msg.SELL)
}

func TestAddRemoveRandom(t *testing.T) {
	// Buys
	testAddRemoveRandom(t, 1, 1, 1, msg.BUY)
	testAddRemoveRandom(t, 100, 1, 1, msg.BUY)
	testAddRemoveRandom(t, 100, 10, 20, msg.BUY)
	testAddRemoveRandom(t, 100, 100, 10000, msg.BUY)
	testAddRemoveRandom(t, 1000, 100, 10000, msg.BUY)
	// Sells
	testAddRemoveRandom(t, 1, 1, 1, msg.SELL)
	testAddRemoveRandom(t, 100, 1, 1, msg.SELL)
	testAddRemoveRandom(t, 100, 10, 20, msg.SELL)
	testAddRemoveRandom(t, 100, 100, 10000, msg.SELL)
	testAddRemoveRandom(t, 1000, 100, 10000, msg.SELL)
}

func testPushAscDesc(t *testing.T, pushCount int, kind msg.MsgKind) {
	priceTree := &rbtree{}
	guidTree := &rbtree{}
	validate(t, priceTree, guidTree)
	for i := 0; i < pushCount; i++ {
		o := &OrderNode{}
		o.CopyFrom(msgMkr.MkPricedOrder(uint64(i), kind))
		priceTree.push(&o.priceNode)
		guidTree.push(&o.guidNode)
		validate(t, priceTree, guidTree)
	}
	for i := pushCount - 1; i >= 0; i-- {
		o := &OrderNode{}
		o.CopyFrom(msgMkr.MkPricedOrder(uint64(i), kind))
		priceTree.push(&o.priceNode)
		guidTree.push(&o.guidNode)
		validate(t, priceTree, guidTree)
	}
}

func testPushSimple(t *testing.T, pushCount int, lowPrice, highPrice uint64, kind msg.MsgKind) {
	priceTree := &rbtree{}
	guidTree := &rbtree{}
	validate(t, priceTree, guidTree)
	for i := 0; i < pushCount; i++ {
		o := &OrderNode{}
		o.CopyFrom(msgMkr.MkPricedOrder(msgMkr.Between(lowPrice, highPrice), kind))
		priceTree.push(&o.priceNode)
		guidTree.push(&o.guidNode)
		validate(t, priceTree, guidTree)
	}
}

func testPushPopSimple(t *testing.T, pushCount int, lowPrice, highPrice uint64, kind msg.MsgKind, popper popperFun) {
	priceTree := &rbtree{}
	guidTree := &rbtree{}
	validate(t, priceTree, guidTree)
	q := mkPrioq(lowPrice, highPrice)
	for i := 0; i < pushCount; i++ {
		o := &OrderNode{}
		o.CopyFrom(msgMkr.MkPricedOrder(msgMkr.Between(lowPrice, highPrice), kind))
		priceTree.push(&o.priceNode)
		guidTree.push(&o.guidNode)
		validate(t, priceTree, guidTree)
		q.push(o)
	}
	for i := 0; i < pushCount; i++ {
		popCheck(t, priceTree, guidTree, q, popper)
	}
}

func testPushPopRandom(t *testing.T, pushCount int, lowPrice, highPrice uint64, kind msg.MsgKind, popper popperFun) {
	priceTree := &rbtree{}
	guidTree := &rbtree{}
	validate(t, priceTree, guidTree)
	q := mkPrioq(lowPrice, highPrice)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || priceTree.peekMin() == nil {
			o := &OrderNode{}
			o.CopyFrom(msgMkr.MkPricedOrder(msgMkr.Between(lowPrice, highPrice), kind))
			priceTree.push(&o.priceNode)
			guidTree.push(&o.guidNode)
			validate(t, priceTree, guidTree)
			q.push(o)
			i++
		} else {
			popCheck(t, priceTree, guidTree, q, popper)
		}
	}
	for priceTree.peekMin() != nil {
		po := priceTree.popMax().getOrderNode()
		fo := q.popMax()
		if fo != po {
			t.Errorf("Mismatched Push/Pop pair")
			return
		}
		ensureFreed(t, po)
		validate(t, priceTree, guidTree)
	}
}

func testAddRemoveSimple(t *testing.T, pushCount int, lowPrice, highPrice uint64, kind msg.MsgKind) {
	priceTree := &rbtree{}
	guidTree := &rbtree{}
	validate(t, priceTree, guidTree)
	orderMap := make(map[uint64]*OrderNode)
	for i := 0; i < pushCount; i++ {
		o := &OrderNode{}
		o.CopyFrom(msgMkr.MkPricedOrder(msgMkr.Between(lowPrice, highPrice), kind))
		priceTree.push(&o.priceNode)
		guidTree.push(&o.guidNode)
		validate(t, priceTree, guidTree)
		orderMap[o.Guid()] = o
	}
	drainTree(t, priceTree, guidTree, orderMap)
}

func testAddRemoveRandom(t *testing.T, pushCount int, lowPrice, highPrice uint64, kind msg.MsgKind) {
	priceTree := &rbtree{}
	guidTree := &rbtree{}
	validate(t, priceTree, guidTree)
	orderMap := make(map[uint64]*OrderNode)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || guidTree.peekMin() == nil {
			o := &OrderNode{}
			o.CopyFrom(msgMkr.MkPricedOrder(msgMkr.Between(lowPrice, highPrice), kind))
			priceTree.push(&o.priceNode)
			guidTree.push(&o.guidNode)
			validate(t, priceTree, guidTree)
			orderMap[o.Guid()] = o
			i++
		} else {
			for g, o := range orderMap {
				po := guidTree.cancel(g).getOrderNode()
				delete(orderMap, g)
				if po != o {
					t.Errorf("Bad pop")
				}
				ensureFreed(t, po)
				validate(t, priceTree, guidTree)
				break
			}
		}
	}
	drainTree(t, priceTree, guidTree, orderMap)
}

func drainTree(t *testing.T, priceTree, guidTree *rbtree, orderMap map[uint64]*OrderNode) {
	for g := range orderMap {
		o := orderMap[g]
		po := guidTree.cancel(o.Guid()).getOrderNode()
		if po != o {
			t.Errorf("Bad pop")
		}
		ensureFreed(t, po)
		validate(t, priceTree, guidTree)
	}
}

func ensureFreed(t *testing.T, o *OrderNode) {
	if !o.priceNode.isFree() {
		t.Errorf("Price Node was not freed")
	}
	if !o.guidNode.isFree() {
		t.Errorf("Guid Node was not freed")
	}
}

// Quick check to ensure the rbtree's internal structure is valid
func validate(t *testing.T, priceTree, guidTree *rbtree) {
	if err := validateRBT(priceTree); err != nil {
		t.Errorf("%s", err.Error())
	}
	if err := validateRBT(guidTree); err != nil {
		t.Errorf("%s", err.Error())
	}
	checkStructure(t, priceTree.root)
	checkStructure(t, guidTree.root)
}

func checkStructure(t *testing.T, n *node) {
	if n == nil {
		return
	}
	checkQueue(t, n)
	if *n.pp != n {
		t.Errorf("Parent pointer does not point to child node")
	}
	if n.left != nil {
		if n.val <= n.left.val {
			t.Errorf("Left value is greater than or equal to node value. Left value: %d Node value %d", n.left.val, n.val)
		}
		checkStructure(t, n.left)
	}
	if n.right != nil {
		if n.val >= n.right.val {
			t.Errorf("Right value is less than or equal to node value. Right value: %d Node value %d", n.right.val, n.val)
		}
		checkStructure(t, n.right)
	}
}

func checkQueue(t *testing.T, n *node) {
	curr := n.next
	prev := n
	for curr != n {
		if curr.prev != prev {
			t.Errorf("Bad queue next/prev pair")
		}
		if curr.pp != nil {
			t.Errorf("Internal queue node with non-nil parent pointer")
		}
		if curr.left != nil {
			t.Errorf("Internal queue node has non-nil left child")
		}
		if curr.right != nil {
			t.Errorf("Internal queue node has non-nil right child")
		}
		if curr.order == nil {
			t.Errorf("Internal queue node has nil OrderNode")
		}
		prev = curr
		curr = curr.next
	}
}

// Function to pop and peek and check that everything is in order
func popCheck(t *testing.T, priceTree, guidTree *rbtree, q *pqueue, popper popperFun) {
	peek, pop, check := popper(t, priceTree, guidTree, q)
	if pop != check {
		t.Errorf("Mismatched push/pop pair")
		return
	}
	if pop != peek {
		t.Errorf("Mismatched peek/pop pair")
		return
	}
	validate(t, priceTree, guidTree)
}

// Helper functions for popping either the max or the min from our queues
func maxPopper(t *testing.T, priceTree, guidTree *rbtree, q *pqueue) (peek, pop, check *OrderNode) {
	peek = priceTree.peekMax().getOrderNode()
	if !guidTree.Has(peek.Guid()) {
		t.Errorf("Guid rbtree does not contain peeked order")
	}
	pop = priceTree.popMax().getOrderNode()
	if guidTree.Has(peek.Price()) {
		t.Errorf("Guid rbtree still contains popped order")
		return
	}
	check = q.popMax()
	ensureFreed(t, pop)
	return
}

func minPopper(t *testing.T, priceTree, guidTree *rbtree, q *pqueue) (peek, pop, check *OrderNode) {
	peek = priceTree.peekMin().getOrderNode()
	if !guidTree.Has(peek.Guid()) {
		t.Errorf("Guid rbtree does not contain peeked order")
	}
	pop = priceTree.popMin().getOrderNode()
	check = q.popMin()
	ensureFreed(t, pop)
	return
}
