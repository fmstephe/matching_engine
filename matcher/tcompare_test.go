package matcher

import (
	"github.com/fmstephe/matching_engine/cbuf"
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var tcompareOrderMaker = trade.NewOrderMaker()

func TestCompareMatchers(t *testing.T) {
	compareMatchers(t, 100, 1, 1)
	compareMatchers(t, 100, 1, 2)
	compareMatchers(t, 100, 10, 20)
	compareMatchers(t, 100, 100, 2000)
	compareMatchers(t, 1000, 100, 2000)
	compareMatchers(t, 1000, 0, 2000)
}

func compareMatchers(t *testing.T, orderPairs int, lowPrice, highPrice int64) {
	srb := cbuf.New(orderPairs * 2)
	sm := newRefmatcher(lowPrice, highPrice, srb)
	rb := cbuf.New(orderPairs * 2)
	m := NewMatcher(orderPairs*2, rb)
	for i := 0; i < orderPairs; i++ {
		b := tcompareOrderMaker.MkPricedBuyData(tcompareOrderMaker.Between(lowPrice, highPrice))
		s := tcompareOrderMaker.MkPricedSellData(tcompareOrderMaker.Between(lowPrice, highPrice))
		sm.submit(b)
		sm.submit(s)
		m.Submit(b)
		m.Submit(s)
		checkBuffers(t, srb, rb)
	}
}

func checkBuffers(t *testing.T, srb, rb *cbuf.Response) {
	if srb.Writes() != rb.Writes() {
		t.Errorf("Different number of writes detected. Simple: %d, Real: %d", srb.Writes(), rb.Writes())
	}
	for srb.Reads() < srb.Writes() {
		sr, serr := srb.GetForRead()
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
