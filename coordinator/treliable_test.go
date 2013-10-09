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

func (c chanWriter) Write(b []byte) (int, error) {
	r := &RMessage{}
	r.WriteFrom(b)
	c.out <- r
	return len(b), nil
}

func (c chanWriter) Close() error {
	return nil
}

// TODO we are creating RMessage structs in chaotic ways. This should be cleaned up
// Two response messages with the same traderId/tradeId should both be resent (until acked)
// When this test was written the CANCELLED message would overwrite the PARTIAL, and only the CANCELLED would be resent
func TestServerAckNotOverwrittenByCancel(t *testing.T) {
	out := make(chan *RMessage, 100)
	w := chanWriter{out}
	r := &reliableResponder{writer: w, unacked: newSet()}
	p := &RMessage{route: APP, originId: 1, msgId: 1, message: Message{Kind: PARTIAL, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1}}
	c := &RMessage{route: APP, originId: 1, msgId: 2, message: Message{Kind: CANCELLED, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1}}
	// Add buy server-ack to unacked list
	r.addToUnacked(p)
	r.resend()
	allResent(t, out, p)
	// Add cancel server-ack to unacked list
	r.addToUnacked(c)
	r.resend()
	allResent(t, out, p, c)
}

func TestUnackedInDetail(t *testing.T) {
	out := make(chan *RMessage, 100)
	w := chanWriter{out}
	r := &reliableResponder{writer: w, unacked: newSet()}
	// Pre-canned message/ack pairs
	m1 := &RMessage{message: Message{Kind: FULL, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1}, route: APP, originId: 1, msgId: 1}
	a1 := &RMessage{message: Message{Kind: FULL, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1}, route: ACK, originId: 1, msgId: 1}
	m2 := &RMessage{message: Message{Kind: FULL, TraderId: 123, TradeId: 2000, StockId: 1, Price: 1, Amount: 1}, route: APP, originId: 1, msgId: 2}
	a2 := &RMessage{message: Message{Kind: FULL, TraderId: 123, TradeId: 2000, StockId: 1, Price: 1, Amount: 1}, route: ACK, originId: 1, msgId: 2}
	m3 := &RMessage{message: Message{Kind: FULL, TraderId: 777, TradeId: 5432, StockId: 1, Price: 1, Amount: 1}, route: APP, originId: 1, msgId: 3}
	a3 := &RMessage{message: Message{Kind: FULL, TraderId: 777, TradeId: 5432, StockId: 1, Price: 1, Amount: 1}, route: ACK, originId: 1, msgId: 3}
	m4 := &RMessage{message: Message{Kind: FULL, TraderId: 371, TradeId: 999, StockId: 1, Price: 1, Amount: 1}, route: APP, originId: 1, msgId: 4}
	a4 := &RMessage{message: Message{Kind: FULL, TraderId: 371, TradeId: 999, StockId: 1, Price: 1, Amount: 1}, route: ACK, originId: 1, msgId: 4}
	m5 := &RMessage{message: Message{Kind: FULL, TraderId: 87, TradeId: 50, StockId: 1, Price: 1, Amount: 1}, route: APP, originId: 1, msgId: 5}
	a5 := &RMessage{message: Message{Kind: FULL, TraderId: 87, TradeId: 50, StockId: 1, Price: 1, Amount: 1}, route: ACK, originId: 1, msgId: 5}
	m6 := &RMessage{message: Message{Kind: FULL, TraderId: 40, TradeId: 499, StockId: 1, Price: 1, Amount: 1}, route: APP, originId: 1, msgId: 6}
	a6 := &RMessage{message: Message{Kind: FULL, TraderId: 40, TradeId: 499, StockId: 1, Price: 1, Amount: 1}, route: ACK, originId: 1, msgId: 6}
	m7 := &RMessage{message: Message{Kind: FULL, TraderId: 99, TradeId: 700000, StockId: 1, Price: 1, Amount: 1}, route: APP, originId: 1, msgId: 7}
	a7 := &RMessage{message: Message{Kind: FULL, TraderId: 99, TradeId: 700000, StockId: 1, Price: 1, Amount: 1}, route: ACK, originId: 1, msgId: 7}
	aUnkown := &RMessage{message: Message{Kind: FULL, TraderId: 1, TradeId: 1, StockId: 1, Price: 1, Amount: 1}, route: ACK, originId: 1, msgId: 8}

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
	if writeN > SizeofRMessage || writeN < 0 {
		writeN = SizeofRMessage
	}
	return &chanReader{in: in, shouldErr: shouldErr, writeN: writeN}
}

