package trade

import (
	"math/rand"
	"testing"
)

var limitOrderMaker = NewOrderMaker()

func ensureSize(l *Limit, t *testing.T) {
	count := 0
	o := l.dummy.Lower
	for o != &l.dummy {
		count++
		o = o.Lower
	}
	if count != l.Size {
		t.Errorf("Mis-sized limit following Lower. Expecting %d, found %d", l.Size, count)
	}
	count = 0
	o = l.dummy.Higher
	for o != &l.dummy {
		count++
		o = o.Higher
	}
	if count != l.Size {
		t.Errorf("Mis-sized limit following Higher. Expecting %d, found %d", l.Size, count)
	}
}

func TestPushThenPop(t *testing.T) {
	testPushThenPop(t, 1)
	testPushThenPop(t, 2)
	testPushThenPop(t, 3)
	testPushThenPop(t, 15)
	testPushThenPop(t, 97)
	testPushThenPop(t, 333)
	testPushThenPop(t, 1024)
}

func testPushThenPop(t *testing.T, pushCount int) {
	l := NewLimit(1)
	ensureSize(l, t)
	orders := make([]*Order, pushCount)
	for i := 0; i < pushCount; i++ {
		o := limitOrderMaker.MkPricedBuy(1)
		orders[i] = o
		l.Push(o)
		ensureSize(l, t)
	}
	for i := 0; i < pushCount; i++ {
		pe := l.Peek()
		po := l.Pop()
		if po != orders[i] {
			t.Errorf("Mismatched push/pop pair")
			return
		}
		if po != pe {
			t.Errorf("Mismatched peek/pop pair")
			return
		}
		ensureSize(l, t)
	}
	if l.Size != 0 {
		t.Errorf("Expecting empty limit, got %d remaining orders", l.Size)
	}
}

func TestRandomPushPop(t *testing.T) {
	testRandomPushPop(t, 1)
	testRandomPushPop(t, 2)
	testRandomPushPop(t, 3)
	testRandomPushPop(t, 15)
	testRandomPushPop(t, 97)
	testRandomPushPop(t, 333)
	testRandomPushPop(t, 1024)
}

func testRandomPushPop(t *testing.T, pushCount int) {
	l := NewLimit(1)
	ensureSize(l, t)
	fifo := make(chan *Order, pushCount)
	r := rand.New(rand.NewSource(1))
	for i := 0; i < pushCount; {
		n := r.Int()
		if n%2 == 0 || l.Size == 0 {
			o := limitOrderMaker.MkPricedBuy(1)
			fifo <- o
			l.Push(o)
			ensureSize(l, t)
			i++
		} else {
			fo := <-fifo
			pe := l.Peek()
			po := l.Pop()
			if fo != po {
				t.Errorf("Mismatched Push/Pop pair")
				return
			}
			if pe != po {
				t.Errorf("Mismatched Peek/Pop pair")
				return
			}
			ensureSize(l, t)
		}
	}
	for l.Size > 0 {
		fo := <-fifo
		po := l.Pop()
		if fo != po {
			t.Errorf("Mismatched Push/Pop pair")
			return
		}
		ensureSize(l, t)
	}
}

func TestWrongPricedOrder(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Mis-priced order allowed in limit")
		}
	}()
	l := NewLimit(1)
	o := limitOrderMaker.MkPricedBuy(2)
	l.Push(o)
}
