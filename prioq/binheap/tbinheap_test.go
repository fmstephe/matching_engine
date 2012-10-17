package binheap

import (
	"github.com/fmstephe/matching_engine/prioq"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

func verifyHeap(t *testing.T, h prioq.Q) {
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

func createHeap(kind trade.OrderKind) prioq.Q {
	return New(kind, 100)
}

func TestPushPop(t *testing.T) {
	prioq.PushPopSuite(t, createHeap, verifyHeap)
}
