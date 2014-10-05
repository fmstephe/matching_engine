package client

import (
	"fmt"
	"github.com/fmstephe/matching_engine/coordinator"
	"github.com/fmstephe/matching_engine/msg"
)

type connect struct {
	traderId  uint32
	orders    chan *msg.Message
	responses chan *Response
}

type traderComm struct {
	outOfSvr  chan *msg.Message
	connecter chan connect
}

type Server struct {
	coordinator.AppMsgHelper
	intoSvr     chan *msg.Message
	connecter   chan connect
	traderMap   map[uint32]traderComm
	connectsMap map[uint32][]connect
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
	cons := s.connectsMap[m.TraderId]
	delete(s.connectsMap, m.TraderId)
	for _, con := range cons {
		tc.connecter <- con
	}
}

func (s *Server) connectTrader(con connect) {
	if cc, ok := s.traderMap[con.traderId]; ok {
		cc.connecter <- con
	} else {
		cons := s.connectsMap[con.traderId]
		if cons == nil {
			cons = make([]connect, 1)
		}
		cons = append(cons, con)
		s.connectsMap[con.traderId] = cons
	}
}

type TraderMaker struct {
	intoSvr   chan *msg.Message
	connecter chan connect
}

func (tm *TraderMaker) Make(traderId uint32) (orders chan *msg.Message, responses chan *Response) {
	m := &msg.Message{}
	m.WriteNewTrader(traderId)
	tm.intoSvr <- m // Register this user
	return tm.Connect(traderId)
}

func (tm *TraderMaker) Connect(traderId uint32) (orders chan *msg.Message, responses chan *Response) {
	orders = make(chan *msg.Message)
	responses = make(chan *Response)
	con := connect{traderId: traderId, orders: orders, responses: responses}
	tm.connecter <- con
	return orders, responses
}
