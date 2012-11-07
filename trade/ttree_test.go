package trade

import (
	"math/rand"
	"testing"
)

// A function signature allowing us to switch easily between min and max queues
type popperFun func(*Tree, *prioq) (*Order, *Order, *Order)

var maker = NewOrderMaker()

func TestPushPopSimpleMin(t *testing.T) {
	// buys
	testPushPopSimple(t, 100, 1, 1, BUY, maxPopper)
	testPushPopSimple(t, 100, 10, 20, BUY, maxPopper)
	testPushPopSimple(t, 100, 100, 10000, BUY, maxPopper)
	// sells
	testPushPopSimple(t, 100, 1, 1, SELL, minPopper)
	testPushPopSimple(t, 100, 10, 20, SELL, minPopper)
	testPushPopSimple(t, 100, 100, 10000, SELL, minPopper)
}

func testPushPopSimple(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind, popper popperFun) {
	bst := NewTree()
	validate(t, bst)
	q := mkPrioq(pushCount, lowPrice, highPrice)
	for i := 0; i < pushCount; i++ {
		o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), kind)
		bst.Push(&o.PriceNode)
		validate(t, bst)
		q.push(o)
		if bst.Size() != (i + 1) {
			t.Errorf("Incorrect size. Expecting %d, got %d instead", i+1, bst.Size())
		}
	}
	for i := 0; i < pushCount; i++ {
		popCheck(t, bst, q, popper)
	}
	if bst.Size() != 0 {
		t.Errorf("Expecting empty limit, got %d remaining orders, after pushing %d orders", bst.Size(), pushCount)
	}
}

func TestRandomPushPop(t *testing.T) {
	// buys
	testPushPopRandom(t, 100, 1, 1, BUY, maxPopper)
	testPushPopRandom(t, 100, 10, 20, BUY, maxPopper)
	testPushPopRandom(t, 100, 100, 10000, BUY, maxPopper)
	// sells
	testPushPopRandom(t, 100, 1, 1, SELL, minPopper)
	testPushPopRandom(t, 100, 10, 20, SELL, minPopper)
	testPushPopRandom(t, 100, 100, 10000, SELL, minPopper)
}

func testPushPopRandom(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind, popper popperFun) {
	bst := NewTree()
	validate(t, bst)
	q := mkPrioq(pushCount, lowPrice, highPrice)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || bst.Size() == 0 {
			o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), kind)
			bst.Push(&o.PriceNode)
			validate(t, bst)
			q.push(o)
			i++
		} else {
			popCheck(t, bst, q, popper)
		}
	}
	for bst.Size() > 0 {
		po := bst.PopMax().O
		fo := q.popMax()
		if fo != po {
			t.Errorf("Mismatched Push/Pop pair")
			return
		}
		validate(t, bst)
	}
}

func TestAddRemoveSimple(t *testing.T) {
	// Buys
	testAddRemoveSimple(t, 100, 1, 1, BUY)
	testAddRemoveSimple(t, 100, 10, 20, BUY)
	testAddRemoveSimple(t, 100, 100, 10000, BUY)
	// Sells
	testAddRemoveSimple(t, 100, 1, 1, SELL)
	testAddRemoveSimple(t, 100, 10, 20, SELL)
	testAddRemoveSimple(t, 100, 100, 10000, SELL)
}

func testAddRemoveSimple(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind) {
	bst := NewTree()
	validate(t, bst)
	orderMap := make(map[int64]*Order)
	for i := 0; i < pushCount; i++ {
		o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), kind)
		bst.Push(&o.GuidNode)
		validate(t, bst)
		orderMap[o.Guid] = o
	}
	drainTree(t, bst, orderMap)
}

func TestAddRemoveRandom(t *testing.T) {
	// Buys
	testAddRemoveRandom(t, 100, 1, 1, BUY)
	testAddRemoveRandom(t, 100, 10, 20, BUY)
	testAddRemoveRandom(t, 100, 100, 10000, BUY)
	// Sells
	testAddRemoveRandom(t, 100, 1, 1, SELL)
	testAddRemoveRandom(t, 100, 10, 20, SELL)
	testAddRemoveRandom(t, 100, 100, 10000, SELL)
}

