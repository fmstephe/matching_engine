package matcher

import (
	"github.com/fmstephe/matching_engine/cbuf"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var tcompareOrderMaker = trade.NewOrderMaker()

func TestCompareMatchers(t *testing.T) {
	compareMatchers(t, 100, 1, 1, 1)
	compareMatchers(t, 100, 10, 1, 1)
	compareMatchers(t, 100, 100, 1, 1)

	compareMatchers(t, 100, 1, 1, 2)
	compareMatchers(t, 100, 10, 1, 2)
	compareMatchers(t, 100, 100, 1, 2)

	compareMatchers(t, 100, 1, 10, 20)
	compareMatchers(t, 100, 10, 10, 20)
	compareMatchers(t, 100, 100, 10, 20)

	compareMatchers(t, 100, 1, 100, 2000)
	compareMatchers(t, 100, 10, 100, 2000)
	compareMatchers(t, 100, 100, 100, 2000)
}

func compareMatchers(t *testing.T, orderPairs, depth int, lowPrice, highPrice int64) {
	rrb := cbuf.New(orderPairs * 2)
	rm := newRefmatcher(lowPrice, highPrice, rrb)
	rb := cbuf.New(orderPairs * 2)
	m := NewMatcher(orderPairs*2, rb)
	orders, err := tcompareOrderMaker.RndTradeSet(orderPairs, depth, lowPrice, highPrice)
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(orders); i++ {
		o := &orders[i]
		rm.submit(o)
		rm.submit(o)
		m.Submit(o)
		m.Submit(o)
		checkBuffers(t, rrb, rb)
	}
}

func checkBuffers(t *testing.T, rrb, rb *cbuf.Response) {
	if rrb.Writes() != rb.Writes() {
		t.Errorf("Different number of writes detected. Simple: %d, Real: %d", rrb.Writes(), rb.Writes())
	}
	for rrb.Reads() < rrb.Writes() {
		sr, serr := rrb.GetForRead()
		if serr != nil {
			t.Errorf(serr.Error())
			return
		}
		r, err := rb.GetForRead()
		if err != nil {
			t.Errorf(err.Error())
			return
		}
		if *sr != *r {
			t.Errorf("Different responses read. Simple: %v, Real: %v", sr, r)
			return
		}
	}
}
