package matcher

import (
	"github.com/fmstephe/matching_engine/trade"
	"testing"
)

var tcompareOrderMaker = trade.NewOrderMaker()

func TestCompareMatchers(t *testing.T) {
	compareMatchers(t, 100, 1, 1, 1)
	compareMatchers(t, 100, 10, 1, 1)
	compareMatchers(t, 100, 100, 1, 1)
	//
	compareMatchers(t, 100, 1, 1, 2)
	compareMatchers(t, 100, 10, 1, 2)
	compareMatchers(t, 100, 100, 1, 2)
	//
	compareMatchers(t, 100, 1, 10, 20)
	compareMatchers(t, 100, 10, 10, 20)
	compareMatchers(t, 100, 100, 10, 20)
	//
	compareMatchers(t, 100, 1, 100, 2000)
	compareMatchers(t, 100, 10, 100, 2000)
	compareMatchers(t, 100, 100, 100, 2000)
}

func compareMatchers(t *testing.T, orderPairs, depth int, lowPrice, highPrice int64) {
	refrc := make(chan *trade.Response, orderPairs*2)
	refm := newRefmatcher(lowPrice, highPrice, refrc)
	rc := make(chan *trade.Response, orderPairs*2)
	m := NewMatcher(orderPairs*2, rc)
	orders, err := tcompareOrderMaker.RndTradeSet(orderPairs, depth, lowPrice, highPrice)
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(orders); i++ {
		o := &orders[i]
		refm.submit(o)
		refm.submit(o)
		m.Submit(o)
		m.Submit(o)
		checkBuffers(t, refrc, rc)
	}
}

func checkBuffers(t *testing.T, refrc, rc chan *trade.Response) {
	refrs := drain(refrc)
	rs := drain(rc)
	if len(refrs) != len(rs) {
		t.Errorf("Different number of writes detected. Simple: %d, Real: %d", len(refrs), len(rs))
	}
	for i := 0; i < len(rs); i++ {
		refr := refrs[i]
		r := rs[i]
		if *refr != *r {
			t.Errorf("Different responses read. Simple: %v, Real: %v", refr, r)
			return
		}
	}
}

func drain(c chan *trade.Response) []*trade.Response {
	rs := make([]*trade.Response, 0)
	for {
		select {
		case r := <-c:
			rs = append(rs, r)
		default:
			return rs
		}
	}
	panic("Unreachable")
}
