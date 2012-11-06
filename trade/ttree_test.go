package trade

import (
	"math/rand"
	"testing"
)

var limitOrderMaker = NewOrderMaker()

func TestPushPopSimpleMin(t *testing.T) {
	// buys
	testPushPopSimple(t, 100, 1, 1, BUY, maxPopper)
	testPushPopSimple(t, 10, 100, 200, BUY, maxPopper)
	testPushPopSimple(t, 100, 100, 10000, BUY, maxPopper)
	// sells
	testPushPopSimple(t, 100, 1, 1, SELL, minPopper)
	testPushPopSimple(t, 100, 100, 200, SELL, minPopper)
	testPushPopSimple(t, 100, 100, 10000, SELL, minPopper)
}

func testPushPopSimple(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind, popper popperFun) {
	bst := NewTree()
	validate(t, bst)
	q := mkPrioq(pushCount, lowPrice, highPrice)
	for i := 0; i < pushCount; i++ {
		o := limitOrderMaker.MkPricedOrder(limitOrderMaker.Between(lowPrice, highPrice), kind)
		bst.Push(&o.LimitNode)
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

/*
func TestRandomPushPop(t *testing.T) {
	// buys
	testPushPopRandom(t, 100, 1, 1, BUY, maxPopper)
	testPushPopRandom(t, 100, 100, 1000, BUY, maxPopper)
	testPushPopRandom(t, 100, 100, 1000, BUY, maxPopper)
	// sells
	testPushPopRandom(t, 100, 1, 1, SELL, minPopper)
	testPushPopRandom(t, 100, 100, 1000, SELL, minPopper)
	testPushPopRandom(t, 100, 100, 1000, SELL, minPopper)
}
*/

func testPushPopRandom(t *testing.T, pushCount int, lowPrice, highPrice int64, kind OrderKind, popper popperFun) {
	bst := NewTree()
	validate(t, bst)
	q := mkPrioq(pushCount, lowPrice, highPrice)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || bst.Size() == 0 {
			o := limitOrderMaker.MkPricedOrder(limitOrderMaker.Between(lowPrice, highPrice), kind)
			bst.Push(&o.LimitNode)
			validate(t, bst)
			q.push(o)
			i++
		} else {
			popCheck(t, bst, q, popper)
		}
	}
	for bst.Size() > 0 {
		fo := q.popMax()
		po := bst.PopMax().O
		if fo != po {
			t.Errorf("Mismatched Push/Pop pair")
			return
		}
		validate(t, bst)
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

// A function signature allowing us to switch easily between min and max queues
type popperFun func(*Tree, *prioq) (*Order, *Order, *Order)

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
	chans               []chan *Order
	lowPrice, highPrice int64
}

func mkPrioq(size int, lowPrice, highPrice int64) *prioq {
	chans := make([]chan *Order, highPrice-lowPrice+1)
	for i := range chans {
		chans[i] = make(chan *Order, size)
	}
	return &prioq{chans: chans, lowPrice: lowPrice, highPrice: highPrice}
}

func (q *prioq) push(o *Order) {
	idx := o.Price - q.lowPrice
	q.chans[idx] <- o
}

func (q *prioq) popMax() *Order {
	if len(q.chans) == 0 {
		return nil
	}
	for i := len(q.chans) - 1; i >= 0; i-- {
		select {
		case o := <-q.chans[i]:
			return o
		default:
			continue
		}
	}
	return nil
}

func (q *prioq) popMin() *Order {
	if len(q.chans) == 0 {
		return nil
	}
	for i := 0; i < len(q.chans); i++ {
		select {
		case o := <-q.chans[i]:
			return o
		default:
			continue
		}
	}
	return nil
}
