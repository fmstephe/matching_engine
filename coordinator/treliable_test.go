package coordinator

import (
	"errors"
	. "github.com/fmstephe/matching_engine/msg"
	"runtime"
	"testing"
)

type chanWriter struct {
	out chan *RMessage
}

func newChanWriter(out chan *RMessage) *chanWriter {
	return &chanWriter{out: out}
}

func (c chanWriter) Write(b []byte) (int, error) {
	r := &RMessage{}
	r.Unmarshal(b)
	c.out <- r
	return len(b), nil
}

func (c chanWriter) Close() error {
	return nil
}

func startMockedResponder() (fromApp chan *Message, fromListener chan *RMessage, out chan *RMessage) {
	fromApp = make(chan *Message, 100)
	fromListener = make(chan *RMessage, 100)
	out = make(chan *RMessage, 100)
	w := newChanWriter(out)
	originId := uint32(1)
	r := newReliableResponder(w, fromApp, fromListener, "Mocked Responder", originId, false)
	go r.Run()
	return fromApp, fromListener, out
}

func TestAppMsgWrittenOut(t *testing.T) {
	fromApp, _, out := startMockedResponder()
	m := &Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 1, Amount: 1}
	rm := &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: *m}
	fromApp <- m
	validateRMsg(t, <-out, rm, 1)
	m1 := &Message{Kind: SELL, TraderId: 2, TradeId: 2, StockId: 2, Price: 2, Amount: 2}
	rm1 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 2, message: *m1}
	fromApp <- m1
	validateRMsg(t, <-out, rm1, 1)
	m2 := &Message{Kind: PARTIAL, TraderId: 3, TradeId: 3, StockId: 3, Price: 3, Amount: 3}
	rm2 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 3, message: *m2}
	fromApp <- m2
	validateRMsg(t, <-out, rm2, 1)
	m3 := &Message{Kind: FULL, TraderId: 4, TradeId: 4, StockId: 4, Price: 4, Amount: 4}
	rm3 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 4, message: *m3}
	fromApp <- m3
	validateRMsg(t, <-out, rm3, 1)
}

func TestOutAckWrittenOut(t *testing.T) {
	_, fromListener, out := startMockedResponder()
	a := &RMessage{route: ACK, direction: OUT, originId: 2, msgId: 10, message: Message{Kind: BUY, TraderId: 1, TradeId: 1, StockId: 1, Price: 1, Amount: 1}}
	rm := &RMessage{}
	*rm = *a
	rm.direction = IN
	rm.originId = 1
	rm.msgId = 1
	fromListener <- a
	validateRMsg(t, <-out, rm, 1)
	a1 := &RMessage{route: ACK, direction: OUT, originId: 2, msgId: 11, message: Message{Kind: BUY, TraderId: 2, TradeId: 2, StockId: 2, Price: 2, Amount: 2}}
	rm1 := &RMessage{}
	*rm1 = *a1
	rm1.direction = IN
	rm1.originId = 1
	rm1.msgId = 2
	fromListener <- a1
	validateRMsg(t, <-out, rm1, 1)
	a2 := &RMessage{route: ACK, direction: OUT, originId: 3, msgId: 12, message: Message{Kind: BUY, TraderId: 3, TradeId: 3, StockId: 3, Price: 3, Amount: 3}}
	rm2 := &RMessage{}
	*rm2 = *a2
	rm2.direction = IN
	rm2.originId = 1
	rm2.msgId = 3
	fromListener <- a2
	validateRMsg(t, <-out, rm2, 1)
	a3 := &RMessage{route: ACK, direction: OUT, originId: 4, msgId: 13, message: Message{Kind: BUY, TraderId: 4, TradeId: 4, StockId: 4, Price: 4, Amount: 4}}
	rm3 := &RMessage{}
	*rm3 = *a3
	rm3.direction = IN
	rm3.originId = 1
	rm3.msgId = 4
	fromListener <- a3
	validateRMsg(t, <-out, rm3, 1)
}

// Two response messages with the same traderId/tradeId should both be resent (until acked)
// When this test was written the CANCELLED message would overwrite the PARTIAL, and only the CANCELLED would be resent
func TestServerAckNotOverwrittenByCancel(t *testing.T) {
	out := make(chan *RMessage, 100)
	w := chanWriter{out}
	r := &reliableResponder{writer: w, unacked: newSet()}
	p := &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: PARTIAL, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1}}
	c := &RMessage{route: APP, direction: IN, originId: 1, msgId: 2, message: Message{Kind: CANCELLED, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1}}
	// Add PARTIAL to unacked list
	r.addToUnacked(p)
	r.resend()
	allResent(t, out, p)
	// Add CANCEL to unacked list
	r.addToUnacked(c)
	r.resend()
	allResent(t, out, p, c)
}

