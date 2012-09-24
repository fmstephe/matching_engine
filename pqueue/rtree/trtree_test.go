package rtree

import (
	"fmt"
	"testing"
	"github.com/fmstephe/matching_engine/trade"
)

var orderMaker = trade.NewOrderMaker()

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
	h := newHeap(trade.BUY)
	for i := 20; i > 0; i-- {
		h.Push(orderMaker.MkPricedBuy(1))
	}
	for i := 1; h.heapLen() > 0; i++ {
		x := h.Pop()
		if x.Price != 1 {
			t.Errorf("%d.th Pop got %d; want %d", i, x, 0)
		}
	}
}

func TestAllSameSell(t *testing.T) {
	h := newHeap(trade.SELL)
	for i := 20; i > 0; i-- {
		h.Push(orderMaker.MkPricedSell(1))
	}
	for i := 1; h.heapLen() > 0; i++ {
		x := h.Pop()
		if x.Price != 1 {
			t.Errorf("%d.th Pop got %d; want %d", i, x, 0)
		}
	}
}

func TestDescendingBuy(t *testing.T) {
	h := newHeap(trade.BUY)
	for i := int32(20); i > 0; i-- {
		h.Push(orderMaker.MkPricedBuy(i))
	}
	for i := int32(20); h.heapLen() > 0; i-- {
		x := h.Pop()
		if x.Price != i {
			t.Errorf("%d.th Pop got %d; want %d", i, x, i)
		}
	}
}

func TestDescendingSell(t *testing.T) {
	h := newHeap(trade.SELL)
	for i := int32(20); i > 0; i-- {
		h.Push(orderMaker.MkPricedSell(i))
	}
	for i := int32(1); h.heapLen() > 0; i++ {
		x := h.Pop()
		if x.Price != i {
			t.Errorf("%d.th Pop got %d; want %d", i, x, i)
		}
	}
}

func TestAscendingBuy(t *testing.T) {
	h := newHeap(trade.BUY)
	for i := int32(1); i <= 20; i++ {
		h.Push(orderMaker.MkPricedBuy(i))
	}
	for i := int32(20); h.heapLen() > 0; i-- {
		x := h.Pop()
		if x.Price != i {
			t.Errorf("%d.th Pop got %d; want %d", i, x, i)
		}
	}
}

func TestAscendingSell(t *testing.T) {
	h := newHeap(trade.SELL)
	for i := int32(1); i <= 20; i++ {
		h.Push(orderMaker.MkPricedSell(i))
	}
	for i := int32(1); h.heapLen() > 0; i++ {
		x := h.Pop()
		if x.Price != i {
			t.Errorf("%d.th Pop got %d; want %d", i, x, i)
		}
	}
}

func TestBuyRandomPushPop(t *testing.T) {
	h := newHeap(trade.BUY)
	size := 1000 * 1000
	priceRange := int32(500)
	priceBase := int32(1000)
	buys := make([]*trade.Order, 0, size)
	for i := 0; i < size; i++ {
		b := orderMaker.MkPricedBuy(orderMaker.Rand32(priceRange) + priceBase)
		buys = append(buys, b)
		h.Push(b)
	}
	leastPrice := priceRange + priceBase + 1
	for i := 0; i < size; i++ {
		b := h.Pop()
		if b.Price > leastPrice {
			t.Errorf("Buy Pop reveals out of order buy order")
		}
		leastPrice = b.Price
	}
}

func TestSellRandomPushPop(t *testing.T) {
	h := newHeap(trade.SELL)
	size := 1000 * 1000
	priceRange := int32(500)
	priceBase := int32(1000)
	buys := make([]*trade.Order, 0, size)
	for i := 0; i < size; i++ {
		b := orderMaker.MkPricedSell(orderMaker.Rand32(priceRange) + priceBase)
		buys = append(buys, b)
		h.Push(b)
	}
	greatestPrice := int32(0)
	for i := 0; i < size; i++ {
		s := h.Pop()
		if s.Price < greatestPrice {
			t.Errorf("Sell Pop reveals out of order sell order")
		}
		greatestPrice = s.Price
	}
}
