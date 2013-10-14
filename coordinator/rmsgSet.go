package coordinator

import (
	"github.com/fmstephe/matching_engine/msg"
)

type rmsgSet struct {
	msgMap map[int64]*RMessage
}

func newSet() *rmsgSet {
	return &rmsgSet{msgMap: make(map[int64]*RMessage)}
}

func (s *rmsgSet) add(rm *RMessage) {
	g := msg.MkGuid(rm.originId, rm.msgId)
	s.msgMap[g] = rm
}

func (s *rmsgSet) remove(rm *RMessage) {
	g := msg.MkGuid(rm.originId, rm.msgId)
	delete(s.msgMap, g)
}

func (s *rmsgSet) contains(rm *RMessage) bool {
	g := msg.MkGuid(rm.originId, rm.msgId)
	_, ok := s.msgMap[g]
	return ok
}

func (s *rmsgSet) do(f func(*RMessage)) {
	for _, m := range s.msgMap {
		f(m)
	}
}