func testAddRemoveRandom(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind) {
	bst := NewTree()
	validate(t, bst)
	orderMap := make(map[int64]*Order)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || bst.Size() == 0 {
			o := maker.MkPricedOrder(maker.Between(lowPrice, highPrice), kind)
			bst.Push(&o.GuidNode)
			validate(t, bst)
			orderMap[o.Guid] = o
			i++
		} else {
			for g, o := range orderMap {
				po := bst.Pop(g).O
				delete(orderMap, g)
				if po != o {
					t.Errorf("Bad pop")
				}
				validate(t, bst)
				break
			}
		}
	}
	drainTree(t, bst, orderMap)
}

func drainTree(t *testing.T, bst *Tree, orderMap map[int64]*Order) {
	for g := range orderMap {
		o := orderMap[g]
		po := bst.Pop(o.Guid).O
		if po != o {
			t.Errorf("Bad pop")
		}
		validate(t, bst)
	}
	if bst.Size() != 0 {
		t.Errorf("Expecting empty tree, got %d remaining orders", bst.Size())
	}
}

// Quick check to ensure the tree is as big as it says it is
func validate(t *testing.T, tree *Tree) {
	size := tree.Size()
	countedSize := countSize(t, tree.root)
	if size != countedSize {
		t.Errorf("Wrong size reported, reported %d, counted %d", size, countedSize)
	}
	if tree.root != nil {
		checkStructure(t, tree.root, size)
	}
}

func countSize(t *testing.T, n *Node) int {
	if n == nil {
		return 0
	}
	return countSize(t, n.left) + countSize(t, n.right) + countNodes(t, n)
}

func countNodes(t *testing.T, n *Node) int {
	count := 1
	curr := n.next
	for curr != n {
		curr = curr.next
		count++
	}
	if count != n.size {
		t.Errorf("Limit queue has inconsistent size. Expected %d, found %d", n.size, count)
	}
	return count
}

func checkStructure(t *testing.T, n *Node, size int) {
	if *n.pp != n {
		t.Errorf("Parent pointer does not point to child node %d", size)
	}
	if n.left != nil {
		if n.val <= n.left.val {
			t.Errorf("Left value is greater than or equal to node value. Left value: %d Node value %d", n.left.val, n.val)
		}
		checkStructure(t, n.left, size)
	}
	if n.right != nil {
		if n.val >= n.right.val {
			t.Errorf("Right value is less than or equal to node value. Right value: %d Node value %d", n.right.val, n.val)
		}
		checkStructure(t, n.right, size)
	}
}

// Function to pop and peek and check that everything is in order
func popCheck(t *testing.T, bst *Tree, q *prioq, popper popperFun) {
	peek, pop, check := popper(bst, q)
	if pop != check {
		t.Errorf("Mismatched push/pop pair")
		return
	}
	if pop != peek {
		t.Errorf("Mismatched peek/pop pair")
		return
	}
	validate(t, bst)
}

// Helper functions for popping either the max or the min from our queues
func maxPopper(bst *Tree, q *prioq) (peek, pop, check *Order) {
	peek = bst.PeekMax().O
	pop = bst.PopMax().O
	check = q.popMax()
	return
}

func minPopper(bst *Tree, q *prioq) (peek, pop, check *Order) {
	peek = bst.PeekMin().O
	pop = bst.PopMin().O
	check = q.popMin()
	return
}

// An easy to build priority queue
type prioq struct {
	prios               [][]*Order
	lowPrice, highPrice int64
}

func mkPrioq(size int, lowPrice, highPrice int64) *prioq {
	prios := make([][]*Order, highPrice-lowPrice+1)
	return &prioq{prios: prios, lowPrice: lowPrice, highPrice: highPrice}
}

func (q *prioq) push(o *Order) {
	idx := o.Price - q.lowPrice
	prio := q.prios[idx]
	prio = append(prio, o)
	q.prios[idx] = prio
}

func (q *prioq) popMax() *Order {
	if len(q.prios) == 0 {
		return nil
	}
	for i := len(q.prios) - 1; i >= 0; i-- {
		switch {
		case len(q.prios[i]) > 0:
			return q.pop(i)
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) popMin() *Order {
	if len(q.prios) == 0 {
		return nil
	}
	for i := 0; i < len(q.prios); i++ {
		switch {
		case len(q.prios[i]) > 0:
			return q.pop(i)
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) pop(i int) *Order {
	prio := q.prios[i]
	o := prio[0]
	prio = prio[1:]
	q.prios[i] = prio
	return o
}
