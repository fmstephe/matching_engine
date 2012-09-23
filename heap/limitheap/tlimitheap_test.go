package limitheap

import (
	"testing"
)

var heapTestOrderMaker = newOrderMaker()

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

func verifyLimit(lim *limit, price int32, t *testing.T) {
	if lim.head == nil {
		t.Errorf("Limit with no Orders found")
	}
	for order := lim.head; order != nil; order = order.next {
		if order.Price != price {
			t.Errorf("Limit, with price %d, contains order with price %d", price, order.Price)
		}
	}
}

func TestAllSameBuy(t *testing.T) {
	h := newHeap(BUY)
	for i := 20; i > 0; i-- {
		h.push(heapTestOrderMaker.mkPricedBuy(1))
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
		h.push(heapTestOrderMaker.mkPricedSell(1))
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
	for i := int32(20); i > 0; i-- {
		h.push(heapTestOrderMaker.mkPricedBuy(i))
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
	h := newHeap(SELL)
	for i := int32(20); i > 0; i-- {
		h.push(heapTestOrderMaker.mkPricedSell(i))
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
	h := newHeap(BUY)
	for i := int32(1); i <= 20; i++ {
		h.push(heapTestOrderMaker.mkPricedBuy(i))
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
	h := newHeap(SELL)
	for i := int32(1); i <= 20; i++ {
		h.push(heapTestOrderMaker.mkPricedSell(i))
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
	h := newHeap(BUY)
	size := 10000
	priceRange := int32(500)
	priceBase := int32(1000)
	buys := make([]*Order, 0, size)
	for i := 0; i < size; i++ {
		b := heapTestOrderMaker.mkPricedBuy(rand32(priceRange) + priceBase)
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
	priceRange := int32(500)
	priceBase := int32(1000)
	buys := make([]*Order, 0, size)
	for i := 0; i < size; i++ {
		b := heapTestOrderMaker.mkPricedSell(rand32(priceRange) + priceBase)
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

func TestRemoveBuy(t *testing.T) {
	testSimpleRemove(t, BUY)
}

func TestRemoveSell(t *testing.T) {
	testSimpleRemove(t, SELL)
}

func testSimpleRemove(t *testing.T, buySell TradeType) {
	h := newHeap(buySell)
	size := int32(10)
	orders := make([]*Order, 0, size)
	for i := int32(1); i <= size; i++ {
		order := heapTestOrderMaker.mkPricedOrder(i, buySell)
		h.push(order)
		orders = append(orders, order)
		verifyHeap(h, t)
	}
	for _, order := range orders {
		removed := h.remove(order.GUID(), order.Price)
		if removed != order {
			t.Errorf("Remove() got %v; wanted %v", removed, order)
		}
		verifyHeap(h, t)
	}
}

func TestRemovePopBuy(t *testing.T) {
	h := newHeap(BUY)
	size := int32(10)
	buys := make([]*Order, 0, size)
	for i := size; i > 0; i-- {
		b := heapTestOrderMaker.mkPricedBuy(i)
		h.push(b)
		buys = append(buys, b)
		verifyHeap(h, t)
	}
	for i, buy := range buys {
		var b *Order
		if i%2 == 0 {
			b = h.remove(buy.GUID(), buy.Price)
		} else {
			b = h.pop()
		}
		if b != buy {
			t.Errorf("Remove() got %v; wanted %v", b, buy)
		}
		verifyHeap(h, t)
	}
}

func TestRemovePopSell(t *testing.T) {
	h := newHeap(SELL)
	size := int32(10)
	sells := make([]*Order, 0, size)
	for i := int32(1); i <= size; i++ {
		s := heapTestOrderMaker.mkPricedSell(i)
		h.push(s)
		sells = append(sells, s)
		verifyHeap(h, t)
	}
	for i, sell := range sells {
		var removed *Order
		if i%2 == 0 {
			removed = h.remove(sell.GUID(), sell.Price)
		} else {
			removed = h.pop()
		}
		if removed != sell {
			t.Errorf("Remove() got %v; wanted %v", removed, sell)
		}
		verifyHeap(h, t)
	}
}
