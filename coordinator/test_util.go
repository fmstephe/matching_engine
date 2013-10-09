package coordinator

import (
	"math/rand"
)

func randomUniqueMsgs() []*RMessage {
	uniqueMap := make(map[uint32]bool)
	r := rand.New(rand.NewSource(1))
	msgs := make([]*RMessage, 0)
	for i := 0; i < 100; i++ {
		origin := uint32(r.Int31())
		id := uint32(r.Int31())
		setOnce(uniqueMap, origin, id)
		m := &RMessage{originId: origin, msgId: id}
		msgs = append(msgs, m)
	}
	return msgs
}

func setOnce(uniqueMap map[uint32]bool, origin, id uint32) {
	val := origin + id
	if uniqueMap[val] == true {
		panic("Generated non-unique message")
	}
	uniqueMap[val] = true
}
