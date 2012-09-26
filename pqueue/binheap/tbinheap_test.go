package binheap

import (
	"github.com/fmstephe/matching_engine/trade"
	"github.com/fmstephe/matching_engine/pqueue"
	"testing"
)

func verifyHeap(t *testing.T, h pqueue.Q) {
	verifyHeapRec(t, h.(*H), 0)
}

func verifyHeapRec(t *testing.T, h *H, i int) {
	orders := h.orders
	n := h.Size()
	j1 := 2*i + 1
	j2 := 2*i + 2
	if j1 < n {
		if better(orders[j1], orders[i]) {
			t.Errorf("H invariant inValidated [%d] = %d > [%d] = %d", i, orders[i], j1, orders[j1])
			return
		}
		verifyHeapRec(t, h, j1)
	}
	if j2 < n {
		if better(orders[j2], orders[i]) {
			t.Errorf("H invariant inValidated [%d] = %d > [%d] = %d", i, orders[i], j1, orders[j2])
			return
		}
		verifyHeapRec(t, h, j2)
	}
}

func TestAllSameBuy(t *testing.T) {
	h := New(trade.BUY, 40)
	pqueue.AllSameBuy(t, h, verifyHeap)
}

func TestAllSameSell(t *testing.T) {
	h := New(trade.SELL, 40)
	pqueue.AllSameSell(t, h, verifyHeap)
}

func TestDescendingBuy(t *testing.T) {
	h := New(trade.BUY, 40)
	pqueue.DescendingBuy(t, h, verifyHeap)
}

func TestDescendingSell(t *testing.T) {
	h := New(trade.SELL, 40)
	pqueue.DescendingSell(t, h, verifyHeap)
}

func TestAscendingBuy(t *testing.T) {
	h := New(trade.BUY, 40)
	pqueue.AscendingBuy(t, h, verifyHeap)
}

func TestAscendingSell(t *testing.T) {
	h := New(trade.SELL, 40)
	pqueue.AscendingSell(t, h, verifyHeap)
}

func TestBuyRandomPushPop(t *testing.T) {
	h := New(trade.BUY, 1000)
	pqueue.BuyRandomPushPop(t, h, verifyHeap)
}

func TestSellRandomPushPop(t *testing.T) {
	h := New(trade.SELL, 1000)
	pqueue.SellRandomPushPop(t, h, verifyHeap)
}
