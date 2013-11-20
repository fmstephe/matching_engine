package client

import (
	"container/list"
	"fmt"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
)

type connect struct {
	traderId  uint32
	orders    chan *msg.Message
	responses chan []byte
}

type clientComm struct {
	outOfSvr  chan *msg.Message
	connecter chan connect
}

type Server struct {
	coordinator.AppMsgHelper
	intoSvr   chan *msg.Message
	connecter chan connect
	clientMap map[uint32]clientComm
	connects  list.List
}

func NewServer() (*Server, *ClientMaker) {
	intoSvr := make(chan *msg.Message)
	clientMap := make(map[uint32]clientComm)
	connecter := make(chan connect)
	return &Server{intoSvr: intoSvr, clientMap: clientMap, connecter: connecter}, &ClientMaker{intoSvr: intoSvr, connecter: connecter}
}

func (s *Server) Run() {
	for {
		select {
		case m := <-s.In:
			s.fromServer(m)
		case m := <-s.intoSvr:
			s.fromClient(m)
		case con := <-s.connecter:
			s.connectClient(con)
		}
	}
}

func (s *Server) fromServer(m *msg.Message) {
	if m.Kind == msg.SHUTDOWN {
		return
	}
	if m != nil {
		cc, ok := s.clientMap[m.TraderId]
		if !ok {
			println("Missing traderId", m.TraderId)
			return
		}
		cc.outOfSvr <- m
	}
}

func (s *Server) fromClient(m *msg.Message) {
	if m.Kind == msg.NEW_USER {
		s.newTrader(m)
	} else {
		s.Out <- m
	}
}

func (s *Server) newTrader(m *msg.Message) {
	_, ok := s.clientMap[m.TraderId]
	if ok {
		println(fmt.Sprintf("Attempted to register a trader (%i) twice", m.TraderId))
		return
	}
	outOfSvr := make(chan *msg.Message)
	c, cc := newClient(m.TraderId, s.intoSvr, outOfSvr)
	go c.run()
	s.clientMap[m.TraderId] = cc
	for e := s.connects.Front(); e != nil; {
		con := e.Value.(connect)
		nxt := e.Next()
		if con.traderId == m.TraderId {
			cc.connecter <- con
			s.connects.Remove(e)
		}
		e = nxt
	}
}

func (s *Server) connectClient(con connect) {
	if cc, ok := s.clientMap[con.traderId]; ok {
		cc.connecter <- con
	} else {
		s.connects.InsertAfter(con, s.connects.Back())
	}
}

type ClientMaker struct {
	intoSvr   chan *msg.Message
	connecter chan connect
}

func (cm *ClientMaker) Make(traderId uint32) (orders chan *msg.Message, responses chan []byte) {
	// TODO do we need a separate method to connect to an existing user?
	// TODO new user messages should be pre-canned in the msg package
	m := &msg.Message{Kind: msg.NEW_USER, TraderId: traderId}
	cm.intoSvr <- m // Register this user
	orders = make(chan *msg.Message)
	responses = make(chan []byte)
	con := connect{traderId: traderId, orders: orders, responses: responses}
	cm.connecter <- con
	return orders, responses
}
