package msg

import (
	"testing"
)

var LOCALHOST = [4]byte{127, 0, 0, 1}

var (
	// A full message with every field set
	fullMessage = Message{Direction: IN, OriginId: 1, MsgId: 1, Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	// A full message but with the price set to 0, i.e. a sell that matches any buy price
	openSellMessage = Message{Direction: IN, OriginId: 1, MsgId: 1, Price: 0, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1}
	// Collection of messages with misssing fields (skipping price)
	partialBodyMessages = []Message{
		Message{Direction: IN, OriginId: 0, MsgId: 1, Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1},
		Message{Direction: IN, OriginId: 1, MsgId: 0, Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1},
		Message{Direction: IN, OriginId: 1, MsgId: 1, Price: 1, Amount: 0, TraderId: 1, TradeId: 1, StockId: 1},
		Message{Direction: IN, OriginId: 1, MsgId: 1, Price: 1, Amount: 1, TraderId: 0, TradeId: 1, StockId: 1},
		Message{Direction: IN, OriginId: 1, MsgId: 1, Price: 1, Amount: 1, TraderId: 1, TradeId: 0, StockId: 1},
		Message{Direction: IN, OriginId: 1, MsgId: 1, Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 0},
	}
	// no fields set at all
	blankMessage = Message{}
)

func testFullAndOpenSell(t *testing.T, f func(Message) Message, full, openSell bool) {
	testAllCategories(t, f, full, openSell, false, false)
	fSErr := func(m Message) Message {
		nm := f(m)
		nm.Status = INVALID_MSG_ERROR
		nm.Direction = IN
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
			t.Errorf("\nExpected valid\n%v", &m)
		} else {
			t.Errorf("\nExpected invalid\n%v", &m)
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
		m.WriteApp(BUY)
		return m
	}
	testFullAndOpenSell(t, f, true, false)
}

func TestWriteSell(t *testing.T) {
	f := func(m Message) Message {
		m.WriteApp(SELL)
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

func TestWriteApp(t *testing.T) {
	// can't write no_kind app
	f := func(m Message) Message {
		m.WriteApp(NO_KIND)
		return m
	}
	testFullAndOpenSell(t, f, false, false)
	// can write buy app
	f = func(m Message) Message {
		m.WriteApp(BUY)
		return m
	}
	testFullAndOpenSell(t, f, true, false)
	// can write sell app
	f = func(m Message) Message {
		m.WriteApp(SELL)
		return m
	}
	testFullAndOpenSell(t, f, true, true)
	// can write cancel app
	f = func(m Message) Message {
		m.WriteApp(CANCEL)
		return m
	}
	testFullAndOpenSell(t, f, true, true)
	// can write partial app
	f = func(m Message) Message {
		m.WriteApp(PARTIAL)
		return m
	}
	testFullAndOpenSell(t, f, true, false)
	// can write full app
	f = func(m Message) Message {
		m.WriteApp(FULL)
		return m
	}
	testFullAndOpenSell(t, f, true, false)
	// can write cancelled app
	f = func(m Message) Message {
		m.WriteApp(CANCELLED)
		return m
	}
	testFullAndOpenSell(t, f, true, true)
	// can write not_cancelled app
	f = func(m Message) Message {
		m.WriteApp(NOT_CANCELLED)
		return m
	}
	testFullAndOpenSell(t, f, true, true)
}

func TestWriteCancelled(t *testing.T) {
	// Can cancel standard message and open sell
	f := func(m Message) Message {
		m.WriteApp(CANCELLED)
		return m
	}
	testFullAndOpenSell(t, f, true, true)
}

func TestWriteNotCancelled(t *testing.T) {
	// Can not_cancel standard message and open sell
	f := func(m Message) Message {
		m.WriteApp(NOT_CANCELLED)
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

func TestWriteAckFor(t *testing.T) {
	// can't write blank ack
	m := Message{}
	mo := &Message{}
	m.WriteAckFor(mo)
	expect(t, false, m)
	// Can't write ack for no_kind
	f := func(m Message) Message {
		m.Kind = NO_KIND
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, false, false)
	// Can write ack for buy
	f = func(m Message) Message {
		m.WriteApp(BUY)
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, false)
	// Can write ack for sell
	f = func(m Message) Message {
		m.WriteApp(SELL)
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, true)
	// Can write ack for cancel
	f = func(m Message) Message {
		m.WriteCancelFor(&m)
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, true)
	// Can't write ack for shutdown
	f = func(m Message) Message {
		m.WriteShutdown()
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testAllCategories(t, f, false, false, false, false)
	// Can write ack for partial
	f = func(m Message) Message {
		m.WriteApp(PARTIAL)
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, false)
	// Can write ack for full
	f = func(m Message) Message {
		m.WriteApp(FULL)
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, false)
	// Can write ack for cancelled
	f = func(m Message) Message {
		m.WriteApp(CANCELLED)
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, true)
	// Can write ack for not_cancelled
	f = func(m Message) Message {
		m.WriteApp(NOT_CANCELLED)
		sa := Message{}
		sa.WriteAckFor(&m)
		return sa
	}
	testFullAndOpenSell(t, f, true, true)
}
