package limitheap

import (
	"github.com/fmstephe/matching_engine/prioq"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var orderMaker = trade.NewOrderMaker()

func verifyHeap(t *testing.T, h prioq.Q) {
	verifyHeapRec(h.(*H), t, 0)
}

func verifyHeapRec(h *H, t *testing.T, i int) {
	heap := h.heap
	n := len(h.heap)
	j1 := 2*i + 1
	j2 := 2*i + 2
	if j1 < n {
		if better(heap[j1], heap[i], h.kind) {
			t.Errorf("H invariant invalidated [%d] = %d > [%d] = %d", i, heap[i], j1, heap[j1])
			return
		}
		verifyHeapRec(h, t, j1)
	}
	if j2 < n {
		if better(heap[j2], heap[i], h.kind) {
			t.Errorf("H invariant invalidated [%d] = %d > [%d] = %d", i, heap[i], j1, heap[j2])
			return
		}
		verifyHeapRec(h, t, j2)
	}
}

func createHeap(kind trade.OrderKind) prioq.Q {
	return New(kind, 10, 10)
}

func TestPushPop(t *testing.T) {
	prioq.PushPopSuite(t, createHeap, verifyHeap)
}
