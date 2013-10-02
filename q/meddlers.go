package q

import (
	"container/list"
	"fmt"
	"math"
	"math/rand"
)

type freqDropMeddler struct {
	trigger  int64
	msgCount int64
}

func NewFreqDropMeddler(trigger int64) *freqDropMeddler {
	if trigger < 1 {
		trigger = math.MaxInt64
	}
	return &freqDropMeddler{trigger: trigger, msgCount: 0}
}

func (m *freqDropMeddler) Meddle(buf *list.List) {
	m.msgCount++
	if buf.Len() > 0 && m.msgCount > m.trigger {
		buf.Remove(buf.Front())
		m.msgCount = 0
	}
}

type probDropMeddler struct {
	prob float64
}

func NewProbDropMeddler(prob float64) *probDropMeddler {
	if prob < 0 || prob > 1 {
		panic(fmt.Sprintf("Probability (%d) must be 0.0 <= x <= 1.0", prob))
	}
	return &probDropMeddler{prob: prob}
}

func (m *probDropMeddler) Meddle(buf *list.List) {
	if buf.Len() > 0 && m.prob > rand.Float64() {
		buf.Remove(buf.Front())
	}
}
