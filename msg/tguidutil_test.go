package msg

import (
	"github.com/fmstephe/matching_engine/ints"
	"math/rand"
	"testing"
)

func TestGuidFuns(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 1000; i++ {
		traderId := r.Uint32()
		tradeId := r.Uint32()
		guid := ints.Combine(traderId, tradeId)
		cTraderId := ints.High32(guid)
		cTradeId := ints.Low32(guid)
		if cTraderId != traderId {
			t.Errorf("Expecting traderId '%s' found '%s'", traderId, cTraderId)
		}
		if cTradeId != tradeId {
			t.Errorf("Expecting tradeId '%s' found '%s'", tradeId, cTradeId)
		}
	}
}

// Test that when the most significant bits are set in the trade/rId the guid functions still work
func TestGuidFunsWithBigNumbers(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 1000; i++ {
		traderId := uint32(-r.Int31())
		tradeId := uint32(-r.Int31())
		guid := ints.Combine(traderId, tradeId)
		cTraderId := ints.High32(guid)
		cTradeId := ints.Low32(guid)
		if cTraderId != traderId {
			t.Errorf("Expecting traderId '%s' found '%s'", traderId, cTraderId)
		}
		if cTradeId != tradeId {
			t.Errorf("Expecting tradeId '%s' found '%s'", tradeId, cTradeId)
		}
	}
}
