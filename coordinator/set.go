package coordinator

import (
	"github.com/fmstephe/matching_engine/msg"
)

// TODO make all these methods private
// Change set to rmsgSet

type set struct {
	msgMap map[int64]*RMessage
}

func newSet() *set {
	return &set{msgMap: make(map[int64]*RMessage)}
}

func (s *set) Add(rm *RMessage) {
	g := msg.MkGuid(rm.originId, rm.msgId)
	s.msgMap[g] = rm
}

func (s *set) Remove(rm *RMessage) {
	g := msg.MkGuid(rm.originId, rm.msgId)
	delete(s.msgMap, g)
}

func (s *set) Contains(rm *RMessage) bool {
	g := msg.MkGuid(rm.originId, rm.msgId)
	_, ok := s.msgMap[g]
	return ok
}

func (s *set) Do(f func(*RMessage)) {
	for _, m := range s.msgMap {
		f(m)
	}
}
