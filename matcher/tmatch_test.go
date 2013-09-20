package matcher

import (
	"github.com/fmstephe/matching_engine/msg"
	"runtime"
	"testing"
)

const (
	stockId = 1
	trader1 = 1
	trader2 = 2
	trader3 = 3
)

var matchMaker = msg.NewMessageMaker(100)

type responseVals struct {
	price   int64
	amount  uint32
	tradeId uint32
	stockId uint32
}

func TestPrice(t *testing.T) {
	testPrice(t, 1, 1, 1)
	testPrice(t, 2, 1, 1)
	testPrice(t, 3, 1, 2)
	testPrice(t, 4, 1, 2)
	testPrice(t, 5, 1, 3)
	testPrice(t, 6, 1, 3)
	testPrice(t, 20, 10, 15)
	testPrice(t, 21, 10, 15)
	testPrice(t, 22, 10, 16)
	testPrice(t, 23, 10, 16)
	testPrice(t, 24, 10, 17)
	testPrice(t, 25, 10, 17)
	testPrice(t, 26, 10, 18)
	testPrice(t, 27, 10, 18)
	testPrice(t, 28, 10, 19)
	testPrice(t, 29, 10, 19)
	testPrice(t, 30, 10, 20)
}

func testPrice(t *testing.T, bPrice, sPrice, expected int64) {
	result := price(bPrice, sPrice)
	if result != expected {
		t.Errorf("price(%d,%d) does not equal %d, got %d instead.", bPrice, sPrice, expected, result)
	}
}

type testerMaker struct {
}

func (tm *testerMaker) Make() MatchTester {
	in := make(chan *msg.Message, 30)
	out := make(chan *msg.Message, 30)
	m := NewMatcher(100)
	m.Config("Matcher", in, out)
	go m.Run()
	return &localTester{in: in, out: out}
}

type localTester struct {
	in  chan *msg.Message
	out chan *msg.Message
}

func (lt *localTester) Send(t *testing.T, m *msg.Message) {
	m.Direction = msg.IN
	m.Route = msg.APP
	lt.in <- m
}

func (lt *localTester) Expect(t *testing.T, ref *msg.Message) {
	ref.Direction = msg.OUT
	ref.Route = msg.APP
	var m *msg.Message
	for {
		m = <-lt.out
		if m.Direction == msg.OUT {
			break
		}
	}
	if *ref != *m {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("\nExpecting: %v\nFound:     %v\n%s:%d", ref, m, fname, lnum)
	}
}

func (lt *localTester) ExpectEmpty(t *testing.T, traderId uint32) {
	if len(lt.out) != 0 {
		t.Errorf("\nExpecting Empty:\nFound: %v", <-lt.out)
	}
}

func (lt *localTester) Cleanup(t *testing.T) {}

func TestRunTestSuite(t *testing.T) {
	RunTestSuite(t, &testerMaker{})
}
