package msgutil

import (
	"github.com/fmstephe/matching_engine/msg"
)

type msgMap map[msg.MsgKind]map[int64]*msg.Message

type Set struct {
	kindMap msgMap
}

func NewSet() *Set {
	return &Set{kindMap: make(msgMap)}
}

func (s *Set) Add(m *msg.Message) {
	guidMap := s.kindMap[m.Kind]
	if guidMap == nil {
		guidMap = make(map[int64]*msg.Message)
		s.kindMap[m.Kind] = guidMap
	}
	g := MkGuid(m.TraderId, m.TradeId)
	guidMap[g] = m
}

func (s *Set) Remove(m *msg.Message) {
	guidMap := s.kindMap[m.Kind]
	if guidMap == nil {
		return
	}
	g := MkGuid(m.TraderId, m.TradeId)
	delete(guidMap, g)
}

func (s *Set) Contains(m *msg.Message) bool {
	guidMap := s.kindMap[m.Kind]
	if guidMap == nil {
		return false
	}
	g := MkGuid(m.TraderId, m.TradeId)
	_, ok := guidMap[g]
	return ok
}

func (s *Set) Do(f func(*msg.Message)) {
	for _, guidMap := range s.kindMap {
		for _, m := range guidMap {
			f(m)
		}
	}
}
