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
	m := &Message{}
	for {
		*m = c.In.Read()
		if m.Kind == SHUTDOWN {
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

func sendAll(out MsgWriter) {
	for i := uint32(1); i <= TO_SEND; i++ {
		out.Write(Message{Kind: SELL, TraderId: 1, TradeId: i, StockId: 1, Price: 7, Amount: 1})
	}
}

type echoServer struct {
	AppMsgHelper
}

func (s *echoServer) Run() {
	m := &Message{}
	for {
		*m = s.In.Read()
		if m.Kind == SHUTDOWN {
			return
		}
		r := Message{}
		r = *m
		r.Kind = BUY
		s.Out.Write(r)
	}
}

func testBadNetwork(t *testing.T, dropProb float64, cFunc CoordinatorFunc) {
	complete := make(chan bool)
	c := newEchoClient(complete)
	s := &echoServer{}
	clientToServer := q.NewMeddleQ("clientToServer", q.NewProbDropMeddler(dropProb))
	serverToClient := q.NewMeddleQ("serverToClient", q.NewProbDropMeddler(dropProb))
	cFunc(serverToClient, clientToServer, c, clientOriginId, "Client", false)
	cFunc(clientToServer, serverToClient, s, serverOriginId, "Server", false)
	<-complete
}
