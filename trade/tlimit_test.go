package trade

import (
	"testing"
)

var limitOrderMaker = NewOrderMaker()

func TestPushPop(t *testing.T) {
	l := NewLimit(1)
	o := limitOrderMaker.MkPricedBuy(1)
	l.Push(o)
	po := l.Pop()
	if po != o {
		t.Errorf("Push/Pop yields different objects")
	}
}
