package coordinator

import (
	"errors"
	. "github.com/fmstephe/matching_engine/msg"
	"runtime"
	"testing"
)

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

func startMockedListener(shouldErr bool, writeN int) (in chan *Message, out chan *Message) {
	in = make(chan *Message, 100)
	out = make(chan *Message, 100)
	cr := newChanReader(in, shouldErr, writeN)
	l := newListener(cr)
	l.Config(false, "test listener", out)
	go l.Run()
	return in, out
}

func TestSmallReadError(t *testing.T) {
	in, out := startMockedListener(false, SizeofMessage-1)
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
	validate(t, <-out, a, 1)
	validate(t, <-out, r, 1)
}

func TestReadError(t *testing.T) {
	in, out := startMockedListener(true, SizeofMessage)
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
	validate(t, <-out, a, 1)
	validate(t, <-out, r, 1)
}

func TestDuplicate(t *testing.T) {
	in, out := startMockedListener(false, SizeofMessage)
	m := &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	in <- m
	in <- m
	// Expected server ack
	a := &Message{}
	a.WriteAckFor(m)
	// Expect an ack for both messages but the message is only forwarded on once
	validate(t, <-out, a, 1)
	validate(t, <-out, m, 1)
	validate(t, <-out, a, 1)
	m2 := &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 2, TradeId: 1, StockId: 1}
	in <- m2
	// Expected server ack 2
	a2 := &Message{}
	a2.WriteAckFor(m2)
	// An ack for m2 and m2 (but nothing relating to m)
	validate(t, <-out, a2, 1)
	validate(t, <-out, m2, 1)
}

// Test ACK sent in, twice, expect ACK both times
func TestDuplicateAck(t *testing.T) {
	in, out := startMockedListener(false, SizeofMessage)
	m := &Message{Status: NORMAL, Route: ACK, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	in <- m
	in <- m
	// Expect the ack to be passed through both times
	validate(t, <-out, m, 1)
	validate(t, <-out, m, 1)
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
	in, out := startMockedListener(false, SizeofMessage)
	in <- m
	in <- m
	in <- m
	// Ack
	a := &Message{}
	a.WriteAckFor(m)
	validate(t, <-out, a, 2)
	validate(t, <-out, m, 2)
	validate(t, <-out, a, 2)
	validate(t, <-out, a, 2)
}

func validate(t *testing.T, m, e *Message, stackOffset int) {
	if *m != *e {
		_, fname, lnum, _ := runtime.Caller(stackOffset)
		t.Errorf("\nExpecting: %v\nFound:     %v \n%s:%d", e, m, fname, lnum)
	}
}