func TestUnackedInDetail(t *testing.T) {
	out := make(chan *RMessage, 100)
	w := chanWriter{out}
	r := &reliableResponder{writer: w, unacked: newSet()}
	// Pre-canned message/ack pairs
	m1 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: FULL, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1}}
	a1 := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 1, message: Message{Kind: FULL, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1}}
	m2 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 2, message: Message{Kind: FULL, TraderId: 123, TradeId: 2000, StockId: 1, Price: 1, Amount: 1}}
	a2 := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 2, message: Message{Kind: FULL, TraderId: 123, TradeId: 2000, StockId: 1, Price: 1, Amount: 1}}
	m3 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 3, message: Message{Kind: FULL, TraderId: 777, TradeId: 5432, StockId: 1, Price: 1, Amount: 1}}
	a3 := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 3, message: Message{Kind: FULL, TraderId: 777, TradeId: 5432, StockId: 1, Price: 1, Amount: 1}}
	m4 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 4, message: Message{Kind: FULL, TraderId: 371, TradeId: 999, StockId: 1, Price: 1, Amount: 1}}
	a4 := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 4, message: Message{Kind: FULL, TraderId: 371, TradeId: 999, StockId: 1, Price: 1, Amount: 1}}
	m5 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 5, message: Message{Kind: FULL, TraderId: 87, TradeId: 50, StockId: 1, Price: 1, Amount: 1}}
	a5 := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 5, message: Message{Kind: FULL, TraderId: 87, TradeId: 50, StockId: 1, Price: 1, Amount: 1}}
	m6 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 6, message: Message{Kind: FULL, TraderId: 40, TradeId: 499, StockId: 1, Price: 1, Amount: 1}}
	a6 := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 6, message: Message{Kind: FULL, TraderId: 40, TradeId: 499, StockId: 1, Price: 1, Amount: 1}}
	m7 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 7, message: Message{Kind: FULL, TraderId: 99, TradeId: 700000, StockId: 1, Price: 1, Amount: 1}}
	a7 := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 7, message: Message{Kind: FULL, TraderId: 99, TradeId: 700000, StockId: 1, Price: 1, Amount: 1}}
	aUnkown := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 8, message: Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 1, Amount: 1}}

	// Add m1-5 to unacked list
	r.addToUnacked(m1)
	r.resend()
	allResent(t, out, m1)
	r.resend()
	allResent(t, out, m1)
	r.addToUnacked(m2)
	r.resend()
	allResent(t, out, m1, m2)
	r.addToUnacked(m3)
	r.resend()
	allResent(t, out, m1, m2, m3)
	r.addToUnacked(m4)
	r.resend()
	allResent(t, out, m1, m2, m3, m4)
	r.addToUnacked(m5)
	r.resend()
	allResent(t, out, m1, m2, m3, m4, m5)

	// ack m3
	r.handleInAck(a3)
	r.resend()
	allResent(t, out, m1, m2, m4, m5)

	// ack m1
	r.handleInAck(a1)
	r.resend()
	allResent(t, out, m2, m4, m5)

	// ack unknown
	r.handleInAck(aUnkown)
	r.resend()
	allResent(t, out, m2, m4, m5)

	// Add m6
	r.addToUnacked(m6)
	r.resend()
	allResent(t, out, m2, m4, m5, m6)

	// ack m4 and m6, and add m7
	r.handleInAck(a4)
	r.handleInAck(a6)
	r.addToUnacked(m7)
	r.resend()
	allResent(t, out, m2, m5, m7)

	// ack m2, m5 and m7
	r.handleInAck(a2)
	r.handleInAck(a5)
	r.handleInAck(a7)
	r.resend()
	if len(out) != 0 {
		t.Errorf("Expecting no messages re-sent, found %d", len(out))
	}
}

func allResent(t *testing.T, out chan *RMessage, expect ...*RMessage) {
	received := make([]*RMessage, 0)
	for len(out) > 0 {
		received = append(received, <-out)
	}
	if len(received) != len(expect) {
		_, fname, lnum, _ := runtime.Caller(1)
		t.Errorf("Expecting %d messages, received %d\n%s:%d", len(expect), len(received), fname, lnum)
	}
	allOfIn(t, expect, received)
	allOfIn(t, received, expect)
}

func allOfIn(t *testing.T, first, second []*RMessage) {
	for _, f := range first {
		found := false
		for _, s := range second {
			if *f == *s {
				found = true
				break
			}
		}
		if !found {
			_, fname, lnum, _ := runtime.Caller(2)
			t.Errorf("Expecting %v, not found\n%s:%d", f, fname, lnum)
		}
	}
}

type chanReader struct {
	in        chan *RMessage
	shouldErr bool
	writeN    int
}

func newChanReader(in chan *RMessage, shouldErr bool, writeN int) *chanReader {
	if writeN > rmsgByteSize || writeN < 0 {
		writeN = rmsgByteSize
	}
	return &chanReader{in: in, shouldErr: shouldErr, writeN: writeN}
}

func (r *chanReader) Read(b []byte) (int, error) {
	bb := b[:r.writeN]
	m := <-r.in
	m.Marshal(bb)
	if r.shouldErr {
		return len(bb), errors.New("fake error")
	}
	return len(bb), nil
}

func (r *chanReader) Close() error {
	return nil
}