func (r *chanReader) Read(b []byte) (int, error) {
	bb := b[:r.writeN]
	m := <-r.in
	m.WriteTo(bb)
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
	in, outApp, outResponder := startMockedListener(false, SizeofRMessage-1)
	m := &RMessage{status: NORMAL, route: APP, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	in <- m
	// Expected server ack
	a := &RMessage{}
	a.WriteAckFor(m)
	a.status = SMALL_READ_ERROR
	// Expected response
	r := &Message{}
	*r = m.message
	validateR(t, <-outResponder, a, 1)
	validate(t, <-outApp, r, 1)
}

func TestReadError(t *testing.T) {
	in, outApp, outResponder := startMockedListener(true, SizeofRMessage)
	m := &RMessage{status: NORMAL, route: APP, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	in <- m
	// Expected server ack
	a := &RMessage{}
	a.WriteAckFor(m)
	a.status = READ_ERROR
	// Expected response
	r := &Message{}
	*r = m.message
	validateR(t, <-outResponder, a, 1)
	validate(t, <-outApp, r, 1)
}

func TestDuplicate(t *testing.T) {
	in, outApp, outResponder := startMockedListener(false, SizeofRMessage)
	m := &RMessage{status: NORMAL, route: APP, originId: 1, msgId: 1, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	in <- m
	in <- m
	// Expected server ack
	a := &RMessage{}
	a.WriteAckFor(m)
	// Expected app msgs
	am := &Message{}
	*am = m.message
	// Expect an ack for both messages but the message is only forwarded on once
	validateR(t, <-outResponder, a, 1)
	validate(t, <-outApp, am, 1)
	validateR(t, <-outResponder, a, 1)
	m2 := &RMessage{status: NORMAL, route: APP, originId: 1, msgId: 2, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 2, TradeId: 1, StockId: 1}}
	in <- m2
	// Expected server ack 2
	a2 := &RMessage{}
	a2.WriteAckFor(m2)
	// Expected app msg
	am2 := &Message{}
	*am2 = m2.message
	// An ack for m2 and m2 (but nothing relating to m)
	validateR(t, <-outResponder, a2, 1)
	validate(t, <-outApp, am2, 1)
}

// Test ACK sent in, twice, expect ACK both times
func TestDuplicateAck(t *testing.T) {
	in, _, outResponder := startMockedListener(false, SizeofRMessage)
	m := &RMessage{status: NORMAL, route: ACK, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	in <- m
	in <- m
	// Expect the ack to be passed through both times
	validateR(t, <-outResponder, m, 1)
	validateR(t, <-outResponder, m, 1)
}

func TestOrdersAckedSentAndDeduped(t *testing.T) {
	// Test BUY sent in, twice, expect ACK, BUY and just ACK for the duplicate
	m := &RMessage{status: NORMAL, route: APP, message: Message{Kind: BUY, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test SELL sent in, twice, expect ACK, SELL and just ACK for the duplicate
	m = &RMessage{status: NORMAL, route: APP, message: Message{Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test PARTIAL sent in, twice, expect ACK, PARTIAL and just ACK for the duplicate
	m = &RMessage{status: NORMAL, route: APP, message: Message{Kind: PARTIAL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test FULL sent in, twice, expect ACK, FULL and just ACK for the duplicate
	m = &RMessage{status: NORMAL, route: APP, message: Message{Kind: FULL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test CANCELLED sent in, twice, expect ACK, CANCELLEd and just ACK for the duplicate
	m = &RMessage{status: NORMAL, route: APP, message: Message{Kind: CANCELLED, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test NOT_CANCELLED sent in, twice, expect ACK, NOT_CANCELLED and just ACK for the duplicate
	m = &RMessage{status: NORMAL, route: APP, message: Message{Kind: NOT_CANCELLED, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
	// Test CANCEL sent in, twice, expect ACK, CANCEL and just ACK for the duplicate
	m = &RMessage{status: NORMAL, route: APP, message: Message{Kind: CANCEL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}}
	sendThriceAckMsgAckAck(t, m)
}

func sendThriceAckMsgAckAck(t *testing.T, m *RMessage) {
	in, outApp, outResponder := startMockedListener(false, SizeofRMessage)
	in <- m
	in <- m
	in <- m
	// Ack
	a := &RMessage{}
	a.WriteAckFor(m)
	// App
	am := &Message{}
	*am = m.message
	validateR(t, <-outResponder, a, 2)
	validate(t, <-outApp, am, 2)
	validateR(t, <-outResponder, a, 2)
	validateR(t, <-outResponder, a, 2)
}

func validate(t *testing.T, m, e *Message, stackOffset int) {
	if *m != *e {
		_, fname, lnum, _ := runtime.Caller(stackOffset)
		t.Errorf("\nExpecting: %v\nFound:     %v \n%s:%d", e, m, fname, lnum)
	}
}

func validateR(t *testing.T, m, e *RMessage, stackOffset int) {
	if *m != *e {
		_, fname, lnum, _ := runtime.Caller(stackOffset)
		t.Errorf("\nExpecting: %v\nFound:     %v \n%s:%d", e, m, fname, lnum)
	}
}

func TestBadNetwork(t *testing.T) {
	testBadNetwork(t, 0.5, Reliable)
}
