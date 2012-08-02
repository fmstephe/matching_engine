package matcher

import (
	"math/rand"
	"testing"
)

func verifyHeap(h *heap, t *testing.T) {
	verifyHeapRec(h, t, 0)
}

func verifyHeapRec(h *heap, t *testing.T, i int) {
	limits := h.limits
	n := h.heapLen()
	j1 := 2*i + 1
	j2 := 2*i + 2
	if j1 < n {
		if better(limits[j1], limits[i], h.buySell) {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %d", i, limits[i], j1, limits[j1])
			return
		}
		verifyHeapRec(h, t, j1)
	}
	if j2 < n {
		if better(limits[j2], limits[i], h.buySell) {
			t.Errorf("heap invariant invalidated [%d] = %d > [%d] = %d", i, limits[i], j1, limits[j2])
			return
		}
		verifyHeapRec(h, t, j2)
	}
}

func verifyLimit(lim *limit, price int64, t *testing.T) {
	if lim.head == nil {
		t.Errorf("Limit with no Orders found")
	}
	for order := lim.head; order != nil; order = order.next {
		if order.Price != price {
			t.Errorf("Limit, with price %d, contains order with price %d", price, order.Price)
		}
	}
}

func mkPricedBuy(price int64) *Order {
	return mkPricedOrder(price, BUY)
}

func mkPricedSell(price int64) *Order {
	return mkPricedOrder(price, SELL)
}

func mkPricedOrder(price int64, buySell TradeType) *Order {
	costData := CostData{Price: price, Amount: 1}
	tradeData := TradeData{TraderId: 1, TradeId: 1, StockId: 1}
	return NewOrder(costData, tradeData, nil, buySell)
}

func TestAllSameBuy(t *testing.T) {
	h := newHeap(BUY)
	for i := 20; i > 0; i-- {
		h.push(mkPricedBuy(1))
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
	h := newHeap(SELL)
	for i := 20; i > 0; i-- {
		h.push(mkPricedSell(1))
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
	h := newHeap(BUY)
	for i := int64(20); i > 0; i-- {
		h.push(mkPricedBuy(i))
	}
	verifyHeap(h, t)
	for i := int64(20); h.heapLen() > 0; i-- {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestDescendingSell(t *testing.T) {
	h := newHeap(SELL)
	for i := int64(20); i > 0; i-- {
		h.push(mkPricedSell(i))
	}
	verifyHeap(h, t)
	for i := int64(1); h.heapLen() > 0; i++ {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestAscendingBuy(t *testing.T) {
	h := newHeap(BUY)
	for i := int64(1); i <= 20; i++ {
		h.push(mkPricedBuy(i))
	}
	verifyHeap(h, t)
	for i := int64(20); h.heapLen() > 0; i-- {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestAscendingSell(t *testing.T) {
	h := newHeap(SELL)
	for i := int64(1); i <= 20; i++ {
		h.push(mkPricedSell(i))
	}
	verifyHeap(h, t)
	for i := int64(1); h.heapLen() > 0; i++ {
		x := h.pop()
		verifyHeap(h, t)
		if x.Price != i {
			t.Errorf("%d.th pop got %d; want %d", i, x, i)
		}
	}
}

func TestBuyRandomPushPop(t *testing.T) {
	h := newHeap(BUY)
	size := 10000
	priceRange := int64(500)
	priceBase := int64(1000)
	buys := make([]*Order, 0, size)
	for i := 0; i < size; i++ {
		b := mkPricedBuy(rand.Int63n(priceRange) + priceBase)
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
	h := newHeap(SELL)
	size := 10000
	priceRange := int64(500)
	priceBase := int64(1000)
	buys := make([]*Order, 0, size)
	for i := 0; i < size; i++ {
		b := mkPricedSell(rand.Int63n(priceRange) + priceBase)
		buys = append(buys, b)
		h.push(b)
		verifyHeap(h, t)
	}
	greatestPrice := int64(0)
	for i := 0; i < size; i++ {
		s := h.pop()
		if s.Price < greatestPrice {
			t.Errorf("Sell pop reveals out of order sell order")
		}
		greatestPrice = s.Price
		verifyHeap(h, t)
	}
}

/*
func _TestRemove0(t *testing.T) {
	h := newHeap(BUY)
	for i := 0; i < 10; i++ {
		h.push(elem(i))
	}
	verifyHeap(h, t)

	for h.heapLen() > 0 {
		i := h.heapLen() - 1
		x := h.Remove(i)
		if x != i {
			t.Errorf("Remove(%d) got %d; want %d", i, x, i)
		}
		verifyHeap(h, t)
	}
}

func TestRemove1(t *testing.T) {
	h := newHeap(BUY)
	for i := 0; i < 10; i++ {
		h.push(elem(i))
	}
	verifyHeap(h, t)

	for i := 0; h.heapLen() > 0; i++ {
		x := h.Remove(0)
		if x != i {
			t.Errorf("Remove(0) got %d; want %d", x, i)
		}
		verifyHeap(h, t)
	}
}

func TestRemove2(t *testing.T) {
	N := 10

	h := newHeap(BUY)
	for i := 0; i < N; i++ {
		h.push(elem(i))
	}
	verifyHeap(h, t)

	m := make(map[int]bool)
	for h.heapLen() > 0 {
		m[h.Remove((h.heapLen()-1)/2)] = true
		verifyHeap(h, t)
	}

	if len(m) != N {
		t.Errorf("len(m) = %d; want %d", len(m), N)
	}
	for i := 0; i < len(m); i++ {
		if !m[i] {
			t.Errorf("m[%d] doesn't exist", i)
		}
	}
}
*/
