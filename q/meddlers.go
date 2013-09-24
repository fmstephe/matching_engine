package q

import (
	"container/list"
)

type dropMeddler struct {
	trigger  int
	msgCount int
}

func NewDropMeddler(trigger int) *dropMeddler {
	if trigger < 1 {
		trigger = 1
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
