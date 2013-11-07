package client

import (
	"fmt"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
)

type traderRegMsg struct {
	traderId    uint32
	outOfClient chan *msg.Message
}

type C struct {
	coordinator.AppMsgHelper
	clientMap  map[uint32]chan *msg.Message
	intoClient chan interface{}
}

func NewClient() (*C, *UserMaker) {
	intoClient := make(chan interface{})
	return &C{intoClient: intoClient, clientMap: make(map[uint32]chan *msg.Message)}, &UserMaker{intoClient: intoClient}
}

func (c *C) Run() {
	for {
		select {
		case m := <-c.In:
			if m.Kind == msg.SHUTDOWN {
				c.Out <- m
				return
			}
			if m != nil {
				clientChan := c.clientMap[m.TraderId]
				if clientChan == nil {
					println("Missing traderId", m.TraderId)
					continue
				}
				clientChan <- m
			}
		case i := <-c.intoClient:
			switch i := i.(type) {
			case *traderRegMsg:
				if c.clientMap[i.traderId] != nil {
					panic(fmt.Sprintf("Attempted to register a trader (%i) twice", i.traderId))
				}
				c.clientMap[i.traderId] = i.outOfClient
			case *msg.Message:
				c.Out <- i
			default:
				panic(fmt.Sprintf("%T is not a legal type", i))
			}
		}
	}
}
