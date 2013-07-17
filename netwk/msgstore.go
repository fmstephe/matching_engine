package netwk

import (
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/msg/msgutil"
)

type msgMap map[msg.MsgKind]map[int64]*msg.Message

type MsgSet struct {
	kindMap msgMap
}

func NewMsgSet() *MsgSet {
	return &MsgSet{kindMap: make(msgMap)}
}

func (s *MsgSet) Add(m *msg.Message) {
	guidMap := s.kindMap[m.Kind]
	if guidMap == nil {
		guidMap = make(map[int64]*msg.Message)
		s.kindMap[m.Kind] = guidMap
	}
	g := msgutil.MkGuid(m.TraderId, m.TradeId)
	guidMap[g] = m
}

func (s *MsgSet) Remove(m *msg.Message) {
	guidMap := s.kindMap[m.Kind]
	if guidMap == nil {
		return
	}
	g := msgutil.MkGuid(m.TraderId, m.TradeId)
	delete(guidMap, g)
}

func (s *MsgSet) Do(f func(*msg.Message)) {
	for _, guidMap := range s.kindMap {
		for _, m := range guidMap {
			f(m)
		}
	}
}
