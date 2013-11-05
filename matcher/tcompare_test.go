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

func compareMatchers(t *testing.T, orderPairs, depth int, lowPrice, highPrice uint64) {
	refIn := make(chan *msg.Message)
	refOut := make(chan *msg.Message, orderPairs*2)
	refm := newRefmatcher(lowPrice, highPrice)
	refm.Config("Reference Matcher", refIn, refOut)
	in := make(chan *msg.Message)
	out := make(chan *msg.Message, orderPairs*2)
	m := NewMatcher(orderPairs * 2)
	m.Config("Real Matcher", in, out)
	testSet, err := cmprMaker.RndTradeSet(orderPairs, depth, lowPrice, highPrice)
	go m.Run()
	go refm.Run()
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(testSet); i++ {
		o := &testSet[i]
		refIn <- o
		refIn <- o
		in <- o
		in <- o
		checkBuffers(t, refOut, out)
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
