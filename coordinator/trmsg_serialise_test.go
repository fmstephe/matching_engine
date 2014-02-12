package coordinator

import (
	"github.com/fmstephe/matching_engine/msg"
	"testing"
)

func messageBuffer() []byte {
	return make([]byte, rmsgByteSize)
}

func TestMarshallDoesNotDestroyMesssage(t *testing.T) {
	m := msg.Message{Kind: 7, Price: 8, Amount: 9, StockId: 10, TraderId: 11, TradeId: 12}
	ref := &RMessage{route: APP, direction: IN, originId: 5, msgId: 6, message: m}
	rm1 := &RMessage{}
	*rm1 = *ref
	b := messageBuffer()
	if err := rm1.Marshal(b); err != nil {
		t.Errorf("Unexpected marshalling error %s", err.Error())
	}
	assertEquivalent(t, ref, rm1, b)
}

func TestMarshallUnMarshalPairsProducesSameRMessage(t *testing.T) {
	m1 := msg.Message{Kind: 7, Price: 8, Amount: 9, StockId: 10, TraderId: 11, TradeId: 12}
	rm1 := &RMessage{route: APP, direction: IN, originId: 5, msgId: 6, message: m1}
	b := messageBuffer()
	if err := rm1.Marshal(b); err != nil {
		t.Errorf("Unexpected marshalling error %s", err.Error())
	}
	rm2 := &RMessage{}
	if err := rm2.Unmarshal(b); err != nil {
		t.Errorf("Unexpected unmarshalling error %s", err.Error())
	}
	assertEquivalent(t, rm1, rm2, b)
}

func assertEquivalent(t *testing.T, exp, fnd *RMessage, b []byte) {
	if *fnd != *exp {
		t.Errorf("\nExpected to find %v\nfound %v\nMarshalled from %v", exp, fnd, b)
	}
}

func TestMarshalWithSmallBufferErrors(t *testing.T) {
	m1 := msg.Message{Kind: 3, Price: 4, Amount: 5, StockId: 6, TraderId: 7, TradeId: 8}
	rm1 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 2, message: m1}
	b := make([]byte, rmsgByteSize-1)
	if err := rm1.Marshal(b); err == nil {
		t.Error("Expected marshalling error. Found none")
	}
}

func TestMarshalWithLargeBufferErrors(t *testing.T) {
	m1 := msg.Message{Kind: 3, Price: 4, Amount: 5, StockId: 6, TraderId: 7, TradeId: 8}
	rm1 := &RMessage{route: APP, direction: IN, originId: 1, msgId: 2, message: m1}
	b := make([]byte, rmsgByteSize+1)
	if err := rm1.Marshal(b); err == nil {
		t.Error("Expected marshalling error. Found none")
	}
}

func TestUnmarshalWithSmallBufferErrors(t *testing.T) {
	rm1 := &RMessage{}
	b := make([]byte, rmsgByteSize-1)
	if err := rm1.Unmarshal(b); err == nil {
		t.Error("Expected marshalling error. Found none")
	}
}

func TestUnmarshalWithLargeBufferErrors(t *testing.T) {
	rm1 := &RMessage{}
	b := make([]byte, rmsgByteSize+1)
	if err := rm1.Unmarshal(b); err == nil {
		t.Error("Expected marshalling error. Found none")
	}
}
