package coordinator

import (
	. "github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/q"
	"testing"
)

const TO_SEND = 1000

const (
	clientOriginId = iota
	serverOriginId = iota
)

type echoClient struct {
	AppMsgHelper
	received []*Message
	complete chan bool
}

func newEchoClient(complete chan bool) *echoClient {
	return &echoClient{received: make([]*Message, TO_SEND), complete: complete}
}

func (c *echoClient) Run() {
	go sendAll(c.Out)
	for {
		m, shutdown := c.MsgProcessor(<-c.In, c.Out)
		if shutdown {
			return
		}
		if m != nil {
			if c.received[m.TradeId-1] != nil {
				panic("Duplicate message received")
			}
			c.received[m.TradeId-1] = m
			if full(c.received) {
				c.complete <- true
				return
			}
		}
	}
}

func full(received []*Message) bool {
	for _, rm := range received {
		if rm == nil {
			return false
		}
	}
	return true
}

func sendAll(out chan<- *Message) {
	for i := uint32(1); i <= TO_SEND; i++ {
		m := &Message{Route: APP, Direction: OUT, Kind: SELL, TraderId: 1, TradeId: i, StockId: 1, Price: 7, Amount: 1}
		out <- m
	}
}

type echoServer struct {
	AppMsgHelper
}

func (s *echoServer) Run() {
	for {
		m, shutdown := s.MsgProcessor(<-s.In, s.Out)
		if shutdown {
			return
		}
		if m != nil {
			r := &Message{}
			*r = *m
			r.Direction = OUT
			s.Out <- r
		}
	}
}

func testBadNetwork(t *testing.T, msgDropFreq int64, cFunc CoordinatorFunc) {
	complete := make(chan bool)
	c := newEchoClient(complete)
	s := &echoServer{}
	clientToServer := q.NewMeddleQ("clientToServer", q.NewDropMeddler(msgDropFreq))
	serverToClient := q.NewMeddleQ("serverToClient", q.NewDropMeddler(msgDropFreq))
	cFunc(serverToClient, clientToServer, c, clientOriginId, "Echo Client", false)
	cFunc(clientToServer, serverToClient, s, serverOriginId, "Echo Server", false)
	<-complete
}
