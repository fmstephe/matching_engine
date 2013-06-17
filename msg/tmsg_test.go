package msg

import (
	"runtime"
	"testing"
)

var LOCALHOST = [4]byte{127, 0, 0, 1}

var (
	// A full message with every field set
	fullMessage = Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1, IP: LOCALHOST, Port: 1201}
	// A full message but with the price set to 0, i.e. a sell that matches any buy price
	openSellMessage = Message{Price: 0, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1, IP: LOCALHOST, Port: 1201}
	// Collection of messages with misssing fields (skipping price) but always populated netwk fields
	partialBodyMessages = []Message{
		Message{Price: 1, Amount: 0, TraderId: 1, TradeId: 1, StockId: 1, IP: LOCALHOST, Port: 1201},
		Message{Price: 1, Amount: 1, TraderId: 0, TradeId: 1, StockId: 1, IP: LOCALHOST, Port: 1201},
		Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 0, StockId: 1, IP: LOCALHOST, Port: 1201},
		Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 0, IP: LOCALHOST, Port: 1201},
		Message{Price: 1, IP: LOCALHOST, Port: 1201},
		Message{Amount: 1, IP: LOCALHOST, Port: 1201},
		Message{TraderId: 1, IP: LOCALHOST, Port: 1201},
		Message{TradeId: 1, IP: LOCALHOST, Port: 1201},
		Message{StockId: 1, IP: LOCALHOST, Port: 1201},
	}
	// Collection of messages with misssing fields including netwk data
	partialNetMessages = []Message{
		Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1, IP: [4]byte{}, Port: 1201},
		Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1, IP: LOCALHOST, Port: 0},
		Message{Price: 1},
		Message{Amount: 1},
		Message{TraderId: 1},
		Message{TradeId: 1},
		Message{StockId: 1},
		Message{IP: LOCALHOST},
		Message{Port: 1201},
	}
	// Message with all empty fields other than netwk data
	netOnlyMessage = Message{IP: LOCALHOST, Port: 1201}
	// no fields set at all
	blankMessage = Message{}
)

func testTopThree(t *testing.T, f func(Message) Message, full, openSell, netOnly bool) {
	testAllCategories(t, f, full, openSell, netOnly, false, false, false)
	fSErr := func(m Message) Message {
		nm := f(m)
		nm.WriteStatus(SENDABLE_ERROR)
		return nm
	}
	testAllCategories(t, fSErr, true, true, true, true, false, false)
	fNSErr := func(m Message) Message {
		nm := f(m)
		nm.WriteStatus(NOT_SENDABLE_ERROR)
		return nm
	}
	testAllCategories(t, fNSErr, true, true, true, true, true, true)
}

func testAllCategories(t *testing.T, f func(Message) Message, full, openSell, netOnly, partialBody, partialNet, blank bool) {
	m := f(fullMessage)
	expect(t, full, m)
	m = f(openSellMessage)
	expect(t, openSell, m)
	m = f(netOnlyMessage)
	expect(t, netOnly, m)
	for _, p := range partialBodyMessages {
		m = f(p)
		expect(t, partialBody, m)
	}
	for _, p := range partialNetMessages {
		m = f(p)
		expect(t, partialNet, m)
	}
	m = f(blankMessage)
	expect(t, blank, m)
}

func expect(t *testing.T, isValid bool, m Message) {
	if isValid != m.Valid() {
		_, fname, lnum, _ := runtime.Caller(1)
		if isValid {
			t.Errorf("\nExpected valid\n%v\n%s:%d", m, fname, lnum)
		} else {
			t.Errorf("\nExpected invalid\n%v\n%s:%d", m, fname, lnum)
		}
	}
}

func TestRouteAndKindlessMessages(t *testing.T) {
	f := func(m Message) Message {
		return Message{}
	}
	testAllCategories(t, f, false, false, false, false, false, false)
}

func TestZeroedMessage(t *testing.T) {
	expect(t, false, Message{})
}

func TestWriteBuy(t *testing.T) {
	f := func(m Message) Message {
		m.WriteBuy()
		return m
	}
	testTopThree(t, f, true, false, false)
}

func TestWriteSell(t *testing.T) {
	f := func(m Message) Message {
		m.WriteSell()
		return m
	}
	testTopThree(t, f, true, true, false)
}

func TestWriteCancelFor(t *testing.T) {
	f := func(m Message) Message {
		cm := Message{}
		cm.WriteCancelFor(&m)
		return cm
	}
	testTopThree(t, f, true, true, false)
}

