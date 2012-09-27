package rtree

import (
	"github.com/fmstephe/matching_engine/pqueue"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var orderMaker = trade.NewOrderMaker()

func verifyHeap(t *testing.T, h pqueue.Q) {
}

func createHeap(buySell trade.TradeType) pqueue.Q {
	return New(buySell)
}

func TestPushPop(t *testing.T) {
	pqueue.PushPopSuite(t, createHeap, verifyHeap)
}
