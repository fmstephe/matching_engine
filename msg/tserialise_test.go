package msg

import (
	"testing"
)

func messageBuffer() []byte {
	return make([]byte, ByteSize)
}

func TestMarshallDoesNotDestroyMesssage(t *testing.T) {
	ref := &Message{Kind: 1, Price: 2, Amount: 3, StockId: 4, TraderId: 5, TradeId: 6}
	m1 := &Message{}
	*m1 = *ref
	b := messageBuffer()
	if err := m1.Marshal(b); err != nil {
		t.Errorf("Unexpected marshalling error %s", err.Error())
	}
	assertEquivalent(t, ref, m1, b)
}

func TestMarshallUnMarshalPairsProducesSameMessage(t *testing.T) {
	m1 := &Message{Kind: 1, Price: 2, Amount: 3, StockId: 4, TraderId: 5, TradeId: 6}
	b := messageBuffer()
	if err := m1.Marshal(b); err != nil {
		t.Errorf("Unexpected marshalling error %s", err.Error())
	}
	m2 := &Message{}
	if err := m2.Unmarshal(b); err != nil {
		t.Errorf("Unexpected unmarshalling error %s", err.Error())
	}
	assertEquivalent(t, m1, m2, b)
}

func assertEquivalent(t *testing.T, exp, fnd *Message, b []byte) {
	if *fnd != *exp {
		t.Errorf("\nExpected to find %v\nfound %v\nMarshalled from %v", exp, fnd, b)
	}
}

func TestMarshalWithSmallBufferErrors(t *testing.T) {
	m1 := &Message{Kind: 1, Price: 2, Amount: 3, StockId: 4, TraderId: 5, TradeId: 6}
	b := make([]byte, ByteSize-1)
	if err := m1.Marshal(b); err == nil {
		t.Error("Expected marshalling error. Found none")
	}
}

func TestMarshalWithLargeBufferErrors(t *testing.T) {
	m1 := &Message{Kind: 1, Price: 2, Amount: 3, StockId: 4, TraderId: 5, TradeId: 6}
	b := make([]byte, ByteSize+1)
	if err := m1.Marshal(b); err == nil {
		t.Error("Expected marshalling error. Found none")
	}
}

func TestUnmarshalWithSmallBufferErrors(t *testing.T) {
	m1 := &Message{}
	b := make([]byte, ByteSize-1)
	if err := m1.Unmarshal(b); err == nil {
		t.Error("Expected marshalling error. Found none")
	}
}

func TestUnmarshalWithLargeBufferErrors(t *testing.T) {
	m1 := &Message{}
	b := make([]byte, ByteSize+1)
	if err := m1.Unmarshal(b); err == nil {
		t.Error("Expected marshalling error. Found none")
	}
}
