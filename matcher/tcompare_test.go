package matcher

import (
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

var cmprMaker = msg.NewMessageMaker(1)

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
	refdispatch := make(chan *msg.Message, orderPairs*2)
	reforders := make(chan *msg.Message, 10)
	refm := newRefmatcher(lowPrice, highPrice, refdispatch, reforders)
	dispatch := make(chan *msg.Message, orderPairs*2)
	orders := make(chan *msg.Message, 10)
	m := NewMatcher(orderPairs * 2)
	m.SetDispatch(dispatch)
	m.SetOrders(orders)
	testSet, err := cmprMaker.RndTradeSet(orderPairs, depth, lowPrice, highPrice)
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
		checkBuffers(t, refdispatch, dispatch)
	}
}

func checkBuffers(t *testing.T, refrc, rc chan *msg.Message) {
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

func drain(c chan *msg.Message) []*msg.Message {
	rs := make([]*msg.Message, 0)
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
