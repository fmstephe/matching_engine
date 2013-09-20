package coordinator

import (
	. "github.com/fmstephe/matching_engine/msg"
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
	r := newResponder(w)
	p := &Message{Route: APP, Direction: IN, Kind: PARTIAL, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1, OriginId: 1, MsgId: 1}
	c := &Message{Route: APP, Direction: IN, Kind: CANCELLED, TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1, OriginId: 1, MsgId: 2}
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
	r := newResponder(w)
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
