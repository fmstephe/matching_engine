package rtree

import (
	"github.com/fmstephe/matching_engine/prioq"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var orderMaker = trade.NewOrderMaker()

func verifyHeap(t *testing.T, h prioq.Q) {
}

func createHeap(kind trade.OrderKind) prioq.Q {
	return New(kind)
}

func TestPushPop(t *testing.T) {
	prioq.PushPopSuite(t, createHeap, verifyHeap)
}
