package netwk

import (
	"bytes"
	"encoding/binary"
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

var mkr = newMatchTesterMaker()

func TestRunTestSuite(t *testing.T) {
	matcher.RunTestSuite(t, newMatchTesterMaker())
}

func TestResponseResend(t *testing.T) {
	mt := mkr.Make().(*netwkTester)
	mt.timeout = RESEND_MILLIS * 5
	defer mt.Cleanup(t)
	// Add Sell
	s := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Add Buy
	b := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// server ack buy
	sab := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	sab.Route = msg.RESPONSE
	sab.Kind = msg.FULL
	// server ack sell
	sas := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	sas.Route = msg.RESPONSE
	sas.Kind = msg.FULL
	// We expect that we will keep receiving the RESPONSE messages, because we didn't ack them
	mt.ExpectNoAck(t, sab)
	mt.ExpectNoAck(t, sas)
	mt.ExpectNoAck(t, sab)
	mt.ExpectNoAck(t, sas)
	mt.ExpectNoAck(t, sab)
	mt.ExpectNoAck(t, sas)
	mt.ExpectNoAck(t, sab)
	mt.ExpectNoAck(t, sas)
	mt.ExpectNoAck(t, sab)
	mt.ExpectNoAck(t, sas)
	mt.ExpectNoAck(t, sab)
	mt.ExpectNoAck(t, sas)
}

func TestClientAck(t *testing.T) {
	mt := mkr.Make().(*netwkTester)
	mt.timeout = RESEND_MILLIS * 5
	defer mt.Cleanup(t)
	// Add Sell
	s := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	// Add buy
	b := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	// server ack buy
	sab := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	sab.Route = msg.RESPONSE
	sab.Kind = msg.FULL
	// server ack sell
	sas := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	sas.Route = msg.RESPONSE
	sas.Kind = msg.FULL
	// We expect that we will keep receiving the RESPONSE messages, until we ack them
	mt.ExpectNoAck(t, sab)
	mt.ExpectNoAck(t, sas)
	// client ack buy
	cab := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	cab.Route = msg.CLIENT_ACK
	cab.Kind = msg.FULL
	mt.SendNoAck(t, cab)
	// client ack sell
	cas := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	cas.Route = msg.CLIENT_ACK
	cas.Kind = msg.FULL
	mt.SendNoAck(t, cas)
	// We expect that after we send the client acks we will no longer receive resends
	mt.ExpectEmpty(t, sab.TraderId)
	mt.ExpectEmpty(t, sas.TraderId)
}

type chanIpWriter struct {
	out chan *msg.Message
}

func (c chanIpWriter) Write(data []byte, ip [4]byte, port int) error {
	buf := bytes.NewBuffer(data)
	r := &msg.Message{}
	binary.Read(buf, binary.BigEndian, r)
	c.out <- r
	return nil
}

func TestUnackedAreResent(t *testing.T) {
	out := make(chan *msg.Message, 100)
	w := chanIpWriter{out}
	r := NewResponder(w)
	// Pre-canned message/ack pairs
	m1 := &msg.Message{TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1, Route: msg.RESPONSE, Kind: msg.FULL}
	a1 := &msg.Message{TraderId: 10, TradeId: 43, StockId: 1, Price: 1, Amount: 1, Route: msg.CLIENT_ACK, Kind: msg.FULL}
	m2 := &msg.Message{TraderId: 123, TradeId: 2000, StockId: 1, Price: 1, Amount: 1, Route: msg.RESPONSE, Kind: msg.FULL}
	a2 := &msg.Message{TraderId: 123, TradeId: 2000, StockId: 1, Price: 1, Amount: 1, Route: msg.CLIENT_ACK, Kind: msg.FULL}
	m3 := &msg.Message{TraderId: 777, TradeId: 5432, StockId: 1, Price: 1, Amount: 1, Route: msg.RESPONSE, Kind: msg.FULL}
	a3 := &msg.Message{TraderId: 777, TradeId: 5432, StockId: 1, Price: 1, Amount: 1, Route: msg.CLIENT_ACK, Kind: msg.FULL}
	m4 := &msg.Message{TraderId: 371, TradeId: 999, StockId: 1, Price: 1, Amount: 1, Route: msg.RESPONSE, Kind: msg.FULL}
	a4 := &msg.Message{TraderId: 371, TradeId: 999, StockId: 1, Price: 1, Amount: 1, Route: msg.CLIENT_ACK, Kind: msg.FULL}
	m5 := &msg.Message{TraderId: 87, TradeId: 50, StockId: 1, Price: 1, Amount: 1, Route: msg.RESPONSE, Kind: msg.FULL}
	a5 := &msg.Message{TraderId: 87, TradeId: 50, StockId: 1, Price: 1, Amount: 1, Route: msg.CLIENT_ACK, Kind: msg.FULL}
	m6 := &msg.Message{TraderId: 40, TradeId: 499, StockId: 1, Price: 1, Amount: 1, Route: msg.RESPONSE, Kind: msg.FULL}
	a6 := &msg.Message{TraderId: 40, TradeId: 499, StockId: 1, Price: 1, Amount: 1, Route: msg.CLIENT_ACK, Kind: msg.FULL}
	m7 := &msg.Message{TraderId: 99, TradeId: 700000, StockId: 1, Price: 1, Amount: 1, Route: msg.RESPONSE, Kind: msg.FULL}
	a7 := &msg.Message{TraderId: 99, TradeId: 700000, StockId: 1, Price: 1, Amount: 1, Route: msg.CLIENT_ACK, Kind: msg.FULL}
	aUnkown := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 1, Amount: 1, Route: msg.CLIENT_ACK, Kind: msg.FULL}

	// Add m1-5 to unacked list
	r.addToUnacked(m1)
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
	r.handleClientAck(a3)
	r.resend()
	allResent(t, out, m1, m2, m4, m5)

	// ack m1
	r.handleClientAck(a1)
	r.resend()
	allResent(t, out, m2, m4, m5)

	// ack unknown
	r.handleClientAck(aUnkown)
	r.resend()
	allResent(t, out, m2, m4, m5)

	// Add m6
	r.addToUnacked(m6)
	r.resend()
	allResent(t, out, m2, m4, m5, m6)

	// ack m4 and m6, and add m7
	r.handleClientAck(a4)
	r.handleClientAck(a6)
	r.addToUnacked(m7)
	r.resend()
	allResent(t, out, m2, m5, m7)

	// ack m2, m5 and m7
	r.handleClientAck(a2)
	r.handleClientAck(a5)
	r.handleClientAck(a7)
	r.resend()
	if len(out) != 0 {
		t.Errorf("Expecting no messages re-sent, found %d", len(out))
	}
}

func allResent(t *testing.T, out chan *msg.Message, expect ...*msg.Message) {
	received := make([]*msg.Message, 0)
	for len(out) > 0 {
		received = append(received, <-out)
	}
	if len(received) != len(expect) {
		t.Errorf("Expecting %d messages, received %d", len(expect), len(received))
	}
	allOfIn(t, expect, received)
	allOfIn(t, received, expect)
}

func allOfIn(t *testing.T, first, second []*msg.Message) {
	for _, f := range first {
		found := false
		for _, s := range second {
			if *f == *s {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expecting %v, not found", f)
		}
	}
}
