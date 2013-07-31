package coordinator

import (
	"errors"
	. "github.com/fmstephe/matching_engine/msg"
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

func startMockedListener(shouldErr bool, writeN int) (in chan *Message, dispatch chan *Message) {
	in = make(chan *Message, 100)
	dispatch = make(chan *Message, 100)
	cr := newChanReader(in, shouldErr, writeN)
	l := newListener(cr)
	l.SetDispatch(dispatch)
	go l.Run()
	return in, dispatch
}

func TestSmallReadError(t *testing.T) {
	in, dispatch := startMockedListener(false, SizeofMessage-1)
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
	validate(t, <-dispatch, a, 1)
	validate(t, <-dispatch, r, 1)
}

func TestReadError(t *testing.T) {
	in, dispatch := startMockedListener(true, SizeofMessage)
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
	validate(t, <-dispatch, a, 1)
	validate(t, <-dispatch, r, 1)
}

func TestDuplicate(t *testing.T) {
	in, dispatch := startMockedListener(false, SizeofMessage)
	m := &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	in <- m
	in <- m
	// Expected server ack
	a := &Message{}
	a.WriteAckFor(m)
	// Expect an ack for both messages but the message is only forwarded on once
	validate(t, <-dispatch, a, 1)
	validate(t, <-dispatch, m, 1)
	validate(t, <-dispatch, a, 1)
	m2 := &Message{Status: NORMAL, Route: APP, Kind: SELL, Price: 7, Amount: 1, TraderId: 2, TradeId: 1, StockId: 1}
	in <- m2
	// Expected server ack 2
	a2 := &Message{}
	a2.WriteAckFor(m2)
	// An ack for m2 and m2 (but nothing relating to m)
	validate(t, <-dispatch, a2, 1)
	validate(t, <-dispatch, m2, 1)
}
