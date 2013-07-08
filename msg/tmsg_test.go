package msg

import (
	"testing"
)

var LOCALHOST = [4]byte{127, 0, 0, 1}

var (
	// A full message with every field set
	fullMessage = Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	// A full message but with the price set to 0, i.e. a sell that matches any buy price
	openSellMessage = Message{Price: 0, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	// Collection of messages with misssing fields (skipping price)
	partialBodyMessages = []Message{
		Message{Price: 1, Amount: 0, TraderId: 1, TradeId: 1, StockId: 1},
		Message{Price: 1, Amount: 1, TraderId: 0, TradeId: 1, StockId: 1},
		Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 0, StockId: 1},
		Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 0},
	}
	// no fields set at all
	blankMessage = Message{}
)

func testFullAndOpenSell(t *testing.T, f func(Message) Message, full, openSell bool) {
	testAllCategories(t, f, full, openSell, false, false)
	fSErr := func(m Message) Message {
		nm := f(m)
		nm.WriteStatus(INVALID_MSG_ERROR)
		return nm
	}
	testAllCategories(t, fSErr, true, true, true, true)
}

func testAllCategories(t *testing.T, f func(Message) Message, full, openSell, partialBody, blank bool) {
	m := f(fullMessage)
	expect(t, full, m)
	m = f(openSellMessage)
	expect(t, openSell, m)
	for _, p := range partialBodyMessages {
		m = f(p)
		expect(t, partialBody, m)
	}
	m = f(blankMessage)
	expect(t, blank, m)
}

func expect(t *testing.T, isValid bool, m Message) {
	if isValid != m.Valid() {
		if isValid {
			t.Errorf("\nExpected valid\n%v", m)
		} else {
			t.Errorf("\nExpected invalid\n%v", m)
		}
	}
}

func TestRouteAndKindlessMessages(t *testing.T) {
	f := func(m Message) Message {
		m.Route = NO_ROUTE
		m.Kind = NO_KIND
		return m
	}
	testFullAndOpenSell(t, f, false, false)
}

func TestZeroedMessage(t *testing.T) {
	expect(t, false, Message{})
}

func TestWriteBuy(t *testing.T) {
	f := func(m Message) Message {
		m.WriteBuy()
		return m
	}
	testFullAndOpenSell(t, f, true, false)
}

func TestWriteSell(t *testing.T) {
	f := func(m Message) Message {
		m.WriteSell()
		return m
	}
	testFullAndOpenSell(t, f, true, true)
}

func TestWriteCancelFor(t *testing.T) {
	f := func(m Message) Message {
		cm := Message{}
		cm.WriteCancelFor(&m)
		return cm
	}
	testFullAndOpenSell(t, f, true, true)
}

func TestWriteResponse(t *testing.T) {
	// can't write no_kind response
	f := func(m Message) Message {
		m.WriteResponse(NO_KIND)
		return m
	}
	testFullAndOpenSell(t, f, false, false)
	// can't write buy response
	f = func(m Message) Message {
		m.WriteResponse(BUY)
		return m
	}
	testFullAndOpenSell(t, f, false, false)
	// can't write sell response
	f = func(m Message) Message {
		m.WriteResponse(SELL)
		return m
	}
	testFullAndOpenSell(t, f, false, false)
	// can't write cancel response
	f = func(m Message) Message {
		m.WriteResponse(CANCEL)
		return m
	}
	testFullAndOpenSell(t, f, false, false)
	// can't write shutdown response
	f = func(m Message) Message {
		m.WriteResponse(SHUTDOWN)
		return m
	}
	testFullAndOpenSell(t, f, false, false)
	// can write partial response
	f = func(m Message) Message {
		m.WriteResponse(PARTIAL)
		return m
	}
	testFullAndOpenSell(t, f, true, false)
	// can write full response
	f = func(m Message) Message {
		m.WriteResponse(FULL)
		return m
	}
	testFullAndOpenSell(t, f, true, false)
	// can write cancelled response
	f = func(m Message) Message {
		m.WriteResponse(CANCELLED)
		return m
	}
	testFullAndOpenSell(t, f, true, true)
	// can write not_cancelled response
	f = func(m Message) Message {
		m.WriteResponse(NOT_CANCELLED)
		return m
	}
	testFullAndOpenSell(t, f, true, true)
}

func TestWriteCancelled(t *testing.T) {
	// Can cancel standard message and open sell
	f := func(m Message) Message {
		m.WriteCancelled()
		return m
	}
	testFullAndOpenSell(t, f, true, true)
}

func TestWriteNotCancelled(t *testing.T) {
	// Can not_cancel standard message and open sell
	f := func(m Message) Message {
		m.WriteNotCancelled()
		return m
	}
	testFullAndOpenSell(t, f, true, true)
}

func TestWriteShutdown(t *testing.T) {
	// Shutdown must have zero values other than Route and Kind
	f := func(m Message) Message {
		m.WriteShutdown()
		return m
	}
	testAllCategories(t, f, false, false, false, true)
}

func TestWriteServerAckFor(t *testing.T) {
	// can't write blank server ack
	m := Message{}
	mo := &Message{}
	m.WriteServerAckFor(mo)
	expect(t, false, m)
	// Can't write server ack for no_kind
	f := func(m Message) Message {
		m.Kind = NO_KIND
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Can write server ack for buy
	f = func(m Message) Message {
		m.WriteBuy()
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, false)
	// Can write server ack for sell
	f = func(m Message) Message {
		m.WriteSell()
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, true)
	// Can write server ack for cancel
	f = func(m Message) Message {
		m.WriteCancelFor(&m)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, true)
	// Can write server ack for shutdown
	f = func(m Message) Message {
		m.WriteShutdown()
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testAllCategories(t, f, false, false, false, true)
	// Can't write server ack for partial
	f = func(m Message) Message {
		m.WriteResponse(PARTIAL)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Can't write server ack for full
	f = func(m Message) Message {
		m.WriteResponse(FULL)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Can't write server ack for cancelled
	f = func(m Message) Message {
		m.WriteResponse(CANCELLED)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Can't write server ack for not_cancelled
	f = func(m Message) Message {
		m.WriteResponse(NOT_CANCELLED)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
}

func TestWriteClientAckFor(t *testing.T) {
	// can't write blank client ack
	m := Message{}
	mo := &Message{}
	m.WriteClientAckFor(mo)
	expect(t, false, m)
	// Cannot write client ack for no_kind
	f := func(m Message) Message {
		m.Kind = NO_KIND
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Cannot write client ack for buy
	f = func(m Message) Message {
		m.WriteBuy()
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Cannot write client ack for sell
	f = func(m Message) Message {
		m.WriteSell()
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Cannot write client ack for cancel
	f = func(m Message) Message {
		m.WriteCancelFor(&m)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Cannot write client ack for shutdown
	f = func(m Message) Message {
		m.WriteShutdown()
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Can write client ack for partial
	f = func(m Message) Message {
		m.WriteResponse(PARTIAL)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, false)
	// Can write client ack for full
	f = func(m Message) Message {
		m.WriteResponse(FULL)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, false)
	// Can write client ack for cancelled
	f = func(m Message) Message {
		m.WriteResponse(CANCELLED)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, true)
	// Can write client ack for not_cancelled
	f = func(m Message) Message {
		m.WriteResponse(NOT_CANCELLED)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, true)
}
