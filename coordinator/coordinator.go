package coordinator

import (
	"io"
)

type CoordinatorFunc func(reader io.ReadCloser, writer io.WriteCloser, app AppMsgRunner, originId uint32, name string, log bool)

type AppMsgRunner interface {
	Config(name string, in MsgReader, out MsgWriter)
	Run()
}

type AppMsgHelper struct {
	Name string
	In   MsgReader
	Out  MsgWriter
}

func (a *AppMsgHelper) Config(name string, in MsgReader, out MsgWriter) {
	a.Name = name
	a.In = in
	a.Out = out
}
