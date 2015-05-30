package matcher

import (
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

var cmprMaker = msg.NewMessageMaker(1)

func TestCompareMatchers(t *testing.T) {
	compareMatchers(t, 100, 1, 1, 1)
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
}

func compareMatchers(t *testing.T, orderPairs, depth int, lowPrice, highPrice uint64) {
	refIn := coordinator.NewChanReaderWriter(1)
	refOut := coordinator.NewChanReaderWriter(orderPairs * 4)
	refm := newRefmatcher(lowPrice, highPrice)
	refm.Config("Reference Matcher", refIn, refOut)
	in := coordinator.NewChanReaderWriter(1)
	out := coordinator.NewChanReaderWriter(orderPairs * 4)
	m := NewMatcher(orderPairs * 4)
	m.Config("Real Matcher", in, out)
	testSet, err := cmprMaker.RndTradeSet(orderPairs, depth, lowPrice, highPrice)
	if err != nil {
		panic(err.Error())
	}
	go m.Run()
	go refm.Run()
	for i := 0; i < len(testSet); i++ {
		refIn.Write(testSet[i])
		in.Write(testSet[i])
	}
	refIn.Write(msg.Message{Kind: msg.SHUTDOWN})
	in.Write(msg.Message{Kind: msg.SHUTDOWN})
	checkBuffers(t, refOut, out)
}

func checkBuffers(t *testing.T, refrc, rc coordinator.MsgReader) {
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

func drain(r coordinator.MsgReader) []*msg.Message {
	ms := make([]*msg.Message, 0)
	for {
		m := &msg.Message{}
		*m = r.Read()
		ms = append(ms, m)
		if m.Kind == msg.SHUTDOWN {
			return ms
		}
	}
}
