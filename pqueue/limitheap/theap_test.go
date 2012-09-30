package limitheap

import (
	"github.com/fmstephe/matching_engine/pqueue"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var orderMaker = trade.NewOrderMaker()

func verifyHeap(t *testing.T, h pqueue.Q) {
	//verifyHeapRec(h.(*H), t, 0)
}
/*
func verifyHeapRec(h *H, t *testing.T, i int) {
	limits := h.limits
	n := len(h.limits)
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
*/
func createHeap(buySell trade.TradeType) pqueue.Q {
	return New(buySell, 100*1000)
}

func TestPushPop(t *testing.T) {
	pqueue.PushPopSuite(t, createHeap, verifyHeap)
}
