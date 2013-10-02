package coordinator

import (
	"github.com/fmstephe/matching_engine/msg"
	"io"
)

type CoordinatorFunc func(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, originId uint32, name string, log bool)

type msgRunner interface {
	Config(originId uint32, log bool, name string, msgs chan *msg.Message)
	Run()
}

type AppMsgRunner interface {
	Config(name string, in, out chan *msg.Message)
	Run()
}

type AppMsgHelper struct {
	Name string
	In   <-chan *msg.Message
	Out  chan<- *msg.Message
}

func (a *AppMsgHelper) Config(name string, in, out chan *msg.Message) {
	a.Name = name
	a.In = in
	a.Out = out
}
