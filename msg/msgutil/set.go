package msgutil

import (
	"github.com/fmstephe/matching_engine/msg"
)

type Set struct {
	msgMap map[int64]*msg.Message
}

func NewSet() *Set {
	return &Set{msgMap: make(map[int64]*msg.Message)}
}

func (s *Set) Add(m *msg.Message) {
	g := MkGuid(m.OriginId, m.MsgId)
	s.msgMap[g] = m
}

func (s *Set) Remove(m *msg.Message) {
	g := MkGuid(m.OriginId, m.MsgId)
	delete(s.msgMap, g)
}

func (s *Set) Contains(m *msg.Message) bool {
	g := MkGuid(m.OriginId, m.MsgId)
	_, ok := s.msgMap[g]
	return ok
}

func (s *Set) Do(f func(*msg.Message)) {
	for _, m := range s.msgMap {
		f(m)
	}
}
