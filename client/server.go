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

type traderComm struct {
	outOfSvr  chan *msg.Message
	connecter chan connect
}

type Server struct {
	coordinator.AppMsgHelper
	intoSvr   chan *msg.Message
	connecter chan connect
	traderMap map[uint32]traderComm
	connects  list.List
}

func NewServer() (*Server, *TraderMaker) {
	intoSvr := make(chan *msg.Message)
	traderMap := make(map[uint32]traderComm)
	connecter := make(chan connect)
	return &Server{intoSvr: intoSvr, traderMap: traderMap, connecter: connecter}, &TraderMaker{intoSvr: intoSvr, connecter: connecter}
}

func (s *Server) Run() {
	for {
		select {
		case m := <-s.In:
			s.fromServer(m)
		case m := <-s.intoSvr:
			s.fromTrader(m)
		case con := <-s.connecter:
			s.connectTrader(con)
		}
	}
}

func (s *Server) fromServer(m *msg.Message) {
	if m.Kind == msg.SHUTDOWN {
		return
	}
	if m != nil {
		cc, ok := s.traderMap[m.TraderId]
		if !ok {
			println("Missing traderId", m.TraderId)
			return
		}
		cc.outOfSvr <- m
	}
}

func (s *Server) fromTrader(m *msg.Message) {
	if m.Kind == msg.NEW_TRADER {
		s.newTrader(m)
	} else {
		s.Out <- m
	}
}

func (s *Server) newTrader(m *msg.Message) {
	_, ok := s.traderMap[m.TraderId]
	if ok {
		println(fmt.Sprintf("Attempted to register a trader (%i) twice", m.TraderId))
		return
	}
	outOfSvr := make(chan *msg.Message)
	t, tc := newTrader(m.TraderId, s.intoSvr, outOfSvr)
	go t.run()
	s.traderMap[m.TraderId] = tc
	for e := s.connects.Front(); e != nil; {
		con := e.Value.(connect)
		nxt := e.Next()
		if con.traderId == m.TraderId {
			tc.connecter <- con
			s.connects.Remove(e)
		}
		e = nxt
	}
}

func (s *Server) connectTrader(con connect) {
	if cc, ok := s.traderMap[con.traderId]; ok {
		cc.connecter <- con
	} else {
		s.connects.InsertAfter(con, s.connects.Back())
	}
}

type TraderMaker struct {
	intoSvr   chan *msg.Message
	connecter chan connect
}

func (tm *TraderMaker) Make(traderId uint32) (orders chan *msg.Message, responses chan []byte) {
	// TODO do we need a separate method to connect to an existing user?
	// TODO new user messages should be pre-canned in the msg package
	m := &msg.Message{}
	m.WriteNewTrader(traderId)
	tm.intoSvr <- m // Register this user
	orders = make(chan *msg.Message)
	responses = make(chan []byte)
	con := connect{traderId: traderId, orders: orders, responses: responses}
	tm.connecter <- con
	return orders, responses
}
