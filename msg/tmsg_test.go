package msg

import (
	"testing"
	"runtime"
)

func getTestMessage() *Message {
	return &Message{Price: 1, Amount: 1, TraderId: 1, TradeId: 1, StockId: 1, IP: [4]byte{127,0,0,1}, Port: 1201}
}

func expect(t *testing.T, isValid bool, m *Message) {
	if isValid != m.IsValid() {
		_, fname, lnum, _ := runtime.Caller(1)
		if isValid {
			t.Errorf("\nExpected valid\n%v\n%s:%d", m, fname, lnum)
		} else {
			t.Errorf("\nExpected invalid\n%v\n%s:%d", m, fname, lnum)
		}
	}
}

func TestNakedTestMessage(t *testing.T) {
	m := getTestMessage()
	expect(t, false, m)
}

func TestZeroedMessage(t *testing.T) {
	m := &Message{}
	expect(t, false, m)
}

func TestWriteBuy(t *testing.T) {
	// Can't write buy for empty message
	m := &Message{}
	m.WriteBuy()
	expect(t, false, m)
	// Can write buy for test message
	m = getTestMessage()
	m.WriteBuy()
	expect(t, true, m)
}

func TestWriteSell(t *testing.T) {
	// Can't write sell for empty message
	m := &Message{}
	m.WriteSell()
	expect(t, false, m)
	// Can write sell for test message
	m = getTestMessage()
	m.WriteSell()
	expect(t, true, m)
}

func TestWriteCancelFor(t *testing.T) {
	// Can't write blank cancel
	m := &Message{}
	mo := &Message{}
	m.WriteCancelFor(mo)
	expect(t, false, m)
	// Can cancel test message
	m = &Message{}
	mo = getTestMessage()
	m.WriteCancelFor(mo)
	expect(t, true, m)
}

func TestWriteResponse(t *testing.T) {
	// Can't write blank response
	m := &Message{}
	m.WriteResponse(PARTIAL)
	expect(t, false, m)
	// can write partial response
	m = getTestMessage()
	m.WriteResponse(PARTIAL)
	expect(t, true, m)
	// can write full response
	m = getTestMessage()
	m.WriteResponse(FULL)
	expect(t, true, m)
	// can write cancelled response
	m = getTestMessage()
	m.WriteResponse(CANCELLED)
	expect(t, true, m)
	// can write not_cancelld response
	m = getTestMessage()
	m.WriteResponse(NOT_CANCELLED)
	expect(t, true, m)
	// can't write buy response
	m = getTestMessage()
	m.WriteResponse(BUY)
	expect(t, false, m)
	// can't write buy response
	m = getTestMessage()
	m.WriteResponse(SELL)
	expect(t, false, m)
	// can't write buy response
	m = getTestMessage()
	m.WriteResponse(CANCEL)
	expect(t, false, m)
	// can't write buy response
	m = getTestMessage()
	m.WriteResponse(SHUTDOWN)
	expect(t, false, m)
}

func TestWriteCancelled(t *testing.T) {
	// Can't write blank cancelled
	m := &Message{}
	m.WriteCancelled()
	expect(t, false, m)
	// Can write populated cancelled
	m = getTestMessage()
	m.WriteCancelled()
	expect(t, true, m)
}

func TestWriteNotCancelled(t *testing.T) {
	// Can't write blank not_cancelled
	m := &Message{}
	m.WriteNotCancelled()
	expect(t, false, m)
	// Can write populated not_cancelled
	m = getTestMessage()
	m.WriteNotCancelled()
	expect(t, true, m)
}

func TestWriteShutdown(t *testing.T) {
	// Can't write blank shutdown
	m := &Message{}
	m.WriteShutdown()
	expect(t, false, m)
	// Can write blankish shutdown
	m = &Message{IP: [4]byte{127,0,0,0}, Port: 1201}
	m.WriteShutdown()
	expect(t, true, m)
	// Can't write non-blank shutdown
	m = getTestMessage()
	m.WriteShutdown()
	expect(t, false, m)
}

func TestWriteServerAckFor(t *testing.T) {
	// can't write blank server ack
	m := &Message{}
	mo := &Message{}
	m.WriteServerAckFor(mo)
	expect(t, false, m)
	// Can write server ack for buy
	m = &Message{}
	mo = getTestMessage()
	mo.WriteBuy()
	m.WriteServerAckFor(mo)
	expect(t, true, m)
	// Can write server ack for sell
	m = &Message{}
	mo = getTestMessage()
	mo.WriteSell()
	m.WriteServerAckFor(mo)
	expect(t, true, m)
	// Can write server ack for cancel
	m = &Message{}
	mo = getTestMessage()
	mo.WriteCancelFor(mo)
	m.WriteServerAckFor(mo)
	expect(t, true, m)
	// Can write server ack for shutdown
	m = &Message{}
	mo = &Message{IP: [4]byte{127,0,0,0}, Port: 1201}
	mo.WriteShutdown()
	m.WriteServerAckFor(mo)
	expect(t, true, m)
	// Can't write server ack for partial
	m = &Message{}
	mo = getTestMessage()
	mo.WriteResponse(PARTIAL)
	m.WriteServerAckFor(mo)
	expect(t, false, m)
	// Can't write server ack for full
	m = &Message{}
	mo = getTestMessage()
	mo.WriteResponse(FULL)
	m.WriteServerAckFor(mo)
	expect(t, false, m)
	// Can't write server ack for cancelled
	m = &Message{}
	mo = getTestMessage()
	mo.WriteCancelled()
	m.WriteServerAckFor(mo)
	expect(t, false, m)
	// Can't write server ack for not_cancelled
	m = &Message{}
	mo = getTestMessage()
	mo.WriteNotCancelled()
	m.WriteServerAckFor(mo)
	expect(t, false, m)
}

func TestWriteClientAckFor(t *testing.T) {
	// can't write blank client ack
	m := &Message{}
	mo := &Message{}
	m.WriteClientAckFor(mo)
	expect(t, false, m)
	// Can't write client ack for buy
	m = &Message{}
	mo = getTestMessage()
	mo.WriteBuy()
	m.WriteClientAckFor(mo)
	expect(t, false, m)
	// Can't write client ack for sell
	m = &Message{}
	mo = getTestMessage()
	mo.WriteSell()
	m.WriteClientAckFor(mo)
	expect(t, false, m)
	// Can't write client ack for cancel
	m = &Message{}
	mo = getTestMessage()
	mo.WriteCancelFor(mo)
	m.WriteClientAckFor(mo)
	expect(t, false, m)
	// Can't write client ack for shutdown
	m = &Message{}
	mo = &Message{IP: [4]byte{0,0,0,0}, Port: 1201}
	mo.WriteShutdown()
	m.WriteClientAckFor(mo)
	expect(t, false, m)
	// Can write client ack for partial
	m = &Message{}
	mo = getTestMessage()
	mo.WriteResponse(PARTIAL)
	m.WriteClientAckFor(mo)
	expect(t, true, m)
	// Can write client ack for full
	m = &Message{}
	mo = getTestMessage()
	mo.WriteResponse(FULL)
	m.WriteClientAckFor(mo)
	expect(t, true, m)
	// Can write client ack for cancelled
	m = &Message{}
	mo = getTestMessage()
	mo.WriteCancelled()
	m.WriteClientAckFor(mo)
	expect(t, true, m)
	// Can write client ack for not_cancelled
	m = &Message{}
	mo = getTestMessage()
	mo.WriteNotCancelled()
	m.WriteClientAckFor(mo)
	expect(t, true, m)
}
