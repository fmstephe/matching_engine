package coordinator

import (
	"math/rand"
	"testing"
)

func TestAddThenRemove(t *testing.T) {
	s := newSet()
	msgs := randomUniqueMsgs()
	for i, m := range msgs {
		s.add(m)
		expectAll(t, s, msgs[0:i+1])
		selfConsistent(t, s)
	}
	jMsgs := scramble(msgs)
	for i, m := range jMsgs {
		s.remove(m)
		expectNone(t, s, jMsgs[0:i+1])
		expectAll(t, s, jMsgs[i+1:len(jMsgs)])
		selfConsistent(t, s)
	}
}

func expectAll(t *testing.T, s *rmsgSet, msgs []*RMessage) {
	allOrNone(t, s, msgs, true)
	sameContent(t, msgs, extractAll(s))
}

func expectNone(t *testing.T, s *rmsgSet, msgs []*RMessage) {
	allOrNone(t, s, msgs, false)
}

func selfConsistent(t *testing.T, s *rmsgSet) {
	allOrNone(t, s, extractAll(s), true)
}

func allOrNone(t *testing.T, s *rmsgSet, msgs []*RMessage, shouldFind bool) {
	for _, m := range msgs {
		if found := s.contains(m); found != shouldFind {
			t.Errorf("Expecting message to be found (%v), %v", shouldFind, m)
		}
	}
}

func extractAll(s *rmsgSet) []*RMessage {
	msgs := make([]*RMessage, 0)
	f := func(m *RMessage) {
		msgs = append(msgs, m)
	}
	s.do(f)
	return msgs
}

func scramble(msgs []*RMessage) []*RMessage {
	jMsgs := make([]*RMessage, len(msgs))
	copy(jMsgs, msgs)
	r := rand.New(rand.NewSource(1))
	for i := range jMsgs {
		idx := r.Int() % len(jMsgs)
		jMsgs[i], jMsgs[idx] = jMsgs[idx], jMsgs[i]
	}
	return jMsgs
}

func sameContent(t *testing.T, msgs1 []*RMessage, msgs2 []*RMessage) {
	if len(msgs1) != len(msgs2) {
		t.Errorf("Slices are of different lengths, msgs1: %d. msgs2: %d", len(msgs1), len(msgs2))
	}
	compareElements(t, msgs1, msgs2)
	compareElements(t, msgs2, msgs1)
}

func compareElements(t *testing.T, msgs1 []*RMessage, msgs2 []*RMessage) {
OUTER:
	for _, m1 := range msgs1 {
		for _, m2 := range msgs2 {
			if m1 == m2 {
				continue OUTER
			}
		}
		t.Errorf("Missing element %v", m1)
	}
}
