package binheap

import (
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var orderMaker = trade.NewOrderMaker()

func verifyHeap(h *heap, t *testing.T) {
	verifyHeapRec(h, t, 0)
}

func verifyHeapRec(h *heap, t *testing.T, i int) {
	orders := h.orders
	n := h.heapLen()
	j1 := 2*i + 1
	j2 := 2*i + 2
	if j1 < n {
		if better(orders[j1], orders[i]) {
			t.Errorf("heap invariant inValidated [%d] = %d > [%d] = %d", i, orders[i], j1, orders[j1])
			return
		}
		verifyHeapRec(h, t, j1)
	}
	if j2 < n {
		if better(orders[j2], orders[i]) {
			t.Errorf("heap invariant inValidated [%d] = %d > [%d] = %d", i, orders[i], j1, orders[j2])
			return
		}
		verifyHeapRec(h, t, j2)
	}
}

func TestAllSameBuy(t *testing.T) {
	h := NewHeap(trade.BUY, 40)
	for i := 20; i > 0; i-- {
		h.push(orderMaker.MkPricedBuy(1))
	}
	verifyHeap(h, t)
	for i := 1; h.heapLen() > 0; i++ {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != 1 {
			t.Errorf("%d.th pop got %d; want %d", i, x, 0)
		}
	}
}

func TestAllSameSell(t *testing.T) {
	h := NewHeap(trade.SELL, 40)
	for i := 20; i > 0; i-- {
		h.push(orderMaker.MkPricedSell(1))
	}
	verifyHeap(h, t)
	for i := 1; h.heapLen() > 0; i++ {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != 1 {
			t.Errorf("%d.th pop got %d; want %d", i, x, 0)
		}
	}
}

func TestDescendingBuy(t *testing.T) {
	h := NewHeap(trade.BUY, 40)
	for i := int32(20); i > 0; i-- {
		h.push(orderMaker.MkPricedBuy(i))
	}
	verifyHeap(h, t)
	for i := int32(20); h.heapLen() > 0; i-- {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestDescendingSell(t *testing.T) {
	h := NewHeap(trade.SELL, 40)
	for i := int32(20); i > 0; i-- {
		h.push(orderMaker.MkPricedSell(i))
	}
	verifyHeap(h, t)
	for i := int32(1); h.heapLen() > 0; i++ {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestAscendingBuy(t *testing.T) {
	h := NewHeap(trade.BUY, 40)
	for i := int32(1); i <= 20; i++ {
		h.push(orderMaker.MkPricedBuy(i))
	}
	verifyHeap(h, t)
	for i := int32(20); h.heapLen() > 0; i-- {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestAscendingSell(t *testing.T) {
	h := NewHeap(trade.SELL, 40)
	for i := int32(1); i <= 20; i++ {
		h.push(orderMaker.MkPricedSell(i))
	}
	verifyHeap(h, t)
	for i := int32(1); h.heapLen() > 0; i++ {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestBuyRandomPushPop(t *testing.T) {
	size := 10000
	h := NewHeap(trade.BUY, size)
	priceRange := int32(500)
	priceBase := int32(1000)
	buys := make([]*trade.Order, 0, size)
	for i := 0; i < size; i++ {
		b := orderMaker.MkPricedBuy(orderMaker.Rand32(priceRange) + priceBase)
		buys = append(buys, b)
		h.push(b)
		verifyHeap(h, t)
	}
	leastPrice := priceRange + priceBase + 1
	for i := 0; i < size; i++ {
		b := h.pop()
		if b.Price > leastPrice {
			t.Errorf("Buy pop reveals out of order buy order")
		}
		leastPrice = b.Price
		verifyHeap(h, t)
	}
}

func TestSellRandomPushPop(t *testing.T) {
	size := 10000
	h := NewHeap(trade.SELL, size)
	priceRange := int32(500)
	priceBase := int32(1000)
	buys := make([]*trade.Order, 0, size)
	for i := 0; i < size; i++ {
		b := orderMaker.MkPricedSell(orderMaker.Rand32(priceRange) + priceBase)
		buys = append(buys, b)
		h.push(b)
		verifyHeap(h, t)
	}
	greatestPrice := int32(0)
	for i := 0; i < size; i++ {
		s := h.pop()
		if s.Price < greatestPrice {
			t.Errorf("Sell pop reveals out of order sell order")
		}
		greatestPrice = s.Price
		verifyHeap(h, t)
	}
}