func TestWriteResponse(t *testing.T) {
	// can write partial response
	f := func(m Message) Message {
		m.WriteResponse(PARTIAL)
		return m
	}
	testTopThree(t, f, true, false, false)
	// can write full response
	f = func(m Message) Message {
		m.WriteResponse(FULL)
		return m
	}
	testTopThree(t, f, true, false, false)
	// can write cancelled response
	f = func(m Message) Message {
		m.WriteResponse(CANCELLED)
		return m
	}
	testTopThree(t, f, true, true, false)
	// can write not_cancelled response
	f = func(m Message) Message {
		m.WriteResponse(NOT_CANCELLED)
		return m
	}
	testTopThree(t, f, true, true, false)
	// can't write buy response
	f = func(m Message) Message {
		m.WriteResponse(BUY)
		return m
	}
	testTopThree(t, f, false, false, false)
	// can't write sell response
	f = func(m Message) Message {
		m.WriteResponse(SELL)
		return m
	}
	testTopThree(t, f, false, false, false)
	// can't write cancel response
	f = func(m Message) Message {
		m.WriteResponse(CANCEL)
		return m
	}
	testTopThree(t, f, false, false, false)
	// can't write shutdown response
	f = func(m Message) Message {
		m.WriteResponse(SHUTDOWN)
		return m
	}
	testTopThree(t, f, false, false, false)
}

func TestWriteCancelled(t *testing.T) {
	f := func(m Message) Message {
		m.WriteCancelled()
		return m
	}
	testTopThree(t, f, true, true, false)
}

func TestWriteNotCancelled(t *testing.T) {
	f := func(m Message) Message {
		m.WriteNotCancelled()
		return m
	}
	testTopThree(t, f, true, true, false)
}

func TestWriteShutdown(t *testing.T) {
	f := func(m Message) Message {
		m.WriteShutdown()
		return m
	}
	testTopThree(t, f, false, false, true)
}

func TestWriteServerAckFor(t *testing.T) {
	// can't write blank server ack
	m := Message{}
	mo := &Message{}
	m.WriteServerAckFor(mo)
	expect(t, false, m)
	// Can write server ack for buy
	f := func(m Message) Message {
		m.WriteBuy()
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testTopThree(t, f, true, false, false)
	// Can write server ack for sell
	f = func(m Message) Message {
		m.WriteSell()
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testTopThree(t, f, true, true, false)
	// Can write server ack for cancel
	f = func(m Message) Message {
		m.WriteCancelFor(&m)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testTopThree(t, f, true, true, false)
	// Can write server ack for shutdown
	f = func(m Message) Message {
		m.WriteShutdown()
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, true)
	// Can't write server ack for partial
	f = func(m Message) Message {
		m.WriteResponse(PARTIAL)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, false)
	// Can't write server ack for full
	f = func(m Message) Message {
		m.WriteResponse(FULL)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, false)
	// Can't write server ack for cancelled
	f = func(m Message) Message {
		m.WriteResponse(CANCELLED)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, false)
	// Can't write server ack for not_cancelled
	f = func(m Message) Message {
		m.WriteResponse(NOT_CANCELLED)
		sa := Message{}
		sa.WriteServerAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, false)
}

func TestWriteClientAckFor(t *testing.T) {
	// can't write blank client ack
	m := Message{}
	mo := &Message{}
	m.WriteClientAckFor(mo)
	expect(t, false, m)
	// Cannot write client ack for buy
	f := func(m Message) Message {
		m.WriteBuy()
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, false)
	// Cannot write client ack for sell
	f = func(m Message) Message {
		m.WriteSell()
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, false)
	// Cannot write client ack for cancel
	f = func(m Message) Message {
		m.WriteCancelFor(&m)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, false)
	// Cannot write client ack for shutdown
	f = func(m Message) Message {
		m.WriteShutdown()
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testTopThree(t, f, false, false, false)
	// Can write client ack for partial
	f = func(m Message) Message {
		m.WriteResponse(PARTIAL)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testTopThree(t, f, true, false, false)
	// Can write client ack for full
	f = func(m Message) Message {
		m.WriteResponse(FULL)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testTopThree(t, f, true, false, false)
	// Can write client ack for cancelled
	f = func(m Message) Message {
		m.WriteResponse(CANCELLED)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testTopThree(t, f, true, true, false)
	// Can write client ack for not_cancelled
	f = func(m Message) Message {
		m.WriteResponse(NOT_CANCELLED)
		sa := Message{}
		sa.WriteClientAckFor(&m)
		return sa
	}
	testTopThree(t, f, true, true, false)
}
