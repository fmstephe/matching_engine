package msgutil

import (
	"github.com/fmstephe/matching_engine/msg"
	"math/rand"
)

func randomUniqueMsgs() []*msg.Message {
	uniqueMap := make(map[uint32]bool)
	r := rand.New(rand.NewSource(1))
	msgs := make([]*msg.Message, 0)
	for i := 0; i < 100; i++ {
		kind := msg.MsgKind(r.Int31n(msg.NUM_OF_KIND))
		traderId := uint32(r.Int31())
		tradeId := uint32(r.Int31())
		setOnce(uniqueMap, kind, traderId, tradeId)
		m := &msg.Message{Kind: kind, TraderId: traderId, TradeId: tradeId}
		msgs = append(msgs, m)
	}
	return msgs
}

func setOnce(uniqueMap map[uint32]bool, kind msg.MsgKind, traderId, tradeId uint32) {
	val := traderId + tradeId + uint32(kind)
	if uniqueMap[val] == true {
		panic("Generated non-unique message")
	}
	uniqueMap[val] = true
}
