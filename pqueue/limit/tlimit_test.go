package limit

import (
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var orderMaker = trade.NewOrderMaker()

func TestPushPop(t *testing.T) {
	l := New(1)
	o := orderMaker.MkPricedBuy(1)
	l.Push(o)
	po := l.Pop()
	if po != o {
		t.Errorf("Push/Pop yields different objects")
	}
}
