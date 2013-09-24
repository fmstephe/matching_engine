package q

import (
	"container/list"
	"math"
)

type dropMeddler struct {
	trigger  int64
	msgCount int64
}

func NewDropMeddler(trigger int64) *dropMeddler {
	if trigger < 1 {
		trigger = math.MaxInt64
	}
	return &dropMeddler{trigger: trigger, msgCount: 0}
}

func (m *dropMeddler) Meddle(buf *list.List) {
	m.msgCount++
	if buf.Len() > 0 && m.msgCount > m.trigger {
		buf.Remove(buf.Front())
		m.msgCount = 0
	}
}
