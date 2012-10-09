package rtree

import (
	"github.com/fmstephe/matching_engine/pqueue"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var orderMaker = trade.NewOrderMaker()

func verifyHeap(t *testing.T, h pqueue.Q) {
}

func createHeap(kind trade.OrderKind) pqueue.Q {
	return New(kind)
}

func TestPushPop(t *testing.T) {
	pqueue.PushPopSuite(t, createHeap, verifyHeap)
}
