package trade

import (
	"math/rand"
	"testing"
)

var limitOrderMaker = NewOrderMaker()

func ensureSize(tree *Tree, t *testing.T) {
	size := tree.Size()
	countedSize := countSize(tree.root)
	if size != countedSize {
		t.Errorf("Wrong size reported, reported %d, counted %d", size, countedSize)
	}
}

func countSize(n *Node) int {
	if n == nil {
		return 0
	}
	return countSize(n.left) + countSize(n.right) + countNodes(n)
}

func countNodes(n *Node) int {
	count := 1
	curr := n.next
	for curr != n {
		curr = curr.next
		count++
	}
	return count
}

func TestRandomPushPopBuy(t *testing.T) {
	size := 1000
	lowPrice := int64(1500)
	highPrice := int64(2000)
	bst := NewTree()
	q := pushOrders(t, bst, size, lowPrice, highPrice, BUY)
	leastPrice := highPrice + 1
	for i := 0; i < size; i++ {
		o1 := bst.PopMax().O
		o2 := q.popMax()
		if o1 != o2 {
			t.Errorf("Mismatched push/pop found")
		}
		if bst.Size() != size-(i+1) {
			t.Errorf("Incorrect size found in RandomPushPopBuy pop phase. Expecting %d, got %d instead", size-(i+1), bst.Size())
		}
		if o1.Price > leastPrice {
			t.Errorf("Buy Pop reveals out of order buy order")
		}
		leastPrice = o1.Price
	}
}

func TestRandomPushPopSell(t *testing.T) {
	size := 1000
	lowPrice := int64(1500)
	highPrice := int64(2000)
	bst := NewTree()
	q := pushOrders(t, bst, size, lowPrice, highPrice, SELL)
	greatestPrice := int64(0)
	for i := 0; i < size; i++ {
		o1 := bst.PopMin().O
		o2 := q.popMin()
		if o1 != o2 {
			t.Errorf("Mismatched push/pop found")
		}
		if bst.Size() != size-(i+1) {
			t.Errorf("Incorrect size found in RandomPushPopBuy pop phase. Expecting %d, got %d instead", size-(i+1), bst.Size())
		}
		if o1.Price < greatestPrice {
			t.Errorf("Buy Pop reveals out of order sell order")
		}
		greatestPrice = o1.Price
	}
}

func TestPushThenPopOnePrice(t *testing.T) {
	testPushPopSimple(t, 100, 1, 1)
	testPushPopSimple(t, 100, 100, 1000)
	testPushPopSimple(t, 1000, 100, 1000)
}

func testPushPopSimple(t *testing.T, pushCount int, lowPrice, highPrice int64) {
	bst := NewTree()
	ensureSize(bst, t)
	q := pushOrders(t, bst, pushCount, lowPrice, highPrice, BUY)
	for i := 0; i < pushCount; i++ {
		pe := bst.PeekMax().O
		po := bst.PopMax().O
		fo := q.popMax()
		if po != fo {
			t.Errorf("Mismatched push/pop pair, after popping %d of %d orders", i+1, pushCount)
			return
		}
		if po != pe {
			t.Errorf("Mismatched peek/pop pair, after popping %d of %d orders", i+1, pushCount)
			return
		}
		ensureSize(bst, t)
	}
	if bst.Size() != 0 {
		t.Errorf("Expecting empty limit, got %d remaining orders, after pushing %d orders", bst.Size(), pushCount)
	}
}

func TestRandomPushPop(t *testing.T) {
	testPushPopRandom(t, 100, 1, 1)
	testPushPopRandom(t, 100, 100, 1000)
	testPushPopRandom(t, 1000, 100, 1000)
}

func testPushPopRandom(t *testing.T, pushCount int, lowPrice, highPrice int64) {
	bst := NewTree()
	ensureSize(bst, t)
	q := mkPrioq(pushCount, lowPrice, highPrice)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || bst.Size() == 0 {
			o := limitOrderMaker.MkPricedBuy(limitOrderMaker.Between(lowPrice, highPrice))
			bst.Push(&o.LimitNode)
			q.push(o)
			ensureSize(bst, t)
			i++
		} else {
			pe := bst.PeekMax().O
			po := bst.PopMax().O
			fo := q.popMax()
			if fo != po {
				t.Errorf("Mismatched Push/Pop pair")
				return
			}
			if pe != po {
				t.Errorf("Mismatched Peek/Pop pair")
				return
			}
			ensureSize(bst, t)
		}
	}
	for bst.Size() > 0 {
		fo := q.popMax()
		po := bst.PopMax().O
		if fo != po {
			t.Errorf("Mismatched Push/Pop pair")
			return
		}
		ensureSize(bst, t)
	}
}

func pushOrders(t *testing.T, bst *Tree, size int, lowPrice, highPrice int64, kind OrderKind) *prioq {
	q := mkPrioq(size, lowPrice, highPrice)
	for i := 0; i < size; i++ {
		o := limitOrderMaker.MkPricedOrder(limitOrderMaker.Between(lowPrice, highPrice), kind)
		bst.Push(&o.LimitNode)
		q.push(o)
		if bst.Size() != (i + 1) {
			t.Errorf("Incorrect size found in RandomPushPopBuy push phase. Expecting %d, got %d instead", i+1, bst.Size())
		}
	}
	return q
}

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
