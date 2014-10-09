package matcher

import (
	"unsafe"
	"github.com/fmstephe/flib/queues/spscq"
	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

var cmprMaker = msg.NewMessageMaker(1)

func TestCompareMatchers(t *testing.T) {
	compareMatchers(t, 100, 1, 1, 1)
	/*
	compareMatchers(t, 100, 10, 1, 1)
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
	*/
}

func compareMatchers(t *testing.T, orderPairs, depth int64, lowPrice, highPrice uint64) {
	refIn, _ := spscq.NewPointerQ(1, 0)
	refOut, _ := spscq.NewPointerQ(fmath.NxtPowerOfTwo(orderPairs*4), 0)
	refm := newRefmatcher(lowPrice, highPrice)
	refm.Config("Reference Matcher", refIn, refOut)
	in, _ := spscq.NewPointerQ(1, 0)
	out, _ := spscq.NewPointerQ(fmath.NxtPowerOfTwo(orderPairs*4), 0)
	m := NewMatcher(orderPairs * 4)
	m.Config("Real Matcher", in, out)
	testSet, err := cmprMaker.RndTradeSet(orderPairs, depth, lowPrice, highPrice)
	if err != nil {
		panic(err.Error())
	}
	go m.Run()
	go refm.Run()
	for i := 0; i < len(testSet); i++ {
		o := &testSet[i]
		refIn.WriteSingleBlocking(unsafe.Pointer(o))
		in.WriteSingleBlocking(unsafe.Pointer(o))
	}
	shutdown := &msg.Message{Kind: msg.SHUTDOWN}
	refIn.WriteSingleBlocking(unsafe.Pointer(shutdown))
	in.WriteSingleBlocking(unsafe.Pointer(shutdown))
	checkBuffers(t, refOut, out)
}

func checkBuffers(t *testing.T, refrc, rc *spscq.PointerQ) {
	refrs := drain(refrc)
	rs := drain(rc)
	if len(refrs) != len(rs) {
		t.Errorf("Different number of writes detected. Simple: %d, Real: %d", len(refrs), len(rs))
		return
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

func drain(q *spscq.PointerQ) []*msg.Message {
	rs := make([]*msg.Message, 0)
	for {
		r := (*msg.Message)(q.ReadSingleBlocking())
		if r.Kind == msg.SHUTDOWN {
			return rs
		}
		rs = append(rs, r)
	}
}
