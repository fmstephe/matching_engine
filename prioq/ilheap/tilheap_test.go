package ilheap

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

// NB: This func is covering up the fact that ilheap does not currently resize its internal slice
func createHeap(kind trade.OrderKind) prioq.Q {
	return New(kind, 10*1000)
}

func TestPushPop(t *testing.T) {
	prioq.PushPopSuite(t, createHeap, verifyHeap)
}
