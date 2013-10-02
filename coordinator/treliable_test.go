package coordinator

import (
	"errors"
	. "github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/msg/msgutil"
	"runtime"
	"testing"
)

type chanWriter struct {
	out chan *Message
}

func (c chanWriter) Write(b []byte) (int, error) {
	r := &Message{}
	r.WriteFrom(b)
	c.out <- r
	return len(b), nil
}

func (c chanWriter) Close() error {
	return nil
}

// Two response messages with the same traderId/tradeId should both be resent (until acked)
// When this test was written the CANCELLED message would overwrite the PARTIAL, and only the CANCELLED would be resent
func TestServerAckNotOverwrittenByCancel(t *testing.T) {
	out := make(chan *Message, 100)
	w := chanWriter{out}
	r := &reliableResponder{writer: w, unacked: msgutil.NewSet()}
	p := &Message{Route: APP, Kind: PARTIAL, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1, OriginId: 1, MsgId: 1}
	c := &Message{Route: APP, Kind: CANCELLED, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1, OriginId: 1, MsgId: 2}
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
	out := make(chan *Message, 100)
	w := chanWriter{out}
	r := &reliableResponder{writer: w, unacked: msgutil.NewSet()}
	// Pre-canned message/ack pairs
	m1 := &Message{TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1, Route: APP, Kind: FULL, OriginId: 1, MsgId: 1}
	a1 := &Message{TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1, Route: ACK, Kind: FULL, OriginId: 1, MsgId: 1}
	m2 := &Message{TraderId: 123, TradeId: 2000, StockId: 1, Price: 1, Amount: 1, Route: APP, Kind: FULL, OriginId: 1, MsgId: 2}
	a2 := &Message{TraderId: 123, TradeId: 2000, StockId: 1, Price: 1, Amount: 1, Route: ACK, Kind: FULL, OriginId: 1, MsgId: 2}
	m3 := &Message{TraderId: 777, TradeId: 5432, StockId: 1, Price: 1, Amount: 1, Route: APP, Kind: FULL, OriginId: 1, MsgId: 3}
	a3 := &Message{TraderId: 777, TradeId: 5432, StockId: 1, Price: 1, Amount: 1, Route: ACK, Kind: FULL, OriginId: 1, MsgId: 3}
	m4 := &Message{TraderId: 371, TradeId: 999, StockId: 1, Price: 1, Amount: 1, Route: APP, Kind: FULL, OriginId: 1, MsgId: 4}
	a4 := &Message{TraderId: 371, TradeId: 999, StockId: 1, Price: 1, Amount: 1, Route: ACK, Kind: FULL, OriginId: 1, MsgId: 4}
	m5 := &Message{TraderId: 87, TradeId: 50, StockId: 1, Price: 1, Amount: 1, Route: APP, Kind: FULL, OriginId: 1, MsgId: 5}
	a5 := &Message{TraderId: 87, TradeId: 50, StockId: 1, Price: 1, Amount: 1, Route: ACK, Kind: FULL, OriginId: 1, MsgId: 5}
	m6 := &Message{TraderId: 40, TradeId: 499, StockId: 1, Price: 1, Amount: 1, Route: APP, Kind: FULL, OriginId: 1, MsgId: 6}
	a6 := &Message{TraderId: 40, TradeId: 499, StockId: 1, Price: 1, Amount: 1, Route: ACK, Kind: FULL, OriginId: 1, MsgId: 6}
	m7 := &Message{TraderId: 99, TradeId: 700000, StockId: 1, Price: 1, Amount: 1, Route: APP, Kind: FULL, OriginId: 1, MsgId: 7}
	a7 := &Message{TraderId: 99, TradeId: 700000, StockId: 1, Price: 1, Amount: 1, Route: ACK, Kind: FULL, OriginId: 1, MsgId: 7}
	aUnkown := &Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 1, Amount: 1, Route: ACK, Kind: FULL, OriginId: 1, MsgId: 8}

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

