package coordinator

import (
	"github.com/fmstephe/matching_engine/msg"
	"math/rand"
	"testing"
)

func TestStructure(t *testing.T) {
	tk := NewTicker()
	msgs := randomUniqueMsgs()
	for _, m := range msgs {
		tk.Tick(m)
		err := validateRBT(tk)
		if err != nil {
			println(err.Error())
		}
	}
}

func TestTickMany(t *testing.T) {
	tk := NewTicker()
	msgs := randomUniqueMsgs()
	for _, m := range msgs {
		ticked := tk.Tick(m)
		if !ticked {
			t.Errorf("Failure to tick new message")
		}
	}
}

func TestTickRepeated(t *testing.T) {
	tk := NewTicker()
	msgs := randomUniqueMsgs()
	for _, m := range msgs {
		ticked := tk.Tick(m)
		if !ticked {
			t.Errorf("Failure to tick new message")
		}
	}
	for _, m := range msgs {
		ticked := tk.Tick(m)
		if ticked {
			t.Errorf("Ticked repeated message")
		}
	}
}

func TestTickRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	st := &simpleTicker{make(map[int64]bool)}
	tk := NewTicker()
	msgs := randomUniqueMsgs()
	for i := 0; i < 1000; i++ {
		idx := r.Int31n(int32(len(msgs)))
		m := msgs[idx]
		ticked := tk.Tick(m)
		simpleTicked := st.tick(m)
		if ticked != simpleTicked {
			t.Errorf("Initial tick failure. ticked: %v, simpleTicked: %v", ticked, simpleTicked)
		}
	}
}

type simpleTicker struct {
	guidMap map[int64]bool
}

func (t *simpleTicker) tick(m *RMessage) bool {
	g := msg.MkGuid(m.originId, m.msgId)
	if t.guidMap[g] {
		return false
	}
	t.guidMap[g] = true
	return true

}
