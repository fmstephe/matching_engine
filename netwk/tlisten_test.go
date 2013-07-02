package netwk

import (
	. "github.com/fmstephe/matching_engine/msg"
	"testing"
)

var mkr = newMatchTesterMaker()

func TestFoo(t *testing.T) {
	nt := mkr.Make().(*netwkTester)
	defer nt.Cleanup(t)
	m := &Message{Status: NORMAL, Route: ORDER, Kind: SELL, Price: 7, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	// Make a slice which is too small
	b := make([]byte, SizeofMessage-1)
	m.WriteTo(b)
	nt.write.Write(b)
	// Expected server ack
	a := &Message{}
	a.WriteServerAckFor(m)
	a.WriteStatus(SMALL_READ_ERROR)
	// Expected response
	r := &Message{}
	*r = *m
	r.WriteStatus(SMALL_READ_ERROR)
	nt.Expect(t, a)
	nt.Expect(t, r)
}
