package coordinator

import (
	"github.com/fmstephe/flib/queues/spscq"
	"io"
)

type CoordinatorFunc func(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, originId uint32, name string, log bool)

type msgRunner interface {
	Config(originId uint32, log bool, name string, msgs *spscq.PointerQ)
	Run()
}

type AppMsgRunner interface {
	Config(name string, in, out *spscq.PointerQ)
	Run()
}

type AppMsgHelper struct {
	Name string
	In   *spscq.PointerQ
	Out  *spscq.PointerQ
}

func (a *AppMsgHelper) Config(name string, in, out *spscq.PointerQ) {
	a.Name = name
	a.In = in
	a.Out = out
}
