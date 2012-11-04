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

func TestPushThenPopOnePrice(t *testing.T) {
	testPushThenPopOnePrice(t, 1)
	testPushThenPopOnePrice(t, 2)
	testPushThenPopOnePrice(t, 3)
	testPushThenPopOnePrice(t, 15)
	testPushThenPopOnePrice(t, 97)
	testPushThenPopOnePrice(t, 333)
	testPushThenPopOnePrice(t, 1024)
}

func testPushThenPopOnePrice(t *testing.T, pushCount int) {
	bst := NewTree()
	ensureSize(bst, t)
	fifo := make(chan *Order, pushCount)
	for i := 0; i < pushCount; i++ {
		o := limitOrderMaker.MkPricedBuy(1)
		fifo <- o
		bst.Push(&o.LimitNode)
		ensureSize(bst, t)
	}
	for i := 0; i < pushCount; i++ {
		pe := bst.PeekMax().O
		po := bst.PopMax().O
		fo := <-fifo
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

func TestRandomPushPopOnePrice(t *testing.T) {
	testRandomPushPopOnePrice(t, 1)
	testRandomPushPopOnePrice(t, 2)
	testRandomPushPopOnePrice(t, 3)
	testRandomPushPopOnePrice(t, 15)
	testRandomPushPopOnePrice(t, 97)
	testRandomPushPopOnePrice(t, 333)
	testRandomPushPopOnePrice(t, 1024)
}

func testRandomPushPopOnePrice(t *testing.T, pushCount int) {
	bst := NewTree()
	ensureSize(bst, t)
	fifo := make(chan *Order, pushCount)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || bst.Size() == 0 {
			o := limitOrderMaker.MkPricedBuy(1)
			fifo <- o
			bst.Push(&o.LimitNode)
			ensureSize(bst, t)
			i++
		} else {
			fo := <-fifo
			pe := bst.PeekMax().O
			po := bst.PopMax().O
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
		fo := <-fifo
		po := bst.PopMax().O
		if fo != po {
			t.Errorf("Mismatched Push/Pop pair")
			return
		}
		ensureSize(bst, t)
	}
}

func TestRandomPushPopBuy(t *testing.T) {
	size := 3
	priceRange := int64(500)
	priceBase := int64(1000)
	bst := NewTree()
	for i := 0; i < size; i++ {
		b := limitOrderMaker.MkPricedBuy(limitOrderMaker.Rand32(priceRange) + priceBase)
		bst.Push(&b.LimitNode)
		if bst.Size() != (i + 1) {
			t.Errorf("Incorrect size found in RandomPushPopBuy push phase. Expecting %d, got %d instead", i+1, bst.Size())
		}
	}
	leastPrice := priceRange + priceBase + 1
	for i := 0; i < size; i++ {
		b := bst.PopMax().O
		if bst.Size() != size-(i+1) {
			t.Errorf("Incorrect size found in RandomPushPopBuy pop phase. Expecting %d, got %d instead", size-(i+1), bst.Size())
		}
		if b.Price > leastPrice {
			t.Errorf("Buy Pop reveals out of order buy order")
		}
		leastPrice = b.Price
	}
}