func allResent(t *testing.T, out chan *Message, expect ...*Message) {
	received := make([]*Message, 0)
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

func allOfIn(t *testing.T, first, second []*Message) {
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
	in        chan *Message
	shouldErr bool
	writeN    int
}

func newChanReader(in chan *Message, shouldErr bool, writeN int) *chanReader {
	if writeN > SizeofMessage || writeN < 0 {
		writeN = SizeofMessage
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

func startMockedListener(shouldErr bool, writeN int) (in, outApp, outResponder chan *Message) {
	in = make(chan *Message, 100)
	outApp = make(chan *Message, 100)
	outResponder = make(chan *Message, 100)
	r := newChanReader(in, shouldErr, writeN)
	originId := uint32(1)
	l := newReliableListener(r, outApp, outResponder, "Mocked Listener", originId, false)
	go l.Run()
	return in, outApp, outResponder
}

func TestSmallReadError(t *testing.T) {
	in, outApp, outResponder := startMockedListener(false, SizeofMessage-1)
	m := &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	in <- m
	// Expected server ack
	a := &Message{}
	a.WriteAckFor(m)
	a.Status = SMALL_READ_ERROR
	// Expected response
	r := &Message{}
	*r = *m
	r.Status = SMALL_READ_ERROR
	validate(t, <-outResponder, a, 1)
	validate(t, <-outApp, r, 1)
}

func TestReadError(t *testing.T) {
	in, outApp, outResponder := startMockedListener(true, SizeofMessage)
	m := &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	in <- m
	// Expected server ack
	a := &Message{}
	a.WriteAckFor(m)
	a.Status = READ_ERROR
	// Expected response
	r := &Message{}
	*r = *m
	r.Status = READ_ERROR
	validate(t, <-outResponder, a, 1)
	validate(t, <-outApp, r, 1)
}

func TestDuplicate(t *testing.T) {
	in, outApp, outResponder := startMockedListener(false, SizeofMessage)
	m := &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1, OriginId: 1, MsgId: 1}
	in <- m
	in <- m
	// Expected server ack
	a := &Message{}
	a.WriteAckFor(m)
	// Expect an ack for both messages but the message is only forwarded on once
	validate(t, <-outResponder, a, 1)
	validate(t, <-outApp, m, 1)
	validate(t, <-outResponder, a, 1)
	m2 := &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 2, TradeId: 1, StockId: 1, OriginId: 1, MsgId: 2}
	in <- m2
	// Expected server ack 2
	a2 := &Message{}
	a2.WriteAckFor(m2)
	// An ack for m2 and m2 (but nothing relating to m)
	validate(t, <-outResponder, a2, 1)
	validate(t, <-outApp, m2, 1)
}

// Test ACK sent in, twice, expect ACK both times
func TestDuplicateAck(t *testing.T) {
	in, _, outResponder := startMockedListener(false, SizeofMessage)
	m := &Message{Status: NORMAL, Route: ACK, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	in <- m
	in <- m
	// Expect the ack to be passed through both times
	validate(t, <-outResponder, m, 1)
	validate(t, <-outResponder, m, 1)
}

func TestOrdersAckedSentAndDeduped(t *testing.T) {
	// Test BUY sent in, twice, expect ACK, BUY and just ACK for the duplicate
	m := &Message{Status: NORMAL, Route: APP, Kind: BUY, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	sendTwiceAckMsgAck(t, m)
	// Test SELL sent in, twice, expect ACK, SELL and just ACK for the duplicate
	m = &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	sendTwiceAckMsgAck(t, m)
	// Test PARTIAL sent in, twice, expect ACK, PARTIAL and just ACK for the duplicate
	m = &Message{Status: NORMAL, Route: APP, Kind: PARTIAL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	sendTwiceAckMsgAck(t, m)
	// Test FULL sent in, twice, expect ACK, FULL and just ACK for the duplicate
	m = &Message{Status: NORMAL, Route: APP, Kind: FULL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	sendTwiceAckMsgAck(t, m)
	// Test CANCELLED sent in, twice, expect ACK, CANCELLEd and just ACK for the duplicate
	m = &Message{Status: NORMAL, Route: APP, Kind: CANCELLED, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	sendTwiceAckMsgAck(t, m)
	// Test NOT_CANCELLED sent in, twice, expect ACK, NOT_CANCELLED and just ACK for the duplicate
	m = &Message{Status: NORMAL, Route: APP, Kind: NOT_CANCELLED, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	sendTwiceAckMsgAck(t, m)
	// Test CANCEL sent in, twice, expect ACK, CANCEL and just ACK for the duplicate
	m = &Message{Status: NORMAL, Route: APP, Kind: CANCEL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	sendTwiceAckMsgAck(t, m)
}

func sendTwiceAckMsgAck(t *testing.T, m *Message) {
	in, outApp, outResponder := startMockedListener(false, SizeofMessage)
	in <- m
	in <- m
	in <- m
	// Ack
	a := &Message{}
	a.WriteAckFor(m)
	validate(t, <-outResponder, a, 2)
	validate(t, <-outApp, m, 2)
	validate(t, <-outResponder, a, 2)
	validate(t, <-outResponder, a, 2)
}

func validate(t *testing.T, m, e *Message, stackOffset int) {
	if *m != *e {
		_, fname, lnum, _ := runtime.Caller(stackOffset)
		t.Errorf("\nExpecting: %v\nFound:     %v \n%s:%d", e, m, fname, lnum)
	}
}

func TestBadNetwork(t *testing.T) {
	testBadNetwork(t, 0.5, Reliable)
}
