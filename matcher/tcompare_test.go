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
	refsubmit := make(chan interface{}, orderPairs*2)
	reforders := make(chan *trade.OrderData, 10)
	refm := newRefmatcher(lowPrice, highPrice, refsubmit, reforders)
	submit := make(chan interface{}, orderPairs*2)
	orders := make(chan *trade.OrderData, 10)
	m := NewMatcher(orderPairs * 2)
	m.SetSubmit(submit)
	m.SetOrders(orders)
	testSet, err := tcompareOrderMaker.RndTradeSet(orderPairs, depth, lowPrice, highPrice)
	go refm.Run()
	go m.Run()
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(testSet); i++ {
		o := &testSet[i]
		reforders <- o
		reforders <- o
		orders <- o
		orders <- o
		checkBuffers(t, refsubmit, submit)
	}
}

func checkBuffers(t *testing.T, refrc, rc chan interface{}) {
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

func drain(c chan interface{}) []*trade.Response {
	rs := make([]*trade.Response, 0)
	for {
		select {
		case r := <-c:
			rs = append(rs, r.(*trade.Response))
		default:
			return rs
		}
	}
	panic("Unreachable")
}
