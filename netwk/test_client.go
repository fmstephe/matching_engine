package netwk

import (
	"github.com/fmstephe/matching_engine/msg"
)

type client struct {
	dispatch     chan *msg.Message
	appMsgs      chan *msg.Message
	receivedMsgs chan *msg.Message
	toSendMsgs   chan *msg.Message
}

func newClient(receivedMsgs, toSendMsgs chan *msg.Message) *client {
	return &client{receivedMsgs: receivedMsgs, toSendMsgs: toSendMsgs}
}

func (c *client) SetDispatch(dispatch chan *msg.Message) {
	c.dispatch = dispatch
}

func (c *client) SetAppMsgs(appMsgs chan *msg.Message) {
	c.appMsgs = appMsgs
}

func (c *client) Run() {
	for {
		select {
		case m := <-c.appMsgs:
			c.receivedMsgs <- m
		case m := <-c.toSendMsgs:
			c.dispatch <- m
		}
	}
}
