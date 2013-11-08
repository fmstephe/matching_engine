package client

import (
	"fmt"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
)

type traderRegMsg struct {
	traderId uint32
	outOfSvr chan *msg.Message
}

type Server struct {
	coordinator.AppMsgHelper
	clientMap map[uint32]chan *msg.Message
	intoSvr   chan interface{}
}

func NewServer() (*Server, *ClientMaker) {
	intoSvr := make(chan interface{})
	clientMap := make(map[uint32]chan *msg.Message)
	return &Server{intoSvr: intoSvr, clientMap: clientMap}, &ClientMaker{intoSvr: intoSvr}
}

func (s *Server) Run() {
	for {
		select {
		case m := <-s.In:
			if m.Kind == msg.SHUTDOWN {
				s.Out <- m
				return
			}
			if m != nil {
				outOfSvr := s.clientMap[m.TraderId]
				if outOfSvr == nil {
					println("Missing traderId", m.TraderId)
					continue
				}
				outOfSvr <- m
			}
		case i := <-s.intoSvr:
			switch i := i.(type) {
			case *traderRegMsg:
				if s.clientMap[i.traderId] != nil {
					panic(fmt.Sprintf("Attempted to register a trader (%i) twice", i.traderId))
				}
				s.clientMap[i.traderId] = i.outOfSvr
			case *msg.Message:
				s.Out <- i
			default:
				panic(fmt.Sprintf("%T is not a legal type", i))
			}
		}
	}
}

type ClientMaker struct {
	intoSvr chan interface{}
}

func (cm *ClientMaker) Connect(traderId uint32) (orders chan ClientMessage, responses chan []byte) {
	// TODO in the future this should take an existing user and establish a connection with it
	outOfSvr := make(chan *msg.Message)
	cr := &traderRegMsg{traderId: traderId, outOfSvr: outOfSvr}
	cm.intoSvr <- cr // Register this user
	u := newClient(traderId, cm.intoSvr, outOfSvr)
	go u.run()
	return u.orders, u.responses
}
