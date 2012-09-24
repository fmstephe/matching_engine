package limitheap

import (
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var orderMaker = trade.NewOrderMaker()

func verifyHeap(h *H, t *testing.T) {
	verifyHeapRec(h, t, 0)
}

func verifyHeapRec(h *H, t *testing.T, i int) {
	limits := h.limits
	n := h.HLen()
	j1 := 2*i + 1
	j2 := 2*i + 2
	if j1 < n {
		if better(limits[j1], limits[i], h.buySell) {
			t.Errorf("H invariant invalidated [%d] = %d > [%d] = %d", i, limits[i], j1, limits[j1])
			return
		}
		verifyHeapRec(h, t, j1)
	}
	if j2 < n {
		if better(limits[j2], limits[i], h.buySell) {
			t.Errorf("H invariant invalidated [%d] = %d > [%d] = %d", i, limits[i], j1, limits[j2])
			return
		}
		verifyHeapRec(h, t, j2)
	}
}

func verifyLimit(lim *limit, price int32, t *testing.T) {
	if lim.head == nil {
		t.Errorf("Limit with no trade.Orders found")
	}
	for order := lim.head; order != nil; order = order.Next {
		if order.Price != price {
			t.Errorf("Limit, with price %d, contains order with price %d", price, order.Price)
		}
	}
}

func TestAllSameBuy(t *testing.T) {
	h := NewHeap(trade.BUY)
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
	h := NewHeap(trade.SELL)
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
	h := NewHeap(trade.BUY)
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
	h := NewHeap(trade.SELL)
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
	h := NewHeap(trade.BUY)
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
	h := NewHeap(trade.SELL)
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
	h := NewHeap(trade.BUY)
	size := 10000
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
	h := NewHeap(trade.SELL)
	size := 10000
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

func TestRemoveBuy(t *testing.T) {
	testSimpleRemove(t, trade.BUY)
}

func TestRemoveSell(t *testing.T) {
	testSimpleRemove(t, trade.SELL)
}

func testSimpleRemove(t *testing.T, buySell trade.TradeType) {
	h := NewHeap(buySell)
	size := int32(10)
	orders := make([]*trade.Order, 0, size)
	for i := int32(1); i <= size; i++ {
		order := orderMaker.MkPricedOrder(i, buySell)
		h.Push(order)
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
	h := NewHeap(trade.BUY)
	size := int32(10)
	buys := make([]*trade.Order, 0, size)
	for i := size; i > 0; i-- {
		b := orderMaker.MkPricedBuy(i)
		h.Push(b)
		buys = append(buys, b)
		verifyHeap(h, t)
	}
	for i, buy := range buys {
		var b *trade.Order
		if i%2 == 0 {
			b = h.remove(buy.GUID(), buy.Price)
		} else {
			b = h.Pop()
		}
		if b != buy {
			t.Errorf("Remove() got %v; wanted %v", b, buy)
		}
		verifyHeap(h, t)
	}
}

func TestRemovePopSell(t *testing.T) {
	h := NewHeap(trade.SELL)
	size := int32(10)
	sells := make([]*trade.Order, 0, size)
	for i := int32(1); i <= size; i++ {
		s := orderMaker.MkPricedSell(i)
		h.Push(s)
		sells = append(sells, s)
		verifyHeap(h, t)
	}
	for i, sell := range sells {
		var removed *trade.Order
		if i%2 == 0 {
			removed = h.remove(sell.GUID(), sell.Price)
		} else {
			removed = h.Pop()
		}
		if removed != sell {
			t.Errorf("Remove() got %v; wanted %v", removed, sell)
		}
		verifyHeap(h, t)
	}
}