func startMockedListener(shouldErr bool, writeN int) (in chan *RMessage, outApp chan *Message, outResponder chan *RMessage) {
	in = make(chan *RMessage, 100)
	outApp = make(chan *Message, 100)
	outResponder = make(chan *RMessage, 100)
	r := newChanReader(in, shouldErr, writeN)
	originId := uint32(1)
	l := newReliableListener(r, outApp, outResponder, "Mocked Listener", originId, false)
	go l.Run()
	return in, outApp, outResponder
}

func TestSmallReadError(t *testing.T) {
	in, outApp, outResponder := startMockedListener(false, rmsgByteSize-1)
	m := &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	in <- m
	// Expected server ack
	a := &RMessage{}
	a.WriteAckFor(m)
	a.status = SMALL_READ_ERROR
	// Expected response
	r := &Message{}
	*r = m.message
	validateRMsg(t, <-outResponder, a, 1)
	validateMsg(t, <-outApp, r, 1)
}

func TestReadError(t *testing.T) {
	in, outApp, outResponder := startMockedListener(true, rmsgByteSize)
	m := &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	in <- m
	// Expected server ack
	a := &RMessage{}
	a.WriteAckFor(m)
	a.status = READ_ERROR
	// Expected response
	r := &Message{}
	*r = m.message
	validateRMsg(t, <-outResponder, a, 1)
	validateMsg(t, <-outApp, r, 1)
}

func TestDuplicate(t *testing.T) {
	in, outApp, outResponder := startMockedListener(false, rmsgByteSize)
	m := &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	in <- m
	in <- m
	// Expected server ack
	a := &RMessage{}
	a.WriteAckFor(m)
	// Expected app msgs
	am := &Message{}
	*am = m.message
	// Expect an ack for both messages but the message is only forwarded on once
	validateRMsg(t, <-outResponder, a, 1)
	validateMsg(t, <-outApp, am, 1)
	validateRMsg(t, <-outResponder, a, 1)
	m2 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 2, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 2, TradeId: 1, StockId: 1}}
	in <- m2
	// Expected server ack 2
	a2 := &RMessage{}
	a2.WriteAckFor(m2)
	// Expected app msg
	am2 := &Message{}
	*am2 = m2.message
	// An ack for m2 and m2 (but nothing relating to m)
	validateRMsg(t, <-outResponder, a2, 1)
	validateMsg(t, <-outApp, am2, 1)
}

// Test ACK sent in, twice, expect ACK both times
func TestDuplicateAck(t *testing.T) {
	in, _, outResponder := startMockedListener(false, rmsgByteSize)
	m := &RMessage{route: ACK, direction: IN, originId: 1, msgId: 1, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	in <- m
	in <- m
	in <- m
	// Expect the ack to be passed through both times
	validateRMsg(t, <-outResponder, m, 1)
	validateRMsg(t, <-outResponder, m, 1)
	validateRMsg(t, <-outResponder, m, 1)
}

func TestOrdersAckedSentAndDeduped(t *testing.T) {
	// Test BUY sent in, twice, expect ACK, BUY and just ACK for the duplicate
	m := &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: BUY, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test SELL sent in, twice, expect ACK, SELL and just ACK for the duplicate
	m = &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test PARTIAL sent in, twice, expect ACK, PARTIAL and just ACK for the duplicate
	m = &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: PARTIAL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test FULL sent in, twice, expect ACK, FULL and just ACK for the duplicate
	m = &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: FULL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test CANCELLED sent in, twice, expect ACK, CANCELLEd and just ACK for the duplicate
	m = &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: CANCELLED, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test NOT_CANCELLED sent in, twice, expect ACK, NOT_CANCELLED and just ACK for the duplicate
	m = &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: NOT_CANCELLED, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test CANCEL sent in, twice, expect ACK, CANCEL and just ACK for the duplicate
	m = &RMessage{route: APP, direction: IN, originId: 1, msgId: 1, message: Message{Kind: CANCEL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
}

func sendThriceAckMsgAckAck(t *testing.T, m *RMessage) {
	in, outApp, outResponder := startMockedListener(false, rmsgByteSize)
	in <- m
	in <- m
	in <- m
	// Ack
	a := &RMessage{}
	a.WriteAckFor(m)
	// App
	am := &Message{}
	*am = m.message
	validateRMsg(t, <-outResponder, a, 2)
	validateMsg(t, <-outApp, am, 2)
	validateRMsg(t, <-outResponder, a, 2)
	validateRMsg(t, <-outResponder, a, 2)
}

func validateMsg(t *testing.T, m, e *Message, stackOffset int) {
	if *m != *e {
		_, fname, lnum, _ := runtime.Caller(stackOffset)
		t.Errorf("\nExpecting: %v\nFound:     %v \n%s:%d", e, m, fname, lnum)
	}
}

func validateRMsg(t *testing.T, m, e *RMessage, stackOffset int) {
	if *m != *e {
		_, fname, lnum, _ := runtime.Caller(stackOffset)
		t.Errorf("\nExpecting: %v\nFound:     %v \n%s:%d", e, m, fname, lnum)
	}
}

func TestBadNetwork(t *testing.T) {
	testBadNetwork(t, 0.5, Reliable)
}
