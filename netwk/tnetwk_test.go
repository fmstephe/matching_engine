package netwk

import (
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

var mkr = newMatchTesterMaker()

// Tests to cover - Client Acks, Message resends

func TestRunTestSuite(t *testing.T) {
	matcher.RunTestSuite(t, newMatchTesterMaker())
}

func TestResponseResend(t *testing.T) {
	mt := mkr.Make()
	mt.(*netwkTester).timeout = RESEND_MILLIS * 2
	defer mt.Cleanup(t)
	// Add Sell
	s := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	s.Route = msg.ORDER
	s.Kind = msg.SELL
	mt.Send(t, s)
	b := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	b.Route = msg.ORDER
	b.Kind = msg.BUY
	mt.Send(t, b)
	eb := &msg.Message{TraderId: 2, TradeId: 1, StockId: 1, Price: -7, Amount: 1}
	eb.Route = msg.RESPONSE
	eb.Kind = msg.FULL
	es := &msg.Message{TraderId: 1, TradeId: 1, StockId: 1, Price: 7, Amount: 1}
	es.Route = msg.RESPONSE
	es.Kind = msg.FULL
	// We expect that we will keep receiving the RESPONSE messages, because we didn't ack them
	mt.Expect(t, eb)
	mt.Expect(t, es)
	mt.Expect(t, eb)
	mt.Expect(t, es)
	mt.Expect(t, eb)
	mt.Expect(t, es)
	mt.Expect(t, eb)
	mt.Expect(t, es)
}
