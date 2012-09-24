package ilheap

import (
	"testing"
	"github.com/fmstephe/matching_engine/trade"
)

var orderMaker = trade.NewOrderMaker()

func verifyHeap(h *H, t *testing.T) {
	verifyHeapRec(h, t, 0)
}

func verifyHeapRec(h *H, t *testing.T, i int) {
	elems := h.elems
	n := h.idx
	j1 := 2*i + 1
	j2 := 2*i + 2
	if j1 < n {
		if elems[j1].val > elems[i].val {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %d", i, elems[i], j1, elems[j1])
			return
		}
		verifyHeapRec(h, t, j1)
	}
	if j2 < n {
		if elems[j2].val > elems[i].val {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %d", i, elems[i], j1, elems[j2])
			return
		}
		verifyHeapRec(h, t, j2)
	}
}

func TestAllSameBuy(t *testing.T) {
	h := newHeap(trade.BUY, 40)
	for i := 20; i > 0; i-- {
		h.Push(orderMaker.MkPricedBuy(1))
	}
	verifyHeap(h, t)
	for i := 1; h.HLen() > 0; i++ {
		x := h.Pop()
		verifyHeap(h, t)
		if x.Price != 1 {
			t.Errorf("%d.th Pop got %d; want %d", i, x, 0)
		}
	}
}

func TestAllSameSell(t *testing.T) {
	h := newHeap(trade.SELL, 40)
	for i := 20; i > 0; i-- {
		h.Push(orderMaker.MkPricedSell(1))
	}
	verifyHeap(h, t)
	for i := 1; h.HLen() > 0; i++ {
		x := h.Pop()
		verifyHeap(h, t)
		if x.Price != 1 {
			t.Errorf("%d.th Pop got %d; want %d", i, x, 0)
		}
	}
}

func TestDescendingBuy(t *testing.T) {
	h := newHeap(trade.BUY, 40)
	for i := int32(20); i > 0; i-- {
		h.Push(orderMaker.MkPricedBuy(i))
	}
	verifyHeap(h, t)
	for i := int32(20); h.HLen() > 0; i-- {
		x := h.Pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th Pop got %d; want %d", i, x, i)
		}
	}
}

func TestDescendingSell(t *testing.T) {
	h := newHeap(trade.SELL, 40)
	for i := int32(20); i > 0; i-- {
		h.Push(orderMaker.MkPricedSell(i))
	}
	verifyHeap(h, t)
	for i := int32(1); h.HLen() > 0; i++ {
		x := h.Pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th Pop got %d; want %d", i, x, i)
		}
	}
}

func TestAscendingBuy(t *testing.T) {
	h := newHeap(trade.BUY, 40)
	for i := int32(1); i <= 20; i++ {
		h.Push(orderMaker.MkPricedBuy(i))
	}
	verifyHeap(h, t)
	for i := int32(20); h.HLen() > 0; i-- {
		x := h.Pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th Pop got %d; want %d", i, x, i)
		}
	}
}

func TestAscendingSell(t *testing.T) {
	h := newHeap(trade.SELL, 40)
	for i := int32(1); i <= 20; i++ {
		h.Push(orderMaker.MkPricedSell(i))
	}
	verifyHeap(h, t)
	for i := int32(1); h.HLen() > 0; i++ {
		x := h.Pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th Pop got %d; want %d", i, x, i)
		}
	}
}

func TestBuyRandomPushPop(t *testing.T) {
	size := 10000
	h := newHeap(trade.BUY, size)
	priceRange := int32(500)
	priceBase := int32(1000)
	buys := make([]*trade.Order, 0, size)
	for i := 0; i < size; i++ {
		b := orderMaker.MkPricedBuy(orderMaker.Rand32(priceRange) + priceBase)
		buys = append(buys, b)
		h.Push(b)
		verifyHeap(h, t)
	}
	leastPrice := priceRange + priceBase + 1
	for i := 0; i < size; i++ {
		b := h.Pop()
		if b.Price > leastPrice {
			t.Errorf("Buy Pop reveals out of order buy order")
		}
		leastPrice = b.Price
		verifyHeap(h, t)
	}
}

func TestSellRandomPushPop(t *testing.T) {
	size := 10000
	h := newHeap(trade.SELL, size)
	priceRange := int32(500)
	priceBase := int32(1000)
	buys := make([]*trade.Order, 0, size)
	for i := 0; i < size; i++ {
		b := orderMaker.MkPricedSell(orderMaker.Rand32(priceRange) + priceBase)
		buys = append(buys, b)
		h.Push(b)
		verifyHeap(h, t)
	}
	greatestPrice := int32(0)
	for i := 0; i < size; i++ {
		s := h.Pop()
		if s.Price < greatestPrice {
			t.Errorf("Sell Pop reveals out of order sell order")
		}
		greatestPrice = s.Price
		verifyHeap(h, t)
	}
}
