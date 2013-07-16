package guid

import (
	"github.com/fmstephe/matching_engine/msg"
	"math"
	"math/rand"
	"testing"
)

func TestPushMany(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	m := make(map[int64]bool)
	s := NewStore()
	for i := 0; i < 100; i++ {
		v := r.Int63n(math.MaxInt64)
		// TODO this test is incomplete
		pushed := s.Push(msg.NO_KIND, v)
		present := m[v]
		m[v] = true
		if pushed == present {
			t.Errorf("Pushed: %v, Present: %v, Value: %v, i: %v", pushed, present, v, i)
		}
		err := validateRBT(s)
		if err != nil {
			println(err.Error())
		}
	}
}
