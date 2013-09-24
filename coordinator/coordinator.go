package coordinator

import (
	"github.com/fmstephe/matching_engine/msg"
)

type msgRunner interface {
	Config(originId uint32, log bool, name string, msgs chan *msg.Message)
	Run()
}

type msgHelper struct {
	originId uint32
	log      bool
	name     string
	msgs     chan *msg.Message
}

func (h *msgHelper) Config(originId uint32, log bool, name string, msgs chan *msg.Message) {
	h.originId = originId
	h.log = log
	h.name = name
	h.msgs = msgs
}

type MsgProcessFunc func(m *msg.Message, out chan<- *msg.Message) (appMsg *msg.Message, shutdown bool)

func DefaultMsgProcessor(m *msg.Message, out chan<- *msg.Message) (*msg.Message, bool) {
	return m, (m.Route == msg.SHUTDOWN)
}

func reliableMsgProcessor(m *msg.Message, out chan<- *msg.Message) (AppMsg *msg.Message, shutdown bool) {
	out <- m
	switch {
	case m.Route == msg.APP && m.Status == msg.NORMAL:
		return m, false
	case m.Route == msg.SHUTDOWN:
		return nil, true
	default:
		return nil, false
	}
}

type AppMsgRunner interface {
	Config(name string, in, out chan *msg.Message, msgProcessor MsgProcessFunc)
	Run()
}

type AppMsgHelper struct {
	Name         string
	In           <-chan *msg.Message
	Out          chan<- *msg.Message
	MsgProcessor MsgProcessFunc
}

func (a *AppMsgHelper) Config(name string, in, out chan *msg.Message, msgProcessor MsgProcessFunc) {
	a.Name = name
	a.In = in
	a.Out = out
	a.MsgProcessor = msgProcessor
}
