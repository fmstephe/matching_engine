package netwk

import (
	"github.com/fmstephe/matching_engine/matcher"
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

var mkr = newMatchTesterMaker()

// TODO we are getting races on resent messages - a solution could be to CLIENT_ACK a message inside the Expect method
// Then drain the port of all repeat messages.
func TestRunTestSuite(t *testing.T) {
	matcher.RunTestSuite(t, newMatchTesterMaker())
}

func TestResponseResend(t *testing.T) {
	mt := mkr.Make()
	mt.(*netwkTester).timeout = RESEND_MILLIS * 5
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
	mt.Expect(t, sab)
	mt.Expect(t, sas)
	mt.Expect(t, sab)
	mt.Expect(t, sas)
	mt.Expect(t, sab)
	mt.Expect(t, sas)
	mt.Expect(t, sab)
	mt.Expect(t, sas)
	mt.Expect(t, sab)
	mt.Expect(t, sas)
	mt.Expect(t, sab)
	mt.Expect(t, sas)
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
	mt.Expect(t, sab)
	mt.Expect(t, sas)
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
	mt.ExpectFiniteResends(t, sab)
	mt.ExpectFiniteResends(t, sas)
}
